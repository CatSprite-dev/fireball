package convert

import (
	"errors"
	"fmt"

	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
	"github.com/shopspring/decimal"
)

var ErrZeroNominal = errors.New("convert: zero nominal")

func ToPoints(candles []tinvest.Candle, nominal decimal.Decimal) (series.Series, error) {
	if !nominal.IsPositive() {
		return nil, fmt.Errorf("%w: %s", ErrZeroNominal, nominal)
	}

	points := make(series.Series, 0, len(candles))
	for _, c := range candles {
		price, err := ToDecimal(c.Close.Units, c.Close.Nano)
		if err != nil {
			return nil, fmt.Errorf("candle %s: close: %w", c.Time.Format("2006-01-02"), err)
		}
		points = append(points, series.Point{Date: series.Day(c.Time), Value: price.Div(nominal)})
	}
	return points, nil
}
