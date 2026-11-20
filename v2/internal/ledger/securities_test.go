package ledger

import (
	"slices"
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/shopspring/decimal"
)

type point struct {
	date  time.Time
	value string
}

func TestSecurityFlows(t *testing.T) {
	at := func(kind OperationKind, instrument InstrumentKind, amount string, date time.Time) Event {
		e := ev("", "", kind, "", amount, "")
		e.InstrumentKind = instrument
		e.Date = date
		return e
	}
	msk := func(d, h int) time.Time { return time.Date(2026, time.March, d, h, 0, 0, 0, moscow) }
	day := func(d int) time.Time { return time.Date(2026, time.March, d, 0, 0, 0, 0, time.UTC) }

	check := func(t *testing.T, name string, got series.Series, want []point) {
		t.Helper()
		if len(got) != len(want) {
			t.Errorf("%s: got %d points, want %d: %+v", name, len(got), len(want), got)
			return
		}
		for i, w := range want {
			if !got[i].Date.Equal(w.date) || !got[i].Value.Equal(decimal.RequireFromString(w.value)) {
				t.Errorf("%s: point %d: %s %s, want %s %s",
					name, i, got[i].Date.Format(time.DateOnly), got[i].Value, w.date.Format(time.DateOnly), w.value)
			}
		}
	}

	cases := []struct {
		name   string
		events []Event
		trades []point
		income []point
	}{
		{
			name: "покупка и комиссия в один день",
			events: []Event{
				at(KindBuy, InstrumentShare, "-10000", msk(2, 10)),
				at(KindFee, InstrumentShare, "-3", msk(2, 10)),
			},
			trades: []point{{day(2), "-10000"}},
			income: []point{{day(2), "-3"}},
		},
		{
			name: "продажа облигации и купон в разные дни",
			events: []Event{
				at(KindCoupon, InstrumentBond, "120", msk(5, 12)),
				at(KindSell, InstrumentBond, "5000", msk(3, 11)),
			},
			trades: []point{{day(3), "5000"}},
			income: []point{{day(5), "120"}},
		},
		{
			name: "погашение облигации",
			events: []Event{
				at(KindRedemption, InstrumentBond, "1000", msk(4, 9)),
			},
			trades: []point{{day(4), "1000"}},
		},
		{
			name: "не бумаги игнорируются",
			events: []Event{
				at(KindBuy, InstrumentFutures, "-200", msk(2, 10)),
				at(KindBuy, InstrumentOption, "-50", msk(2, 11)),
				at(KindBuy, InstrumentCurrency, "-9000", msk(2, 12)),
				at(KindDeposit, "", "100000", msk(2, 13)),
			},
		},
		{
			name: "дивиденд и налог в один день",
			events: []Event{
				at(KindDividend, InstrumentShare, "500", msk(6, 10)),
				at(KindTax, InstrumentShare, "-65", msk(6, 10)),
			},
			income: []point{{day(6), "435"}},
		},
		{
			name: "пусто",
		},
	}
	for _, c := range cases {
		original := slices.Clone(c.events)
		trades, income := SecurityFlows(c.events)

		check(t, c.name+": trades", trades, c.trades)
		check(t, c.name+": income", income, c.income)
		for i := range original {
			if !c.events[i].Date.Equal(original[i].Date) || !c.events[i].BaseAmount.Amount.Equal(original[i].BaseAmount.Amount) {
				t.Errorf("%s: event %d changed", c.name, i)
			}
		}
	}
}
