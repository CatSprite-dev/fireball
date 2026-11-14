package ledger

import (
	"time"

	"github.com/shopspring/decimal"
)

var moscow = time.FixedZone("MSK", 3*60*60)

type split struct {
	date  time.Time       // дата сплита: сделки раньше — в старых единицах
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
	// Полюс (PLZL), 1:10, 27.03.2025.
	"6d9a99d0-73e5-4971-84d6-b90729f2400d": {
		{date: time.Date(2025, 3, 27, 0, 0, 0, 0, moscow), ratio: decimal.NewFromInt(10)},
	},
	// Транснефть, прив. (TRNFP), 1:100, 19.02.2024.
	"e1448f66-ee8f-47dd-8421-af2009f6236c": {
		{date: time.Date(2024, 2, 19, 0, 0, 0, 0, moscow), ratio: decimal.NewFromInt(100)},
	},
	// Русолово (ROLO), 1:10, 17.01.2023.
	"88d6ef5c-f342-4ac9-acbb-bf7167ba01fe": {
		{date: time.Date(2023, 1, 17, 0, 0, 0, 0, moscow), ratio: decimal.NewFromInt(10)},
	},
	// Росгосстрах (RGSS), 1:15,00150015, 27.04.2016.
	"9358770b-0e09-46fd-bde5-1b2f28b7759c": {
		{date: time.Date(2016, 4, 27, 0, 0, 0, 0, moscow), ratio: decimal.RequireFromString("15.00150015")},
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
