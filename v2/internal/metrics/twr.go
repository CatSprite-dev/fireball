// Package metrics — показатели портфеля, посчитанные по рядам стоимости и пополнений.
package metrics

import (
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/shopspring/decimal"
)

// TWR — накопленный рост портфеля, взвешенный по времени: на from — 1, дальше по дням до to включительно.
// values — стоимость на конец дня, deposits — чистые пополнения по дням (ledger.NetDeposits).
// Пополнение дня считается пришедшим утром: доходность дня d = V(d) / (V(d-1) + F(d)) − 1.
// Пустой портфель (V(d-1) + F(d) = 0) рост не меняет.
func TWR(values, deposits series.Series, from, to time.Time) series.Series {
	growth := decimal.NewFromInt(1)
	result := series.Series{{Date: from, Value: growth}}
	for d := from.AddDate(0, 0, 1); !d.After(to); d = d.AddDate(0, 0, 1) {
		base := valueAt(values, d.AddDate(0, 0, -1)).Add(depositOn(deposits, d))
		if !base.IsZero() {
			growth = growth.Mul(valueAt(values, d).Div(base)).Round(16)
		}
		result = append(result, series.Point{Date: d, Value: growth})
	}
	return result
}

// valueAt — стоимость на конец дня d; до первой точки — 0.
func valueAt(values series.Series, d time.Time) decimal.Decimal {
	p, ok := values.At(d)
	if !ok {
		return decimal.Zero
	}
	return p.Value
}

// depositOn — пополнение ровно в день d; 0, если в этот день его не было.
func depositOn(deposits series.Series, d time.Time) decimal.Decimal {
	p, ok := deposits.At(d)
	if !ok || !p.Date.Equal(d) {
		return decimal.Zero
	}
	return p.Value
}
