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
