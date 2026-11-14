package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strings"
	"text/tabwriter"
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

	events, err := buildEvents(data)
	if err != nil {
		return err
	}

	currencies, err := client.GetCurrencies(ctx, tinvest.InstrumentStatusAll)
	if err != nil {
		return fmt.Errorf("get currencies: %w", err)
	}

	rates, err := loadRates(ctx, client, currencies, events)
	if err != nil {
		return fmt.Errorf("load rates: %w", err)
	}
	if err := ledger.ApplyBase(events, rates); err != nil {
		return err
	}
	ledger.InheritInstruments(events)
	currencyByAsset := convert.CurrencyByAsset(currencies)
	events = ledger.AddCurrencyLegs(events, currencyByAsset)
	logger.Info("loaded", "accounts", len(data), "events", len(events), "duration", time.Since(start))

	holdings, err := allHoldings(data)
	if err != nil {
		return err
	}

	printBook(data, holdings, events, currencyByAsset, currencyByUID(currencies))

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

func allHoldings(data []accountData) ([]ledger.Holding, error) {
	var all []ledger.Holding
	for _, d := range data {
		if d.Portfolio == nil {
			continue
		}
		holdings, err := convert.ToHoldings(d.Account.ID, d.Portfolio)
		if err != nil {
			return nil, fmt.Errorf("account %s: %w", d.Account.Name, err)
		}
		all = append(all, holdings...)
	}
	return all, nil
}

func loadRates(ctx context.Context, client *tinvest.Client, currencies []tinvest.Currency, events []ledger.Event) (*fx.Rates, error) {
	needed := make(map[string]bool)
	var earliest time.Time
	for _, e := range events {
		if earliest.IsZero() || e.Date.Before(earliest) {
			earliest = e.Date
		}
		if e.Amount.Currency != baseCurrency {
			needed[e.Amount.Currency] = true
		}
	}
	if len(needed) == 0 {
		return fx.New(baseCurrency, nil), nil
	}

	instruments, err := convert.ExchangeCurrencies(currencies)
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

	return fx.New(baseCurrency, points), nil
}

func currencyByUID(currencies []tinvest.Currency) map[string]string {
	result := make(map[string]string, len(currencies))
	for _, c := range currencies {
		result[c.UID] = strings.ToUpper(c.IsoCurrencyName)
	}
	return result
}

// printBook — временная проверка: все страницы журнала.
func printBook(data []accountData, holdings []ledger.Holding, events []ledger.Event, currencyByAsset, currencyByUID map[string]string) {
	pageKeyByInstrumentUID := metrics.PageKeys(events, currencyByAsset)

	tickers := make(map[string]string)
	for _, d := range data {
		for _, op := range d.Operations {
			if key := pageKeyByInstrumentUID[op.InstrumentUID]; key != "" && op.Ticker != "" {
				tickers[key] = op.Ticker
			}
		}
	}

	values := make(map[string]decimal.Decimal)
	for _, h := range holdings {
		key, ok := pageKeyByInstrumentUID[h.InstrumentUID]
		if !ok {
			key = h.InstrumentUID
			if code, isCurrency := currencyByUID[h.InstrumentUID]; isCurrency && h.InstrumentKind == ledger.InstrumentCurrency {
				key = code
			}
		}
		values[key] = values[key].Add(h.Value())
		tickers[key] = h.Ticker
	}

	type row struct {
		key    string
		f      *metrics.Flows
		value  decimal.Decimal
		profit decimal.Decimal
	}
	var rows []row
	total := decimal.Zero
	for key, f := range metrics.ByInstrument(events, currencyByAsset) {
		v := values[key]
		profit := v.Add(f.Bought).Add(f.Sold).Add(f.Income).Add(f.Fees).Add(f.Taxes).Add(f.Clearing)
		total = total.Add(profit)
		rows = append(rows, row{key, f, v, profit})
	}
	slices.SortFunc(rows, func(a, b row) int { return a.profit.Cmp(b.profit) })

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "тикер\tкуплено\tпродано\tдоходы\tкомиссии\tналоги\tклиринг\tшт\tстоимость\tприбыль\t")
	for _, r := range rows {
		name := tickers[r.key]
		if name == "" {
			name = r.key
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
			name,
			r.f.Bought.StringFixed(2), r.f.Sold.StringFixed(2), r.f.Income.StringFixed(2),
			r.f.Fees.StringFixed(2), r.f.Taxes.StringFixed(2), r.f.Clearing.StringFixed(2),
			r.f.Quantity.String(), r.value.StringFixed(2), r.profit.StringFixed(2))
	}
	netIn := decimal.Zero
	fees := decimal.Zero
	taxes := decimal.Zero
	other := decimal.Zero
	for _, e := range events {
		amount := e.BaseAmount.Amount
		switch e.OperationKind {
		case ledger.KindDeposit, ledger.KindWithdrawal, ledger.KindTransferIn, ledger.KindTransferOut:
			netIn = netIn.Add(amount)
			continue
		case ledger.KindExchange:
			continue
		}
		if e.InstrumentUID != "" {
			continue
		}
		switch e.OperationKind {
		case ledger.KindFee:
			fees = fees.Add(amount)
		case ledger.KindTax:
			taxes = taxes.Add(amount)
		default:
			other = other.Add(amount)
		}
	}

	value := decimal.Zero
	for _, h := range holdings {
		value = value.Add(h.Value())
	}
	profit := value.Sub(netIn)
	unexplained := profit.Sub(total).Sub(fees).Sub(taxes).Sub(other)

	fmt.Fprintf(w, "по бумагам\t\t\t\t\t\t\t\t\t%s\t\n", total.StringFixed(2))
	fmt.Fprintf(w, "комиссии без бумаги\t\t\t\t\t\t\t\t\t%s\t\n", fees.StringFixed(2))
	fmt.Fprintf(w, "налоги без бумаги\t\t\t\t\t\t\t\t\t%s\t\n", taxes.StringFixed(2))
	fmt.Fprintf(w, "прочее без бумаги\t\t\t\t\t\t\t\t\t%s\t\n", other.StringFixed(2))
	fmt.Fprintf(w, "необъяснённое\t\t\t\t\t\t\t\t\t%s\t\n", unexplained.StringFixed(2))
	fmt.Fprintf(w, "ИТОГО\t\t\t\t\t\t\t\t\t%s\t\n", profit.StringFixed(2))
	w.Flush()
}
