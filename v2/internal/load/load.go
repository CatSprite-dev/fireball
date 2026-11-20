// Package load — загрузка данных брокера: счета, операции, курсы, цены, индекс.
// Вместе с tinvest ходит в сеть; остальные пакеты получают данные готовыми.
package load

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/CatSprite-dev/fireball/v2/internal/convert"
	"github.com/CatSprite-dev/fireball/v2/internal/fx"
	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

var excludedAccountTypes = map[string]bool{
	"ACCOUNT_TYPE_DFA": true,
}

type Loader struct {
	client *tinvest.Client
	base   string
	logger *slog.Logger
}

// New — загрузчик с клиентом брокера, базовой валютой base и логгером.
func New(client *tinvest.Client, base string, logger *slog.Logger) *Loader {
	return &Loader{client: client, base: base, logger: logger}
}

// Account — счёт с операциями и портфелем; у закрытого счёта портфеля нет.
type Account struct {
	Account    tinvest.Account
	Operations []tinvest.Operation
	Portfolio  *tinvest.Portfolio
}

// Portfolio — все счета и книга событий в базовой валюте.
type Portfolio struct {
	Accounts   []Account
	Events     []ledger.Event
	Currencies convert.Currencies
	Rates      *fx.Rates
}

// Portfolio — счета, операции, валюты и курсы; события книги уже в базовой валюте.
func (l *Loader) Portfolio(ctx context.Context) (*Portfolio, error) {
	start := time.Now()
	accounts, err := l.client.GetAccounts(ctx, tinvest.AccountStatusAll)
	if err != nil {
		return nil, fmt.Errorf("get accounts: %w", err)
	}

	accounts = filterAccounts(accounts)
	if len(accounts) == 0 {
		return nil, errors.New("no supported accounts")
	}

	data, err := l.loadAccounts(ctx, accounts)
	if err != nil {
		return nil, fmt.Errorf("load accounts: %w", err)
	}

	events, err := buildEvents(data)
	if err != nil {
		return nil, err
	}

	currencies, err := l.client.GetCurrencies(ctx, tinvest.InstrumentStatusAll)
	if err != nil {
		return nil, fmt.Errorf("get currencies: %w", err)
	}
	cur, err := convert.NewCurrencies(currencies)
	if err != nil {
		return nil, fmt.Errorf("currencies: %w", err)
	}

	rates, err := l.loadRates(ctx, cur, events)
	if err != nil {
		return nil, fmt.Errorf("load rates: %w", err)
	}
	if err := ledger.ApplyBase(events, rates); err != nil {
		return nil, err
	}
	ledger.InheritInstruments(events)
	events = ledger.AddCurrencyLegs(events, cur.ByAsset)

	l.logger.Info("loaded", "accounts", len(data), "events", len(events), "duration", time.Since(start))
	return &Portfolio{Accounts: data, Events: events, Currencies: cur, Rates: rates}, nil
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

func (l *Loader) loadAccounts(ctx context.Context, accounts []tinvest.Account) ([]Account, error) {
	results := make([]Account, len(accounts))
	now := time.Now()

	g, gctx := errgroup.WithContext(ctx)
	for i, acc := range accounts {
		g.Go(func() error {
			ops, err := l.client.GetOperationsByCursor(gctx, tinvest.OperationsFilter{
				AccountID: acc.ID,
				From:      &acc.OpenedDate,
				To:        &now,
			})
			if err != nil {
				return fmt.Errorf("account %s: %w", acc.ID, err)
			}
			var portfolio *tinvest.Portfolio
			if acc.Status != "ACCOUNT_STATUS_CLOSED" {
				portfolio, err = l.client.GetPortfolio(gctx, acc.ID, l.base)
				if err != nil {
					return fmt.Errorf("account %s: %w", acc.ID, err)
				}
			}
			results[i] = Account{Account: acc, Operations: ops, Portfolio: portfolio}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

func buildEvents(data []Account) ([]ledger.Event, error) {
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

func (l *Loader) loadRates(ctx context.Context, cur convert.Currencies, events []ledger.Event) (*fx.Rates, error) {
	needed := make(map[string]bool)
	var earliest time.Time
	for _, e := range events {
		if earliest.IsZero() || e.Date.Before(earliest) {
			earliest = e.Date
		}
		if e.Amount.Currency != l.base {
			needed[e.Amount.Currency] = true
		}
		if code, ok := cur.ByAsset[e.AssetUID]; ok {
			needed[code] = true
		}
	}
	if len(needed) == 0 {
		return fx.New(l.base, nil), nil
	}

	from := earliest.AddDate(0, 0, -45)
	to := time.Now()

	points := make(map[string]series.Series, len(needed))
	for currency := range needed {
		instrument, ok := cur.Exchange[currency]
		if !ok {
			return nil, fmt.Errorf("no exchange instrument for currency %s", currency)
		}

		candles, err := l.client.GetCandles(ctx, instrument.UID, from, to, tinvest.CandleIntervalDay)
		if err != nil {
			return nil, fmt.Errorf("candles %s: %w", currency, err)
		}

		currencyPoints, err := convert.ToPoints(candles.Candles, instrument.Nominal)
		if err != nil {
			return nil, fmt.Errorf("points %s: %w", currency, err)
		}
		points[currency] = currencyPoints
	}

	return fx.New(l.base, points), nil
}
