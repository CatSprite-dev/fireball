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
		instrumentKind string
		wantOperation  ledger.OperationKind
		wantInstrument ledger.InstrumentKind
	}{
		{"покупка акции", "OPERATION_TYPE_BUY", "INSTRUMENT_TYPE_SHARE", ledger.KindBuy, ledger.InstrumentShare},
		{"пополнение", "OPERATION_TYPE_INPUT", "INSTRUMENT_TYPE_UNSPECIFIED", ledger.KindDeposit, ""},
		{"налог на дивиденды", "OPERATION_TYPE_DIVIDEND_TAX", "INSTRUMENT_TYPE_SHARE", ledger.KindTax, ledger.InstrumentShare},
		{"амортизация", "OPERATION_TYPE_BOND_REPAYMENT", "INSTRUMENT_TYPE_BOND", ledger.KindAmortization, ledger.InstrumentBond},
		{"покупка валюты", "OPERATION_TYPE_BUY", "INSTRUMENT_TYPE_CURRENCY", ledger.KindBuy, ledger.InstrumentCurrency},
		{"покупка фьючерса", "OPERATION_TYPE_BUY", "INSTRUMENT_TYPE_FUTURES", ledger.KindBuy, ledger.InstrumentFutures},
		{"комиссия по опциону", "OPERATION_TYPE_BROKER_FEE", "INSTRUMENT_TYPE_OPTION", ledger.KindFee, ledger.InstrumentOption},
		{"вариационная маржа", "OPERATION_TYPE_ACCRUING_VARMARGIN", "INSTRUMENT_TYPE_UNSPECIFIED", ledger.KindDerivatives, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := tinvest.Operation{
				State:          stateExecuted,
				Type:           tt.opType,
				InstrumentKind: tt.instrumentKind,
				Payment:        rubValue("100"),
				QuantityDone:   "1",
			}
			got, err := ToEvent("acc", op)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.OperationKind != tt.wantOperation {
				t.Errorf("operation kind = %q, want %q", got.OperationKind, tt.wantOperation)
			}
			if got.InstrumentKind != tt.wantInstrument {
				t.Errorf("instrument kind = %q, want %q", got.InstrumentKind, tt.wantInstrument)
			}
		})
	}
}

func TestToInstrumentKind(t *testing.T) {
	tests := []struct {
		in   string
		want ledger.InstrumentKind
	}{
		{"INSTRUMENT_TYPE_BOND", ledger.InstrumentBond},
		{"INSTRUMENT_TYPE_CURRENCY", ledger.InstrumentCurrency},
		{"INSTRUMENT_TYPE_FUTURES", ledger.InstrumentFutures},
		{"INSTRUMENT_TYPE_UNSPECIFIED", ""},
		{"", ""},
		{"INSTRUMENT_TYPE_SP", ledger.InstrumentOther},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := toInstrumentKind(tt.in); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToEventFuturesTradeHasNoMoney(t *testing.T) {
	// Покупка 5 фьючерсов Si-9.26 с живого счёта: payment −409 040 — это размер контракта.
	buy := tinvest.Operation{
		State:          stateExecuted,
		Type:           "OPERATION_TYPE_BUY",
		InstrumentKind: "INSTRUMENT_TYPE_FUTURES",
		Payment:        rubValue("-409040"),
		QuantityDone:   "5",
	}
	got, err := ToEvent("acc", buy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Amount.Amount.IsZero() || got.Amount.Currency != "RUB" {
		t.Errorf("amount = %s %s, want 0 RUB", got.Amount.Amount, got.Amount.Currency)
	}
	if got.Quantity.String() != "5" {
		t.Errorf("quantity = %s, want 5", got.Quantity)
	}

	// Вариационная маржа — настоящие деньги, её не трогаем.
	margin := tinvest.Operation{
		State:   stateExecuted,
		Type:    "OPERATION_TYPE_ACCRUING_VARMARGIN",
		Payment: rubValue("2885"),
	}
	got, err = ToEvent("acc", margin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Amount.Amount.String() != "2885" {
		t.Errorf("margin amount = %s, want 2885", got.Amount.Amount)
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

func TestToEvents(t *testing.T) {
	ops := []tinvest.Operation{
		{State: stateExecuted, Type: "OPERATION_TYPE_INPUT", ID: "1", Payment: rubValue("1000")},
		{State: "OPERATION_STATE_CANCELED", Type: "OPERATION_TYPE_BUY", ID: "2", Payment: rubValue("-500")},
		{State: stateExecuted, Type: "OPERATION_TYPE_COUPON", ID: "3", Payment: rubValue("50")},
	}

	events, err := ToEvents("acc", ops)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 || events[0].OperationID != "1" || events[1].OperationID != "3" {
		t.Fatalf("got %+v, want events 1 and 3", events)
	}
}

func TestToEventsStopsOnError(t *testing.T) {
	ops := []tinvest.Operation{
		{State: stateExecuted, Type: "OPERATION_TYPE_INPUT", ID: "1", Payment: rubValue("1000")},
		{State: stateExecuted, Type: "OPERATION_TYPE_SOMETHING_NEW", ID: "2", Payment: rubValue("1")},
	}

	events, err := ToEvents("acc", ops)
	if !errors.Is(err, ErrUnknownOperationType) {
		t.Fatalf("want ErrUnknownOperationType, got %v", err)
	}
	if events != nil {
		t.Fatalf("want nil events on error, got %d", len(events))
	}
}

func TestPositionKind(t *testing.T) {
	tests := []struct {
		in   string
		want ledger.InstrumentKind
	}{
		{"bond", ledger.InstrumentBond},
		{"currency", ledger.InstrumentCurrency},
		{"etf", ledger.InstrumentETF},
		{"", ""},
		{"sp", ledger.InstrumentOther},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := positionKind(tt.in); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
