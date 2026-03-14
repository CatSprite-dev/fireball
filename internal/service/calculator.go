package service

import (
	"errors"
	"log"
	"net/http"
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

func (calc *Calculator) GetInstrumentInfo(token string, instrumentIdType pkg.InstrumentIdType, classCode pkg.ClassCode, instrumentId string) (domain.Instrument, error) {
	if classCode == "" {
		classCode = pkg.ClassCodeUnspecified
	}
	rawInstrument, err := calc.ApiClient.GetInstrumentBy(token, instrumentIdType, classCode, instrumentId)
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
	candleSourceType pkg.CandleSource) ([]domain.Candle, error) {

	rawCandles, err := calc.ApiClient.GetCandles(token, &from, &to, interval, instrumentId, candleSourceType, 0)
	if err != nil {
		return []domain.Candle{}, err
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
	if len(indexCandles) == 0 {
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
			pkg.OperationTypeInpMulti,
			pkg.OperationTypeOutMulti,
		},
		pkg.OperationStateExecuted,
		true,
	)
	if err != nil {
		return domain.MoneyValue{}, err
	}

	sumIn := domain.MoneyValue{}
	sumOut := domain.MoneyValue{}
	transIn := domain.MoneyValue{}
	transOut := domain.MoneyValue{}
	netCashFlow := domain.MoneyValue{}
	for _, block := range operations {
		for _, item := range block.Items {
			netCashFlow = AddMoneyValue(netCashFlow, domain.MoneyValue(item.Payment))
			switch item.Type {
			case string(pkg.OperationTypeInput):
				sumIn = AddMoneyValue(sumIn, domain.MoneyValue(item.Payment))
			case string(pkg.OperationTypeOutput):
				sumOut = AddMoneyValue(sumOut, domain.MoneyValue(item.Payment))
			case string(pkg.OperationTypeInpMulti):
				transIn = AddMoneyValue(transIn, domain.MoneyValue(item.Payment))
			case string(pkg.OperationTypeOutMulti):
				transOut = AddMoneyValue(transOut, domain.MoneyValue(item.Payment))
			}
		}
	}

	totalReturn := SubtractMoneyValue(portfolio.TotalAmountPortfolio, netCashFlow)

	log.Printf("Input: %v\n", sumIn.Units)
	log.Printf("Output: %v\n", sumOut.Units)
	log.Printf("TransInput: %v\n", transIn.Units)
	log.Printf("TransOutput: %v\n", transOut.Units)
	log.Printf("CashFlow: %v\n", netCashFlow.Units)
	log.Printf("TotalReturn: %v", totalReturn.Units)

	return totalReturn, nil
}

func (calc *Calculator) CalculateHistoricalHoldings(token string, portfolio domain.UserFullPortfolio, from time.Time, to time.Time, candleInterval pkg.CandleInterval) (map[time.Time]map[string]domain.Quotation, error) {
	// Взять текущую сумму портфеля
	// Взять историю операций по портфелю
	// Взять историю цен по каждому инструменту в портфеле

	// От сегодняшнего дня шагнуть назад на интервал свечи
	// Посмотреть были ли продажи и покупки на этом интервале
	// Если были продажи то прибавить количество проданных бумаг к текущему количеству в портфеле
	// Если были покупки то отнять количество купленных бумаг от текущего количества в портфеле
	// Умножаем получившееся количество бумаг на цену закрытия свечи и получаем стоиомсть портфеля в этот интервал

	operations, err := calc.ApiClient.GetUserOperationsByCursor(
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
		return map[time.Time]map[string]domain.Quotation{}, err
	}

	candlesOfPositions, err := calc.FetchCandlesForPositions(token, portfolio, from, to, candleInterval)
	if err != nil {
		return map[time.Time]map[string]domain.Quotation{}, err
	}

	start := candlesOfPositions[portfolio.Positions[0].InstrumentUID][len(candlesOfPositions[portfolio.Positions[0].InstrumentUID])-1].Time
	end := candlesOfPositions[portfolio.Positions[0].InstrumentUID][0].Time

	// Инициализируем начальное состояние (сегодня)
	positionsQuantity := make(map[time.Time]map[string]domain.Quotation)
	positionsQuantity[start] = make(map[string]domain.Quotation)
	for _, pos := range portfolio.Positions {
		if _, exists := candlesOfPositions[pos.Figi]; !exists {
			// Догружаем свечи если в истории нет
			candles, err := calc.GetCandles(token, pos.Figi, from, to, candleInterval, pkg.CandleSourceExchange)
			if err != nil {
				log.Printf("Ошибка загрузки свечей для %s: %v", pos.Ticker, err)
				continue
			}
			candlesOfPositions[pos.InstrumentUID] = candles
		}
	}

	currentDate := start
	for currentDate.After(end) || currentDate.Equal(end) {
		yesterday := currentDate.AddDate(0, 0, -1)

		// Копируем состояние из текущего дня во вчера
		positionsQuantity[yesterday] = make(map[string]domain.Quotation)
		for ticker, qty := range positionsQuantity[currentDate] {
			positionsQuantity[yesterday][ticker] = qty
		}

		// Обрабатываем операции этого дня
		for _, block := range operations {
			for _, item := range block.Items {
				opDate := item.Date.Truncate(24 * time.Hour)

				if opDate.Equal(currentDate) {
					switch item.Type {
					case string(pkg.OperationTypeBuy):
						// Купили сегодня → вчера было меньше
						current := positionsQuantity[yesterday][item.Ticker]
						positionsQuantity[yesterday][item.Ticker] = SubtractQuotations(current, domain.Quotation{Units: item.Quantity})

					case string(pkg.OperationTypeSell):
						// Продали сегодня → вчера было больше
						current := positionsQuantity[yesterday][item.Ticker]
						positionsQuantity[yesterday][item.Ticker] = AddQuotations(current, domain.Quotation{Units: item.Quantity})
					}
				}
			}
		}

		currentDate = yesterday
	}

	return positionsQuantity, nil
}

func (calc *Calculator) FetchCandlesForPositions(
	token string,
	portfolio domain.UserFullPortfolio,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval) (map[string][]domain.Candle, error) {

	type candleResult struct {
		figi    string
		candles []domain.Candle
		err     error
	}

	resultCh := make(chan candleResult, len(portfolio.Positions))

	for _, pos := range portfolio.Positions {
		go func(p domain.Position) {
			candles, err := calc.GetCandles(token, p.Figi, from, to, candleInterval, pkg.CandleSourceExchange)
			resultCh <- candleResult{
				figi:    p.Figi,
				candles: candles,
				err:     err,
			}
		}(pos)
	}

	candlesOfPositions := make(map[string][]domain.Candle)
	for i := 0; i < len(portfolio.Positions); i++ {
		res := <-resultCh
		if res.err != nil {
			log.Printf("Ошибка при получении свечей для инструмента %s: %v\n", res.figi, res.err)
			continue
		}
		candlesOfPositions[res.figi] = res.candles
	}

	return candlesOfPositions, nil
}
