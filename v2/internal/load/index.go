package load

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/convert"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

// Index — значения индекса ticker по торговым дням с from по to.
func (l *Loader) Index(ctx context.Context, ticker string, from, to time.Time) (series.Series, error) {
	indicatives, err := l.client.GetIndicatives(ctx)
	if err != nil {
		return nil, fmt.Errorf("indicatives: %w", err)
	}
	var uid string
	for _, ind := range indicatives {
		if ind.Ticker == ticker {
			uid = ind.UID
		}
	}
	if uid == "" {
		return nil, fmt.Errorf("%s not found in indicatives", ticker)
	}

	candles, err := l.client.GetCandles(ctx, uid, from, to, tinvest.CandleIntervalDay)
	if err != nil {
		return nil, fmt.Errorf("%s candles: %w", ticker, err)
	}
	index, err := convert.ToPoints(candles.Candles, decimal.NewFromInt(1))
	if err != nil {
		return nil, fmt.Errorf("%s points: %w", ticker, err)
	}
	return index, nil
}
