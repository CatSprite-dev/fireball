package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/convert"
	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/load"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/snapshot"
)

// Position — строка таблицы бумаг. Деньги и проценты — строкой.
type Position struct {
	Key       string `json:"key"`       // ключ позиции из книги
	Ticker    string `json:"ticker"`    // тикер из последней операции по позиции; нет операций — ключ
	Name      string `json:"name"`      // имя из последней операции по позиции
	Kind      string `json:"kind"`      // группа: etf, bond, share, futures, option, currency, other
	Quantity  string `json:"quantity"`  // количество на сегодня
	Closed    bool   `json:"closed"`    // количество 0
	Value     string `json:"value"`     // стоимость сегодня
	Share     string `json:"share"`     // доля от стоимости всех выбранных счетов, %
	Invested  string `json:"invested"`  // стоимость остатка на сегодня (по средней цене)
	Income    string `json:"income"`    // выплаты за период за вычетом налогов на них
	Profit    string `json:"profit"`    // зафиксировано + изменение бумажной + выплаты + клиринг за период
	ProfitPct string `json:"profitPct"` // пока пусто
}

// PositionTable — таблица бумаг: позиции, рубли отдельной строкой, то, что на уровне счёта, и итоги.
// Итоги считаются здесь: окно деньги не складывает.
type PositionTable struct {
	Positions    []Position   `json:"positions"`
	Groups       []GroupTotal `json:"groups"`       // итоги по группам, вместе с закрытыми позициями
	AccountLevel string       `json:"accountLevel"` // комиссии и налоги без бумаги за период
	Unexplained  string       `json:"unexplained"`  // рубли у брокера минус наши — в стоимости, в любом периоде
	// UnexplainedProfit — то же расхождение в прибыли: только за всё время, иначе пусто.
	UnexplainedProfit string `json:"unexplainedProfit"`
	TotalValue        string `json:"totalValue"`  // стоимость всех строк с расхождением; совпадает с карточкой «Стоимость»
	TotalProfit       string `json:"totalProfit"` // прибыль всех строк и уровня счёта; совпадает с карточкой «Прибыль»
}

// GroupTotal — итог группы таблицы.
type GroupTotal struct {
	Kind   string `json:"kind"`
	Value  string `json:"value"`
	Profit string `json:"profit"`
}

// Positions — таблица бумаг по счетам accountIDs за период period (как в Overview).
func (s *PortfolioService) Positions(ctx context.Context, accountIDs []string, period string) (PositionTable, error) {
	d, err := s.data(ctx)
	if err != nil {
		return PositionTable{}, err
	}
	from, err := periodStart(period, d.today, ledger.Filter(d.portfolio.Events, accountIDs))
	if err != nil {
		return PositionTable{}, err
	}
	return positionTable(d.portfolio, d.market, s.base, from, d.today, accountIDs, period == "all")
}

// position — позиция за период в decimal, до перевода в строки.
type position struct {
	kind     ledger.InstrumentKind
	quantity decimal.Decimal
	value    decimal.Decimal // стоимость сегодня
	cost     decimal.Decimal // стоимость остатка сегодня
	income   decimal.Decimal // выплаты за период
	profit   decimal.Decimal // прибыль за период
}

// positionTable — позиции по счетам accountIDs: каждый счёт отдельно, строки с одним ключом складываются.
// За период — «на сегодня» минус «накануне from», как в summarize. Сверяется с summarize:
// сумма стоимостей — наша стоимость, сумма прибылей с уровнем счёта — наша прибыль за период.
// Необъяснённое расхождение входит в стоимость всегда, а в прибыль — только за всё время (allTime), как в overview.
func positionTable(p *load.Portfolio, m *load.Market, base string, from, today time.Time, accountIDs []string, allTime bool) (PositionTable, error) {
	byKey := make(map[string]*position)
	get := func(key string, kind ledger.InstrumentKind) *position {
		if byKey[key] == nil {
			byKey[key] = &position{kind: kind}
		}
		return byKey[key]
	}
	accountLevel := decimal.Zero
	before := from.AddDate(0, 0, -1)

	for _, id := range accountIDs {
		events := ledger.Filter(p.Events, []string{id})
		var earlier []ledger.Event
		for _, e := range events {
			if series.Day(e.Date).Before(from) {
				earlier = append(earlier, e)
			}
		}

		end, endValues, err := positionsAt(p, m, base, events, today, id)
		if err != nil {
			return PositionTable{}, err
		}
		start, startValues, err := positionsAt(p, m, base, earlier, before, id)
		if err != nil {
			return PositionTable{}, err
		}

		for key, f := range end {
			pos := get(key, f.Kind)
			pos.quantity = pos.quantity.Add(f.Quantity)
			pos.value = pos.value.Add(endValues[key])
			pos.cost = pos.cost.Add(f.Cost)

			was := start[key]
			if was == nil {
				was = &ledger.Flows{}
			}
			income := f.Income.Add(f.Taxes).Sub(was.Income).Sub(was.Taxes)
			pos.income = pos.income.Add(income)
			pos.profit = pos.profit.
				Add(f.Realized).Add(f.Fees).Sub(was.Realized).Sub(was.Fees).
				Add(endValues[key].Sub(f.Cost)).Sub(startValues[key].Sub(was.Cost)).
				Add(income).
				Add(f.Clearing).Sub(was.Clearing)
		}

		cash := get(base, ledger.InstrumentCurrency)
		cash.quantity = cash.quantity.Add(endValues[base])
		cash.value = cash.value.Add(endValues[base])
		cash.cost = cash.cost.Add(endValues[base])

		for _, e := range events {
			if series.Day(e.Date).Before(from) || e.InstrumentUID != "" || e.AssetUID != "" {
				continue
			}
			if e.OperationKind == ledger.KindFee || e.OperationKind == ledger.KindTax {
				accountLevel = accountLevel.Add(e.BaseAmount.Amount)
			}
		}
	}

	totalValue, totalProfit := decimal.Zero, accountLevel
	for _, pos := range byKey {
		totalValue = totalValue.Add(pos.value)
		totalProfit = totalProfit.Add(pos.profit)
	}
	end, change, err := summarize(p, m, base, from, today, accountIDs)
	if err != nil {
		return PositionTable{}, err
	}
	if !totalValue.Equal(end.Value) {
		return PositionTable{}, fmt.Errorf("positions: value %s, overview value %s", totalValue.StringFixed(2), end.Value.StringFixed(2))
	}
	if profit := change.Value.Sub(change.Invested); !totalProfit.Equal(profit) {
		return PositionTable{}, fmt.Errorf("positions: profit with account level %s, overview profit %s", totalProfit.StringFixed(2), profit.StringFixed(2))
	}

	gap, err := unexplainedCash(p, base, today, accountIDs)
	if err != nil {
		return PositionTable{}, err
	}
	result := PositionTable{AccountLevel: accountLevel.StringFixed(2), Unexplained: gap.StringFixed(2)}
	totalValue = totalValue.Add(gap)
	if allTime {
		result.UnexplainedProfit = gap.StringFixed(2)
		totalProfit = totalProfit.Add(gap)
	}
	result.TotalValue, result.TotalProfit = totalValue.StringFixed(2), totalProfit.StringFixed(2)

	groups := make(map[string]*position)
	names := instrumentNames(p)
	for key, pos := range byKey {
		row := Position{
			Key:       key,
			Ticker:    key,
			Kind:      string(pos.kind),
			Quantity:  pos.quantity.String(),
			Closed:    pos.quantity.IsZero(),
			Value:     pos.value.StringFixed(2),
			Invested:  pos.cost.StringFixed(2),
			Income:    pos.income.StringFixed(2),
			Profit:    pos.profit.StringFixed(2),
			ProfitPct: "",
		}
		if n, ok := names[key]; ok {
			row.Ticker, row.Name = n.ticker, n.name
		}
		if key == base {
			row.Ticker, row.Name = base, "Рубли"
		}
		if row.Kind == "" {
			row.Kind = string(ledger.InstrumentOther)
		}
		if !end.Value.IsZero() {
			row.Share = pos.value.Div(end.Value).Mul(decimal.NewFromInt(100)).StringFixed(2)
		}
		result.Positions = append(result.Positions, row)

		g := groups[row.Kind]
		if g == nil {
			g = &position{}
			groups[row.Kind] = g
		}
		g.value, g.profit = g.value.Add(pos.value), g.profit.Add(pos.profit)
	}
	for kind, g := range groups {
		result.Groups = append(result.Groups, GroupTotal{Kind: kind, Value: g.value.StringFixed(2), Profit: g.profit.StringFixed(2)})
	}
	return result, nil
}

// positionsAt — позиции счёта по событиям events и их стоимость на день day по ключам.
// Под ключом base — рубли. Если у позиции нет цены — ошибка, как в summarizeAt.
func positionsAt(p *load.Portfolio, m *load.Market, base string, events []ledger.Event, day time.Time, accountID string) (map[string]*ledger.Flows, map[string]decimal.Decimal, error) {
	positions := ledger.ByInstrument(events, p.Currencies.ByAsset)
	cash := ledger.CashHistory(events, base)
	rows, gaps := snapshot.Build(positions, m.Prices, m.Bonds, cash, base, day, day)
	if len(gaps) > 0 {
		return nil, nil, fmt.Errorf("account %s: value on %s: no price for %d positions, first %s", accountID, day.Format(time.DateOnly), len(gaps), gaps[0].Key)
	}
	values := make(map[string]decimal.Decimal)
	for _, r := range rows {
		values[r.Key] = values[r.Key].Add(r.Value)
	}
	return positions, values, nil
}

type instrumentName struct {
	ticker, name string
	date         time.Time
}

// instrumentNames — тикер и имя по ключу позиции из последней операции брокера с этим инструментом.
func instrumentNames(p *load.Portfolio) map[string]instrumentName {
	keys := ledger.PageKeys(p.Events, p.Currencies.ByAsset)
	result := make(map[string]instrumentName)
	for _, a := range p.Accounts {
		for _, op := range a.Operations {
			key, ok := keys[op.InstrumentUID]
			if !ok || op.Ticker == "" {
				continue
			}
			if was, seen := result[key]; seen && !op.Date.After(was.date) {
				continue
			}
			result[key] = instrumentName{ticker: op.Ticker, name: op.Name, date: op.Date}
		}
	}
	return result
}

// unexplainedCash — рубли у брокера сейчас минус наши рубли на сегодня, по счетам accountIDs.
// Сравниваются только деньги в базовой валюте: так видно операции, которых брокер не отдал в API.
// У закрытого счёта портфеля нет — у брокера на нём 0.
func unexplainedCash(p *load.Portfolio, base string, today time.Time, accountIDs []string) (decimal.Decimal, error) {
	selected := make(map[string]bool, len(accountIDs))
	for _, id := range accountIDs {
		selected[id] = true
	}

	total := decimal.Zero
	for _, a := range p.Accounts {
		if !selected[a.Account.ID] {
			continue
		}
		broker := decimal.Zero
		if a.Portfolio != nil {
			for _, pos := range a.Portfolio.Positions {
				if p.Currencies.ByUID[pos.InstrumentUID] != base {
					continue
				}
				quantity, err := convert.ToDecimal(pos.Quantity.Units, pos.Quantity.Nano)
				if err != nil {
					return decimal.Zero, fmt.Errorf("account %s: %s position: %w", a.Account.ID, base, err)
				}
				broker = broker.Add(quantity)
			}
		}
		ours := ledger.CashHistory(ledger.Filter(p.Events, []string{a.Account.ID}), base).At(series.EndOfDay(today))
		total = total.Add(broker.Sub(ours))
	}
	return total, nil
}
