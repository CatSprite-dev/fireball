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

	rows, gaps := positionRows("t", ledger.InstrumentShare, history, prices, Bond{}, d(1), d(7))

	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2: %+v", len(rows), rows)
	}
	wantValues := []string{"2725.8", "2793"}
	for i, r := range rows {
		if !r.Date.Equal(d(4+i)) || !r.Value.Equal(decimal.RequireFromString(wantValues[i])) {
			t.Errorf("row %d: %s value %s, want %s value %s",
				i, r.Date.Format(time.DateOnly), r.Value, d(4+i).Format(time.DateOnly), wantValues[i])
		}
		if r.Kind != ledger.InstrumentShare {
			t.Errorf("row %d: kind %q, want share", i, r.Kind)
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

	rows, gaps := positionRows("si", ledger.InstrumentFutures, history, nil, Bond{}, d(2), d(3))

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
		if r.Kind != ledger.InstrumentFutures {
			t.Errorf("%s: kind %q, want futures", r.Date.Format(time.DateOnly), r.Kind)
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
		if r.Kind != ledger.InstrumentCurrency {
			t.Errorf("row %d: kind %q, want currency", i, r.Kind)
		}
	}
}

func TestPositionRowsBond(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	day := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)

	history := ledger.History{
		{Date: time.Date(2025, 11, 20, 12, 0, 0, 0, msk), Quantity: decimal.NewFromInt(112)},
	}
	prices := series.Series{{Date: day, Value: decimal.RequireFromString("102.16")}}
	bond := Bond{
		Nominal: series.Series{{Date: day, Value: decimal.NewFromInt(1000)}},
		NKD:     series.Series{{Date: day, Value: decimal.RequireFromString("11.38")}},
	}

	rows, gaps := positionRows("dbm5", ledger.InstrumentBond, history, prices, bond, day, day)

	if len(gaps) != 0 || len(rows) != 1 {
		t.Fatalf("got %d rows and %d gaps, want 1 and 0", len(rows), len(gaps))
	}
	r := rows[0]
	checks := []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		{"цена в рублях", r.Price.Value, "1021.6"},
		{"НКД", r.NKD, "11.38"},
		{"стоимость", r.Value, "115693.76"},
	}
	for _, c := range checks {
		if !c.got.Equal(decimal.RequireFromString(c.want)) {
			t.Errorf("%s = %s, want %s", c.name, c.got, c.want)
		}
	}
}

func TestPositionRowsBondWithoutNominalIsGap(t *testing.T) {
	day := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	history := ledger.History{{Date: day, Quantity: decimal.NewFromInt(1)}}
	prices := series.Series{{Date: day, Value: decimal.RequireFromString("102.16")}}

	rows, gaps := positionRows("bond", ledger.InstrumentBond, history, prices, Bond{}, day, day)

	if len(rows) != 0 || len(gaps) != 1 {
		t.Errorf("got %d rows and %d gaps, want 0 and 1", len(rows), len(gaps))
	}
}

func TestTotal(t *testing.T) {
	d := func(day int) time.Time { return time.Date(2026, 8, day, 0, 0, 0, 0, time.UTC) }
	row := func(day int, key string, kind ledger.InstrumentKind, value string) Row {
		return Row{Date: d(day), Key: key, Kind: kind, Value: decimal.RequireFromString(value)}
	}

	rows := []Row{
		row(4, "t", ledger.InstrumentShare, "2793"),
		row(2, "sber", ledger.InstrumentShare, "3100.5"),
		row(2, "t", ledger.InstrumentShare, "2725.8"),
		row(2, "RUB", ledger.InstrumentCurrency, "1000.25"),
		row(4, "si", ledger.InstrumentFutures, "0"),
	}

	got := Total(rows, d(2), d(4))

	want := []string{"6826.55", "0", "2793"}
	if len(got) != len(want) {
		t.Fatalf("got %d points, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if !got[i].Date.Equal(d(2+i)) || !got[i].Value.Equal(decimal.RequireFromString(w)) {
			t.Errorf("point %d: %s %s, want %s %s",
				i, got[i].Date.Format(time.DateOnly), got[i].Value, d(2+i).Format(time.DateOnly), w)
		}
	}
}
