package metrics

import (
	"testing"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/shopspring/decimal"
)

func ev(id, parent string, kind ledger.OperationKind, uid, amount, qty string) ledger.Event {
	e := ledger.Event{
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
	events := []ledger.Event{
		ev("dep", "", ledger.KindDeposit, "", "200000", ""),
		ev("buy1", "", ledger.KindBuy, "etln", "-100510", "100"),
		ev("fee1", "buy1", ledger.KindFee, "etln", "-50.26", ""),
		ev("cpn", "", ledger.KindCoupon, "etln", "2000", ""),
		ev("tax", "", ledger.KindTax, "etln", "-260", ""),
		ev("sell1", "", ledger.KindSell, "etln", "100789", "100"),
		ev("fee2", "sell1", ledger.KindFee, "etln", "-301.53", ""),
		ev("svc", "", ledger.KindFee, "", "-99", ""),
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

	oldDollar := ev("buy-old", "", ledger.KindBuy, "usd-cngd", "-810.15", "11")
	oldDollar.InstrumentKind = ledger.InstrumentCurrency
	oldDollar.AssetUID = "usd-asset"

	newDollar := ev("buy-new", "", ledger.KindBuy, "usd-tom", "-6722.35", "74")
	newDollar.InstrumentKind = ledger.InstrumentCurrency
	newDollar.AssetUID = "usd-asset"

	share := ev("buy-share", "", ledger.KindBuy, "etln", "-100510", "100")
	share.InstrumentKind = ledger.InstrumentShare

	got := ByInstrument([]ledger.Event{oldDollar, newDollar, share}, currencyByAsset)

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
	buyUSD := ev("buy-usd", "", ledger.KindBuy, "usd-tom", "-810.15", "11")
	buyUSD.InstrumentKind = ledger.InstrumentCurrency
	buyUSD.AssetUID = "usd-asset"

	leg := ev("buy-usd:leg", "buy-usd", ledger.KindExchange, "usd-tom", "810.15", "")
	leg.InstrumentKind = ledger.InstrumentCurrency
	leg.AssetUID = "usd-asset"
	leg.Amount = money.Money{Amount: decimal.RequireFromString("11"), Currency: "USD"}

	buyShare := ev("buy-aapl", "", ledger.KindBuy, "aapl", "-740", "1")
	buyShare.Amount = money.Money{Amount: decimal.RequireFromString("-10"), Currency: "USD"}

	dividend := ev("div-aapl", "", ledger.KindDividend, "aapl", "37.5", "")
	dividend.Amount = money.Money{Amount: decimal.RequireFromString("0.5"), Currency: "USD"}

	got := ByInstrument([]ledger.Event{buyUSD, leg, buyShare, dividend}, map[string]string{"usd-asset": "USD"})

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
	events := []ledger.Event{
		ev("vm:0", "vm", ledger.KindVariationMargin, "fut", "2885", ""),
		ev("vm2:0", "vm2", ledger.KindVariationMargin, "fut", "-1687", ""),
		ev("exp", "", ledger.KindExpiration, "opt", "264", ""),
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
	events := []ledger.Event{
		ev("buy-bond", "", ledger.KindBuy, "bond", "-1028.33", "1"),
		ev("redeem", "", ledger.KindRedemption, "bond", "1035.10", ""),

		ev("buy-opt", "", ledger.KindBuy, "opt", "-305.90", "322"),
		ev("exp", "", ledger.KindExpiration, "opt", "0", ""),

		ev("buy-amort", "", ledger.KindBuy, "amort", "-10000", "10"),
		ev("amort", "", ledger.KindAmortization, "amort", "2500", ""),
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
	oldUID := ev("buy", "", ledger.KindBuy, "tspx-tqtd", "-6174.67", "700")
	oldUID.InstrumentKind = ledger.InstrumentETF
	oldUID.AssetUID = "tspx-asset"

	newUID := ev("sell", "", ledger.KindSell, "tspx-tqtf", "5497.80", "700")
	newUID.InstrumentKind = ledger.InstrumentETF
	newUID.AssetUID = "tspx-asset"

	option := ev("opt", "", ledger.KindBuy, "gz130", "-500.80", "80")
	option.InstrumentKind = ledger.InstrumentOption
	option.AssetUID = "gazp-asset"

	share := ev("gazp", "", ledger.KindBuy, "gazp-uid", "-25821.30", "100")
	share.InstrumentKind = ledger.InstrumentShare
	share.AssetUID = "gazp-asset"

	got := ByInstrument([]ledger.Event{oldUID, newUID, option, share}, nil)

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
	oldTSPX := ev("buy", "", ledger.KindBuy, "tspx-tqtd", "-6174.67", "700")
	oldTSPX.InstrumentKind = ledger.InstrumentETF
	oldTSPX.AssetUID = "tspx-asset"

	newTSPX := ev("sell", "", ledger.KindSell, "tspx-tqtf", "5497.80", "700")
	newTSPX.InstrumentKind = ledger.InstrumentETF
	newTSPX.AssetUID = "tspx-asset"

	usd := ev("buy-usd", "", ledger.KindBuy, "usd-tom", "-810.15", "11")
	usd.InstrumentKind = ledger.InstrumentCurrency
	usd.AssetUID = "usd-asset"

	option := ev("opt", "", ledger.KindBuy, "gz130", "-500.80", "80")
	option.InstrumentKind = ledger.InstrumentOption
	option.AssetUID = "gazp-asset"

	got := PageKeys([]ledger.Event{oldTSPX, newTSPX, usd, option}, map[string]string{"usd-asset": "USD"})

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
