// Package snapshot — стоимость портфеля по дням: количество из книги × цена.
package snapshot

import (
	"slices"
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
func positionRows(key string, kind ledger.InstrumentKind, history ledger.History, prices series.Series, from, to time.Time) ([]Row, []Gap) {
	var rows []Row
	var gaps []Gap
	inGap := false

	for d := series.Day(from); !d.After(series.Day(to)); d = d.AddDate(0, 0, 1) {
		quantity := history.At(series.EndOfDay(d))
		if quantity.IsZero() {
			inGap = false
			continue
		}

		if kind == ledger.InstrumentFutures {
			rows = append(rows, Row{Date: d, Key: key, Quantity: quantity})
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

// Build — строки всех позиций за каждый день с from по to и дни без цены.
// prices — ряд цен на каждый ключ позиции.
func Build(positions map[string]*ledger.Flows, prices map[string]series.Series, from, to time.Time) ([]Row, []Gap) {
	keys := make([]string, 0, len(positions))
	for key := range positions {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	var rows []Row
	var gaps []Gap
	for _, key := range keys {
		r, g := positionRows(key, positions[key].Kind, positions[key].History, prices[key], from, to)
		rows = append(rows, r...)
		gaps = append(gaps, g...)
	}
	return rows, gaps
}
