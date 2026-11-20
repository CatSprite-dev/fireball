package ledger

import (
	"github.com/CatSprite-dev/fireball/v2/internal/series"
)

// DepositsAndWithdrawals — приходы и уходы денег по московским дням в базовой валюте, раздельно:
// in — пополнения и входящие переводы (плюс), out — выводы и исходящие переводы (минус).
func DepositsAndWithdrawals(events []Event) (in, out series.Series) {
	for _, e := range byDate(events) {
		switch e.OperationKind {
		case KindDeposit, KindWithdrawal, KindTransferIn, KindTransferOut:
		default:
			continue
		}

		if e.BaseAmount.Amount.IsPositive() {
			in = addToDay(in, e)
		} else {
			out = addToDay(out, e)
		}
	}
	return dropZeroDays(in), dropZeroDays(out)
}
