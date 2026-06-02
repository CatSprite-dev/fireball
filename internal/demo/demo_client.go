package demo

import (
	"context"
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/service"
)

type DemoClient struct {
	real  service.APIClient
	token string
}

func NewDemoClient(real service.APIClient, token string) *DemoClient {
	return &DemoClient{real: real, token: token}
}

func (d *DemoClient) GetPortfolio(ctx context.Context, token string, accountID string) (api.UserPortfolio, error) {

	return api.UserPortfolio{}, nil
}

func (d *DemoClient) GetOperationsByCursor(
	ctx context.Context,
	token string,
	accountId string,
	instrumentId string,
	from *time.Time,
	to *time.Time,
	operationTypes []pkg.OperationType,
	operationState pkg.OperationState,
	WithoutCommissions bool,
) ([]api.UserOperations, error) {

	return nil, nil
}

func (d *DemoClient) BondBy(
	ctx context.Context,
	token string,
	idType pkg.InstrumentIdType,
	classCode pkg.ClassCode,
	id string,
) (api.InstrumentBond, error) {
	return d.real.BondBy(ctx, token, idType, classCode, id)
}

func (d *DemoClient) GetInstrumentBy(
	ctx context.Context,
	token string,
	idType pkg.InstrumentIdType,
	classCode pkg.ClassCode,
	id string,
) (api.InstrumentResponse, error) {
	return d.real.GetInstrumentBy(ctx, token, idType, classCode, id)
}

func (d *DemoClient) Indicatives(ctx context.Context, token string) (api.IndicativeInstruments, error) {
	return d.real.Indicatives(ctx, token)
}

func (d *DemoClient) GetCandles(
	ctx context.Context,
	token string,
	from *time.Time,
	to *time.Time,
	interval pkg.CandleInterval,
	instrumentId string,
	candleSourceType pkg.CandleSource,
	limit int,
) (api.Candles, error) {
	return d.real.GetCandles(ctx, token, from, to, interval, instrumentId, candleSourceType, limit)
}

func (d *DemoClient) GetClosePrices(
	ctx context.Context,
	token string,
	instrumentID string,
	instrumentStatus pkg.InstrumentStatus,
) (api.ClosePrices, error) {
	return d.real.GetClosePrices(ctx, token, instrumentID, instrumentStatus)
}

func (d *DemoClient) Currencies(
	ctx context.Context,
	token string,
	instrumentStatus pkg.InstrumentStatus,
	instrumentExhange pkg.InstrumentExchange,
) (api.Currencies, error) {
	return d.real.Currencies(ctx, token, instrumentStatus, instrumentExhange)
}
