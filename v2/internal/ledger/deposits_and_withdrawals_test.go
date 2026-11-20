package ledger

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestDepositsAndWithdrawals(t *testing.T) {
	at := func(kind OperationKind, amount string, hour int) Event {
		e := ev("", "", kind, "", amount, "")
		e.Date = time.Date(2026, time.March, 2, hour, 0, 0, 0, moscow)
		return e
	}
	events := []Event{
		at(KindDeposit, "23000", 10),
		at(KindWithdrawal, "-22541", 15),
		at(KindBuy, "-400", 11),
	}

	in, out := DepositsAndWithdrawals(events)

	day := time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC)
	if len(in) != 1 || !in[0].Date.Equal(day) || !in[0].Value.Equal(decimal.RequireFromString("23000")) {
		t.Errorf("in = %+v, want one point 23000 on %s", in, day.Format(time.DateOnly))
	}
	if len(out) != 1 || !out[0].Date.Equal(day) || !out[0].Value.Equal(decimal.RequireFromString("-22541")) {
		t.Errorf("out = %+v, want one point -22541 on %s", out, day.Format(time.DateOnly))
	}
}
