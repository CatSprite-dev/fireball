package convert

import (
	"errors"
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
	"github.com/shopspring/decimal"
)

func TestToPoints(t *testing.T) {
	day := time.Date(2023, 12, 29, 0, 0, 0, 0, time.UTC)
	candles := []tinvest.Candle{
		{Time: day, Close: tinvest.Quotation{Units: "90", Nano: 360000000}},
	}

	points, err := ToPoints(candles, decimal.NewFromInt(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 1 {
		t.Fatalf("got %d points, want 1", len(points))
	}
	if !points[0].Date.Equal(day) || !points[0].Rate.Equal(decimal.RequireFromString("90.36")) {
		t.Errorf("got %+v, want 2023-12-29 90.36", points[0])
	}
}

func TestToPointsEmptyClose(t *testing.T) {
	candles := []tinvest.Candle{{Time: time.Now()}}

	_, err := ToPoints(candles, decimal.NewFromInt(1))
	if !errors.Is(err, ErrNoValue) {
		t.Errorf("got %v, want ErrNoValue", err)
	}
}

func TestToPointsDividesByNominal(t *testing.T) {
	candles := []tinvest.Candle{
		{Time: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC), Close: tinvest.Quotation{Units: "18", Nano: 50000000}},
	}

	points, err := ToPoints(candles, decimal.NewFromInt(100))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !points[0].Rate.Equal(decimal.RequireFromString("0.1805")) {
		t.Errorf("rate = %s, want 0.1805", points[0].Rate)
	}
}

func TestToPointsZeroNominal(t *testing.T) {
	_, err := ToPoints(nil, decimal.Zero)
	if !errors.Is(err, ErrZeroNominal) {
		t.Errorf("got %v, want ErrZeroNominal", err)
	}
}
