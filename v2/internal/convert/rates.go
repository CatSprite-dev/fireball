package convert

import (
	"fmt"

	"github.com/CatSprite-dev/fireball/v2/internal/fx"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

func ToPoints(candles []tinvest.Candle) ([]fx.Point, error) {
	points := make([]fx.Point, 0, len(candles))
	for _, c := range candles {
		rate, err := ToDecimal(c.Close.Units, c.Close.Nano)
		if err != nil {
			return nil, fmt.Errorf("candle %s: close: %w", c.Time.Format("2006-01-02"), err)
		}
		points = append(points, fx.Point{Date: c.Time.UTC(), Rate: rate})
	}
	return points, nil
}
