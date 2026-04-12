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

type PortfolioRequest struct {
	Token      string
	AccountID  string
	OpenedDate time.Time
}

var ErrNotFound = errors.New("not found")

type Calculator struct {
	ApiClient *api.Client
}

func NewCalculator(apiClient *api.Client) *Calculator {
	return &Calculator{ApiClient: apiClient}
}

func (calc *Calculator) GetFullPortfolio(session PortfolioRequest) (domain.Portfolio, error) {
	t := time.Now()

	rawPortfolio, err := calc.ApiClient.GetPortfolio(session.Token, session.AccountID)
	if err != nil {
		return domain.Portfolio{}, err
	}
	portfolio := convertFullPortfolio(rawPortfolio)
	portfolio, err = enrichFullPortfolio(calc, portfolio, session.Token, session.AccountID, session.OpenedDate)
	if err != nil {
		return domain.Portfolio{}, err
	}

	log.Printf("Время выполнения GetFullPortfolio: %.2f сек\n", time.Since(t).Seconds())
	return portfolio, nil
}

func (calc *Calculator) GetOperations(
	token string,
	accountId string,
	instrumentId string,
	from *time.Time,
	to *time.Time,
	operationTypes []pkg.OperationType,
	operationState pkg.OperationState,
	withoutCommissions bool,
) (domain.UserOperations, error) {
	rawOperations, err := calc.ApiClient.GetUserOperationsByCursor(
		token, accountId, instrumentId, from, to,
		operationTypes, operationState, withoutCommissions,
	)
	if err != nil {
		return domain.UserOperations{}, err
	}
	return convertOperations(rawOperations), nil
}

func (calc *Calculator) GetDividends(
	token string,
	accountID string,
	instrumentId string,
	from time.Time,
	to time.Time,
) (map[string]domain.MoneyValue, error) {
	operations, err := calc.ApiClient.GetUserOperationsByCursor(
		token, accountID, instrumentId, &from, &to,
		[]pkg.OperationType{pkg.OperationTypeDividend, pkg.OperationTypeCoupon},
		pkg.OperationStateExecuted, false,
	)
	if err != nil {
		return nil, err
	}

	result := make(map[string]domain.MoneyValue)
	for _, block := range operations {
		for _, item := range block.Items {
			if item.Ticker == "" {
				continue
			}
			result[item.Ticker] = AddMoneyValue(result[item.Ticker], domain.MoneyValue(item.Payment))
		}
	}
	return result, nil
}

func (calc *Calculator) GetTotalReturn(
	token string,
	portfolio domain.Portfolio,
	accountID string,
	openedDate time.Time,
) (domain.MoneyValue, domain.Quotation, domain.MoneyValue, error) {
	now := time.Now()
	operations, err := calc.ApiClient.GetUserOperationsByCursor(
		token, accountID, "", &openedDate, &now,
		[]pkg.OperationType{
			pkg.OperationTypeInput,
			pkg.OperationTypeOutput,
			pkg.OperationTypeInpMulti,
			pkg.OperationTypeOutMulti,
		},
		pkg.OperationStateExecuted, true,
	)
	if err != nil {
		return domain.MoneyValue{}, domain.Quotation{}, domain.MoneyValue{}, err
	}

	var totalInvested domain.MoneyValue
	for _, block := range operations {
		for _, item := range block.Items {
			totalInvested = AddMoneyValue(totalInvested, domain.MoneyValue(item.Payment))
		}
	}

	totalReturn := SubtractMoneyValue(portfolio.TotalAmountPortfolio, totalInvested)
	coef, err := DivideMoneyValueToQuotation(totalReturn, totalInvested)
	if err != nil {
		return domain.MoneyValue{}, domain.Quotation{}, domain.MoneyValue{}, err
	}
	totalReturnRelative := MultiplyQuotation(coef, domain.Quotation{Units: "100"})

	return totalReturn, totalReturnRelative, totalInvested, nil
}

func (calc *Calculator) GetInstrument(token string, instrumentIdType pkg.InstrumentIdType, classCode pkg.ClassCode, instrumentId string) (domain.Instrument, error) {
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

func (calc *Calculator) BondBy(token string, instrumentIdType pkg.InstrumentIdType, classCode pkg.ClassCode, instrumentId string) (domain.Bond, error) {
	if classCode == "" {
		classCode = pkg.ClassCodeUnspecified
	}

	rawBond, err := calc.ApiClient.BondBy(token, instrumentIdType, classCode, instrumentId)
	if err != nil {
		var requestErr api.RequestError
		if errors.As(err, &requestErr) && requestErr.StatusCode == http.StatusNotFound {
			return domain.Bond{}, ErrNotFound
		}
		return domain.Bond{}, err
	}

	bond := convertBond(rawBond)
	return bond, nil
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

	for chunkFrom := from; chunkFrom.Before(to); {
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
	portfolio domain.Portfolio,
	indexTicker string,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
	candleSource pkg.CandleSource,
) (domain.ChartData, error) {
	t := time.Now()
	operations, err := calc.GetOperations(
		token, portfolio.AccountID, "", &from, &to,
		[]pkg.OperationType{
			pkg.OperationTypeBuy,
			pkg.OperationTypeSell,
			pkg.OperationTypeDividend,
			pkg.OperationTypeCoupon,
		},
		pkg.OperationStateExecuted, false,
	)
	if err != nil {
		return domain.ChartData{}, fmt.Errorf("failed to get operations: %w", err)
	}

	index, err := calc.GetIndexByTicker(token, indexTicker)
	if err != nil {
		return domain.ChartData{}, err
	}

	portfolioCandles, err := calc.GetCandlesForPortfolio(token, portfolio, operations, from, to, candleInterval, candleSource)
	if err != nil {
		log.Printf("failed to get portfolio candles: %v", err)
		portfolioCandles = []domain.Candle{}
	}

	indexCandles, err := calc.GetCandles(token, index.UID, from, to, candleInterval, candleSource)
	if err != nil {
		log.Printf("failed to get index candles: %v", err)
		indexCandles = []domain.Candle{}
	}

	virtualPortfolioCandles, err := buildIndexPortfolioCandles(operations, indexCandles, portfolioCandles, candleInterval)
	if err != nil {
		log.Printf("failed to get virtual portfolio candles: %v", err)
		virtualPortfolioCandles = []domain.Candle{}
	}

	times := make([]time.Time, 0, len(portfolioCandles))
	portfolioClose := make([]domain.Quotation, 0, len(portfolioCandles))
	indexClose := make([]domain.Quotation, 0, len(virtualPortfolioCandles))

	for _, c := range portfolioCandles {
		times = append(times, c.Time)
		portfolioClose = append(portfolioClose, c.Close)
	}
	for _, c := range virtualPortfolioCandles {
		indexClose = append(indexClose, c.Close)
	}

	log.Printf("Время выполнения GetChartData: %.2f сек\n", time.Since(t).Seconds())
	return domain.ChartData{
		Times:     times,
		Index:     indexClose,
		Portfolio: portfolioClose,
	}, nil
}

// GetCandlesForPortfolio builds historical value candles for the portfolio
// Close = sum(qty * price) for each interval + dividends/coupons received that interval
// Bond prices are converted from % of face value using a multiplier
func (calc *Calculator) GetCandlesForPortfolio(
	token string,
	portfolio domain.Portfolio,
	operations domain.UserOperations,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
	candleSource pkg.CandleSource,
) ([]domain.Candle, error) {
	if portfolio.OpenedDate.After(from) {
		from = portfolio.OpenedDate
	}

	// Reconstruct historical qty per instrument per interval
	historicalHoldings, err := calc.CalculateHistoricalHoldings(operations, portfolio, from, to, candleInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical holdings: %w", err)
	}

	figis := extractUniqueFigis(historicalHoldings)

	historicalCandles, err := calc.FetchHistoricalCandlesForPortfolio(token, figis, from, to, candleInterval, candleSource)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch candles: %w", err)
	}

	paymentsByInterval, err := getPaymentsByInterval(operations, candleInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by interval: %w", err)
	}

	// Bond price multiplier
	figiToMultiplier := calc.FetchBondMultipliers(token, portfolio.Positions, operations)

	// Index candles by truncated interval time for O(1) lookup
	candleIndex := make(map[string]map[time.Time]domain.Candle)
	for figi, candles := range historicalCandles {
		candleIndex[figi] = make(map[time.Time]domain.Candle)
		for _, c := range candles {
			candleIndex[figi][truncateToInterval(c.Time, candleInterval)] = c
		}
	}

	intervals := make([]time.Time, 0, len(historicalHoldings))
	for interval := range historicalHoldings {
		intervals = append(intervals, interval)
	}
	slices.SortFunc(intervals, time.Time.Compare)

	result := make([]domain.Candle, 0, len(intervals))
	lastPrice := make(map[string]domain.Candle) // forward-fill cache
	// Pre-fill lastPrice from candles fetched before 'from'
	for figi, candles := range historicalCandles {
		for _, c := range candles {
			if c.Time.Before(from) {
				lastPrice[figi] = c
			}
		}
	}

	initial := AddMoneyValue(portfolio.TotalAmountShares, portfolio.TotalAmountBonds)
	initial = AddMoneyValue(initial, portfolio.TotalAmountEtf)
	initial = AddMoneyValue(initial, portfolio.TotalAmountSp)

	for i, interval := range intervals {
		if i == len(intervals)-1 {
			closeVal := domain.Quotation{Units: initial.Units, Nano: initial.Nano}
			result = append(result, domain.Candle{Time: interval, Close: closeVal})
			break
		}

		var closeVal domain.Quotation
		for figi, qty := range historicalHoldings[interval] {
			candle, ok := candleIndex[figi][interval]
			if ok {
				lastPrice[figi] = candle
			} else if candle, ok = lastPrice[figi]; ok {
				// forward fill
			} else {
				continue
			}

			close_ := candle.Close
			if m, ok := figiToMultiplier[figi]; ok {
				close_ = MultiplyQuotation(close_, m)
			}

			posVal := MultiplyQuotation(qty, close_)
			closeVal = AddQuotations(closeVal, posVal)
		}

		if payment, ok := paymentsByInterval[interval]; ok {
			closeVal = AddQuotations(closeVal, domain.Quotation{Units: payment.Units, Nano: payment.Nano})
		}

		result = append(result, domain.Candle{Time: interval, Close: closeVal})
	}

	return result, nil
}

// CalculateHistoricalHoldings reconstructs portfolio positions for each interval
// by walking backwards from `to` and reversing buy/sell operations.
func (calc *Calculator) CalculateHistoricalHoldings(
	operations domain.UserOperations,
	portfolio domain.Portfolio,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
) (map[time.Time]map[string]domain.Quotation, error) {
	start := truncateToInterval(to, candleInterval)
	end := truncateToInterval(from, candleInterval)

	positionsQuantity := make(map[time.Time]map[string]domain.Quotation)
	positionsQuantity[start] = make(map[string]domain.Quotation)

	// Seed with current positions
	for _, pos := range portfolio.Positions {
		if !isInvestmentInstrument(pos.InstrumentType) {
			continue
		}
		positionsQuantity[start][pos.Figi] = pos.Quantity
	}

	currentTime := start
	for currentTime.After(end) {
		var prevTime time.Time
		switch candleInterval {
		case pkg.CandleIntervalWeek:
			prevTime = currentTime.AddDate(0, 0, -7)
		case pkg.CandleIntervalMonth:
			prevTime = currentTime.AddDate(0, -1, 0)
		default:
			prevTime = currentTime.Add(-candleIntervalDuration(candleInterval))
		}

		// Copy forward positions, then reverse operations that fall in this interval
		positionsQuantity[prevTime] = make(map[string]domain.Quotation)
		for figi, qty := range positionsQuantity[currentTime] {
			positionsQuantity[prevTime][figi] = qty
		}

		for _, item := range operations.Items {
			if !truncateToInterval(item.Date, candleInterval).Equal(currentTime) {
				continue
			}
			switch pkg.OperationType(item.Type) {
			case pkg.OperationTypeBuy:
				if !isInvestmentInstrument(item.InstrumentType) {
					continue
				}
				positionsQuantity[prevTime][item.Figi] = SubtractQuotations(
					positionsQuantity[prevTime][item.Figi],
					domain.Quotation{Units: item.Quantity},
				)
			case pkg.OperationTypeSell:
				if !isInvestmentInstrument(item.InstrumentType) {
					continue
				}
				positionsQuantity[prevTime][item.Figi] = AddQuotations(
					positionsQuantity[prevTime][item.Figi],
					domain.Quotation{Units: item.Quantity},
				)
			}
		}

		currentTime = prevTime
	}

	return positionsQuantity, nil
}

// FetchHistoricalCandlesForPortfolio fetches candles for all figis in parallel.
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
			candles, err := calc.GetCandles(token, f, from.AddDate(0, 0, -10), to, candleInterval, candleSource)
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

// fetchBondMultipliers fetches nominal/100 multiplier for each bond figi in parallel.
// Falls back to 10 (1000 RUB nominal) if bond info is unavailable.
func (calc *Calculator) FetchBondMultipliers(token string, positions []domain.Position, operations domain.UserOperations) map[string]domain.Quotation {
	// Collect unique bond figis
	bondFigis := make(map[string]struct{})

	for _, pos := range positions {
		if isBond(pos.InstrumentType) && pos.Figi != "" {
			bondFigis[pos.Figi] = struct{}{}
		}
	}

	for _, item := range operations.Items {
		if isBond(item.InstrumentType) && item.Figi != "" {
			bondFigis[item.Figi] = struct{}{}
		}
	}

	type bondResult struct {
		figi       string
		multiplier domain.Quotation
	}

	resultCh := make(chan bondResult, len(bondFigis))
	for figi := range bondFigis {
		go func(f string) {
			bond, err := calc.BondBy(token, pkg.InstrumentIdTypeFigi, "", f)
			if err != nil {
				log.Printf("failed to get bond info for %s, using default multiplier: %v", f, err)
				resultCh <- bondResult{figi: f, multiplier: domain.Quotation{Units: "10", Nano: 0}}
				return
			}
			multiplier, err := DivideMoneyValue(bond.Nominal, domain.MoneyValue{Units: "100", Nano: 0})
			if err != nil {
				resultCh <- bondResult{figi: f, multiplier: domain.Quotation{Units: "10", Nano: 0}}
				return
			}
			resultCh <- bondResult{figi: f, multiplier: domain.Quotation{Units: multiplier.Units, Nano: multiplier.Nano}}
		}(figi)
	}

	result := make(map[string]domain.Quotation, len(bondFigis))
	for range bondFigis {
		r := <-resultCh
		result[r.figi] = r.multiplier
	}
	return result
}
