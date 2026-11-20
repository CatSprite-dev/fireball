package metrics

import (
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/shopspring/decimal"
)

// valueAt — стоимость на конец дня d; до первой точки — 0.
func valueAt(values series.Series, d time.Time) decimal.Decimal {
	p, ok := values.At(d)
	if !ok {
		return decimal.Zero
	}
	return p.Value
}

// valueOn — значение ровно в день d; 0, если точки этого дня нет.
func valueOn(s series.Series, d time.Time) decimal.Decimal {
	p, ok := s.At(d)
	if !ok || !p.Date.Equal(d) {
		return decimal.Zero
	}
	return p.Value
}
