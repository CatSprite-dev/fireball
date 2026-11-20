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
	Date     time.Time             // метка дня
	Key      string                // ключ позиции из книги
	Kind     ledger.InstrumentKind // тип инструмента
	Quantity decimal.Decimal       // количество на конец дня
	Price    series.Point          // цена одной бумаги в базовой валюте и день, от которого она
	NKD      decimal.Decimal       // НКД одной бумаги; не у облигаций — 0
	Value    decimal.Decimal       // количество × (цена + НКД)
}

// Gap — дни подряд, когда позиция была, а цены нет.
type Gap struct {
	Key  string
	From time.Time
	To   time.Time
}

// Bond — номинал и НКД облигации по дням, по возрастанию даты.
type Bond struct {
	Nominal series.Series
	NKD     series.Series
}

// positionRows — строки одной позиции за каждый день с from по to и дни без цены.
func positionRows(key string, kind ledger.InstrumentKind, history ledger.History, prices series.Series, bond Bond, from, to time.Time) ([]Row, []Gap) {
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
			rows = append(rows, Row{Date: d, Key: key, Kind: kind, Quantity: quantity})
			continue
		}

		price, ok := prices.At(d)
		nkd := decimal.Zero
		if ok && kind == ledger.InstrumentBond {
			price, nkd, ok = bondPrice(price, bond, d)
		}
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
			Kind:     kind,
			Quantity: quantity,
			Price:    price,
			NKD:      nkd,
			Value:    quantity.Mul(price.Value.Add(nkd)),
		})
	}
	return rows, gaps
}

// cashRows — деньги в базовой валюте base на конец каждого дня; цена денег в базе всегда 1.
func cashRows(base string, cash ledger.History, from, to time.Time) []Row {
	var rows []Row
	for d := series.Day(from); !d.After(series.Day(to)); d = d.AddDate(0, 0, 1) {
		amount := cash.At(series.EndOfDay(d))
		if amount.IsZero() {
			continue
		}
		rows = append(rows, Row{
			Date:     d,
			Key:      base,
			Kind:     ledger.InstrumentCurrency,
			Quantity: amount,
			Price:    series.Point{Date: d, Value: decimal.NewFromInt(1)},
			Value:    amount,
		})
	}
	return rows
}

// bondPrice — цена облигации в рублях из цены в процентах от номинала и НКД на день d.
// false — номинала на этот день нет, оценить нельзя.
func bondPrice(percent series.Point, bond Bond, d time.Time) (series.Point, decimal.Decimal, bool) {
	nominal, ok := bond.Nominal.At(d)
	if !ok {
		return series.Point{}, decimal.Zero, false
	}
	price := percent.Value.Mul(nominal.Value).Div(decimal.NewFromInt(100))

	nkd := decimal.Zero
	if a, ok := bond.NKD.At(d); ok {
		nkd = a.Value
	}
	return series.Point{Date: percent.Date, Value: price}, nkd, true
}

// Build — строки всех позиций и денег за каждый день с from по to и дни без цены.
// prices — ряд цен на каждый ключ позиции, по возрастанию даты; у облигаций — в процентах от номинала.
// bonds — номинал и НКД облигаций по ключу позиции.
// cash — деньги в базовой валюте base.
func Build(positions map[string]*ledger.Flows, prices map[string]series.Series, bonds map[string]Bond, cash ledger.History, base string, from, to time.Time) ([]Row, []Gap) {
	keys := make([]string, 0, len(positions))
	for key := range positions {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	var rows []Row
	var gaps []Gap
	for _, key := range keys {
		p := positions[key]
		r, g := positionRows(key, p.Kind, p.History, prices[key], bonds[key], from, to)
		rows = append(rows, r...)
		gaps = append(gaps, g...)
	}
	rows = append(rows, cashRows(base, cash, from, to)...)

	return rows, gaps
}
