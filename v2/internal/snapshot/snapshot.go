// Package snapshot — стоимость портфеля по дням: количество из книги × цена.
package snapshot

import (
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/shopspring/decimal"
)

// Row — одна позиция за один день.
type Row struct {
	Date     time.Time       // метка дня
	Key      string          // ключ позиции из книги
	Quantity decimal.Decimal // количество на конец дня
	Price    series.Point    // свеча, по которой взята цена
	Value    decimal.Decimal // количество × цена
}

// Gap — дни подряд, когда позиция была, а цены нет.
type Gap struct {
	Key  string
	From time.Time
	To   time.Time
}

// positionRows — строки одной позиции за каждый день с from по to и дни без цены.
func positionRows(key string, history ledger.History, prices series.Series, from, to time.Time) ([]Row, []Gap) {
	var rows []Row
	var gaps []Gap
	inGap := false

	for d := series.Day(from); !d.After(series.Day(to)); d = d.AddDate(0, 0, 1) {
		quantity := history.At(series.EndOfDay(d))
		if quantity.IsZero() {
			inGap = false
			continue
		}

		price, ok := prices.At(d)
		if !ok {
			if inGap {
				gaps[len(gaps)-1].To = d
			} else {
				gaps = append(gaps, Gap{Key: key, From: d, To: d})
				inGap = true
			}
			continue
		}

		inGap = false
		rows = append(rows, Row{
			Date:     d,
			Key:      key,
			Quantity: quantity,
			Price:    price,
			Value:    quantity.Mul(price.Value),
		})
	}
	return rows, gaps
}
