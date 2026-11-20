package ledger

import (
	"github.com/CatSprite-dev/fireball/v2/internal/series"
)

// NetDeposits — чистые пополнения по московским дням в базовой валюте:
// пополнения и входящие переводы с плюсом, выводы и исходящие переводы с минусом.
// Дни, где движения взаимно погасились, в ряд не входят.
func NetDeposits(events []Event) series.Series {
	var result series.Series
	for _, e := range byDate(events) {
		switch e.OperationKind {
		case KindDeposit, KindWithdrawal, KindTransferIn, KindTransferOut:
			result = addToDay(result, e)
		}
	}
	return dropZeroDays(result)
}
