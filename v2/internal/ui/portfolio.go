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

// Overview — сводка портфеля для окна: стоимость на сегодня, остальное — за период. Деньги — строкой, чтобы не терять копейки во float.
type Overview struct {
	Value    string `json:"value"`    // стоимость в базовой валюте с поправкой на необъяснённое расхождение, "1044651.48"
	AsOf     string `json:"asOf"`     // день оценки, "2006-01-02"
	Invested string `json:"invested"` // пополнения минус выводы за период
	Profit   string `json:"profit"`   // стоимость сегодня минус стоимость накануне периода минус вложено за период
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
	mu     sync.Mutex
	loaded *loaded
}

// loaded — всё, что сервис загрузил у брокера.
type loaded struct {
	portfolio *load.Portfolio
	market    *load.Market
	indexes   map[string]series.Series // тикер индекса → значения по торговым дням
	today     time.Time
}

// indexTickers — индексы, с которыми окно умеет сравнивать.
var indexTickers = []string{"MCFTR"}

// NewPortfolioService — сервис портфеля для окна поверх загрузчика load.
func NewPortfolioService(loader *load.Loader, base string, logger *slog.Logger) *PortfolioService {
	return &PortfolioService{loader: loader, base: base, logger: logger}
}

// Overview — стоимость выбранных счетов на сегодня и прибыль за период: period — как в Performance.
// Пустой список счетов — нули.
func (s *PortfolioService) Overview(ctx context.Context, accountIDs []string, period string) (Overview, error) {
	d, err := s.data(ctx)
	if err != nil {
		return Overview{}, err
	}
	from, err := periodStart(period, d.today, ledger.Filter(d.portfolio.Events, accountIDs))
	if err != nil {
		return Overview{}, err
	}
	return overview(d.portfolio, d.market, s.base, from, d.today, accountIDs, period == "all")
}

// overview — карточки по счетам accountIDs. Стоимость поправлена на необъяснённое расхождение
// с брокером и совпадает с ним. За всё время расхождение входит и в прибыль; в другом периоде
// стоимость на начало тоже считается поправленной, и прибыль за период оно не задевает.
func overview(p *load.Portfolio, m *load.Market, base string, from, today time.Time, accountIDs []string, allTime bool) (Overview, error) {
	end, change, err := summarize(p, m, base, from, today, accountIDs)
	if err != nil {
		return Overview{}, err
	}
	gap, err := unexplainedCash(p, base, today, accountIDs)
	if err != nil {
		return Overview{}, err
	}

	profit := change.Value.Sub(change.Invested)
	if allTime {
		profit = profit.Add(gap)
	}
	return Overview{
		Value:    end.Value.Add(gap).StringFixed(2),
		AsOf:     today.Format(time.DateOnly),
		Invested: change.Invested.StringFixed(2),
		Profit:   profit.StringFixed(2),
		Realized: change.Realized.StringFixed(2),
		Paper:    change.Paper.StringFixed(2),
		Income:   change.Income.StringFixed(2),
	}, nil
}

// Accounts — счета портфеля в порядке, как у брокера.
func (s *PortfolioService) Accounts(ctx context.Context) ([]Account, error) {
	d, err := s.data(ctx)
	if err != nil {
		return nil, err
	}
	return toAccounts(d.portfolio.Accounts), nil
}

// data — загруженные данные; при первом вызове грузит их: рынок и индексы — с первой операции
// минус 30 дней по сегодня. Ошибка не запоминается: следующий вызов пробует загрузить снова.
func (s *PortfolioService) data(ctx context.Context) (*loaded, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loaded != nil {
		return s.loaded, nil
	}

	start := time.Now()
	today := series.Day(time.Now())
	p, err := s.loader.Portfolio(ctx)
	if err != nil {
		return nil, err
	}
	from := today
	if len(p.Events) > 0 {
		from = series.Day(p.Events[0].Date)
	}
	from = from.AddDate(0, 0, -30)

	m, err := s.loader.Market(ctx, p, from, time.Now())
	if err != nil {
		return nil, err
	}
	// Индекса может не быть у брокера: без него окно работает, сравнение с ним вернёт ошибку.
	indexes := make(map[string]series.Series, len(indexTickers))
	for _, ticker := range indexTickers {
		index, err := s.loader.Index(ctx, ticker, from, time.Now())
		if err != nil {
			s.logger.Warn("index not loaded", "index", ticker, "error", err)
			continue
		}
		indexes[ticker] = index
	}

	s.loaded = &loaded{portfolio: p, market: m, indexes: indexes, today: today}
	s.logger.Info("data loaded", "duration", time.Since(start))
	return s.loaded, nil
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

// summary — стоимость счетов на день и из чего складывается прибыль к этому дню.
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

// summarize — сводка по счетам accountIDs: end — на сегодня, change — изменение за период с from
// по today («на сегодня» минус «накануне from»). Каждый счёт считается отдельно, потом всё складывается:
// так средняя цена у каждого счёта своя, как у брокера и налоговой — если одна бумага лежит
// на двух счетах, продажа на одном не списывается по общей средней.
func summarize(p *load.Portfolio, m *load.Market, base string, from, today time.Time, accountIDs []string) (end, change summary, err error) {
	before := from.AddDate(0, 0, -1)
	for _, id := range accountIDs {
		events := ledger.Filter(p.Events, []string{id})
		e, err := summarizeAt(p, m, base, events, today, id)
		if err != nil {
			return summary{}, summary{}, err
		}
		var earlier []ledger.Event
		for _, ev := range events {
			if series.Day(ev.Date).Before(from) {
				earlier = append(earlier, ev)
			}
		}
		b, err := summarizeAt(p, m, base, earlier, before, id)
		if err != nil {
			return summary{}, summary{}, err
		}
		end, change = end.add(e), change.add(e.sub(b))
	}

	if parts, profit := change.parts(), change.Value.Sub(change.Invested); !parts.Equal(profit) {
		return summary{}, summary{}, fmt.Errorf("summary %s — %s: profit parts %s, profit %s",
			from.Format(time.DateOnly), today.Format(time.DateOnly), parts.StringFixed(2), profit.StringFixed(2))
	}
	return end, change, nil
}

// summarizeAt — сводка одного счёта по событиям events на день day. Если у позиции нет цены — ошибка:
// без неё стоимость молча вышла бы меньше настоящей. Если части прибыли не сходятся
// с прибылью — тоже ошибка: значит, какой-то поток денег не попал ни в одну часть.
func summarizeAt(p *load.Portfolio, m *load.Market, base string, events []ledger.Event, day time.Time, accountID string) (summary, error) {
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

	if parts, profit := sum.parts(), sum.Value.Sub(sum.Invested); !parts.Equal(profit) {
		return summary{}, fmt.Errorf("account %s: summary on %s: profit parts %s, profit %s", accountID, day.Format(time.DateOnly), parts.StringFixed(2), profit.StringFixed(2))
	}
	return sum, nil
}

// parts — сумма частей прибыли; по тождеству равна Value − Invested.
func (s summary) parts() decimal.Decimal {
	return s.Realized.Add(s.Paper).Add(s.Income).Add(s.Clearing).Add(s.Account)
}

func (s summary) sub(o summary) summary {
	return s.add(summary{
		Value:    o.Value.Neg(),
		Invested: o.Invested.Neg(),
		Realized: o.Realized.Neg(),
		Paper:    o.Paper.Neg(),
		Income:   o.Income.Neg(),
		Clearing: o.Clearing.Neg(),
		Account:  o.Account.Neg(),
	})
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
