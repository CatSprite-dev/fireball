package snapshot

import (
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/shopspring/decimal"
)

func TestPositionRows(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	d := func(day int) time.Time { return time.Date(2026, 8, day, 0, 0, 0, 0, time.UTC) }
	at := func(day, hour int) time.Time { return time.Date(2026, 8, day, hour, 0, 0, 0, msk) }

	history := ledger.History{
		{Date: at(2, 12), Quantity: decimal.NewFromInt(10)},
		{Date: at(6, 15), Quantity: decimal.Zero},
	}
	prices := series.Series{
		{Date: d(4), Value: decimal.RequireFromString("272.58")},
		{Date: d(5), Value: decimal.RequireFromString("279.3")},
	}

	rows, gaps := positionRows("t", history, prices, d(1), d(7))

	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2: %+v", len(rows), rows)
	}
	wantValues := []string{"2725.8", "2793"}
	for i, r := range rows {
		if !r.Date.Equal(d(4+i)) || !r.Value.Equal(decimal.RequireFromString(wantValues[i])) {
			t.Errorf("row %d: %s value %s, want %s value %s",
				i, r.Date.Format(time.DateOnly), r.Value, d(4+i).Format(time.DateOnly), wantValues[i])
		}
	}

	if len(gaps) != 1 {
		t.Fatalf("got %d gaps, want 1: %+v", len(gaps), gaps)
	}
	if !gaps[0].From.Equal(d(2)) || !gaps[0].To.Equal(d(3)) {
		t.Errorf("gap %s — %s, want 2026-08-02 — 2026-08-03",
			gaps[0].From.Format(time.DateOnly), gaps[0].To.Format(time.DateOnly))
	}
}
