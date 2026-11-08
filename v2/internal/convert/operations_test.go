package convert

import (
	"errors"
	"testing"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

func rubValue(units string) tinvest.MoneyValue {
	return tinvest.MoneyValue{Currency: "rub", Units: units}
}

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
				Payment:        rubValue("100"),
				QuantityDone:   "1",
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
			op := tinvest.Operation{State: state, Type: "OPERATION_TYPE_BUY", Payment: rubValue("100")}
			_, err := ToEvent("acc", op)
			if !errors.Is(err, ErrSkipped) {
				t.Fatalf("want ErrSkipped, got %v", err)
			}
		})
	}
}

func TestToEventUnknownType(t *testing.T) {
	op := tinvest.Operation{State: stateExecuted, Type: "OPERATION_TYPE_SOMETHING_NEW", Payment: rubValue("100")}
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
		Payment:           rubValue("1911"),
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

func TestToEventAmounts(t *testing.T) {
	// Продажа Эталон-Финанс 002Р-03 с живого счёта: payment уже включает НКД.
	sell := tinvest.Operation{
		State:        stateExecuted,
		Type:         "OPERATION_TYPE_SELL",
		Payment:      rubValue("100789"),
		AccruedInt:   rubValue("279"),
		QuantityDone: "100",
	}
	got, err := ToEvent("acc", sell)
	if err != nil {
		t.Fatalf("sell: unexpected error: %v", err)
	}
	if got.Amount.Amount.String() != "100789" || got.AccruedInterest.Amount.String() != "279" ||
		got.Quantity.String() != "100" {
		t.Fatalf("sell: got amount %s, accrued %s, quantity %s",
			got.Amount.Amount, got.AccruedInterest.Amount, got.Quantity)
	}
	if got.Amount.Currency != "RUB" {
		t.Fatalf("sell: currency = %q, want RUB", got.Amount.Currency)
	}

	// Комиссия за эту продажу: нет НКД и нет количества — это не ошибка.
	fee := tinvest.Operation{
		State:   stateExecuted,
		Type:    "OPERATION_TYPE_BROKER_FEE",
		Payment: tinvest.MoneyValue{Currency: "rub", Units: "-301", Nano: -530000000},
	}
	got, err = ToEvent("acc", fee)
	if err != nil {
		t.Fatalf("fee: unexpected error: %v", err)
	}
	if got.Amount.Amount.String() != "-301.53" || !got.Quantity.IsZero() {
		t.Fatalf("fee: got amount %s, quantity %s", got.Amount.Amount, got.Quantity)
	}
}

func TestToEventEmptyPayment(t *testing.T) {
	op := tinvest.Operation{State: stateExecuted, Type: "OPERATION_TYPE_INPUT"}
	_, err := ToEvent("acc", op)
	if !errors.Is(err, ErrNoValue) {
		t.Fatalf("want ErrNoValue for empty payment, got %v", err)
	}
}
