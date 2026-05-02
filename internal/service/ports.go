package service

import (
	"context"
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

type APIClient interface {
	GetPortfolio(ctx context.Context, token string, accountID string) (api.UserPortfolio, error)
	GetOperationsByCursor(ctx context.Context,
		token string,
		accountId string,
		instrumentId string,
		from *time.Time,
		to *time.Time,
		operationTypes []pkg.OperationType,
		operationState pkg.OperationState,
		WithoutCommissions bool) ([]api.UserOperations, error)
	BondBy(ctx context.Context, token string, idType pkg.InstrumentIdType, classCode pkg.ClassCode, id string) (api.Bond, error)
	GetInstrumentBy(ctx context.Context, token string, idType pkg.InstrumentIdType, classCode pkg.ClassCode, id string) (api.Instrument, error)
	Indicatives(ctx context.Context, token string) (api.IndicativeInstruments, error)
	GetCandles(ctx context.Context, token string,
		from *time.Time,
		to *time.Time,
		interval pkg.CandleInterval,
		instrumentId string,
		candleSourceType pkg.CandleSource,
		limit int) (api.Candles, error)
}
