package convert

import (
	"errors"
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

func TestToEventKind(t *testing.T) {
	tests := []struct {
		name           string
		opType         string
		instrumentType string
		want           ledger.Kind
	}{
		{"покупка акции", "OPERATION_TYPE_BUY", "share", ledger.KindBuy},
		{"пополнение", "OPERATION_TYPE_INPUT", "", ledger.KindDeposit},
		{"налог на дивиденды", "OPERATION_TYPE_DIVIDEND_TAX", "share", ledger.KindTax},
		{"амортизация", "OPERATION_TYPE_BOND_REPAYMENT", "bond", ledger.KindAmortization},
		{"покупка фьючерса", "OPERATION_TYPE_BUY", "futures", ledger.KindDerivatives},
		{"комиссия по опциону", "OPERATION_TYPE_BROKER_FEE", "option", ledger.KindDerivatives},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := tinvest.Operation{
				State:          stateExecuted,
				Type:           tt.opType,
				InstrumentType: tt.instrumentType,
			}
			got, err := ToEvent("acc", op)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Kind != tt.want {
				t.Fatalf("kind = %q, want %q", got.Kind, tt.want)
			}
		})
	}
}

func TestToEventSkipsNotExecuted(t *testing.T) {
	for _, state := range []string{"OPERATION_STATE_CANCELED", "OPERATION_STATE_PROGRESS"} {
		t.Run(state, func(t *testing.T) {
			op := tinvest.Operation{State: state, Type: "OPERATION_TYPE_BUY"}
			_, err := ToEvent("acc", op)
			if !errors.Is(err, ErrSkipped) {
				t.Fatalf("want ErrSkipped, got %v", err)
			}
		})
	}
}

func TestToEventUnknownType(t *testing.T) {
	op := tinvest.Operation{State: stateExecuted, Type: "OPERATION_TYPE_SOMETHING_NEW"}
	_, err := ToEvent("acc", op)
	if !errors.Is(err, ErrUnknownOperationType) {
		t.Fatalf("want ErrUnknownOperationType, got %v", err)
	}
}

func TestToEventFields(t *testing.T) {
	moscow := time.FixedZone("MSK", 3*60*60)
	op := tinvest.Operation{
		State:             stateExecuted,
		Type:              "OPERATION_TYPE_COUPON",
		ID:                "op-1",
		ParentOperationID: "parent-1",
		InstrumentUID:     "uid-1",
		Date:              time.Date(2026, 10, 6, 9, 20, 0, 0, moscow),
	}

	got, err := ToEvent("acc-1", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AccountID != "acc-1" || got.OperationID != "op-1" ||
		got.ParentID != "parent-1" || got.InstrumentUID != "uid-1" {
		t.Fatalf("fields copied wrong: %+v", got)
	}

	wantDate := time.Date(2026, 10, 6, 6, 20, 0, 0, time.UTC)
	if !got.Date.Equal(wantDate) || got.Date.Location() != time.UTC {
		t.Fatalf("date = %v, want %v in UTC", got.Date, wantDate)
	}
}
