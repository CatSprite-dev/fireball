package convert

import (
	"fmt"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

// ToAccrued раскладывает НКД облигации по дням на два ряда: номинал и НКД одной бумаги.
func ToAccrued(list []tinvest.AccruedInterest) (nominal, nkd series.Series, err error) {
	nominal = make(series.Series, 0, len(list))
	nkd = make(series.Series, 0, len(list))

	for _, a := range list {
		day := series.Day(a.Date)

		n, err := ToDecimal(a.Nominal.Units, a.Nominal.Nano)
		if err != nil {
			return nil, nil, fmt.Errorf("accrued %s: nominal: %w", day.Format(time.DateOnly), err)
		}
		v, err := ToDecimal(a.Value.Units, a.Value.Nano)
		if err != nil {
			return nil, nil, fmt.Errorf("accrued %s: value: %w", day.Format(time.DateOnly), err)
		}

		nominal = append(nominal, series.Point{Date: day, Value: n})
		nkd = append(nkd, series.Point{Date: day, Value: v})
	}
	return nominal, nkd, nil
}
