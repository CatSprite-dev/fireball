package metrics

import (
	"time"

	"github.com/shopspring/decimal"
)

var moscow = time.FixedZone("MSK", 3*60*60)

type split struct {
	date  time.Time       // первый день торгов после сплита: сделки раньше — в старых единицах
	ratio decimal.Decimal // сколько новых бумаг за одну старую
}

var splitsByAssetUID = map[string][]split{
	// Т-Технологии (T), 1:10, торги возобновились 17.04.2026.
	"4393bdb6-9017-466e-a3bb-ddb55d24159f": {
		{date: time.Date(2026, 4, 17, 0, 0, 0, 0, moscow), ratio: decimal.NewFromInt(10)},
	},
	// Банк ВТБ (VTBR), консолидация 5000:1, торги возобновились 15.07.2024.
	"94f18cc6-9e01-4db6-9d73-8c3ba6ab9655": {
		{date: time.Date(2024, 7, 15, 0, 0, 0, 0, moscow), ratio: decimal.RequireFromString("0.0002")},
	},
	// Норильский никель (GMKN), 1:100, торги возобновились 08.04.2024.
	"ba7b5a1b-8515-4134-8c26-b69a2372fe82": {
		{date: time.Date(2024, 4, 8, 0, 0, 0, 0, moscow), ratio: decimal.NewFromInt(100)},
	},
}

func adjustForSplits(positionKey string, date time.Time, quantity decimal.Decimal) decimal.Decimal {
	for _, s := range splitsByAssetUID[positionKey] {
		if date.Before(s.date) {
			quantity = quantity.Mul(s.ratio)
		}
	}
	return quantity
}
