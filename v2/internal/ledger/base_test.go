package ledger

import (
	"errors"
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/shopspring/decimal"
)

type fakeConverter map[string]decimal.Decimal

func (f fakeConverter) ToBase(m money.Money, _ time.Time) (money.Money, error) {
	rate, ok := f[m.Currency]
	if !ok {
		return money.Money{}, errors.New("no rate")
	}
	return money.Money{Amount: m.Amount.Mul(rate), Currency: "RUB"}, nil
}

func mny(amount, currency string) money.Money {
	return money.Money{Amount: decimal.RequireFromString(amount), Currency: currency}
}

func TestApplyRUB(t *testing.T) {
	conv := fakeConverter{"RUB": decimal.NewFromInt(1), "USD": decimal.NewFromInt(80)}
	events := []Event{
		{OperationID: "rub-op", Amount: mny("-1000", "RUB")},
		{OperationID: "usd-op", Amount: mny("-9.2", "USD")},
	}

	if err := ApplyBase(events, conv); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !events[0].BaseAmount.Amount.Equal(decimal.RequireFromString("-1000")) {
		t.Errorf("rub-op = %s, want -1000", events[0].BaseAmount.Amount)
	}
	if !events[1].BaseAmount.Amount.Equal(decimal.RequireFromString("-736")) {
		t.Errorf("usd-op = %s, want -736", events[1].BaseAmount.Amount)
	}
}

func TestApplyRUBStopsOnError(t *testing.T) {
	conv := fakeConverter{"RUB": decimal.NewFromInt(1)}
	events := []Event{{OperationID: "eur-op", Amount: mny("5", "EUR")}}

	if err := ApplyBase(events, conv); err == nil {
		t.Fatal("want error for currency without rate")
	}
}
