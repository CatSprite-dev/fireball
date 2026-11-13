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

	rates, err := loadRates(ctx, client, events)
	if err != nil {
		return fmt.Errorf("load rates: %w", err)
	}
	if err := ledger.ApplyBase(events, rates); err != nil {
		return err
	}
	logger.Info("loaded", "accounts", len(data), "events", len(events), "duration", time.Since(start))

	holdings, err := allHoldings(data)
	if err != nil {
		return err
	}

	checkTotal(logger, holdings, events)
	checkCash(logger, data, holdings, events)
	checkFutures(logger, data, events)

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

// checkTotal — временная проверка: общая прибыль «сверху» и «снизу».
func checkTotal(logger *slog.Logger, holdings []ledger.Holding, events []ledger.Event) {
	value := decimal.Zero
	values := make(map[string]decimal.Decimal)
	for _, h := range holdings {
		value = value.Add(h.Value())
		values[h.InstrumentUID] = values[h.InstrumentUID].Add(h.Value())
	}

	instrumentOf := make(map[string]string)
	for _, e := range events {
		if e.InstrumentUID != "" {
			instrumentOf[e.OperationID] = e.InstrumentUID
		}
	}

	netIn := decimal.Zero
	derivatives := decimal.Zero
	unlinked := decimal.Zero
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
			unlinked = unlinked.Add(amount)
		}
	}

	instruments := decimal.Zero
	for uid, f := range metrics.ByInstrument(events) {
		instruments = instruments.Add(values[uid]).Add(f.Bought).Add(f.Sold).Add(f.Income).Add(f.Fees).Add(f.Taxes)
	}

	top := value.Sub(netIn)
	bottom := instruments.Add(derivatives).Add(unlinked)
	logger.Info("total", "value_now", value.Round(2), "net_in", netIn.Round(2), "profit", top.Round(2))
	logger.Info("bottom-up",
		"instruments", instruments.Round(2),
		"derivatives", derivatives.Round(2),
		"unlinked", unlinked.Round(2),
		"sum", bottom.Round(2),
		"diff", top.Sub(bottom).Round(2),
	)
}

// checkCash — временная проверка: деньги на счёте против суммы всех движений.
func checkCash(logger *slog.Logger, data []accountData, holdings []ledger.Holding, events []ledger.Event) {
	flows := make(map[string]decimal.Decimal)
	byCurrency := make(map[string]map[string]decimal.Decimal)
	for _, e := range events {
		flows[e.AccountID] = flows[e.AccountID].Add(e.BaseAmount.Amount)
		if byCurrency[e.AccountID] == nil {
			byCurrency[e.AccountID] = make(map[string]decimal.Decimal)
		}
		cur := e.Amount.Currency
		byCurrency[e.AccountID][cur] = byCurrency[e.AccountID][cur].Add(e.Amount.Amount)
	}

	cash := make(map[string]decimal.Decimal)
	for _, h := range holdings {
		if h.Type == "currency" {
			cash[h.AccountID] = cash[h.AccountID].Add(h.Value())
		}
	}

	for _, d := range data {
		id := d.Account.ID
		diff := cash[id].Sub(flows[id])
		if diff.Abs().LessThan(decimal.NewFromInt(1)) {
			continue
		}
		logger.Info("cash mismatch",
			"account", d.Account.Name,
			"cash_broker", cash[id].Round(2),
			"sum_of_events", flows[id].Round(2),
			"diff", diff.Round(2),
		)
		for cur, sum := range byCurrency[id] {
			logger.Info("  events by currency", "currency", cur, "sum", sum.Round(2))
		}
		for _, h := range holdings {
			if h.AccountID == id && h.Type == "currency" {
				logger.Info("  broker cash", "ticker", h.Ticker, "quantity", h.Quantity)
			}
		}
	}
}

// checkFXTrades — временная проверка: сделки с самой валютой (покупка и продажа долларов и евро).
func checkFXTrades(logger *slog.Logger, holdings []ledger.Holding, events []ledger.Event) {
	fxUIDs := map[string]string{
		"a22a1263-8e1b-4546-a1aa-416463f104d3": "USD000UTSTOM (CETS)",
		"b1c06e5e-f5d4-4ff5-8cb3-dcacccd933da": "EUR_RUB__TOM (CETS)",
	}
	for _, h := range holdings {
		if h.Type == "currency" && h.Ticker != "RUB000UTSTOM" {
			fxUIDs[h.InstrumentUID] = h.Ticker + " (portfolio)"
		}
	}

	byOperation := make(map[string]string)
	for _, e := range events {
		if _, ok := fxUIDs[e.InstrumentUID]; ok {
			byOperation[e.OperationID] = e.InstrumentUID
		}
	}

	type total struct{ rub, qty decimal.Decimal }
	totals := make(map[string]*total)

	for _, e := range events {
		uid := e.InstrumentUID
		if uid == "" {
			uid = byOperation[e.ParentID]
		}
		name, ok := fxUIDs[uid]
		if !ok {
			continue
		}

		t := totals[name]
		if t == nil {
			t = &total{}
			totals[name] = t
		}
		t.rub = t.rub.Add(e.BaseAmount.Amount)
		switch e.Kind {
		case ledger.KindBuy:
			t.qty = t.qty.Add(e.Quantity)
		case ledger.KindSell:
			t.qty = t.qty.Sub(e.Quantity)
		}

		logger.Info("fx trade",
			"instrument", name,
			"date", e.Date.Format(time.DateOnly),
			"kind", e.Kind,
			"amount", e.Amount.Amount,
			"currency", e.Amount.Currency,
			"quantity", e.Quantity,
		)
	}

	for name, t := range totals {
		logger.Info("fx total", "instrument", name, "rub", t.rub.Round(2), "quantity", t.qty)
	}
}

// checkDays — временная проверка: все события счёта за выбранные даты.
func checkDays(logger *slog.Logger, events []ledger.Event) {
	days := map[string]bool{
		"2021-08-30": true, // первая покупка за доллары
		"2022-11-24": true, // комиссия за сделку с долларом
		"2023-07-24": true, // комиссия + продажа за 74 $
	}
	for _, e := range events {
		if !days[e.Date.Format(time.DateOnly)] {
			continue
		}
		logger.Info("day",
			"date", e.Date.Format(time.DateTime),
			"account", e.AccountID,
			"kind", e.Kind,
			"instrument", e.InstrumentUID,
			"op", e.OperationID,
			"parent", e.ParentID,
			"amount", e.Amount.Amount,
			"currency", e.Amount.Currency,
			"quantity", e.Quantity,
		)
	}
}

// checkFutures — временная проверка: сделки с фьючерсами против вариационной маржи.
func checkFutures(logger *slog.Logger, data []accountData, events []ledger.Event) {
	trades := make(map[string]decimal.Decimal)
	margin := make(map[string]decimal.Decimal)
	for _, e := range events {
		if e.Kind != ledger.KindDerivatives {
			continue
		}
		if e.InstrumentUID != "" {
			trades[e.AccountID] = trades[e.AccountID].Add(e.BaseAmount.Amount)
		} else {
			margin[e.AccountID] = margin[e.AccountID].Add(e.BaseAmount.Amount)
		}
	}
	for _, d := range data {
		id := d.Account.ID
		if trades[id].IsZero() && margin[id].IsZero() {
			continue
		}
		logger.Info("futures",
			"account", d.Account.Name,
			"trades_payment", trades[id].Round(2),
			"varmargin", margin[id].Round(2),
		)
	}
}
