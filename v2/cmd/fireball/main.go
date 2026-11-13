package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"

	"github.com/CatSprite-dev/fireball/v2/internal/convert"
	"github.com/CatSprite-dev/fireball/v2/internal/fx"
	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

const baseCurrency = "RUB"

var excludedAccountTypes = map[string]bool{
	"ACCOUNT_TYPE_INVEST_BOX": true,
	"ACCOUNT_TYPE_DFA":        true,
}

type accountData struct {
	Account    tinvest.Account
	Operations []tinvest.Operation
	Portfolio  *tinvest.Portfolio
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := run(logger); err != nil {
		logger.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("load .env: %w", err)
		}
	}

	token := os.Getenv("T_INVEST_TOKEN_ALL")
	if token == "" {
		return errors.New("T_INVEST_TOKEN_ALL environment variable is not set")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := tinvest.New(token, logger)

	accounts, err := client.GetAccounts(ctx, tinvest.AccountStatusAll)
	if err != nil {
		return fmt.Errorf("get accounts: %w", err)
	}

	accounts = filterAccounts(accounts)
	if len(accounts) == 0 {
		return errors.New("no supported accounts")
	}

	start := time.Now()
	data, err := loadAccounts(ctx, client, accounts)
	if err != nil {
		return fmt.Errorf("load accounts: %w", err)
	}
	logger.Info("all accounts loaded", "accounts", len(data), "duration", time.Since(start))

	events, err := buildEvents(data)
	if err != nil {
		return err
	}

	checkTransfers(logger, events)

	if err := checkRates(ctx, logger, client, events); err != nil {
		return err
	}

	return nil
}

func buildEvents(data []accountData) ([]ledger.Event, error) {
	var all []ledger.Event
	var accountIDs []string
	for _, d := range data {
		events, err := convert.ToEvents(d.Account.ID, d.Operations)
		if err != nil {
			return nil, fmt.Errorf("account %s: %w", d.Account.Name, err)
		}
		all = append(all, events...)
		accountIDs = append(accountIDs, d.Account.ID)
	}

	selected := ledger.Filter(all, accountIDs)
	ledger.SortByDate(selected)
	return selected, nil
}

// checkTransfers — временная проверка: переводы между выбранными счетами
// должны взаимно гаситься. Показывает переводы, у которых нет второй половины.
func checkTransfers(logger *slog.Logger, events []ledger.Event) {
	total := decimal.Zero
	halves := make(map[string][]ledger.Event)
	for _, e := range events {
		if e.Kind != ledger.KindTransferIn && e.Kind != ledger.KindTransferOut {
			continue
		}
		total = total.Add(e.Amount.Amount)
		halves[e.OperationID] = append(halves[e.OperationID], e)
	}

	logger.Info("transfers", "events", len(events), "transfers_sum", total)

	for id, h := range halves {
		if len(h) == 1 {
			logger.Info("unpaired transfer",
				"id", id,
				"account", h[0].AccountID,
				"date", h[0].Date,
				"kind", h[0].Kind,
				"amount", h[0].Amount.Amount,
			)
		}
	}
}

// checkRates — временная проверка: валютные события в рублях.
func checkRates(ctx context.Context, logger *slog.Logger, client *tinvest.Client, events []ledger.Event) error {
	rates, err := loadRates(ctx, client, events)
	if err != nil {
		return fmt.Errorf("load rates: %w", err)
	}

	if err := ledger.ApplyBase(events, rates); err != nil {
		return err
	}
	for _, e := range events {
		if e.Amount.Currency == baseCurrency {
			continue
		}
		logger.Info("in rubles",
			"date", e.Date.Format(time.DateOnly),
			"amount", e.Amount.Amount,
			"currency", e.Amount.Currency,
			"rub", e.BaseAmount.Amount.Round(2),
		)
	}
	return nil
}

func filterAccounts(accounts []tinvest.Account) []tinvest.Account {
	var supported []tinvest.Account
	for _, acc := range accounts {
		if excludedAccountTypes[acc.Type] {
			continue
		}
		supported = append(supported, acc)
	}
	return supported
}

func loadAccounts(ctx context.Context, client *tinvest.Client, accounts []tinvest.Account) ([]accountData, error) {
	results := make([]accountData, len(accounts))
	now := time.Now()

	g, gctx := errgroup.WithContext(ctx)
	for i, acc := range accounts {
		g.Go(func() error {
			ops, err := client.GetOperationsByCursor(gctx, tinvest.OperationsFilter{
				AccountID: acc.ID,
				From:      &acc.OpenedDate,
				To:        &now,
			})
			if err != nil {
				return fmt.Errorf("account %s: %w", acc.ID, err)
			}
			var portfolio *tinvest.Portfolio
			if acc.Status != "ACCOUNT_STATUS_CLOSED" {
				portfolio, err = client.GetPortfolio(gctx, acc.ID, baseCurrency)
				if err != nil {
					return fmt.Errorf("account %s: %w", acc.ID, err)
				}
			}
			results[i] = accountData{Account: acc, Operations: ops, Portfolio: portfolio}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

func loadRates(ctx context.Context, client *tinvest.Client, events []ledger.Event) (*fx.Rates, error) {
	needed := make(map[string]bool)
	var earliest time.Time
	for _, e := range events {
		if e.Amount.Currency == baseCurrency {
			continue
		}
		needed[e.Amount.Currency] = true
		if earliest.IsZero() || e.Date.Before(earliest) {
			earliest = e.Date
		}
	}
	if len(needed) == 0 {
		return fx.New(nil), nil
	}

	list, err := client.GetCurrencies(ctx, tinvest.InstrumentStatusAll)
	if err != nil {
		return nil, err
	}
	instruments, err := convert.ExchangeCurrencies(list)
	if err != nil {
		return nil, err
	}

	from := earliest.AddDate(0, 0, -14)
	to := time.Now()

	points := make(map[string][]fx.Point, len(needed))
	for currency := range needed {
		instrument, ok := instruments[currency]
		if !ok {
			return nil, fmt.Errorf("no exchange instrument for currency %s", currency)
		}

		candles, err := client.GetCandles(ctx, instrument.UID, from, to, tinvest.CandleIntervalDay)
		if err != nil {
			return nil, fmt.Errorf("candles %s: %w", currency, err)
		}

		currencyPoints, err := convert.ToPoints(candles.Candles, instrument.Nominal)
		if err != nil {
			return nil, fmt.Errorf("points %s: %w", currency, err)
		}
		points[currency] = currencyPoints
	}

	return fx.New(points), nil
}
