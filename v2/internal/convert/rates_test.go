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

	points, err := ToPoints(candles)
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

	_, err := ToPoints(candles)
	if !errors.Is(err, ErrNoValue) {
		t.Errorf("got %v, want ErrNoValue", err)
	}
}
