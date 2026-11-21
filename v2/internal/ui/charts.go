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

// Point — точка графика. Значение — строкой с двумя знаками: рубли или проценты.
type Point struct {
	Date  string `json:"date"` // "2006-01-02"
	Value string `json:"value"`
}

// Shadow — реальная линия и её тень в индексе.
type Shadow struct {
	Actual []Point `json:"actual"`
	Shadow []Point `json:"shadow"`
	Diff   string  `json:"diff"` // реальная минус тень на последнюю дату, ₽
}

// Charts — ряды для графиков «Обзора» по дням.
type Charts struct {
	TWR            []Point `json:"twr"`            // рост портфеля, % с CompareFrom из Performance; разница последних точек с Index — карточка «Против индекса»
	Index          []Point `json:"index"`          // рост индекса, % с CompareFrom; последняя точка — карточка «Индекс»
	Value          []Point `json:"value"`          // стоимость с поправкой на необъяснённое расхождение; последняя точка — карточка «Стоимость»
	ShadowTrades   Shadow  `json:"shadowTrades"`   // акции, облигации, фонды и их тень по сделкам, с CompareFrom
	ShadowDeposits Shadow  `json:"shadowDeposits"` // стоимость (как Value) и тень по пополнениям, с CompareFrom
}

// Charts — графики по счетам accountIDs за период period (как в Performance) против индекса index.
func (s *PortfolioService) Charts(ctx context.Context, accountIDs []string, period, index string) (Charts, error) {
	d, err := s.data(ctx)
	if err != nil {
		return Charts{}, err
	}
	values, ok := d.indexes[index]
	if !ok {
		return Charts{}, fmt.Errorf("index %s is not available from the broker", index)
	}
	return charts(d.portfolio, d.market, values, s.base, accountIDs, period, d.today)
}

// charts — ряды графиков. Всё, что про стоимость портфеля, считается от ряда с поправкой
// на необъяснённое расхождение (dailyValues); после расчёта ничего не прибавляется.
func charts(p *load.Portfolio, m *load.Market, index series.Series, base string, accountIDs []string, period string, today time.Time) (Charts, error) {
	events := ledger.Filter(p.Events, accountIDs)
	from, err := periodStart(period, today, events)
	if err != nil {
		return Charts{}, err
	}
	rows, values, err := dailyValues(p, m, base, accountIDs, from, today)
	if err != nil {
		return Charts{}, err
	}
	cmpFrom := compareStart(index, from)

	var securityRows []snapshot.Row
	for _, r := range rows {
		switch r.Kind {
		case ledger.InstrumentShare, ledger.InstrumentBond, ledger.InstrumentETF:
			securityRows = append(securityRows, r)
		}
	}
	trades, income := ledger.SecurityFlows(events)
	securities, shadowTrades, err := metrics.ShadowTrades(snapshot.Total(securityRows, from, today), index, trades, income, cmpFrom, today)
	if err != nil {
		return Charts{}, err
	}
	shadowDeposits, err := metrics.ShadowDeposits(values, index, ledger.NetDeposits(events), cmpFrom, today)
	if err != nil {
		return Charts{}, err
	}

	in, out := ledger.DepositsAndWithdrawals(events)
	return Charts{
		TWR:            percentPoints(metrics.TWR(values, in, out, cmpFrom, today)),
		Index:          percentPoints(metrics.TWR(index, nil, nil, cmpFrom, today)),
		Value:          moneyPoints(values, from),
		ShadowTrades:   toShadow(securities, shadowTrades, cmpFrom),
		ShadowDeposits: toShadow(values, shadowDeposits, cmpFrom),
	}, nil
}

// toShadow — линия и тень с дня from и разница между ними на последнюю дату.
// Разница — из значений, округлённых до копеек, как в последних точках: равна разнице их подписей ровно.
func toShadow(actual, shadow series.Series, from time.Time) Shadow {
	result := Shadow{Actual: moneyPoints(actual, from), Shadow: moneyPoints(shadow, from)}
	if len(actual) > 0 && len(shadow) > 0 {
		a, s := actual[len(actual)-1].Value.Round(2), shadow[len(shadow)-1].Value.Round(2)
		result.Diff = a.Sub(s).StringFixed(2)
	}
	return result
}

// percentPoints — ряд роста в процентах: 1.155 → "15.50".
func percentPoints(growth series.Series) []Point {
	result := make([]Point, 0, len(growth))
	for _, pt := range growth {
		result = append(result, Point{
			Date:  pt.Date.Format(time.DateOnly),
			Value: pt.Value.Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100)).StringFixed(2),
		})
	}
	return result
}

// moneyPoints — точки ряда с дня from.
func moneyPoints(s series.Series, from time.Time) []Point {
	result := make([]Point, 0, len(s))
	for _, pt := range s {
		if pt.Date.Before(from) {
			continue
		}
		result = append(result, Point{Date: pt.Date.Format(time.DateOnly), Value: pt.Value.StringFixed(2)})
	}
	return result
}
