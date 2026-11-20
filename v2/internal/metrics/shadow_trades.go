package metrics

import (
	"fmt"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
)

// ShadowTrades — реальные бумаги и их тень в паях индекса.
// Тень повторяет каждое движение денег между бумагами и кошельком: покупка и комиссия —
// тень покупает индекс; продажа, погашение, купон и дивиденд — тень продаёт.
// real — стоимость бумаг; shadow — паи индекса × значение индекса.
// Паи тени могут уйти в минус, если продано больше, чем выросла тень.
// События дня from уже внутри стоимости на from и не учитываются.
func ShadowTrades(securities, index, trades, income series.Series, from, to time.Time) (actual, shadow series.Series, err error) {
	start, ok := index.At(from)
	if !ok {
		return nil, nil, fmt.Errorf("shadow trades: no index value on %s", from.Format(time.DateOnly))
	}

	units := valueAt(securities, from).Div(start.Value)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		i := valueAt(index, d)
		if d.After(from) {
			units = units.Sub(valueOn(trades, d).Add(valueOn(income, d)).Div(i))
		}
		actual = append(actual, series.Point{Date: d, Value: valueAt(securities, d)})
		shadow = append(shadow, series.Point{Date: d, Value: units.Mul(i)})
	}
	return actual, shadow, nil
}
