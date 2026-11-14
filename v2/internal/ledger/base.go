package ledger

import (
	"fmt"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
)

type BaseConverter interface {
	ToBase(m money.Money, date time.Time) (money.Money, error)
}

func ApplyBase(events []Event, conv BaseConverter) error {
	for i := range events {
		amountBase, err := conv.ToBase(events[i].Amount, events[i].Date)
		if err != nil {
			return fmt.Errorf("operation %s: %w", events[i].OperationID, err)
		}
		events[i].BaseAmount = amountBase
	}
	return nil
}
