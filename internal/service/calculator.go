package service

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

var ErrNotFound = errors.New("not found")

type Calculator struct {
	ApiClient *api.Client
}

func NewCalculator(apiClient *api.Client) *Calculator {
	return &Calculator{ApiClient: apiClient}
}

func (calc *Calculator) GetFullPortfolio(token string) (domain.UserFullPortfolio, error) {
	t := time.Now()

	userAccounts, err := calc.ApiClient.GetAccounts(token, pkg.AccountStatusOpen)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}
	if len(userAccounts.Accounts) == 0 {
		return domain.UserFullPortfolio{}, errors.New("found no accounts")
	}

	accountID := userAccounts.Accounts[0].ID
	openedDate := userAccounts.Accounts[0].OpenedDate

	rawPortfolio, err := calc.ApiClient.GetPortfolio(token, accountID)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}
	fullEmptyPortfolio := convertFullPortfolio(rawPortfolio)
	enrichedFullPortfolio, err := enrichFullPortfolio(calc, fullEmptyPortfolio, token, accountID, openedDate)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}

	log.Printf("Время выполнения GetFullPortfolio: %.2f сек\n", time.Since(t).Seconds())
	_, err = calc.GetTotalReturn(token, enrichedFullPortfolio, accountID, openedDate)
	return enrichedFullPortfolio, nil
}

func (calc *Calculator) GetDividends(
	token string,
	accountID string,
	instrumentId string,
	from time.Time,
	to time.Time) (map[string]domain.MoneyValue, error) {

	operations, err := calc.ApiClient.GetUserOperationsByCursor(
		token,
		accountID,
		instrumentId,
		&from,
		&to,
		[]pkg.OperationType{pkg.OperationTypeDividend, pkg.OperationTypeCoupon},
		pkg.OperationStateExecuted,
		false,
	)
	if err != nil {
		return nil, err
	}

	result := make(map[string]domain.MoneyValue)
	for _, block := range operations {
		for _, item := range block.Items {
			key := item.Ticker
			if key == "" {
				continue
			}
			current := result[key]
			result[key] = AddMoneyValue(current, domain.MoneyValue(item.Payment))
		}
	}
	return result, nil
}

func (calc *Calculator) GetInstrumentInfo(token string, instrumentIdType pkg.InstrumentIdType, instrumentId string) (domain.Instrument, error) {
	rawInstrument, err := calc.ApiClient.GetInstrumentBy(token, instrumentIdType, pkg.ClassCodeUnspecified, instrumentId)
	if err != nil {
		var requestErr api.RequestError
		if errors.As(err, &requestErr) && requestErr.StatusCode == http.StatusNotFound {
			return domain.Instrument{}, ErrNotFound
		}
		return domain.Instrument{}, err
	}
	instrument := convertInstrument(rawInstrument)
	return instrument, nil
}

func (calc *Calculator) GetIndexByTicker(token string, ticker string) (domain.Instrument, error) {
	result := domain.Instrument{}

	rawInstruments, err := calc.ApiClient.Indicatives(token)
	if err != nil {
		return domain.Instrument{}, err
	}
	indicativeInstruments := convertIndicativeInstrument(rawInstruments)
	for _, instr := range indicativeInstruments.Instruments {
		if instr.Ticker == ticker {
			result = instr
			break
		}
	}
	return result, nil
}

func (calc *Calculator) GetCandles(token string,
	instrumentId string,
	from time.Time,
	to time.Time,
	interval pkg.CandleInterval,
	candleSourceType pkg.CandleSource) (domain.Candles, error) {

	rawCandles, err := calc.ApiClient.GetCandles(token, &from, &to, interval, instrumentId, candleSourceType, 0)
	if err != nil {
		return domain.Candles{}, err
	}

	candles := convertCandles(rawCandles)
	return candles, nil
}

func (calc *Calculator) GetChartData(token string, indexTicker string, from time.Time, to time.Time, candleInterval pkg.CandleInterval) (domain.ChartData, error) {
	index, err := calc.GetIndexByTicker(token, indexTicker)
	if err != nil {
		return domain.ChartData{}, err
	}

	// portfolioCandles, err := calc.GetCandlesForPortfolio(token)

	indexCandles, err := calc.GetCandles(token, index.UID, from, to, candleInterval, pkg.CandleSourceExchange)
	if err != nil {
		return domain.ChartData{}, err
	}
	if len(indexCandles.Candles) == 0 {
		return domain.ChartData{}, errors.New("no candles data available")
	}

	return domain.ChartData{
		IndexCandles: indexCandles,
	}, nil
}

func (calc *Calculator) GetTotalReturn(token string, portfolio domain.UserFullPortfolio, accountID string, openedDate time.Time) (domain.MoneyValue, error) {
	now := time.Now()
	operations, err := calc.ApiClient.GetUserOperationsByCursor(
		token,
		accountID,
		"",
		&openedDate,
		&now,
		[]pkg.OperationType{
			pkg.OperationTypeInput,
			pkg.OperationTypeOutput,
		},
		pkg.OperationStateExecuted,
		true,
	)
	if err != nil {
		return domain.MoneyValue{}, err
	}

	netCashFlow := domain.MoneyValue{}
	for _, block := range operations {
		for _, item := range block.Items {
			netCashFlow = AddMoneyValue(netCashFlow, domain.MoneyValue(item.Payment))
		}
	}

	totalReturn := SubtractMoneyValue(portfolio.TotalAmountPortfolio, netCashFlow)

	return totalReturn, nil
}

func (calc *Calculator) GetCandlesForPortfolio(token string, portfolio domain.UserFullPortfolio, from time.Time, to time.Time, candleInterval pkg.CandleInterval) (domain.Candles, error) {
	// Взять текущую сумму портфеля
	// Взять историю операций по портфелю
	// Взять историю цен по каждому инструменту в портфеле

	// От сегодняшнего дня шагнуть назад на интервал свечи
	// Посмотреть были ли продажи и покупки на этом интервале
	// Если были продажи то прибавить количество проданных бумаг к текущему количеству в портфеле
	// Если были покупки то отнять количество купленных бумаг от текущего количества в портфеле
	// Умножаем получившееся количество бумаг на цену закрытия свечи и получаем стоиомсть портфеля в этот интервал

	wg := sync.WaitGroup{}

	_, err := calc.ApiClient.GetUserOperationsByCursor(
		token,
		portfolio.AccountID,
		"",
		&from,
		&to,
		[]pkg.OperationType{
			pkg.OperationTypeBuy,
			pkg.OperationTypeSell,
		},
		pkg.OperationStateExecuted,
		true,
	)
	if err != nil {
		return domain.Candles{}, err
	}

	type candleResult struct {
		instrumentID string
		candles      domain.Candles
		err          error
	}

	resultCh := make(chan candleResult, len(portfolio.Positions))

	candlesOfPositions := make(map[string]domain.Candles)
	for _, pos := range portfolio.Positions {
		wg.Add(1)
		go func(p domain.Position) {
			candles, err := calc.GetCandles(token, p.InstrumentUID, from, to, candleInterval, pkg.CandleSourceExchange)
			resultCh <- candleResult{
				instrumentID: p.InstrumentUID,
				candles:      candles,
				err:          err,
			}
		}(pos)
		wg.Done()
	}

	for i := 0; i < len(portfolio.Positions); i++ {
		res := <-resultCh
		if res.err != nil {
			log.Printf("Ошибка при получении свечей для инструмента %s: %v\n", res.instrumentID, res.err)
			continue
		}
		candlesOfPositions[res.instrumentID] = res.candles
	}

	close(resultCh)

	wg.Wait()
	return domain.Candles{}, nil
}
