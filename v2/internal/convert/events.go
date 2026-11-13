package convert

import (
	"errors"
	"fmt"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

var (
	ErrSkipped              = errors.New("convert: operation skipped")
	ErrUnknownOperationType = errors.New("convert: unknown operation type")
)

const stateExecuted = "OPERATION_STATE_EXECUTED"

// ToEvent переводит одну операцию брокера в событие журнала
func ToEvent(accountID string, op tinvest.Operation) (ledger.Event, error) {
	if op.State != stateExecuted {
		return ledger.Event{}, ErrSkipped
	}

	kind, ok := operationKinds[op.Type]
	if !ok {
		return ledger.Event{}, fmt.Errorf("operation %s: %w: %s", op.ID, ErrUnknownOperationType, op.Type)
	}
	instrumentKind := toInstrumentKind(op.InstrumentKind)
	if instrumentKind == ledger.InstrumentFutures || instrumentKind == ledger.InstrumentOption {
		kind = ledger.KindDerivatives
	}

	amount, err := ToMoney(op.Payment)
	if err != nil {
		return ledger.Event{}, fmt.Errorf("operation %s: payment: %w", op.ID, err)
	}

	accruedInterest, err := ToMoney(op.AccruedInt)
	if err != nil && !errors.Is(err, ErrNoValue) {
		return ledger.Event{}, fmt.Errorf("operation %s: accrued interest: %w", op.ID, err)
	}

	event := ledger.Event{
		AccountID:       accountID,
		OperationID:     op.ID,
		ParentID:        op.ParentOperationID,
		Date:            op.Date.UTC(),
		Kind:            kind,
		InstrumentUID:   op.InstrumentUID,
		InstrumentKind:  instrumentKind,
		Amount:          amount,
		AccruedInterest: accruedInterest,
	}

	// Количество нужно только сделкам. Берём исполненное, а не заявленное
	if kind == ledger.KindBuy || kind == ledger.KindSell {
		quantity, err := ToDecimal(op.QuantityDone, 0)
		if err != nil {
			return ledger.Event{}, fmt.Errorf("operation %s: quantity: %w", op.ID, err)
		}
		event.Quantity = quantity
	}

	return event, nil
}

func ToEvents(accountID string, ops []tinvest.Operation) ([]ledger.Event, error) {
	events := make([]ledger.Event, 0, len(ops))
	for _, op := range ops {
		event, err := ToEvent(accountID, op)
		if errors.Is(err, ErrSkipped) {
			continue
		}
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}
