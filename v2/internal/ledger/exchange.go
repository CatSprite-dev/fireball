package ledger

import "github.com/CatSprite-dev/fireball/v2/internal/money"

func AddCurrencyLegs(events []Event, currencyByAsset map[string]string) []Event {
	result := make([]Event, 0, len(events))
	for _, e := range events {
		result = append(result, e)

		if e.InstrumentKind != InstrumentCurrency || (e.OperationKind != KindBuy && e.OperationKind != KindSell) {
			continue
		}
		currency, ok := currencyByAsset[e.AssetUID]
		if !ok {
			continue
		}

		quantity := e.Quantity
		if e.OperationKind == KindSell {
			quantity = quantity.Neg()
		}

		result = append(result, Event{
			AccountID:      e.AccountID,
			OperationID:    e.OperationID + ":leg",
			ParentID:       e.OperationID,
			Date:           e.Date,
			OperationKind:  KindExchange,
			InstrumentUID:  e.InstrumentUID,
			InstrumentKind: e.InstrumentKind,
			AssetUID:       e.AssetUID,
			Amount:         money.Money{Amount: quantity, Currency: currency},
			BaseAmount:     money.Money{Amount: e.BaseAmount.Amount.Neg(), Currency: e.BaseAmount.Currency},
		})
	}
	return result
}
