package ledger

import (
	"time"

	"github.com/shopspring/decimal"
)

var moscow = time.FixedZone("MSK", 3*60*60)

type split struct {
	date   time.Time // дата сплита: сделки раньше — в старых единицах
	before int64     // было бумаг
	after  int64     // стало бумаг
}

var splitsByAssetUID = map[string][]split{
	// Росгосстрах (RGSS), 3333:50000, 27.04.2016.
	"9358770b-0e09-46fd-bde5-1b2f28b7759c": {
		{date: time.Date(2016, 4, 27, 0, 0, 0, 0, moscow), before: 3333, after: 50000},
	},
}

func adjustForSplits(positionKey string, date time.Time, quantity decimal.Decimal) decimal.Decimal {
	for _, table := range []map[string][]split{splitsByAssetUID, generatedSplitsByAssetUID} {
		for _, s := range table[positionKey] {
			if date.Before(s.date) {
				quantity = quantity.Mul(decimal.NewFromInt(s.after)).Div(decimal.NewFromInt(s.before))
			}
		}
	}
	return quantity
}
