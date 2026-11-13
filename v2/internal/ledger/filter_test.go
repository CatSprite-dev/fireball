package ledger

import (
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/shopspring/decimal"
)

func rub(s string) money.Money {
	return money.Money{Amount: decimal.RequireFromString(s), Currency: "RUB"}
}

func day(d int) time.Time {
	return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC)
}

// Два счёта и перевод между ними — как на живых данных: один id, одно время, суммы с разным знаком.
func testEvents() []Event {
	return []Event{
		{AccountID: "broker", OperationID: "1", Date: day(1), OperationKind: KindDeposit, Amount: rub("10000")},
		{AccountID: "robot", OperationID: "2", Date: day(2), OperationKind: KindDeposit, Amount: rub("5000")},
		{AccountID: "broker", OperationID: "t1", Date: day(6), OperationKind: KindTransferOut, Amount: rub("-2096")},
		{AccountID: "robot", OperationID: "t1", Date: day(6), OperationKind: KindTransferIn, Amount: rub("2096")},
	}
}

func TestFilterOneAccount(t *testing.T) {
	got := Filter(testEvents(), []string{"broker"})

	if len(got) != 2 {
		t.Fatalf("got %d events, want 2", len(got))
	}
	for _, e := range got {
		if e.AccountID != "broker" {
			t.Fatalf("got event from %q", e.AccountID)
		}
	}
}

func TestFilterKeepsOrderAndInput(t *testing.T) {
	events := testEvents()
	got := Filter(events, []string{"broker", "robot"})

	if len(got) != len(events) {
		t.Fatalf("got %d events, want %d", len(got), len(events))
	}
	for i := range events {
		if got[i].OperationID != events[i].OperationID || got[i].AccountID != events[i].AccountID {
			t.Fatalf("event %d: order changed", i)
		}
	}
}

func TestFilterNoAccounts(t *testing.T) {
	if got := Filter(testEvents(), nil); len(got) != 0 {
		t.Fatalf("got %d events, want 0", len(got))
	}
}

// Перевод внутри выбранных счетов в сумме даёт ноль, а для одного счёта остаётся выводом.
func TestTransfersNetOut(t *testing.T) {
	sumTransfers := func(events []Event) decimal.Decimal {
		total := decimal.Zero
		for _, e := range events {
			if e.OperationKind == KindTransferIn || e.OperationKind == KindTransferOut {
				total = total.Add(e.Amount.Amount)
			}
		}
		return total
	}

	both := Filter(testEvents(), []string{"broker", "robot"})
	if got := sumTransfers(both); !got.IsZero() {
		t.Fatalf("both accounts: transfers sum = %s, want 0", got)
	}

	onlyBroker := Filter(testEvents(), []string{"broker"})
	if got := sumTransfers(onlyBroker); !got.Equal(decimal.RequireFromString("-2096")) {
		t.Fatalf("one account: transfers sum = %s, want -2096", got)
	}
}

func TestSortByDateStable(t *testing.T) {
	events := []Event{
		{OperationID: "late", Date: day(5)},
		{OperationID: "trade", Date: day(3)},
		{OperationID: "fee", Date: day(3)},
		{OperationID: "early", Date: day(1)},
	}

	SortByDate(events)

	want := []string{"early", "trade", "fee", "late"}
	for i, id := range want {
		if events[i].OperationID != id {
			t.Fatalf("position %d: got %s, want %s", i, events[i].OperationID, id)
		}
	}
}
