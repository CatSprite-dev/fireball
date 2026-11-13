package metrics

import (
	"testing"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/shopspring/decimal"
)

func ev(id, parent string, kind ledger.Kind, uid, amount, qty string) ledger.Event {
	e := ledger.Event{
		OperationID:   id,
		ParentID:      parent,
		Kind:          kind,
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
		ev("fee1", "buy1", ledger.KindFee, "", "-50.26", ""),
		ev("cpn", "", ledger.KindCoupon, "etln", "2000", ""),
		ev("tax", "", ledger.KindTax, "etln", "-260", ""),
		ev("sell1", "", ledger.KindSell, "etln", "100789", "100"),
		ev("fee2", "sell1", ledger.KindFee, "", "-301.53", ""),
		ev("svc", "", ledger.KindFee, "", "-99", ""),
	}

	got := ByInstrument(events)

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
