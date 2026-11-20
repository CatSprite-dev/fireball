package load

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/CatSprite-dev/fireball/v2/internal/convert"
	"github.com/CatSprite-dev/fireball/v2/internal/fx"
	"github.com/CatSprite-dev/fireball/v2/internal/ledger"
	"github.com/CatSprite-dev/fireball/v2/internal/series"
	"github.com/CatSprite-dev/fireball/v2/internal/snapshot"
	"github.com/CatSprite-dev/fireball/v2/internal/tinvest"
)

// Market — цены и номиналы с НКД по ключам позиций.
type Market struct {
	Prices map[string]series.Series
	Bonds  map[string]snapshot.Bond
}

// Market — цены в базовой валюте и номиналы с НКД облигаций с from по to на все позиции портфеля.
func (l *Loader) Market(ctx context.Context, p *Portfolio, from, to time.Time) (*Market, error) {
	positions := ledger.ByInstrument(p.Events, p.Currencies.ByAsset)
	uidsByKey := uidsByKey(p.Events, p.Currencies.ByAsset)
	tradePrices := ledger.TradePrices(p.Events, p.Currencies.ByAsset)

	prices, err := l.loadPrices(ctx, positions, uidsByKey, tradePrices, p.Rates, from, to)
	if err != nil {
		return nil, err
	}
	bonds, err := l.loadBonds(ctx, positions, uidsByKey, from, to)
	if err != nil {
		return nil, err
	}
	return &Market{Prices: prices, Bonds: bonds}, nil
}

// uidsByKey — инструменты каждой позиции: сначала тот, по которому была последняя операция,
// при равных датах — по uid. Так series.Merge первой берёт свечу инструмента, которым пользуются сейчас.
func uidsByKey(events []ledger.Event, currencyByAsset map[string]string) map[string][]string {
	lastUsed := make(map[string]time.Time)
	for _, e := range events {
		if e.InstrumentUID != "" && e.Date.After(lastUsed[e.InstrumentUID]) {
			lastUsed[e.InstrumentUID] = e.Date
		}
	}

	result := make(map[string][]string)
	for uid, key := range ledger.PageKeys(events, currencyByAsset) {
		result[key] = append(result[key], uid)
	}
	for _, uids := range result {
		slices.SortFunc(uids, func(a, b string) int {
			if c := lastUsed[b].Compare(lastUsed[a]); c != 0 {
				return c
			}
			return strings.Compare(a, b)
		})
	}
	return result
}

func (l *Loader) loadPrices(ctx context.Context, positions map[string]*ledger.Flows, uidsByKey map[string][]string, tradePrices map[string]series.Series, rates *fx.Rates, from, to time.Time) (map[string]series.Series, error) {
	prices := make(map[string]series.Series)
	for key, f := range positions {
		if f.Kind == ledger.InstrumentCurrency {
			if s, ok := rates.Series(key); ok {
				prices[key] = s
			}
			continue
		}
		var parts []series.Series
		for _, uid := range uidsByKey[key] {
			candles, err := l.client.GetCandles(ctx, uid, from, to, tinvest.CandleIntervalDay)
			if err != nil {
				continue
			}
			points, err := convert.ToPoints(candles.Candles, decimal.NewFromInt(1))
			if err != nil {
				return nil, fmt.Errorf("prices %s: %w", uid, err)
			}
			if f.Kind != ledger.InstrumentBond {
				inst, err := l.client.GetInstrumentBy(ctx, tinvest.InstrumentIDUID, "", uid)
				if err != nil {
					return nil, fmt.Errorf("instrument %s: %w", uid, err)
				}
				points, err = rates.Convert(points, strings.ToUpper(inst.Currency))
				if err != nil {
					return nil, fmt.Errorf("prices %s: %w", uid, err)
				}
			}
			parts = append(parts, points)
		}
		prices[key] = series.Merge(append(parts, tradePrices[key])...)
	}
	return prices, nil
}

func (l *Loader) loadBonds(ctx context.Context, positions map[string]*ledger.Flows, uidsByKey map[string][]string, from, to time.Time) (map[string]snapshot.Bond, error) {
	bonds := make(map[string]snapshot.Bond)
	for key, f := range positions {
		if f.Kind != ledger.InstrumentBond {
			continue
		}
		var nominals, nkds []series.Series
		for _, uid := range uidsByKey[key] {
			list, err := l.client.GetAccruedInterests(ctx, uid, from, to)
			if err != nil {
				continue
			}
			nominal, nkd, err := convert.ToAccrued(list)
			if err != nil {
				return nil, fmt.Errorf("accrued %s: %w", uid, err)
			}
			nominals, nkds = append(nominals, nominal), append(nkds, nkd)
		}
		bonds[key] = snapshot.Bond{Nominal: series.Merge(nominals...), NKD: series.Merge(nkds...)}
	}
	return bonds, nil
}
