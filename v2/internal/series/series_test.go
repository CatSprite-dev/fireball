package series

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestAt(t *testing.T) {
	d := func(day int) time.Time { return time.Date(2026, 8, day, 0, 0, 0, 0, time.UTC) }

	s := Series{
		{Date: d(3), Value: decimal.RequireFromString("272.32")},
		{Date: d(5), Value: decimal.RequireFromString("279.3")},
	}

	checks := []struct {
		name  string
		day   time.Time
		ok    bool
		value string
		from  time.Time
	}{
		{"до первой точки", d(2), false, "", time.Time{}},
		{"точка этого дня", d(3), true, "272.32", d(3)},
		{"точки нет — вчерашняя", d(4), true, "272.32", d(3)},
		{"точка этого дня", d(5), true, "279.3", d(5)},
		{"после последней — последняя", d(20), true, "279.3", d(5)},
	}
	for _, c := range checks {
		p, ok := s.At(c.day)
		if ok != c.ok {
			t.Errorf("%s: ok = %v, want %v", c.name, ok, c.ok)
			continue
		}
		if !ok {
			continue
		}
		if !p.Value.Equal(decimal.RequireFromString(c.value)) || !p.Date.Equal(c.from) {
			t.Errorf("%s: got %s от %s, want %s от %s",
				c.name, p.Value, p.Date.Format(time.DateOnly), c.value, c.from.Format(time.DateOnly))
		}
	}
}

func TestDayIsMoscowDate(t *testing.T) {
	// 3 января 01:30 по Москве = 2 января 22:30 UTC
	got := Day(time.Date(2024, 1, 2, 22, 30, 0, 0, time.UTC))
	want := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Day = %s, want %s", got, want)
	}
}

func TestEndOfDay(t *testing.T) {
	end := EndOfDay(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC))

	lastMinute := time.Date(2024, 1, 3, 23, 59, 0, 0, moscow)
	nextMidnight := time.Date(2024, 1, 4, 0, 0, 0, 0, moscow)

	if end.Before(lastMinute) {
		t.Errorf("23:59 по Москве должно входить в день, end = %s", end)
	}
	if !end.Before(nextMidnight) {
		t.Errorf("полночь следующего дня не должна входить в день, end = %s", end)
	}
}
