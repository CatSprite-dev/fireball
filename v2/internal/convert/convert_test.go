package convert

import (
	"errors"
	"testing"

	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
	"github.com/shopspring/decimal"
)

func TestToDecimal(t *testing.T) {
	tests := []struct {
		name  string
		units string
		nano  int32
		want  string
	}{
		{"целое", "1234", 0, "1234"},
		{"с копейками", "1234", 560000000, "1234.56"},
		{"только дробная часть", "0", 170000000, "0.17"},
		{"отрицательное", "-21", -250000000, "-21.25"},
		{"ноль", "0", 0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToDecimal(tt.units, tt.nano)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(decimal.RequireFromString(tt.want)) {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestToDecimalEmpty(t *testing.T) {
	_, err := ToDecimal("", 0)
	if !errors.Is(err, ErrNoValue) {
		t.Fatalf("want ErrNoValue, got %v", err)
	}
}

func TestToDecimalInvalid(t *testing.T) {
	_, err := ToDecimal("abc", 0)
	if err == nil || errors.Is(err, ErrNoValue) {
		t.Fatalf("want parse error, got %v", err)
	}
}

func TestToMoneyUppercasesCurrency(t *testing.T) {
	got, err := ToMoney(tinvest.MoneyValue{Currency: "rub", Units: "100", Nano: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Currency != "RUB" {
		t.Fatalf("currency = %q, want RUB", got.Currency)
	}
}
