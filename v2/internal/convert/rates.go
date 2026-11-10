package convert

import (
	"errors"
	"fmt"

	"github.com/CatSprite-dev/fireball/v2/internal/fx"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
	"github.com/shopspring/decimal"
)

var ErrZeroNominal = errors.New("convert: zero nominal")

func ToPoints(candles []tinvest.Candle, nominal decimal.Decimal) ([]fx.Point, error) {
	if !nominal.IsPositive() {
		return nil, fmt.Errorf("%w: %s", ErrZeroNominal, nominal)
	}

	points := make([]fx.Point, 0, len(candles))
	for _, c := range candles {
		price, err := ToDecimal(c.Close.Units, c.Close.Nano)
		if err != nil {
			return nil, fmt.Errorf("candle %s: close: %w", c.Time.Format("2006-01-02"), err)
		}
		points = append(points, fx.Point{Date: c.Time.UTC(), Rate: price.Div(nominal)})
	}
	return points, nil
}
