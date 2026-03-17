package service

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
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

	accounts, err := calc.ApiClient.GetAccounts(token, pkg.AccountStatusOpen)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}
	if len(accounts.Accounts) == 0 {
		return domain.UserFullPortfolio{}, errors.New("found no accounts")
	}

	accountID := accounts.Accounts[0].ID
	openedDate := accounts.Accounts[0].OpenedDate

	rawPortfolio, err := calc.ApiClient.GetPortfolio(token, accountID)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}
	portfolio := convertFullPortfolio(rawPortfolio)
	portfolio, err = enrichFullPortfolio(calc, portfolio, token, accountID, openedDate)

	if err != nil {
		return domain.UserFullPortfolio{}, err
	}

	log.Printf("Время выполнения GetFullPortfolio: %.2f сек\n", time.Since(t).Seconds())
	return portfolio, nil
}

func (calc *Calculator) GetOperations(token string,
	accountId string,
	instrumentId string,
	from *time.Time,
	to *time.Time,
	operationTypes []pkg.OperationType,
	operationState pkg.OperationState,
	WithoutCommissions bool) (domain.UserOperations, error) {

	rawOperations, err := calc.ApiClient.GetUserOperationsByCursor(
		token,
		accountId,
		instrumentId,
		from,
		to,
		operationTypes,
		operationState,
		WithoutCommissions,
	)
	if err != nil {
		return domain.UserOperations{}, err
	}

	operations := convertOperations(rawOperations)
	return operations, nil
}

func (calc *Calculator) GetDividends(
	token string,
	accountID string,
	instrumentId string,
	from time.Time,
	to time.Time,
) (map[string]domain.MoneyValue, error) {

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

func (calc *Calculator) GetTotalReturn(
	token string,
	portfolio domain.UserFullPortfolio,
	accountID string,
	openedDate time.Time,
) (domain.MoneyValue, error) {
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

func (calc *Calculator) GetInstrumentInfo(
	token string,
	instrumentIdType pkg.InstrumentIdType,
	classCode pkg.ClassCode,
	instrumentId string,
) (domain.Instrument, error) {
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
	return convertInstrument(rawInstrument), nil
}

func (calc *Calculator) GetIndexByTicker(token string, ticker string) (domain.Instrument, error) {
	rawInstruments, err := calc.ApiClient.Indicatives(token)
	if err != nil {
		return domain.Instrument{}, err
	}
	indicativeInstruments := convertIndicativeInstrument(rawInstruments)
	for _, instr := range indicativeInstruments.Instruments {
		if instr.Ticker == ticker {
			return instr, nil
		}
	}
	return domain.Instrument{}, nil
}

func (calc *Calculator) GetCandles(
	token string,
	instrumentId string,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
	candleSourceType pkg.CandleSource,
) ([]domain.Candle, error) {
	maxRange := maxIntervalRange(candleInterval)
	var allCandles []domain.Candle

	chunkFrom := from
	for chunkFrom.Before(to) {
		chunkTo := chunkFrom.Add(maxRange)
		if chunkTo.After(to) {
			chunkTo = to
		}

		rawCandles, err := calc.ApiClient.GetCandles(token, &chunkFrom, &chunkTo, candleInterval, instrumentId, candleSourceType, 0)
		if err != nil {
			return nil, err
		}
		allCandles = append(allCandles, convertCandles(rawCandles)...)

		chunkFrom = chunkTo
	}

	return allCandles, nil
}

func (calc *Calculator) GetChartData(
	token string,
	portfolio domain.UserFullPortfolio,
	indexTicker string,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
	candleSource pkg.CandleSource,
) (domain.ChartData, error) {
	index, err := calc.GetIndexByTicker(token, indexTicker)
	if err != nil {
		return domain.ChartData{}, err
	}

	portfolioCandles, err := calc.GetCandlesForPortfolio(token, portfolio, from, to, candleInterval, candleSource)
	if err != nil {
		log.Printf("failed to get portfolio candles: %v", err)
		portfolioCandles = []domain.Candle{}
	}

	indexCandles, err := calc.GetCandles(token, index.UID, from, to, candleInterval, candleSource)
	if err != nil {
		log.Printf("failed to get index candles: %v", err)
		indexCandles = []domain.Candle{}
	}

	return domain.ChartData{
		IndexCandles:     indexCandles,
		PortfolioCandles: portfolioCandles,
	}, nil
}

// GetCandlesForPortfolio builds daily TWR candles for investment instruments only.
// TWR day factor = closeVal / openVal (same qty, different prices) — new purchases don't distort returns.
// Bond prices are converted from % of face value using nominal computed from current position data.
func (calc *Calculator) GetCandlesForPortfolio(
	token string,
	portfolio domain.UserFullPortfolio,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
	candleSource pkg.CandleSource,
) ([]domain.Candle, error) {

	if portfolio.OpenedDate.After(from) {
		from = portfolio.OpenedDate
	}

	historicalHoldings, err := calc.CalculateHistoricalHoldings(token, portfolio, from, to, candleInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical holdings: %w", err)
	}

	figis := extractUniqueFigis(historicalHoldings)

	historicalCandlesForPortfolio, err := calc.FetchHistoricalCandlesForPortfolio(token, figis, from, to, candleInterval, candleSource)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch candles: %w", err)
	}

	dividendsByDay, err := calc.getPaymentsByDay(token, portfolio.AccountID, from, to, candleInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to get dividends: %w", err)
	}

	// Compute bond multiplier: nominal/100 = currentPrice / lastCandleClose
	// so that pct_price * multiplier = rub_price
	figiToMultiplier := make(map[string]domain.Quotation)
	for _, pos := range portfolio.Positions {
		if !isBond(pos.InstrumentType) {
			continue
		}
		candles, ok := historicalCandlesForPortfolio[pos.Figi]
		if !ok || len(candles) == 0 {
			log.Printf("Not found candles for %v, figi: %v", pos.Name, pos.Figi)
			continue
		}
		lastClose := candles[len(candles)-1].Close
		if lastClose.Units == "0" && lastClose.Nano == 0 {
			continue
		}
		// multiplier = currentPrice / lastClose  (both in same scale → gives nominal/100)
		multiplier, err := DivideQuotation(
			domain.Quotation{Units: pos.CurrentPrice.Units, Nano: pos.CurrentPrice.Nano},
			lastClose,
		)
		if err != nil {
			log.Printf("failed to compute bond multiplier for %s: %v", pos.Figi, err)
			continue
		}
		figiToMultiplier[pos.Figi] = multiplier
	}

	// Build day-keyed candle index for fast lookup
	candleIndex := make(map[string]map[time.Time]domain.Candle)
	for figi, candles := range historicalCandlesForPortfolio {
		candleIndex[figi] = make(map[time.Time]domain.Candle)
		for _, c := range candles {
			t := truncateToInterval(c.Time, candleInterval)
			candleIndex[figi][t] = c
		}
	}

	intervals := make([]time.Time, 0, len(historicalHoldings))
	for interval := range historicalHoldings {
		intervals = append(intervals, interval)
	}
	slices.SortFunc(intervals, time.Time.Compare)

	result := make([]domain.Candle, 0, len(intervals))
	lastPrice := make(map[string]domain.Candle)
	twrCumulative := domain.Quotation{Units: "1", Nano: 0}

	for _, interval := range intervals {
		positions := historicalHoldings[interval]

		var openVal, closeVal domain.Quotation
		hasAnyCandle := false

		for figi, qty := range positions {
			if !parseDecimal(qty.Units, qty.Nano).IsPositive() {
				continue
			}

			var candle domain.Candle
			var ok bool
			candle, ok = candleIndex[figi][interval]
			if ok {
				lastPrice[figi] = candle
				hasAnyCandle = true
			} else if candle, ok = lastPrice[figi]; ok {
				// forward fill
			} else {
				continue
			}

			// Apply bond multiplier if needed to convert % price to RUB
			open, close_ := candle.Open, candle.Close
			if m, ok := figiToMultiplier[figi]; ok {
				open = MultiplyQuotation(open, m)
				close_ = MultiplyQuotation(close_, m)
			}

			openVal = AddQuotations(openVal, MultiplyQuotation(qty, open))
			closeVal = AddQuotations(closeVal, MultiplyQuotation(qty, close_))
		}

		if div, ok := dividendsByDay[interval]; ok {
			divDec := parseDecimal(div.Units, div.Nano)
			if divDec.IsPositive() {
				divQ := domain.Quotation{Units: div.Units, Nano: div.Nano}
				closeVal = AddQuotations(closeVal, divQ)
			}
		}

		if !hasAnyCandle {
			continue
		}

		result = append(result, domain.Candle{
			Time:  interval,
			Open:  openVal,  // вчерашний TWR = открытие сегодня
			Close: closeVal, // накопленный TWR
		})

	}
	if len(result) > 0 {
		log.Printf("TWR final: %s.%d", twrCumulative.Units, twrCumulative.Nano)
		log.Printf("First candle time: %v, open: %v, close: %v", result[0].Time, result[0].Open, result[0].Close)
		log.Printf("Last candle time: %v, open: %v, close: %v", result[len(result)-1].Time, result[len(result)-1].Open, result[len(result)-1].Close)
	}

	return result, nil
}

// CalculateHistoricalHoldings reconstructs daily portfolio positions by walking
// backwards from `to` and reversing buy/sell operations.
func (calc *Calculator) CalculateHistoricalHoldings(
	token string,
	portfolio domain.UserFullPortfolio,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
) (map[time.Time]map[string]domain.Quotation, error) {

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
		return nil, err
	}

	start := truncateToInterval(to, candleInterval)
	end := truncateToInterval(from, candleInterval)

	positionsQuantity := make(map[time.Time]map[string]domain.Quotation)
	positionsQuantity[start] = make(map[string]domain.Quotation)
	for _, pos := range portfolio.Positions {
		if !isInvestmentInstrument(pos.InstrumentType) {
			continue
		}
		positionsQuantity[start][pos.Figi] = pos.Quantity
	}

	currentTime := start
	for currentTime.After(end) {
		// Define prevTime
		var prevTime time.Time
		switch candleInterval {
		case pkg.CandleIntervalWeek:
			prevTime = currentTime.AddDate(0, 0, -7)
		case pkg.CandleIntervalMonth:
			prevTime = currentTime.AddDate(0, -1, 0)
		default:
			prevTime = currentTime.Add(-candleIntervalDuration(candleInterval))
		}

		positionsQuantity[prevTime] = make(map[string]domain.Quotation)
		for figi, qty := range positionsQuantity[currentTime] {
			positionsQuantity[prevTime][figi] = qty
		}

		for _, block := range operations {
			for _, item := range block.Items {
				if item.Figi == "" {
					continue
				}
				// Операция попадает в интервал [prevTime, currentTime)
				opTime := truncateToInterval(item.Date, candleInterval)
				if !opTime.Equal(currentTime) {
					continue
				}
				switch item.Type {
				case string(pkg.OperationTypeBuy):
					if !isInvestmentInstrument(item.InstrumentKind) {
						continue
					}
					current := positionsQuantity[prevTime][item.Figi]
					positionsQuantity[prevTime][item.Figi] = SubtractQuotations(current, domain.Quotation{Units: item.Quantity})
				case string(pkg.OperationTypeSell):
					if !isInvestmentInstrument(item.InstrumentKind) {
						continue
					}
					current := positionsQuantity[prevTime][item.Figi]
					positionsQuantity[prevTime][item.Figi] = AddQuotations(current, domain.Quotation{Units: item.Quantity})
				}
			}
		}

		currentTime = prevTime
	}

	return positionsQuantity, nil
}

// FetchHistoricalCandlesForPortfolio fetches candles for all figis in parallel.
// Uses CandleSourceUnspecified which covers all instrument types including bonds.
func (calc *Calculator) FetchHistoricalCandlesForPortfolio(
	token string,
	figis []string,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
	candleSource pkg.CandleSource,
) (map[string][]domain.Candle, error) {

	type candleResult struct {
		figi    string
		candles []domain.Candle
		err     error
	}

	resultCh := make(chan candleResult, len(figis))

	for _, figi := range figis {
		go func(f string) {
			candles, err := calc.GetCandles(token, f, from, to, candleInterval, candleSource)
			resultCh <- candleResult{figi: f, candles: candles, err: err}
		}(figi)
	}

	result := make(map[string][]domain.Candle)
	for range figis {
		res := <-resultCh
		if res.err != nil {
			log.Printf("failed to fetch candles for %s: %v", res.figi, res.err)
			continue
		}
		result[res.figi] = res.candles
	}

	return result, nil
}
