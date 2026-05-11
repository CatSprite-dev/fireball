package service

import (
	"context"
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

type MockAPIClient struct {
	MockGetPortfolio          func(ctx context.Context, token string, accountID string) (api.UserPortfolio, error)
	MockGetOperationsByCursor func(
		ctx context.Context,
		token string,
		accountId string,
		instrumentId string,
		from *time.Time,
		to *time.Time,
		operationTypes []pkg.OperationType,
		operationState pkg.OperationState,
		WithoutCommissions bool) ([]api.UserOperations, error)
	MockBondBy          func(ctx context.Context, token string, idType pkg.InstrumentIdType, classCode pkg.ClassCode, id string) (api.InstrumentBond, error)
	MockGetInstrumentBy func(ctx context.Context, token string, idType pkg.InstrumentIdType, classCode pkg.ClassCode, id string) (api.InstrumentResponse, error)
	MockIndicatives     func(ctx context.Context, token string) (api.IndicativeInstruments, error)
	MockGetCandles      func(ctx context.Context, token string,
		from *time.Time,
		to *time.Time,
		interval pkg.CandleInterval,
		instrumentId string,
		candleSourceType pkg.CandleSource,
		limit int) (api.Candles, error)
}

func (m *MockAPIClient) GetPortfolio(ctx context.Context, token string, accountID string) (api.UserPortfolio, error) {
	if m.MockGetPortfolio != nil {
		return m.MockGetPortfolio(ctx, token, accountID)
	}
	return api.UserPortfolio{}, nil
}

func (m *MockAPIClient) GetOperationsByCursor(
	ctx context.Context,
	token string,
	accountId string,
	instrumentId string,
	from *time.Time,
	to *time.Time,
	operationTypes []pkg.OperationType,
	operationState pkg.OperationState,
	WithoutCommissions bool) ([]api.UserOperations, error) {
	if m.MockGetOperationsByCursor != nil {
		return m.MockGetOperationsByCursor(ctx, token, accountId, instrumentId, from, to, operationTypes, operationState, WithoutCommissions)
	}
	return nil, nil
}

func (m *MockAPIClient) Indicatives(ctx context.Context, token string) (api.IndicativeInstruments, error) {
	if m.MockIndicatives != nil {
		return m.MockIndicatives(ctx, token)
	}
	return api.IndicativeInstruments{}, nil
}

func (m *MockAPIClient) GetCandles(
	ctx context.Context,
	token string,
	from *time.Time,
	to *time.Time,
	interval pkg.CandleInterval,
	instrumentId string,
	candleSourceType pkg.CandleSource,
	limit int) (api.Candles, error) {
	if m.MockGetCandles != nil {
		return m.MockGetCandles(ctx, token, from, to, interval, instrumentId, candleSourceType, limit)
	}
	return api.Candles{}, nil
}

func (m *MockAPIClient) BondBy(ctx context.Context, token string, idType pkg.InstrumentIdType, classCode pkg.ClassCode, id string) (api.InstrumentBond, error) {
	if m.MockBondBy != nil {
		return m.MockBondBy(ctx, token, idType, classCode, id)
	}
	return api.InstrumentBond{}, nil
}

func (m *MockAPIClient) GetInstrumentBy(ctx context.Context, token string, idType pkg.InstrumentIdType, classCode pkg.ClassCode, id string) (api.InstrumentResponse, error) {
	if m.MockGetInstrumentBy != nil {
		return m.MockGetInstrumentBy(ctx, token, idType, classCode, id)
	}
	return api.InstrumentResponse{}, nil
}
