package ledger

import (
	"slices"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestNetDeposits(t *testing.T) {
	at := func(kind OperationKind, amount string, date time.Time) Event {
		e := ev("", "", kind, "", amount, "")
		e.Date = date
		return e
	}
	msk := func(m time.Month, d, h int) time.Time { return time.Date(2026, m, d, h, 0, 0, 0, moscow) }
	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }

	type point struct {
		date  time.Time
		value string
	}
	cases := []struct {
		name   string
		events []Event
		want   []point
	}{
		{
			name: "пополнение и вывод в один день",
			events: []Event{
				at(KindDeposit, "100000", msk(time.March, 2, 10)),
				at(KindWithdrawal, "-30000", msk(time.March, 2, 15)),
			},
			want: []point{{day(time.March, 2), "70000"}},
		},
		{
			name: "переводы погасились — точки нет",
			events: []Event{
				at(KindTransferOut, "-5000", msk(time.March, 2, 10)),
				at(KindTransferIn, "5000", msk(time.March, 2, 11)),
			},
		},
		{
			name: "23:30 UTC — уже следующий московский день",
			events: []Event{
				at(KindDeposit, "1000", time.Date(2026, time.March, 2, 23, 30, 0, 0, time.UTC)),
			},
			want: []point{{day(time.March, 3), "1000"}},
		},
		{
			name: "покупка и дивиденд не пополнения",
			events: []Event{
				at(KindBuy, "-2000", msk(time.March, 2, 10)),
				at(KindDividend, "150", msk(time.March, 3, 10)),
				at(KindDeposit, "500", msk(time.March, 3, 11)),
			},
			want: []point{{day(time.March, 3), "500"}},
		},
		{
			name: "события не по порядку",
			events: []Event{
				at(KindDeposit, "300", msk(time.March, 5, 10)),
				at(KindDeposit, "100", msk(time.March, 1, 10)),
				at(KindWithdrawal, "-200", msk(time.March, 3, 10)),
			},
			want: []point{
				{day(time.March, 1), "100"},
				{day(time.March, 3), "-200"},
				{day(time.March, 5), "300"},
			},
		},
		{
			name: "пусто",
		},
	}
	for _, c := range cases {
		original := slices.Clone(c.events)
		got := NetDeposits(c.events)

		if len(got) != len(c.want) {
			t.Errorf("%s: got %d points, want %d: %+v", c.name, len(got), len(c.want), got)
			continue
		}
		for i, w := range c.want {
			if !got[i].Date.Equal(w.date) || !got[i].Value.Equal(decimal.RequireFromString(w.value)) {
				t.Errorf("%s: point %d: %s %s, want %s %s",
					c.name, i, got[i].Date.Format(time.DateOnly), got[i].Value, w.date.Format(time.DateOnly), w.value)
			}
		}
		for i := range original {
			if !c.events[i].Date.Equal(original[i].Date) || !c.events[i].BaseAmount.Amount.Equal(original[i].BaseAmount.Amount) {
				t.Errorf("%s: event %d changed", c.name, i)
			}
		}
	}
}
