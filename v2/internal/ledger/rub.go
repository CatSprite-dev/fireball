package ledger

import (
	"fmt"
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
)

type RUBConverter interface {
	ToRUB(m money.Money, date time.Time) (money.Money, error)
}

func ApplyRUB(events []Event, conv RUBConverter) error {
	for i := range events {
		amountRUB, err := conv.ToRUB(events[i].Amount, events[i].Date)
		if err != nil {
			return fmt.Errorf("operation %s: %w", events[i].OperationID, err)
		}
		events[i].AmountRUB = amountRUB
	}
	return nil
}
