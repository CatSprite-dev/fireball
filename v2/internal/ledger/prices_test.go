package ledger

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestTradePrices(t *testing.T) {
	const tAsset = "t-asset"
	withSplit(t, tAsset, split{date: time.Date(2026, 4, 17, 0, 0, 0, 0, moscow), before: 1, after: 10})

	trade := func(id string, kind OperationKind, instrument InstrumentKind, amount, qty string, date time.Time) Event {
		e := ev(id, "", kind, "t-uid", amount, qty)
		e.InstrumentKind = instrument
		e.AssetUID = tAsset
		e.Date = date
		return e
	}
	at := func(m time.Month, d, h int) time.Time { return time.Date(2026, m, d, h, 0, 0, 0, moscow) }

	events := []Event{
		trade("b1", KindBuy, InstrumentShare, "-3107.10", "1", time.Date(2025, 4, 10, 12, 0, 0, 0, moscow)),
		trade("b2", KindBuy, InstrumentShare, "-1598", "5", at(time.April, 17, 10)),
		trade("b3", KindBuy, InstrumentShare, "-1610", "5", at(time.April, 17, 15)),
		trade("s1", KindSell, InstrumentShare, "960", "3", at(time.May, 20, 12)),
	}

	bond := ev("bond", "", KindBuy, "bond-uid", "-1028.33", "1")
	bond.InstrumentKind = InstrumentBond
	bond.AssetUID = "bond-asset"
	fut := ev("fut", "", KindBuy, "si-uid", "0", "1")
	fut.InstrumentKind = InstrumentFutures
	events = append(events, bond, fut)

	got := TradePrices(events, nil)

	if len(got) != 1 {
		t.Fatalf("got %d positions, want 1 (only T): %v", len(got), got)
	}
	day := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }
	want := []struct {
		date  time.Time
		price string
	}{
		{day(2025, time.April, 10), "310.71"},
		{day(2026, time.April, 17), "322"},
		{day(2026, time.May, 20), "320"},
	}
	s := got[tAsset]
	if len(s) != len(want) {
		t.Fatalf("got %d points, want %d: %+v", len(s), len(want), s)
	}
	for i, w := range want {
		if !s[i].Date.Equal(w.date) || !s[i].Value.Equal(decimal.RequireFromString(w.price)) {
			t.Errorf("point %d: %s %s, want %s %s",
				i, s[i].Date.Format(time.DateOnly), s[i].Value, w.date.Format(time.DateOnly), w.price)
		}
	}
}
