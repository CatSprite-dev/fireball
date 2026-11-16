package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"

	"github.com/CatSprite-dev/fireball/v2/internal/convert"
	"github.com/CatSprite-dev/fireball/v2/internal/fx"
	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/snapshot"
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
	cur, err := convert.NewCurrencies(currencies)
	if err != nil {
		return fmt.Errorf("currencies: %w", err)
	}

	rates, err := loadRates(ctx, client, cur, events)
	if err != nil {
		return fmt.Errorf("load rates: %w", err)
	}
	if err := ledger.ApplyBase(events, rates); err != nil {
		return err
	}
	ledger.InheritInstruments(events)
	events = ledger.AddCurrencyLegs(events, cur.ByAsset)
	logger.Info("loaded", "accounts", len(data), "events", len(events), "duration", time.Since(start))

	var robotID string
	for _, d := range data {
		if strings.Contains(strings.ToLower(d.Account.Name), "робот") {
			robotID = d.Account.ID
		}
	}
	if robotID == "" {
		return errors.New("account «робот» not found")
	}
	checkCurrencies(ctx, client, data)
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

func loadRates(ctx context.Context, client *tinvest.Client, cur convert.Currencies, events []ledger.Event) (*fx.Rates, error) {
	needed := make(map[string]bool)
	var earliest time.Time
	for _, e := range events {
		if earliest.IsZero() || e.Date.Before(earliest) {
			earliest = e.Date
		}
		if e.Amount.Currency != baseCurrency {
			needed[e.Amount.Currency] = true
		}
		if code, ok := cur.ByAsset[e.AssetUID]; ok {
			needed[code] = true
		}
	}
	if len(needed) == 0 {
		return fx.New(baseCurrency, nil), nil
	}

	from := earliest.AddDate(0, 0, -14)
	to := time.Now()

	points := make(map[string]series.Series, len(needed))
	for currency := range needed {
		instrument, ok := cur.Exchange[currency]
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

// loadPrices — ряд цен на каждую позицию: валюте курс, остальным свечи всех её uid,
// а где свечей нет — цены своих сделок.
func loadPrices(ctx context.Context, client *tinvest.Client, positions map[string]*ledger.Flows, uidsByKey map[string][]string, tradePrices map[string]series.Series, rates *fx.Rates, from, to time.Time) (map[string]series.Series, error) {
	prices := make(map[string]series.Series)
	for key, f := range positions {
		if f.Kind == ledger.InstrumentCurrency {
			if s, ok := rates.Series(key); ok {
				prices[key] = s
			}
			continue
		}

		var parts []series.Series
		for _, uid := range uidsByKey[key] {
			candles, err := client.GetCandles(ctx, uid, from, to, tinvest.CandleIntervalDay)
			if err != nil {
				continue
			}
			points, err := convert.ToPoints(candles.Candles, decimal.NewFromInt(1))
			if err != nil {
				return nil, fmt.Errorf("prices %s: %w", uid, err)
			}
			parts = append(parts, points)
		}
		prices[key] = series.Merge(append(parts, tradePrices[key])...)
	}
	return prices, nil
}

// loadBonds — номинал и НКД по дням для каждой облигации, склеенные по всем её uid.
func loadBonds(ctx context.Context, client *tinvest.Client, positions map[string]*ledger.Flows, uidsByKey map[string][]string, from, to time.Time) (map[string]snapshot.Bond, error) {
	bonds := make(map[string]snapshot.Bond)
	for key, f := range positions {
		if f.Kind != ledger.InstrumentBond {
			continue
		}
		var nominals, nkds []series.Series
		for _, uid := range uidsByKey[key] {
			list, err := client.GetAccruedInterests(ctx, uid, from, to)
			if err != nil {
				continue
			}
			nominal, nkd, err := convert.ToAccrued(list)
			if err != nil {
				return nil, fmt.Errorf("accrued %s: %w", uid, err)
			}
			nominals = append(nominals, nominal)
			nkds = append(nkds, nkd)
		}
		bonds[key] = snapshot.Bond{Nominal: series.Merge(nominals...), NKD: series.Merge(nkds...)}
	}
	return bonds, nil
}

// checkRobot — временная проверка: снимок счёта «робот» за дни, которые есть в снимках proporcia.
func checkRobot(ctx context.Context, client *tinvest.Client, data []accountData, events []ledger.Event, rates *fx.Rates, cur convert.Currencies) error {
	if len(events) == 0 {
		return errors.New("no events on account «робот»")
	}
	from := events[0].Date
	to := time.Now()

	positions := ledger.ByInstrument(events, cur.ByAsset)
	pageKeys := ledger.PageKeys(events, cur.ByAsset)

	uidsByKey := make(map[string][]string)
	for uid, key := range pageKeys {
		uidsByKey[key] = append(uidsByKey[key], uid)
	}

	prices, err := loadPrices(ctx, client, positions, uidsByKey, ledger.TradePrices(events, cur.ByAsset), rates, from, to)
	if err != nil {
		return err
	}
	bonds, err := loadBonds(ctx, client, positions, uidsByKey, from, to)
	if err != nil {
		return err
	}

	rows, gaps := snapshot.Build(positions, prices, bonds, ledger.CashHistory(events, baseCurrency), baseCurrency, from, to)

	tickers := map[string]string{baseCurrency: baseCurrency}
	for _, d := range data {
		for _, op := range d.Operations {
			if key := pageKeys[op.InstrumentUID]; key != "" && op.Ticker != "" {
				tickers[key] = op.Ticker
			}
		}
	}

	fmt.Printf("робот: событий %d, строк %d, дыр %d\n", len(events), len(rows), len(gaps))
	for _, g := range gaps {
		fmt.Printf("дыра  %-8s %s — %s\n", tickers[g.Key], g.From.Format(time.DateOnly), g.To.Format(time.DateOnly))
	}

	first := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	last := time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC)
	for _, r := range rows {
		if r.Date.Before(first) || r.Date.After(last) {
			continue
		}
		fmt.Printf("%s  %-6s шт %-8s цена %-9s стоимость %10s  цена от %s\n",
			r.Date.Format(time.DateOnly), tickers[r.Key], r.Quantity, r.Price.Value.StringFixed(2),
			r.Value.StringFixed(2), r.Price.Date.Format(time.DateOnly))
	}
	return nil
}

// dumpSnapshot — временная выгрузка снимка всех счетов в CSV для проверки.
func dumpSnapshot(ctx context.Context, client *tinvest.Client, data []accountData, events []ledger.Event, rates *fx.Rates, cur convert.Currencies, path string) error {
	from := events[0].Date
	to := time.Now()

	positions := ledger.ByInstrument(events, cur.ByAsset)
	pageKeys := ledger.PageKeys(events, cur.ByAsset)

	uidsByKey := make(map[string][]string)
	for uid, key := range pageKeys {
		uidsByKey[key] = append(uidsByKey[key], uid)
	}

	prices, err := loadPrices(ctx, client, positions, uidsByKey, ledger.TradePrices(events, cur.ByAsset), rates, from, to)
	if err != nil {
		return err
	}
	bonds, err := loadBonds(ctx, client, positions, uidsByKey, from, to)
	if err != nil {
		return err
	}
	rows, gaps := snapshot.Build(positions, prices, bonds, ledger.CashHistory(events, baseCurrency), baseCurrency, from, to)

	tickers := map[string]string{baseCurrency: baseCurrency}
	for _, d := range data {
		for _, op := range d.Operations {
			if key := pageKeys[op.InstrumentUID]; key != "" && op.Ticker != "" {
				tickers[key] = op.Ticker
			}
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Write([]string{"date", "key", "ticker", "kind", "quantity", "price", "price_date", "nkd", "value"})
	for _, r := range rows {
		kind := "cash"
		if p, ok := positions[r.Key]; ok {
			kind = string(p.Kind)
		}
		w.Write([]string{
			r.Date.Format(time.DateOnly), r.Key, tickers[r.Key], kind,
			r.Quantity.String(), r.Price.Value.String(), r.Price.Date.Format(time.DateOnly),
			r.NKD.String(), r.Value.String(),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	fmt.Printf("записано строк: %d, дыр: %d → %s\n", len(rows), len(gaps), path)
	return nil
}

func dumpOps(data []accountData, from, to time.Time) {
	for _, d := range data {
		for _, op := range d.Operations {
			if op.Date.Before(from) || op.Date.After(to) {
				continue
			}
			fmt.Printf("%s  %-8s %-45s qty=%-6s pay=%v  %q  %s\n",
				op.Date.Format("2006-01-02 15:04"), op.Ticker, op.Type,
				op.Quantity, op.Payment, op.Name, op.Description)
		}
	}
}

func checkCurrencies(ctx context.Context, client *tinvest.Client, data []accountData) error {
	want := map[string]bool{"TGRN": true, "TBIO": true, "TEUR": true, "TUSD": true, "TSPX": true}
	seen := map[string]bool{}
	for _, d := range data {
		for _, op := range d.Operations {
			if !want[op.Ticker] || op.InstrumentUID == "" || seen[op.InstrumentUID] {
				continue
			}
			seen[op.InstrumentUID] = true
			inst, err := client.GetInstrumentBy(ctx, tinvest.InstrumentIDUID, "", op.InstrumentUID)
			if err != nil {
				return fmt.Errorf("instrument %s: %w", op.Ticker, err)
			}
			fmt.Printf("%-5s uid=%s class=%-6s currency=%s  payment=%s\n",
				inst.Ticker, inst.UID, inst.ClassCode, inst.Currency, op.Payment.Currency)
		}
	}
	return nil
}
