// Package ledger — журнал событий и книга позиций
//
// Позиция — всё, что было у пользователя по одному активу. Ключ позиции:
//   - валюта — код валюты (USD), потому что одна валюта торгуется под разными uid;
//   - фьючерс и опцион — instrumentUid: у каждого контракта своя позиция;
//   - акции, облигации, фонды — assetUid: при переводе бумаги брокер меняет
//     instrumentUid, а assetUid остаётся прежним;
//   - если assetUid пустой — instrumentUid.
//
// Портфель брокера знает только instrumentUid, поэтому PageKeys даёт карту
// instrumentUid → ключ позиции, чтобы стоимость легла туда же, куда сделки.
package ledger

import (
	"github.com/shopspring/decimal"
)

type Flows struct {
	Bought   decimal.Decimal
	Sold     decimal.Decimal
	Income   decimal.Decimal
	Fees     decimal.Decimal
	Taxes    decimal.Decimal
	Quantity decimal.Decimal
	Clearing decimal.Decimal
}

func ByInstrument(events []Event, currencyByAsset map[string]string) map[string]*Flows {
	result := make(map[string]*Flows)

	for _, e := range events {
		addCurrencyMirror(result, e)
		uid := pageKey(e, currencyByAsset)

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
		case KindBuy:
			f.Bought = f.Bought.Add(amount)
			f.Quantity = f.Quantity.Add(adjustForSplits(uid, e.Date, e.Quantity))
		case KindSell:
			f.Sold = f.Sold.Add(amount)
			f.Quantity = f.Quantity.Sub(adjustForSplits(uid, e.Date, e.Quantity))
		case KindDividend, KindCoupon, KindAmortization:
			f.Income = f.Income.Add(amount)
		case KindRedemption:
			f.Income = f.Income.Add(amount)
			f.Quantity = decimal.Zero
		case KindFee:
			f.Fees = f.Fees.Add(amount)
		case KindTax:
			f.Taxes = f.Taxes.Add(amount)
		case KindVariationMargin:
			f.Clearing = f.Clearing.Add(amount)
		case KindExpiration:
			f.Clearing = f.Clearing.Add(amount)
			f.Quantity = decimal.Zero
		}
	}

	return result
}

// pageKey возвращает ключ позиции, на которую записывается событие
func pageKey(e Event, currencyByAsset map[string]string) string {
	switch e.InstrumentKind {
	case InstrumentCurrency:
		if code, ok := currencyByAsset[e.AssetUID]; ok {
			return code
		}
		return e.InstrumentUID
	case InstrumentFutures, InstrumentOption:
		return e.InstrumentUID
	}
	if e.AssetUID != "" {
		if current, ok := currentAssetUIDByOldAssetUID[e.AssetUID]; ok {
			return current
		}
		return e.AssetUID
	}
	return e.InstrumentUID
}

func addCurrencyMirror(pages map[string]*Flows, e Event) {
	if e.OperationKind == KindExchange || e.Amount.Currency == e.BaseAmount.Currency || e.Amount.Amount.IsZero() {
		return
	}

	f, ok := pages[e.Amount.Currency]
	if !ok {
		f = &Flows{}
		pages[e.Amount.Currency] = f
	}

	mirror := e.BaseAmount.Amount.Neg()
	if mirror.IsNegative() {
		f.Bought = f.Bought.Add(mirror)
	} else {
		f.Sold = f.Sold.Add(mirror)
	}
	f.Quantity = f.Quantity.Add(e.Amount.Amount)
}

func PageKeys(events []Event, currencyCodeByAssetUID map[string]string) (pageKeyByInstrumentUID map[string]string) {
	pageKeyByInstrumentUID = make(map[string]string)
	for _, e := range events {
		if e.InstrumentUID != "" {
			pageKeyByInstrumentUID[e.InstrumentUID] = pageKey(e, currencyCodeByAssetUID)
		}
	}
	return pageKeyByInstrumentUID
}
