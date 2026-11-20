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
		default:
			continue
		}

		day := series.Day(e.Date)
		if n := len(result); n > 0 && result[n-1].Date.Equal(day) {
			result[n-1].Value = result[n-1].Value.Add(e.BaseAmount.Amount)
		} else {
			result = append(result, series.Point{Date: day, Value: e.BaseAmount.Amount})
		}
	}

	nonZero := result[:0]
	for _, p := range result {
		if !p.Value.IsZero() {
			nonZero = append(nonZero, p)
		}
	}
	return nonZero
}
