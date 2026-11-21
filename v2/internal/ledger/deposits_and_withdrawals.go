package ledger

import (
	"github.com/CatSprite-dev/fireball/v2/internal/series"
)

// DepositsAndWithdrawals — приходы и уходы денег по московским дням в базовой валюте, раздельно:
// in — пополнения и входящие переводы (плюс), out — выводы и исходящие переводы (минус).
// Перевод между своими счетами, если в events есть обе его стороны, — не приход и не уход:
// деньги остались внутри выбранных счетов.
func DepositsAndWithdrawals(events []Event) (in, out series.Series) {
	sorted := byDate(events)
	internal := internalTransfers(sorted)
	for i, e := range sorted {
		switch e.OperationKind {
		case KindDeposit, KindWithdrawal, KindTransferIn, KindTransferOut:
		default:
			continue
		}
		if internal[i] {
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

// internalTransfers — номера событий, которые складываются в пары переводов между своими счетами:
// исходящий и входящий перевод в один московский день, на одну сумму, с разных счетов.
// Каждое событие входит не больше чем в одну пару.
func internalTransfers(events []Event) map[int]bool {
	paired := make(map[int]bool)
	for i, out := range events {
		if out.OperationKind != KindTransferOut {
			continue
		}
		for j, in := range events {
			if paired[j] || in.OperationKind != KindTransferIn || in.AccountID == out.AccountID {
				continue
			}
			if series.Day(in.Date).Equal(series.Day(out.Date)) && in.BaseAmount.Amount.Equal(out.BaseAmount.Amount.Neg()) {
				paired[i], paired[j] = true, true
				break
			}
		}
	}
	return paired
}
