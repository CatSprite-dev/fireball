package money

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func rub(s string) Money {
	return Money{Amount: decimal.RequireFromString(s), Currency: "RUB"}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b Money
		want Money
	}{
		{"простое сложение", rub("1234.56"), rub("0.44"), rub("1235.00")},
		{"с отрицательной", rub("100"), rub("-30.5"), rub("69.5")},
		{"точность копеек", rub("0.1"), rub("0.2"), rub("0.3")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Add(tt.b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Amount.Equal(tt.want.Amount) || got.Currency != tt.want.Currency {
				t.Fatalf("got %s %s, want %s %s", got.Amount, got.Currency, tt.want.Amount, tt.want.Currency)
			}
		})
	}
}

func TestAddCurrencyMismatch(t *testing.T) {
	usd := Money{Amount: decimal.RequireFromString("1"), Currency: "USD"}

	_, err := rub("1").Add(usd)
	if !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("want ErrCurrencyMismatch, got %v", err)
	}
}
