package metrics

import (
	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/shopspring/decimal"
)

type Flows struct {
	Bought   decimal.Decimal
	Sold     decimal.Decimal
	Income   decimal.Decimal
	Fees     decimal.Decimal
	Taxes    decimal.Decimal
	Quantity decimal.Decimal
}

func ByInstrument(events []ledger.Event) map[string]*Flows {
	opsByInstrument := make(map[string]string, len(events))
	for _, e := range events {
		if e.InstrumentUID != "" {
			opsByInstrument[e.OperationID] = e.InstrumentUID
		}
	}

	result := make(map[string]*Flows)

	for _, e := range events {
		uid := e.InstrumentUID
		if uid == "" {
			uid = opsByInstrument[e.ParentID]
		}
		if uid == "" {
			continue
		}

		f, ok := result[uid]
		if !ok {
			f = &Flows{}
			result[uid] = f
		}

		amount := e.BaseAmount.Amount
		switch e.OperationKind {
		case ledger.KindBuy:
			f.Bought = f.Bought.Add(amount)
			f.Quantity = f.Quantity.Add(e.Quantity)
		case ledger.KindSell:
			f.Sold = f.Sold.Add(amount)
			f.Quantity = f.Quantity.Sub(e.Quantity)
		case ledger.KindDividend, ledger.KindCoupon, ledger.KindAmortization, ledger.KindRedemption:
			f.Income = f.Income.Add(amount)
		case ledger.KindFee:
			f.Fees = f.Fees.Add(amount)
		case ledger.KindTax:
			f.Taxes = f.Taxes.Add(amount)
		}
	}

	return result
}
