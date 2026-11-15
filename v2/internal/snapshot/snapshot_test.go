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

	rows, gaps := positionRows("t", ledger.InstrumentShare, history, prices, d(1), d(7))

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

func TestPositionRowsFuturesHaveZeroValue(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	d := func(day int) time.Time { return time.Date(2026, 8, day, 0, 0, 0, 0, time.UTC) }

	history := ledger.History{
		{Date: time.Date(2026, 8, 2, 12, 0, 0, 0, msk), Quantity: decimal.NewFromInt(1)},
	}

	rows, gaps := positionRows("si", ledger.InstrumentFutures, history, nil, d(2), d(3))

	if len(gaps) != 0 {
		t.Errorf("futures must not have price gaps: %+v", gaps)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	for _, r := range rows {
		if !r.Value.IsZero() || !r.Quantity.Equal(decimal.NewFromInt(1)) {
			t.Errorf("%s: quantity %s value %s, want 1 and 0", r.Date.Format(time.DateOnly), r.Quantity, r.Value)
		}
	}
}

func TestCashRows(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	d := func(day int) time.Time { return time.Date(2025, 1, day, 0, 0, 0, 0, time.UTC) }
	at := func(day, hour int) time.Time { return time.Date(2025, 1, day, hour, 0, 0, 0, msk) }

	cash := ledger.History{
		{Date: at(10, 12), Quantity: decimal.RequireFromString("200000")},
		{Date: at(11, 12), Quantity: decimal.RequireFromString("99439.74")},
	}

	rows := cashRows("RUB", cash, d(9), d(11))

	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2: %+v", len(rows), rows)
	}
	want := []string{"200000", "99439.74"}
	for i, r := range rows {
		if r.Key != "RUB" || !r.Value.Equal(decimal.RequireFromString(want[i])) || !r.Price.Value.Equal(decimal.NewFromInt(1)) {
			t.Errorf("row %d: %s %s value %s price %s, want RUB %s price 1",
				i, r.Date.Format(time.DateOnly), r.Key, r.Value, r.Price.Value, want[i])
		}
	}
}
