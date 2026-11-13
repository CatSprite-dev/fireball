package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"time"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"

	"github.com/CatSprite-dev/fireball/v2/internal/convert"
	"github.com/CatSprite-dev/fireball/v2/internal/fx"
	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/metrics"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

const baseCurrency = "RUB"

var excludedAccountTypes = map[string]bool{
	"ACCOUNT_TYPE_DFA": true,
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

	if err := checkProfit(logger, data, events); err != nil {
		return err
	}

	if err := checkSplits(logger, data, events); err != nil {
		return err
	}

	if err := checkTotal(logger, data, events); err != nil {
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

// checkProfit — временная проверка: позиции и прибыль по каждой бумаге.
func checkProfit(logger *slog.Logger, data []accountData, events []ledger.Event) error {
	values := make(map[string]decimal.Decimal)
	brokerQty := make(map[string]decimal.Decimal)
	tickers := make(map[string]string)

	for _, d := range data {
		if d.Portfolio == nil {
			continue
		}
		holdings, err := convert.ToHoldings(d.Account.ID, d.Portfolio)
		if err != nil {
			return fmt.Errorf("account %s: %w", d.Account.Name, err)
		}

		sum := decimal.Zero
		for _, h := range holdings {
			sum = sum.Add(h.Value())
			values[h.InstrumentUID] = values[h.InstrumentUID].Add(h.Value())
			brokerQty[h.InstrumentUID] = brokerQty[h.InstrumentUID].Add(h.Quantity)
			tickers[h.InstrumentUID] = h.Ticker
		}

		total, err := convert.ToDecimal(d.Portfolio.TotalAmountPortfolio.Units, d.Portfolio.TotalAmountPortfolio.Nano)
		if err != nil && !errors.Is(err, convert.ErrNoValue) {
			return fmt.Errorf("account %s: total: %w", d.Account.Name, err)
		}
		logger.Info("account value", "account", d.Account.Name, "ours", sum.Round(2), "broker", total)
	}

	flows := metrics.ByInstrument(events)

	type row struct {
		uid    string
		profit decimal.Decimal
		f      *metrics.Flows
	}
	var rows []row
	for uid, f := range flows {
		profit := values[uid].Add(f.Bought).Add(f.Sold).Add(f.Income).Add(f.Fees).Add(f.Taxes)
		rows = append(rows, row{uid: uid, profit: profit, f: f})
	}
	slices.SortFunc(rows, func(a, b row) int { return a.profit.Cmp(b.profit) })

	total := decimal.Zero
	for _, r := range rows {
		total = total.Add(r.profit)
		name := tickers[r.uid]
		if name == "" {
			name = r.uid
		}
		logger.Info("profit",
			"instrument", name,
			"profit", r.profit.Round(2),
			"value_now", values[r.uid].Round(2),
			"bought", r.f.Bought.Round(2),
			"sold", r.f.Sold.Round(2),
			"income", r.f.Income.Round(2),
			"fees", r.f.Fees.Round(2),
			"taxes", r.f.Taxes.Round(2),
			"qty_ours", r.f.Quantity,
			"qty_broker", brokerQty[r.uid],
		)
	}
	logger.Info("profit total", "instruments", len(rows), "sum", total.Round(2))
	return nil
}

// checkSplits — временная проверка: цена за штуку в сделках по бумагам со сплитом.
func checkSplits(logger *slog.Logger, data []accountData, events []ledger.Event) error {
	watch := map[string]bool{"PLZL": true, "T": true}

	uids := make(map[string]string)
	for _, d := range data {
		if d.Portfolio == nil {
			continue
		}
		holdings, err := convert.ToHoldings(d.Account.ID, d.Portfolio)
		if err != nil {
			return err
		}
		for _, h := range holdings {
			if watch[h.Ticker] {
				uids[h.InstrumentUID] = h.Ticker
			}
		}
	}

	for _, e := range events {
		ticker, ok := uids[e.InstrumentUID]
		if !ok || (e.Kind != ledger.KindBuy && e.Kind != ledger.KindSell) || e.Quantity.IsZero() {
			continue
		}
		logger.Info("trade",
			"ticker", ticker,
			"date", e.Date.Format(time.DateOnly),
			"kind", e.Kind,
			"quantity", e.Quantity,
			"amount", e.BaseAmount.Amount.Round(2),
			"price_per_piece", e.BaseAmount.Amount.Div(e.Quantity).Abs().Round(2),
		)
	}
	return nil
}

// checkTotal — временная проверка: общая прибыль за всё время по всем счетам.
func checkTotal(logger *slog.Logger, data []accountData, events []ledger.Event) error {
	value := decimal.Zero
	for _, d := range data {
		if d.Portfolio == nil {
			continue
		}
		total, err := convert.ToDecimal(d.Portfolio.TotalAmountPortfolio.Units, d.Portfolio.TotalAmountPortfolio.Nano)
		if err != nil && !errors.Is(err, convert.ErrNoValue) {
			return err
		}
		value = value.Add(total)
	}

	instrumentOf := make(map[string]string)
	for _, e := range events {
		if e.InstrumentUID != "" {
			instrumentOf[e.OperationID] = e.InstrumentUID
		}
	}

	netIn := decimal.Zero
	unlinked := make(map[ledger.Kind]decimal.Decimal)
	derivatives := decimal.Zero
	for _, e := range events {
		amount := e.BaseAmount.Amount
		switch e.Kind {
		case ledger.KindDeposit, ledger.KindWithdrawal, ledger.KindTransferIn, ledger.KindTransferOut:
			netIn = netIn.Add(amount)
			continue
		case ledger.KindDerivatives:
			derivatives = derivatives.Add(amount)
			continue
		}
		if e.InstrumentUID == "" && instrumentOf[e.ParentID] == "" {
			unlinked[e.Kind] = unlinked[e.Kind].Add(amount)
		}
	}

	logger.Info("total",
		"value_now", value.Round(2),
		"net_deposits", netIn.Round(2),
		"profit", value.Sub(netIn).Round(2),
	)
	logger.Info("not in instruments", "derivatives", derivatives.Round(2))
	for kind, sum := range unlinked {
		logger.Info("not in instruments", "kind", kind, "sum", sum.Round(2))
	}

	byKind := make(map[ledger.Kind]decimal.Decimal)
	byAccount := make(map[string]decimal.Decimal)
	for _, e := range events {
		switch e.Kind {
		case ledger.KindDeposit, ledger.KindWithdrawal, ledger.KindTransferIn, ledger.KindTransferOut,
			ledger.KindDividend, ledger.KindCoupon:
			byKind[e.Kind] = byKind[e.Kind].Add(e.BaseAmount.Amount)
		}
		switch e.Kind {
		case ledger.KindDeposit, ledger.KindWithdrawal, ledger.KindTransferIn, ledger.KindTransferOut:
			byAccount[e.AccountID] = byAccount[e.AccountID].Add(e.BaseAmount.Amount)
		}
	}
	for kind, sum := range byKind {
		logger.Info("by kind", "kind", kind, "sum", sum.Round(2))
	}
	for _, d := range data {
		logger.Info("net in by account",
			"account", d.Account.Name,
			"type", d.Account.Type,
			"status", d.Account.Status,
			"net_in", byAccount[d.Account.ID].Round(2),
		)
	}

	return nil
}
