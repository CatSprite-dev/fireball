package fx

import (
	"errors"
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/shopspring/decimal"
)

func date(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

func usdRates() *Rates {
	return New(map[string][]Point{
		"USD": {
			{Date: date("2024-01-03"), Rate: decimal.RequireFromString("92")},
			{Date: date("2023-12-29"), Rate: decimal.RequireFromString("90.36")},
		},
	})
}

func TestRate(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		date     string
		want     string
	}{
		{"rub is always one", "RUB", "2020-01-01", "1"},
		{"exact date", "USD", "2024-01-03", "92"},
		{"holiday takes previous", "USD", "2024-01-01", "90.36"},
		{"after last takes last", "USD", "2024-06-01", "92"},
	}

	rates := usdRates()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rates.Rate(tt.currency, date(tt.date))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(decimal.RequireFromString(tt.want)) {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRateErrors(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		date     string
		want     error
	}{
		{"before first point", "USD", "2023-12-28", ErrNoRate},
		{"unknown currency", "CNY", "2024-01-03", ErrUnknownCurrency},
	}

	rates := usdRates()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := rates.Rate(tt.currency, date(tt.date))
			if !errors.Is(err, tt.want) {
				t.Errorf("got error %v, want %v", err, tt.want)
			}
		})
	}
}

func TestNewDoesNotModifyInput(t *testing.T) {
	input := []Point{
		{Date: date("2024-01-03"), Rate: decimal.RequireFromString("92")},
		{Date: date("2023-12-29"), Rate: decimal.RequireFromString("90.36")},
	}

	New(map[string][]Point{"USD": input})

	if !input[0].Date.Equal(date("2024-01-03")) {
		t.Errorf("input was reordered: first date %s", input[0].Date.Format(time.DateOnly))
	}
}
func TestRateUsesMoscowDay(t *testing.T) {
	// 3 января 01:30 по Москве = 2 января 22:30 UTC
	event := time.Date(2024, 1, 2, 22, 30, 0, 0, time.UTC)

	got, err := usdRates().Rate("USD", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(decimal.RequireFromString("92")) {
		t.Errorf("got %s, want 92 (rate of Jan 3)", got)
	}
}

func TestToRUB(t *testing.T) {
	m := money.Money{Amount: decimal.RequireFromString("-9.2"), Currency: "USD"}

	got, err := usdRates().ToBase(m, date("2024-01-03"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Currency != "RUB" || !got.Amount.Equal(decimal.RequireFromString("-846.4")) {
		t.Errorf("got %s %s, want -846.4 RUB", got.Amount, got.Currency)
	}
}
