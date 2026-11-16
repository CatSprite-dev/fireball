package ledger

import (
	"github.com/CatSprite-dev/fireball/v2/internal/series"
)

// TradePrices — цены своих сделок по позициям: сумма сделки / количество, в базовой валюте.
// Нужны там, где у бумаги нет свечей: тогда последняя известная цена — цена своей сделки.
// Облигации, валюта и фьючерсы не входят: у облигации сумма сделки включает НКД,
// валюта оценивается по курсу, а у фьючерса сумма сделки обнулена.
func TradePrices(events []Event, currencyByAsset map[string]string) map[string]series.Series {
	result := make(map[string]series.Series)
	for _, e := range byDate(events) {
		if e.OperationKind != KindBuy && e.OperationKind != KindSell {
			continue
		}
		switch e.InstrumentKind {
		case InstrumentBond, InstrumentCurrency, InstrumentFutures:
			continue
		}

		key := pageKey(e, currencyByAsset)
		quantity := adjustForSplits(key, e.Date, e.Quantity)
		if key == "" || quantity.IsZero() {
			continue
		}

		p := series.Point{Date: series.Day(e.Date), Value: e.BaseAmount.Amount.Abs().Div(quantity)}
		s := result[key]
		if n := len(s); n > 0 && s[n-1].Date.Equal(p.Date) {
			s[n-1] = p
		} else {
			s = append(s, p)
		}
		result[key] = s
	}
	return result
}
