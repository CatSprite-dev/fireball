package ui

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/load"
	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
)

func TestPerformance(t *testing.T) {
	msk := time.FixedZone("MSK", 3*60*60)
	start := time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC)
	half, today := start.AddDate(0, 0, 182), start.AddDate(0, 0, 365)
	at := func(kind ledger.OperationKind, asset, amount, qty string, day time.Time) ledger.Event {
		e := ledger.Event{
			AccountID:     "acc",
			Date:          time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, msk),
			OperationKind: kind,
			Amount:        money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"},
			BaseAmount:    money.Money{Amount: decimal.RequireFromString(amount), Currency: "RUB"},
		}
		if asset != "" {
			e.AssetUID, e.InstrumentUID, e.InstrumentKind = asset, asset+"-uid", ledger.InstrumentShare
			e.Quantity = decimal.RequireFromString(qty)
		}
		return e
	}
	pt := func(day time.Time, v string) series.Point {
		return series.Point{Date: day, Value: decimal.RequireFromString(v)}
	}

	p := &load.Portfolio{Events: []ledger.Event{
		// в первый день внесли 1000 и купили 10 шт. по 100.
		at(ledger.KindDeposit, "", "1000", "", start),
		at(ledger.KindBuy, "a", "-1000", "10", start),
		// через год — ещё 1100 деньгами.
		at(ledger.KindDeposit, "", "1100", "", today),
	}}
	m := &load.Market{Prices: map[string]series.Series{
		// цена 100 → 110 через полгода → 121 через год.
		"a": {pt(start, "100"), pt(half, "110"), pt(today, "121")},
	}}
	// индекс начинается только через полгода: 1000 → 1100.
	index := series.Series{pt(half, "1000"), pt(today, "1100")}

	got, err := performance(p, m, index, "RUB", []string{"acc"}, "all", today)
	if err != nil {
		t.Fatal(err)
	}

	want := Performance{
		From: start.Format(time.DateOnly),
		// за полгода +10%; в последний день (1210 + 1100) / (1100 + 1100) = +5%; 1.1 × 1.05 = 1.155.
		TWR: "15.50",
		// −1000 в начале, −1100 и +2310 через год: 1000 × (1 + r) = 1210, r = 21%.
		MWR: "21.00",
		// индекс с полугода: +10%.
		Index: "10.00",
		// TWR с полугода +5% против индекса +10%.
		VsIndex:     "-5.00",
		CompareFrom: half.Format(time.DateOnly),
	}
	if got != want {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
}

func TestPeriodStart(t *testing.T) {
	today := time.Date(2026, time.November, 21, 0, 0, 0, 0, time.UTC)
	first := ledger.Event{Date: time.Date(2020, time.March, 16, 10, 0, 0, 0, time.UTC)}

	checks := []struct {
		period string
		want   time.Time
	}{
		{"1m", time.Date(2026, time.October, 21, 0, 0, 0, 0, time.UTC)},
		{"12m", time.Date(2025, time.November, 21, 0, 0, 0, 0, time.UTC)},
		{"5y", time.Date(2021, time.November, 21, 0, 0, 0, 0, time.UTC)},
		{"all", time.Date(2020, time.March, 16, 0, 0, 0, 0, time.UTC)},
	}
	for _, c := range checks {
		got, err := periodStart(c.period, today, []ledger.Event{first})
		if err != nil || !got.Equal(c.want) {
			t.Errorf("%s: got %s %v, want %s", c.period, got.Format(time.DateOnly), err, c.want.Format(time.DateOnly))
		}
	}
	if _, err := periodStart("2w", today, nil); err == nil {
		t.Error("unknown period: want error")
	}
}
