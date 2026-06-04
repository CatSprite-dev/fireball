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

func (d *DemoClient) GetPortfolio(_ context.Context, _ string, _ string) (api.UserPortfolio, error) {
	return demoPortfolio(), nil
}

func (d *DemoClient) GetOperationsByCursor(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ *time.Time,
	_ *time.Time,
	operationTypes []pkg.OperationType,
	_ pkg.OperationState,
	_ bool,
) ([]api.UserOperations, error) {
	all := demoOperations()
	if len(operationTypes) == 0 {
		return all, nil
	}
	typeSet := make(map[string]struct{}, len(operationTypes))
	for _, t := range operationTypes {
		typeSet[string(t)] = struct{}{}
	}
	filtered := []api.Item{}
	for _, item := range all[0].Items {
		if _, ok := typeSet[item.Type]; ok {
			filtered = append(filtered, item)
		}
	}
	return []api.UserOperations{{HasNext: false, Items: filtered}}, nil
}

func (d *DemoClient) BondBy(
	ctx context.Context,
	_ string,
	idType pkg.InstrumentIdType,
	classCode pkg.ClassCode,
	id string,
) (api.InstrumentBond, error) {
	return d.real.BondBy(ctx, d.token, idType, classCode, id)
}

func (d *DemoClient) GetInstrumentBy(
	ctx context.Context,
	_ string,
	idType pkg.InstrumentIdType,
	classCode pkg.ClassCode,
	id string,
) (api.InstrumentResponse, error) {
	return d.real.GetInstrumentBy(ctx, d.token, idType, classCode, id)
}

func (d *DemoClient) Indicatives(ctx context.Context, _ string) (api.IndicativeInstruments, error) {
	return d.real.Indicatives(ctx, d.token)
}

func (d *DemoClient) GetCandles(
	ctx context.Context,
	_ string,
	from *time.Time,
	to *time.Time,
	interval pkg.CandleInterval,
	instrumentId string,
	candleSourceType pkg.CandleSource,
	limit int,
) (api.Candles, error) {
	return d.real.GetCandles(ctx, d.token, from, to, interval, instrumentId, candleSourceType, limit)
}

func (d *DemoClient) GetClosePrices(
	ctx context.Context,
	_ string,
	instrumentIDs []string,
	instrumentStatus pkg.InstrumentStatus,
) (api.ClosePrices, error) {
	return d.real.GetClosePrices(ctx, d.token, instrumentIDs, instrumentStatus)
}

func (d *DemoClient) Currencies(
	ctx context.Context,
	_ string,
	instrumentStatus pkg.InstrumentStatus,
	instrumentExhange pkg.InstrumentExchange,
) (api.Currencies, error) {
	return d.real.Currencies(ctx, d.token, instrumentStatus, instrumentExhange)
}
