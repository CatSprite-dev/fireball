package metrics

import (
	"fmt"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
)

// ShadowDeposits — тень портфеля по пополнениям: те же деньги в те же дни, вложенные в индекс.
// Возвращает паи индекса × значение индекса на каждый день с from по to.
// Выводы уменьшают паи, и те могут уйти в минус.
// Пополнения дня from уже внутри стоимости на from и не учитываются.
func ShadowDeposits(values, index, deposits series.Series, from, to time.Time) (series.Series, error) {
	start, ok := index.At(from)
	if !ok {
		return nil, fmt.Errorf("shadow deposits: no index value on %s", from.Format(time.DateOnly))
	}

	units := valueAt(values, from).Div(start.Value)
	var shadow series.Series
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		i := valueAt(index, d)
		if d.After(from) {
			units = units.Add(valueOn(deposits, d).Div(i))
		}
		shadow = append(shadow, series.Point{Date: d, Value: units.Mul(i)})
	}
	return shadow, nil
}
