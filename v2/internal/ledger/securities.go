package ledger

import (
	"github.com/CatSprite-dev/fireball/v2/internal/series"
)

// SecurityFlows — денежные потоки по акциям, облигациям и фондам по московским дням в базовой валюте.
// trades — покупки, продажи, погашения и амортизация; income — дивиденды, купоны, комиссии и налоги.
// Дни, где движения взаимно погасились, в ряды не входят.
func SecurityFlows(events []Event) (trades, income series.Series) {
	for _, e := range byDate(events) {
		switch e.InstrumentKind {
		case InstrumentShare, InstrumentBond, InstrumentETF:
		default:
			continue
		}

		switch e.OperationKind {
		case KindBuy, KindSell, KindRedemption, KindAmortization:
			trades = addToDay(trades, e)
		case KindDividend, KindCoupon, KindFee, KindTax:
			income = addToDay(income, e)
		}
	}
	return dropZeroDays(trades), dropZeroDays(income)
}
