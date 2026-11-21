package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/load"
	"github.com/CatSprite-dev/fireball/v2/internal/metrics"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/snapshot"
)

// Performance — доходность выбранных счетов за период и сравнение с индексом.
// Проценты — строкой с двумя знаками ("-10.53").
type Performance struct {
	From        string `json:"from"`        // начало периода, "2006-01-02"
	TWR         string `json:"twr"`         // доходность, взвешенная по времени, % за период
	MWR         string `json:"mwr"`         // доходность, взвешенная по деньгам, % годовых; пусто, если не посчитать
	Index       string `json:"index"`       // рост индекса, % с CompareFrom (или From) по сегодня
	VsIndex     string `json:"vsIndex"`     // TWR минус рост индекса за тот же отрезок, п.п.
	CompareFrom string `json:"compareFrom"` // день, с которого идёт сравнение, если индекс начинается позже From; иначе пусто
}

// Performance — TWR, MWR и сравнение с индексом по счетам accountIDs.
// period: "1m", "3m", "6m", "12m", "5y" или "all" — с первой операции выбранных счетов.
func (s *PortfolioService) Performance(ctx context.Context, accountIDs []string, period, index string) (Performance, error) {
	d, err := s.data(ctx)
	if err != nil {
		return Performance{}, err
	}
	values, ok := d.indexes[index]
	if !ok {
		return Performance{}, fmt.Errorf("index %s is not available from the broker", index)
	}
	return performance(d.portfolio, d.market, values, s.base, accountIDs, period, d.today)
}

func performance(p *load.Portfolio, m *load.Market, index series.Series, base string, accountIDs []string, period string, today time.Time) (Performance, error) {
	events := ledger.Filter(p.Events, accountIDs)
	from, err := periodStart(period, today, events)
	if err != nil {
		return Performance{}, err
	}

	_, values, err := dailyValues(p, m, base, accountIDs, from, today)
	if err != nil {
		return Performance{}, err
	}
	in, out := ledger.DepositsAndWithdrawals(events)

	result := Performance{
		From: from.Format(time.DateOnly),
		TWR:  growthPercent(metrics.TWR(values, in, out, from, today)).StringFixed(2),
	}
	if r, err := metrics.MWR(values, ledger.NetDeposits(events), from, today); err == nil {
		result.MWR = decimal.NewFromFloat(r * 100).StringFixed(2)
	}

	cmpFrom := compareStart(index, from)
	if cmpFrom.After(from) {
		result.CompareFrom = cmpFrom.Format(time.DateOnly)
	}
	indexGrowth := growthPercent(metrics.TWR(index, nil, nil, cmpFrom, today))
	twrCmp := growthPercent(metrics.TWR(values, in, out, cmpFrom, today))
	result.Index = indexGrowth.StringFixed(2)
	result.VsIndex = twrCmp.Sub(indexGrowth).StringFixed(2)
	return result, nil
}

// dailyValues — строки позиций и стоимость счетов accountIDs на каждый день с from по today.
// К стоимости прибавлено необъяснённое расхождение каждого счёта с дня его последней операции:
// раньше брокер не мог списать то, чего мы не видим. От этого ряда считаются TWR, MWR и тени,
// поэтому несуществующие деньги не вкладываются в индекс и не дают доходности.
// Строки — без поправки: она про деньги счёта, а не про позиции.
// Если у позиции нет цены — ошибка, как в summarizeAt.
func dailyValues(p *load.Portfolio, m *load.Market, base string, accountIDs []string, from, today time.Time) ([]snapshot.Row, series.Series, error) {
	events := ledger.Filter(p.Events, accountIDs)
	positions := ledger.ByInstrument(events, p.Currencies.ByAsset)
	cash := ledger.CashHistory(events, base)
	rows, gaps := snapshot.Build(positions, m.Prices, m.Bonds, cash, base, from, today)
	if len(gaps) > 0 {
		return nil, nil, fmt.Errorf("values from %s: no price for %d positions, first %s on %s",
			from.Format(time.DateOnly), len(gaps), gaps[0].Key, gaps[0].From.Format(time.DateOnly))
	}
	fixes, err := corrections(p, base, today, accountIDs)
	if err != nil {
		return nil, nil, err
	}

	values := snapshot.Total(rows, from, today)
	for i, pt := range values {
		for _, f := range fixes {
			if !pt.Date.Before(f.since) {
				values[i].Value = values[i].Value.Add(f.amount)
			}
		}
	}
	return rows, values, nil
}

// correction — необъяснённое расхождение одного счёта и день, с которого оно входит в стоимость.
type correction struct {
	since  time.Time
	amount decimal.Decimal
}

// corrections — поправки счетов accountIDs: расхождение счёта с дня его последней операции
// (у закрытого счёта — день закрытия). Счета без расхождения не попадают.
func corrections(p *load.Portfolio, base string, today time.Time, accountIDs []string) ([]correction, error) {
	var result []correction
	for _, id := range accountIDs {
		gap, err := unexplainedCash(p, base, today, []string{id})
		if err != nil {
			return nil, err
		}
		if gap.IsZero() {
			continue
		}
		since := time.Time{}
		for _, e := range ledger.Filter(p.Events, []string{id}) {
			if day := series.Day(e.Date); day.After(since) {
				since = day
			}
		}
		result = append(result, correction{since: since, amount: gap})
	}
	return result, nil
}

// compareStart — день, с которого портфель сравнивается с индексом: начало периода,
// а если индекс начинается позже — его первый день.
func compareStart(index series.Series, from time.Time) time.Time {
	if len(index) > 0 && index[0].Date.After(from) {
		return index[0].Date
	}
	return from
}

// periodStart — первый день периода. Для "all" — день первой операции; нет операций — сегодня.
func periodStart(period string, today time.Time, events []ledger.Event) (time.Time, error) {
	switch period {
	case "1m":
		return today.AddDate(0, -1, 0), nil
	case "3m":
		return today.AddDate(0, -3, 0), nil
	case "6m":
		return today.AddDate(0, -6, 0), nil
	case "12m":
		return today.AddDate(-1, 0, 0), nil
	case "5y":
		return today.AddDate(-5, 0, 0), nil
	case "all":
		first := today
		for _, e := range events {
			if day := series.Day(e.Date); day.Before(first) {
				first = day
			}
		}
		return first, nil
	}
	return time.Time{}, fmt.Errorf("unknown period %q", period)
}

// growthPercent — последняя точка ряда роста в процентах: 1.155 → 15.5.
func growthPercent(growth series.Series) decimal.Decimal {
	return growth[len(growth)-1].Value.Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100))
}
