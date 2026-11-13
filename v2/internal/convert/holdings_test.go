package convert

import (
	"testing"

	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
	"github.com/shopspring/decimal"
)

func TestToHoldings(t *testing.T) {
	p := &tinvest.Portfolio{
		Positions: []tinvest.Position{
			{
				InstrumentUID:  "dbm5",
				Ticker:         "RU000A10DBM5",
				InstrumentType: "bond",
				Quantity:       tinvest.Quotation{Units: "112"},
				CurrentPrice:   tinvest.MoneyValue{Currency: "rub", Units: "1020", Nano: 300000000},
				CurrentNkd:     tinvest.MoneyValue{Currency: "rub", Units: "10", Nano: 900000000},
			},
			{
				InstrumentUID:  "rub",
				Ticker:         "RUB000UTSTOM",
				InstrumentType: "currency",
				Quantity:       tinvest.Quotation{Units: "23", Nano: 490000000},
				CurrentPrice:   tinvest.MoneyValue{Currency: "rub", Units: "1"},
			},
		},
	}

	got, err := ToHoldings("self", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d holdings, want 2", len(got))
	}

	if got[0].AccountID != "self" {
		t.Errorf("account = %q, want self", got[0].AccountID)
	}
	if !got[0].NKD.Equal(decimal.RequireFromString("10.9")) {
		t.Errorf("nkd = %s, want 10.9", got[0].NKD)
	}
	if !got[1].NKD.IsZero() || !got[1].BrokerYield.IsZero() {
		t.Errorf("rub: empty nkd and yield should be zero, got nkd %s yield %s", got[1].NKD, got[1].BrokerYield)
	}

	total := got[0].Value().Add(got[1].Value())
	if !total.Equal(decimal.RequireFromString("115517.89")) {
		t.Errorf("total = %s, want 115517.89 (broker's total for this account)", total)
	}
}

func TestToHoldingsEmptyPrice(t *testing.T) {
	p := &tinvest.Portfolio{
		Positions: []tinvest.Position{{Ticker: "BAD", Quantity: tinvest.Quotation{Units: "1"}}},
	}

	if _, err := ToHoldings("acc", p); err == nil {
		t.Fatal("want error for empty price")
	}
}
