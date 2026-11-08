package convert

import (
	"errors"
	"fmt"

	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

var ErrSkipped = errors.New("convert: operation skipped")
var ErrUnknownOperationType = errors.New("convert: unknown operation type")

const stateExecuted = "OPERATION_STATE_EXECUTED"

func ToEvent(accountID string, op tinvest.Operation) (ledger.Event, error) {
	event := ledger.Event{}
	if op.State != stateExecuted {
		return ledger.Event{}, ErrSkipped
	}
	kind, ok := operationKinds[op.Type]
	if !ok {
		return ledger.Event{}, fmt.Errorf("%w: %s", ErrUnknownOperationType, op.Type)
	}
	event.Kind = kind

	if op.InstrumentType == "futures" || op.InstrumentType == "option" {
		event.Kind = ledger.KindDerivatives
	}

	event.AccountID = accountID
	event.OperationID = op.ID
	event.ParentID = op.ParentOperationID
	event.Date = op.Date.UTC()
	event.InstrumentUID = op.InstrumentUID

	return event, nil
}
