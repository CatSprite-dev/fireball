package ledger

import (
	"github.com/CatSprite-dev/fireball/v2/internal/series"
)

func addToDay(s series.Series, e Event) series.Series {
	day := series.Day(e.Date)
	if n := len(s); n > 0 && s[n-1].Date.Equal(day) {
		s[n-1].Value = s[n-1].Value.Add(e.BaseAmount.Amount)
		return s
	}
	return append(s, series.Point{Date: day, Value: e.BaseAmount.Amount})
}

func dropZeroDays(s series.Series) series.Series {
	nonZero := s[:0]
	for _, p := range s {
		if !p.Value.IsZero() {
			nonZero = append(nonZero, p)
		}
	}
	return nonZero
}
