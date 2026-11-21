// Package ui — сервисы и типы для окна. Ядро про него не знает.
package ui

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/load"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/snapshot"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

// Overview — сводка портфеля для окна. Деньги — строкой, чтобы не терять копейки во float.
type Overview struct {
	Value    string `json:"value"`    // стоимость портфеля в базовой валюте, decimal строкой "1044651.48"
	AsOf     string `json:"asOf"`     // день оценки, "2006-01-02"
	Invested string `json:"invested"` // пополнения минус выводы за всё время
	Profit   string `json:"profit"`   // стоимость минус вложено
	Realized string `json:"realized"` // зафиксированная прибыль по сделкам с комиссиями
	Paper    string `json:"paper"`    // бумажная прибыль: стоимость позиций минус их стоимость остатка
	Income   string `json:"income"`   // выплаты за вычетом налогов на них
}

// Account — счёт для окна.
type Account struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Closed bool   `json:"closed"`
}

type PortfolioService struct {
	loader *load.Loader
	base   string
	logger *slog.Logger

	// Временно, до базы данных: данные брокера грузятся один раз при первом вызове.
	// mu не даёт двум вызовам при открытии окна грузить одновременно.
	mu        sync.Mutex
	portfolio *load.Portfolio
	market    *load.Market
	today     time.Time
}

// NewPortfolioService — сервис портфеля для окна поверх загрузчика load.
func NewPortfolioService(loader *load.Loader, base string, logger *slog.Logger) *PortfolioService {
	return &PortfolioService{loader: loader, base: base, logger: logger}
}

// Overview — стоимость выбранных счетов на сегодня. Пустой список — стоимость 0.
func (s *PortfolioService) Overview(ctx context.Context, accountIDs []string) (Overview, error) {
	p, m, today, err := s.data(ctx)
	if err != nil {
		return Overview{}, err
	}
	sum, err := summarize(p, m, s.base, today, accountIDs)
	if err != nil {
		return Overview{}, err
	}

	profit := sum.Value.Sub(sum.Invested)
	return Overview{
		Value:    sum.Value.StringFixed(2),
		AsOf:     today.Format(time.DateOnly),
		Invested: sum.Invested.StringFixed(2),
		Profit:   profit.StringFixed(2),
		Realized: sum.Realized.StringFixed(2),
		Paper:    sum.Paper.StringFixed(2),
		Income:   sum.Income.StringFixed(2),
	}, nil
}

// Accounts — счета портфеля в порядке, как у брокера.
func (s *PortfolioService) Accounts(ctx context.Context) ([]Account, error) {
	p, _, _, err := s.data(ctx)
	if err != nil {
		return nil, err
	}
	return toAccounts(p.Accounts), nil
}

// data — загруженные портфель и рынок; при первом вызове грузит их. Ошибка не запоминается:
// следующий вызов пробует загрузить снова.
func (s *PortfolioService) data(ctx context.Context) (*load.Portfolio, *load.Market, time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.portfolio != nil {
		return s.portfolio, s.market, s.today, nil
	}

	today := series.Day(time.Now())
	p, err := s.loader.Portfolio(ctx)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	m, err := s.loader.Market(ctx, p, today.AddDate(0, 0, -30), time.Now())
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	s.portfolio, s.market, s.today = p, m, today
	return p, m, today, nil
}

func toAccounts(accounts []load.Account) []Account {
	result := make([]Account, 0, len(accounts))
	for _, a := range accounts {
		result = append(result, Account{
			ID:     a.Account.ID,
			Name:   a.Account.Name,
			Closed: a.Account.Status == string(tinvest.AccountStatusClosed),
		})
	}
	return result
}

// summary — стоимость счетов на день и из чего складывается прибыль за всё время.
// Тождество: Realized + Paper + Income + Clearing + Account = Value − Invested.
type summary struct {
	Value    decimal.Decimal // стоимость позиций и денег
	Invested decimal.Decimal // пополнения минус выводы
	Realized decimal.Decimal // зафиксированная прибыль по сделкам с комиссиями
	Paper    decimal.Decimal // стоимость позиций минус их стоимость остатка
	Income   decimal.Decimal // выплаты за вычетом налогов на них
	Clearing decimal.Decimal // вариационная маржа и экспирация
	Account  decimal.Decimal // комиссии и налоги без бумаги
}

// summarize — сводка по счетам accountIDs на день day: каждый счёт считается отдельно, потом
// всё складывается. Так средняя цена у каждого счёта своя, как у брокера и налоговой:
// если одна бумага лежит на двух счетах, продажа на одном не списывается по общей средней.
func summarize(p *load.Portfolio, m *load.Market, base string, day time.Time, accountIDs []string) (summary, error) {
	var total summary
	for _, id := range accountIDs {
		sum, err := summarizeAccount(p, m, base, day, id)
		if err != nil {
			return summary{}, err
		}
		total = total.add(sum)
	}
	return total, nil
}

// summarizeAccount — сводка по одному счёту. Если у позиции нет цены — ошибка:
// без неё стоимость молча вышла бы меньше настоящей. Если части прибыли не сходятся
// с прибылью — тоже ошибка: значит, какой-то поток денег не попал ни в одну часть.
func summarizeAccount(p *load.Portfolio, m *load.Market, base string, day time.Time, accountID string) (summary, error) {
	events := ledger.Filter(p.Events, []string{accountID})
	positions := ledger.ByInstrument(events, p.Currencies.ByAsset)
	cash := ledger.CashHistory(events, base)
	rows, gaps := snapshot.Build(positions, m.Prices, m.Bonds, cash, base, day, day)
	if len(gaps) > 0 {
		return summary{}, fmt.Errorf("account %s: value on %s: no price for %d positions, first %s", accountID, day.Format(time.DateOnly), len(gaps), gaps[0].Key)
	}

	var sum summary
	valueByKey := make(map[string]decimal.Decimal)
	for _, r := range rows {
		sum.Value = sum.Value.Add(r.Value)
		valueByKey[r.Key] = valueByKey[r.Key].Add(r.Value)
	}
	for key, f := range positions {
		sum.Realized = sum.Realized.Add(f.Realized).Add(f.Fees)
		sum.Paper = sum.Paper.Add(valueByKey[key]).Sub(f.Cost)
		sum.Income = sum.Income.Add(f.Income).Add(f.Taxes)
		sum.Clearing = sum.Clearing.Add(f.Clearing)
	}
	for _, e := range events {
		switch e.OperationKind {
		case ledger.KindDeposit, ledger.KindWithdrawal, ledger.KindTransferIn, ledger.KindTransferOut:
			sum.Invested = sum.Invested.Add(e.BaseAmount.Amount)
		case ledger.KindFee, ledger.KindTax:
			if e.InstrumentUID == "" && e.AssetUID == "" {
				sum.Account = sum.Account.Add(e.BaseAmount.Amount)
			}
		}
	}

	parts := sum.Realized.Add(sum.Paper).Add(sum.Income).Add(sum.Clearing).Add(sum.Account)
	if profit := sum.Value.Sub(sum.Invested); !parts.Equal(profit) {
		return summary{}, fmt.Errorf("account %s: summary on %s: profit parts %s, profit %s", accountID, day.Format(time.DateOnly), parts.StringFixed(2), profit.StringFixed(2))
	}
	return sum, nil
}

func (s summary) add(o summary) summary {
	return summary{
		Value:    s.Value.Add(o.Value),
		Invested: s.Invested.Add(o.Invested),
		Realized: s.Realized.Add(o.Realized),
		Paper:    s.Paper.Add(o.Paper),
		Income:   s.Income.Add(o.Income),
		Clearing: s.Clearing.Add(o.Clearing),
		Account:  s.Account.Add(o.Account),
	}
}
