package convert

import (
	"errors"
	"fmt"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
	"github.com/shopspring/decimal"
)

var (
	ErrSkipped              = errors.New("convert: operation skipped")
	ErrUnknownOperationType = errors.New("convert: unknown operation type")
)

const stateExecuted = "OPERATION_STATE_EXECUTED"

// ToEvent переводит одну операцию брокера в событие журнала.
func ToEvent(accountID string, op tinvest.Operation) (ledger.Event, error) {
	if op.State != stateExecuted {
		return ledger.Event{}, ErrSkipped
	}

	operationKind, ok := operationKinds[op.Type]
	if !ok {
		return ledger.Event{}, fmt.Errorf("operation %s: %w: %s", op.ID, ErrUnknownOperationType, op.Type)
	}
	instrumentKind := toInstrumentKind(op.InstrumentKind)

	amount, err := ToMoney(op.Payment)
	if err != nil {
		return ledger.Event{}, fmt.Errorf("operation %s: payment: %w", op.ID, err)
	}
	if isFuturesTrade(operationKind, instrumentKind) {
		// У сделки с фьючерсом payment — размер контракта, а не деньги.
		// Деньги по фьючерсам двигает только вариационная маржа.
		amount = money.Money{Amount: decimal.Zero, Currency: amount.Currency}
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
		OperationKind:   operationKind,
		InstrumentUID:   op.InstrumentUID,
		AssetUID:        op.AssetUID,
		InstrumentKind:  instrumentKind,
		Amount:          amount,
		AccruedInterest: accruedInterest,
	}

	if isTrade(operationKind) {
		quantity, err := ToDecimal(op.QuantityDone, 0)
		if err != nil {
			return ledger.Event{}, fmt.Errorf("operation %s: quantity: %w", op.ID, err)
		}
		event.Quantity = quantity
	}

	return event, nil
}

// ToEvents переводит операции одного счёта, пропуская неисполненные.
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
		if !isDerivativeCash(event.OperationKind) {
			events = append(events, event)
			continue
		}
		split, err := splitByContracts(event, op)
		if err != nil {
			return nil, err
		}
		events = append(events, split...)
	}
	return events, nil
}

func isDerivativeCash(kind ledger.OperationKind) bool {
	return kind == ledger.KindVariationMargin || kind == ledger.KindExpiration
}

func splitByContracts(event ledger.Event, op tinvest.Operation) ([]ledger.Event, error) {
	if len(op.ChildOperations) == 0 {
		return []ledger.Event{event}, nil
	}

	result := make([]ledger.Event, 0, len(op.ChildOperations))
	sum := decimal.Zero
	for i, ch := range op.ChildOperations {
		amount, err := ToMoney(ch.Payment)
		if err != nil {
			return nil, fmt.Errorf("operation %s: child %d: %w", op.ID, i, err)
		}
		sum = sum.Add(amount.Amount)

		child := event
		child.OperationID = fmt.Sprintf("%s:%d", op.ID, i)
		child.ParentID = op.ID
		child.InstrumentUID = ch.InstrumentUID
		child.Amount = amount
		result = append(result, child)
	}

	if !sum.Equal(event.Amount.Amount) {
		return []ledger.Event{event}, nil
	}
	return result, nil
}

func isTrade(kind ledger.OperationKind) bool {
	return kind == ledger.KindBuy || kind == ledger.KindSell
}

func isFuturesTrade(kind ledger.OperationKind, instrument ledger.InstrumentKind) bool {
	return isTrade(kind) && instrument == ledger.InstrumentFutures
}
