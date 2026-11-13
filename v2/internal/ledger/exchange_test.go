package ledger

import (
	"testing"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/shopspring/decimal"
)

func TestAddCurrencyLegs(t *testing.T) {
	const usdAsset = "57880736"
	currencies := map[string]string{usdAsset: "USD"}

	events := []Event{
		{
			OperationID: "buy-usd", OperationKind: KindBuy,
			InstrumentKind: InstrumentCurrency, AssetUID: usdAsset,
			Quantity:   decimal.NewFromInt(11),
			Amount:     mny("-810.15", "RUB"),
			BaseAmount: mny("-810.15", "RUB"),
		},
		{
			OperationID: "fund", OperationKind: KindBuy,
			InstrumentKind: InstrumentETF,
			Quantity:       decimal.NewFromInt(50),
			Amount:         mny("-5.4", "USD"),
		},
		{
			OperationID: "sell-usd", OperationKind: KindSell,
			InstrumentKind: InstrumentCurrency, AssetUID: usdAsset,
			Quantity:   decimal.NewFromInt(74),
			Amount:     mny("6722.35", "RUB"),
			BaseAmount: mny("6722.35", "RUB"),
		},
		{
			OperationID: "fee", OperationKind: KindFee,
			InstrumentKind: InstrumentCurrency, AssetUID: usdAsset,
			Amount: mny("-60.5", "RUB"),
		},
	}

	got := AddCurrencyLegs(events, currencies)

	wantIDs := []string{"buy-usd", "buy-usd:leg", "fund", "sell-usd", "sell-usd:leg", "fee"}
	if len(got) != len(wantIDs) {
		t.Fatalf("got %d events, want %d", len(got), len(wantIDs))
	}
	for i, id := range wantIDs {
		if got[i].OperationID != id {
			t.Errorf("event %d = %s, want %s", i, got[i].OperationID, id)
		}
	}

	checkLeg := func(leg Event, parent, amount, base string) {
		t.Helper()
		if leg.OperationKind != KindExchange || leg.ParentID != parent {
			t.Errorf("%s: kind %s parent %s", leg.OperationID, leg.OperationKind, leg.ParentID)
		}
		if leg.Amount.Currency != "USD" || !leg.Amount.Amount.Equal(decimal.RequireFromString(amount)) {
			t.Errorf("%s: amount %s %s, want %s USD", leg.OperationID, leg.Amount.Amount, leg.Amount.Currency, amount)
		}
		if !leg.BaseAmount.Amount.Equal(decimal.RequireFromString(base)) {
			t.Errorf("%s: base %s, want %s", leg.OperationID, leg.BaseAmount.Amount, base)
		}
	}
	checkLeg(got[1], "buy-usd", "11", "810.15")
	checkLeg(got[4], "sell-usd", "-74", "-6722.35")
}

func TestAddCurrencyLegsUnknownAsset(t *testing.T) {
	events := []Event{{
		OperationID: "buy", OperationKind: KindBuy,
		InstrumentKind: InstrumentCurrency, AssetUID: "unknown",
		Quantity: decimal.NewFromInt(1), Amount: money.Money{Currency: "RUB"},
	}}

	if got := AddCurrencyLegs(events, map[string]string{}); len(got) != 1 {
		t.Errorf("got %d events, want 1 (no leg for unknown asset)", len(got))
	}
}
