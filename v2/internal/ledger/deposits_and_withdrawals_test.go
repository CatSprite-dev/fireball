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

func TestDepositsAndWithdrawalsInternalTransfer(t *testing.T) {
	at := func(account string, kind OperationKind, amount string, hour int) Event {
		e := ev("", "", kind, "", amount, "")
		e.AccountID = account
		e.Date = time.Date(2026, time.July, 24, hour, 0, 0, 0, moscow)
		return e
	}
	// перевод 150000 с брокерского на робота и обычное пополнение робота на 300 в тот же день.
	events := []Event{
		at("broker", KindTransferOut, "-150000", 11),
		at("robot", KindTransferIn, "150000", 11),
		at("robot", KindDeposit, "300", 12),
	}
	day := time.Date(2026, time.July, 24, 0, 0, 0, 0, time.UTC)

	t.Run("оба счёта выбраны — перевод гасится", func(t *testing.T) {
		in, out := DepositsAndWithdrawals(events)
		if len(in) != 1 || !in[0].Date.Equal(day) || !in[0].Value.Equal(decimal.RequireFromString("300")) {
			t.Errorf("in = %+v, want one point 300", in)
		}
		if len(out) != 0 {
			t.Errorf("out = %+v, want empty", out)
		}
	})

	t.Run("выбран только робот — перевод это приход", func(t *testing.T) {
		in, out := DepositsAndWithdrawals(Filter(events, []string{"robot"}))
		if len(in) != 1 || !in[0].Value.Equal(decimal.RequireFromString("150300")) {
			t.Errorf("in = %+v, want one point 150300", in)
		}
		if len(out) != 0 {
			t.Errorf("out = %+v, want empty", out)
		}
	})

	t.Run("выбран только брокерский — перевод это уход", func(t *testing.T) {
		in, out := DepositsAndWithdrawals(Filter(events, []string{"broker"}))
		if len(in) != 0 {
			t.Errorf("in = %+v, want empty", in)
		}
		if len(out) != 1 || !out[0].Value.Equal(decimal.RequireFromString("-150000")) {
			t.Errorf("out = %+v, want one point -150000", out)
		}
	})
}
