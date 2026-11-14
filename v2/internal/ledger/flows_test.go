package ledger

import (
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/shopspring/decimal"
)

func ev(id, parent string, kind OperationKind, uid, amount, qty string) Event {
	e := Event{
		OperationID:   id,
		ParentID:      parent,
		OperationKind: kind,
		InstrumentUID: uid,
		BaseAmount:    money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"},
	}
	if qty != "" {
		e.Quantity = decimal.RequireFromString(qty)
	}
	return e
}

func TestByInstrument(t *testing.T) {
	events := []Event{
		ev("dep", "", KindDeposit, "", "200000", ""),
		ev("buy1", "", KindBuy, "etln", "-100510", "100"),
		ev("fee1", "buy1", KindFee, "etln", "-50.26", ""),
		ev("cpn", "", KindCoupon, "etln", "2000", ""),
		ev("tax", "", KindTax, "etln", "-260", ""),
		ev("sell1", "", KindSell, "etln", "100789", "100"),
		ev("fee2", "sell1", KindFee, "etln", "-301.53", ""),
		ev("svc", "", KindFee, "", "-99", ""),
	}

	got := ByInstrument(events, nil)

	if len(got) != 1 {
		t.Fatalf("got %d instruments, want 1 (etln): %v", len(got), got)
	}
	f := got["etln"]

	checks := []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		{"bought", f.Bought, "-100510"},
		{"sold", f.Sold, "100789"},
		{"income", f.Income, "2000"},
		{"fees", f.Fees, "-351.79"},
		{"taxes", f.Taxes, "-260"},
		{"quantity", f.Quantity, "0"},
	}
	for _, c := range checks {
		if !c.got.Equal(decimal.RequireFromString(c.want)) {
			t.Errorf("%s = %s, want %s", c.name, c.got, c.want)
		}
	}
}

func TestByInstrumentCurrencyPageByCode(t *testing.T) {
	currencyByAsset := map[string]string{"usd-asset": "USD"}

	oldDollar := ev("buy-old", "", KindBuy, "usd-cngd", "-810.15", "11")
	oldDollar.InstrumentKind = InstrumentCurrency
	oldDollar.AssetUID = "usd-asset"

	newDollar := ev("buy-new", "", KindBuy, "usd-tom", "-6722.35", "74")
	newDollar.InstrumentKind = InstrumentCurrency
	newDollar.AssetUID = "usd-asset"

	share := ev("buy-share", "", KindBuy, "etln", "-100510", "100")
	share.InstrumentKind = InstrumentShare

	got := ByInstrument([]Event{oldDollar, newDollar, share}, currencyByAsset)

	if _, ok := got["USD"]; !ok {
		t.Fatalf("no USD page: %v", got)
	}
	if !got["USD"].Quantity.Equal(decimal.RequireFromString("85")) {
		t.Errorf("USD quantity = %s, want 85", got["USD"].Quantity)
	}
	if _, ok := got["etln"]; !ok {
		t.Errorf("share page must stay under its uid")
	}
	if len(got) != 2 {
		t.Errorf("got %d pages, want 2 (USD, etln): %v", len(got), got)
	}
}

func TestByInstrumentCurrencyMirror(t *testing.T) {
	buyUSD := ev("buy-usd", "", KindBuy, "usd-tom", "-810.15", "11")
	buyUSD.InstrumentKind = InstrumentCurrency
	buyUSD.AssetUID = "usd-asset"

	leg := ev("buy-usd:leg", "buy-usd", KindExchange, "usd-tom", "810.15", "")
	leg.InstrumentKind = InstrumentCurrency
	leg.AssetUID = "usd-asset"
	leg.Amount = money.Money{Amount: decimal.RequireFromString("11"), Currency: "USD"}

	buyShare := ev("buy-aapl", "", KindBuy, "aapl", "-740", "1")
	buyShare.Amount = money.Money{Amount: decimal.RequireFromString("-10"), Currency: "USD"}

	dividend := ev("div-aapl", "", KindDividend, "aapl", "37.5", "")
	dividend.Amount = money.Money{Amount: decimal.RequireFromString("0.5"), Currency: "USD"}

	got := ByInstrument([]Event{buyUSD, leg, buyShare, dividend}, map[string]string{"usd-asset": "USD"})

	usd := got["USD"]
	checks := []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		{"usd bought", usd.Bought, "-847.65"},
		{"usd sold", usd.Sold, "740"},
		{"usd quantity", usd.Quantity, "1.5"},
		{"aapl bought", got["aapl"].Bought, "-740"},
		{"aapl income", got["aapl"].Income, "37.5"},
	}
	for _, c := range checks {
		if !c.got.Equal(decimal.RequireFromString(c.want)) {
			t.Errorf("%s = %s, want %s", c.name, c.got, c.want)
		}
	}
}

func TestByInstrumentClearing(t *testing.T) {
	events := []Event{
		ev("vm:0", "vm", KindVariationMargin, "fut", "2885", ""),
		ev("vm2:0", "vm2", KindVariationMargin, "fut", "-1687", ""),
		ev("exp", "", KindExpiration, "opt", "264", ""),
	}

	got := ByInstrument(events, nil)

	if !got["fut"].Clearing.Equal(decimal.RequireFromString("1198")) {
		t.Errorf("fut clearing = %s, want 1198", got["fut"].Clearing)
	}
	if !got["opt"].Clearing.Equal(decimal.RequireFromString("264")) {
		t.Errorf("opt clearing = %s, want 264", got["opt"].Clearing)
	}
}

func TestByInstrumentRedemptionAndExpirationCloseQuantity(t *testing.T) {
	events := []Event{
		ev("buy-bond", "", KindBuy, "bond", "-1028.33", "1"),
		ev("redeem", "", KindRedemption, "bond", "1035.10", ""),

		ev("buy-opt", "", KindBuy, "opt", "-305.90", "322"),
		ev("exp", "", KindExpiration, "opt", "0", ""),

		ev("buy-amort", "", KindBuy, "amort", "-10000", "10"),
		ev("amort", "", KindAmortization, "amort", "2500", ""),
	}

	got := ByInstrument(events, nil)

	checks := []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		{"bond quantity", got["bond"].Quantity, "0"},
		{"bond income", got["bond"].Income, "1035.10"},
		{"option quantity", got["opt"].Quantity, "0"},
		{"amortization keeps quantity", got["amort"].Quantity, "10"},
	}
	for _, c := range checks {
		if !c.got.Equal(decimal.RequireFromString(c.want)) {
			t.Errorf("%s = %s, want %s", c.name, c.got, c.want)
		}
	}
}

func TestByInstrumentMergesByAsset(t *testing.T) {
	oldUID := ev("buy", "", KindBuy, "tspx-tqtd", "-6174.67", "700")
	oldUID.InstrumentKind = InstrumentETF
	oldUID.AssetUID = "tspx-asset"

	newUID := ev("sell", "", KindSell, "tspx-tqtf", "5497.80", "700")
	newUID.InstrumentKind = InstrumentETF
	newUID.AssetUID = "tspx-asset"

	option := ev("opt", "", KindBuy, "gz130", "-500.80", "80")
	option.InstrumentKind = InstrumentOption
	option.AssetUID = "gazp-asset"

	share := ev("gazp", "", KindBuy, "gazp-uid", "-25821.30", "100")
	share.InstrumentKind = InstrumentShare
	share.AssetUID = "gazp-asset"

	got := ByInstrument([]Event{oldUID, newUID, option, share}, nil)

	if !got["tspx-asset"].Quantity.IsZero() {
		t.Errorf("tspx quantity = %s, want 0", got["tspx-asset"].Quantity)
	}
	if _, ok := got["gz130"]; !ok {
		t.Errorf("option must keep its own page")
	}
	if len(got) != 3 {
		t.Errorf("got %d pages, want 3 (tspx-asset, gz130, gazp-asset): %v", len(got), got)
	}
}

func TestPageKeys(t *testing.T) {
	oldTSPX := ev("buy", "", KindBuy, "tspx-tqtd", "-6174.67", "700")
	oldTSPX.InstrumentKind = InstrumentETF
	oldTSPX.AssetUID = "tspx-asset"

	newTSPX := ev("sell", "", KindSell, "tspx-tqtf", "5497.80", "700")
	newTSPX.InstrumentKind = InstrumentETF
	newTSPX.AssetUID = "tspx-asset"

	usd := ev("buy-usd", "", KindBuy, "usd-tom", "-810.15", "11")
	usd.InstrumentKind = InstrumentCurrency
	usd.AssetUID = "usd-asset"

	option := ev("opt", "", KindBuy, "gz130", "-500.80", "80")
	option.InstrumentKind = InstrumentOption
	option.AssetUID = "gazp-asset"

	got := PageKeys([]Event{oldTSPX, newTSPX, usd, option}, map[string]string{"usd-asset": "USD"})

	want := map[string]string{
		"tspx-tqtd": "tspx-asset",
		"tspx-tqtf": "tspx-asset",
		"usd-tom":   "USD",
		"gz130":     "gz130",
	}
	for uid, key := range want {
		if got[uid] != key {
			t.Errorf("%s → %q, want %q", uid, got[uid], key)
		}
	}
}

func TestByInstrumentMergesKnownConversion(t *testing.T) {
	gdr := ev("buy-tcsg", "", KindBuy, "a640e094", "-8353.10", "1")
	gdr.InstrumentKind = InstrumentShare
	gdr.AssetUID = "4ac39e69-fff7-406c-8324-a869d97ec022"

	share := ev("buy-t", "", KindBuy, "87db07bc", "-71683.66", "64")
	share.InstrumentKind = InstrumentShare
	share.AssetUID = "4393bdb6-9017-466e-a3bb-ddb55d24159f"

	got := ByInstrument([]Event{gdr, share}, nil)

	if len(got) != 1 {
		t.Fatalf("got %d positions, want 1 (T): %v", len(got), got)
	}
	pos := got["4393bdb6-9017-466e-a3bb-ddb55d24159f"]
	if pos == nil {
		t.Fatalf("no T position: %v", got)
	}
	if !pos.Bought.Equal(decimal.RequireFromString("-80036.76")) {
		t.Errorf("T bought = %s, want -80036.76", pos.Bought)
	}
}

func TestByInstrumentAdjustsForSplit(t *testing.T) {
	const tAsset = "4393bdb6-9017-466e-a3bb-ddb55d24159f"
	msk := time.FixedZone("MSK", 3*60*60)

	before := ev("buy-before", "", KindBuy, "t-uid", "-3107.10", "1")
	before.InstrumentKind = InstrumentShare
	before.AssetUID = tAsset
	before.Date = time.Date(2025, 4, 10, 12, 0, 0, 0, msk)

	lastDay := ev("buy-last", "", KindBuy, "t-uid", "-3236.20", "1")
	lastDay.InstrumentKind = InstrumentShare
	lastDay.AssetUID = tAsset
	lastDay.Date = time.Date(2026, 4, 14, 18, 0, 0, 0, msk)

	after := ev("buy-after", "", KindBuy, "t-uid", "-1598", "5")
	after.InstrumentKind = InstrumentShare
	after.AssetUID = tAsset
	after.Date = time.Date(2026, 4, 17, 10, 0, 0, 0, msk)

	sell := ev("sell-after", "", KindSell, "t-uid", "960", "3")
	sell.InstrumentKind = InstrumentShare
	sell.AssetUID = tAsset
	sell.Date = time.Date(2026, 5, 20, 12, 0, 0, 0, msk)

	got := ByInstrument([]Event{before, lastDay, after, sell}, nil)

	pos := got[tAsset]
	if !pos.Quantity.Equal(decimal.NewFromInt(22)) {
		t.Errorf("quantity = %s, want 22 (1×10 + 1×10 + 5 − 3)", pos.Quantity)
	}
	if !pos.Bought.Equal(decimal.RequireFromString("-7941.30")) {
		t.Errorf("bought = %s, want -7941.30 (money is not changed by split)", pos.Bought)
	}
}

func TestByInstrumentAdjustsForReverseSplit(t *testing.T) {
	const vtbAsset = "94f18cc6-9e01-4db6-9d73-8c3ba6ab9655"
	msk := time.FixedZone("MSK", 3*60*60)

	before := ev("buy-before", "", KindBuy, "vtbr-uid", "-230", "10000")
	before.InstrumentKind = InstrumentShare
	before.AssetUID = vtbAsset
	before.Date = time.Date(2024, 7, 8, 12, 0, 0, 0, msk)

	after := ev("buy-after", "", KindBuy, "vtbr-uid", "-1150", "10")
	after.InstrumentKind = InstrumentShare
	after.AssetUID = vtbAsset
	after.Date = time.Date(2024, 7, 15, 12, 0, 0, 0, msk)

	got := ByInstrument([]Event{before, after}, nil)

	if !got[vtbAsset].Quantity.Equal(decimal.NewFromInt(12)) {
		t.Errorf("quantity = %s, want 12 (10000 / 5000 + 10)", got[vtbAsset].Quantity)
	}
}
