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
	interval pkg.CandleInterval,
	candleSourceType pkg.CandleSource,
) ([]domain.Candle, error) {
	rawCandles, err := calc.ApiClient.GetCandles(token, &from, &to, interval, instrumentId, candleSourceType, 0)
	if err != nil {
		return []domain.Candle{}, err
	}
	return convertCandles(rawCandles), nil
}

func (calc *Calculator) GetChartData(
	token string,
	portfolio domain.UserFullPortfolio,
	indexTicker string,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
) (domain.ChartData, error) {
	index, err := calc.GetIndexByTicker(token, indexTicker)
	if err != nil {
		return domain.ChartData{}, err
	}

	portfolioCandles, err := calc.GetCandlesForPortfolio(token, portfolio, from, to, candleInterval)
	if err != nil {
		log.Printf("failed to get portfolio candles: %v", err)
		portfolioCandles = []domain.Candle{}
	}

	indexCandles, err := calc.GetCandles(token, index.UID, from, to, candleInterval, pkg.CandleSourceExchange)
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
) ([]domain.Candle, error) {

	historicalHoldings, err := calc.CalculateHistoricalHoldings(token, portfolio, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical holdings: %w", err)
	}

	figis := extractUniqueFigis(historicalHoldings)

	candlesOfPositions, err := calc.FetchHistoricalCandlesForPortfolio(token, figis, from, to, candleInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch candles: %w", err)
	}

	// Compute bond multiplier: nominal/100 = currentPrice / lastCandleClose
	// so that pct_price * multiplier = rub_price
	figiToMultiplier := make(map[string]domain.Quotation)
	for _, pos := range portfolio.Positions {
		if !isBond(pos.InstrumentType) {
			continue
		}
		candles, ok := candlesOfPositions[pos.Figi]
		if !ok || len(candles) == 0 {
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
	for figi, candles := range candlesOfPositions {
		candleIndex[figi] = make(map[time.Time]domain.Candle)
		for _, c := range candles {
			day := c.Time.UTC().Truncate(24 * time.Hour)
			candleIndex[figi][day] = c
		}
	}

	days := make([]time.Time, 0, len(historicalHoldings))
	for day := range historicalHoldings {
		days = append(days, day)
	}
	slices.SortFunc(days, time.Time.Compare)

	result := make([]domain.Candle, 0, len(days))
	lastPrice := make(map[string]domain.Candle)
	twrCumulative := domain.Quotation{Units: "1", Nano: 0}

	for _, day := range days {
		positions := historicalHoldings[day]

		var openVal, highVal, lowVal, closeVal domain.Quotation
		hasAnyCandle := false

		for figi, qty := range positions {
			if !parseDecimal(qty.Units, qty.Nano).IsPositive() {
				continue
			}

			var candle domain.Candle
			var ok bool
			candle, ok = candleIndex[figi][day]
			if ok {
				lastPrice[figi] = candle
				hasAnyCandle = true
			} else if candle, ok = lastPrice[figi]; ok {
				// forward fill
			} else {
				continue
			}

			// Apply bond multiplier if needed to convert % price to RUB
			open, high, low, close_ := candle.Open, candle.High, candle.Low, candle.Close
			if m, ok := figiToMultiplier[figi]; ok {
				open = MultiplyQuotation(open, m)
				high = MultiplyQuotation(high, m)
				low = MultiplyQuotation(low, m)
				close_ = MultiplyQuotation(close_, m)
			}

			openVal = AddQuotations(openVal, MultiplyQuotation(qty, open))
			highVal = AddQuotations(highVal, MultiplyQuotation(qty, high))
			lowVal = AddQuotations(lowVal, MultiplyQuotation(qty, low))
			closeVal = AddQuotations(closeVal, MultiplyQuotation(qty, close_))
		}

		if !hasAnyCandle {
			continue
		}

		// TWR: dayFactor = closeVal / openVal
		dayFactor, err := DivideQuotation(closeVal, openVal)
		if err != nil {
			log.Printf("failed to compute day factor for %s: %v", day.Format("2006-01-02"), err)
			continue
		}
		twrPrev := twrCumulative
		twrCumulative = MultiplyQuotation(twrCumulative, dayFactor)

		twrHigh, err := DivideQuotation(MultiplyQuotation(twrPrev, highVal), openVal)
		if err != nil {
			log.Printf("failed to compute twrHigh for %s: %v", day.Format("2006-01-02"), err)
			continue
		}

		twrLow, err := DivideQuotation(MultiplyQuotation(twrPrev, lowVal), openVal)
		if err != nil {
			log.Printf("failed to compute twrLow for %s: %v", day.Format("2006-01-02"), err)
			continue
		}

		result = append(result, domain.Candle{
			Time:       day,
			Open:       twrPrev,       // вчерашний TWR = открытие сегодня
			High:       twrHigh,       // twrPrev × (high / open)
			Low:        twrLow,        // twrPrev × (low / open)
			Close:      twrCumulative, // накопленный TWR
			IsComplete: true,
		})

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

	start := to.UTC().Truncate(24 * time.Hour)
	end := from.UTC().Truncate(24 * time.Hour)

	positionsQuantity := make(map[time.Time]map[string]domain.Quotation)
	positionsQuantity[start] = make(map[string]domain.Quotation)
	for _, pos := range portfolio.Positions {
		if !isInvestmentInstrument(pos.InstrumentType) {
			continue
		}
		positionsQuantity[start][pos.Figi] = pos.Quantity
	}

	currentDate := start
	for currentDate.After(end) {
		yesterday := currentDate.AddDate(0, 0, -1)

		positionsQuantity[yesterday] = make(map[string]domain.Quotation)
		for figi, qty := range positionsQuantity[currentDate] {
			positionsQuantity[yesterday][figi] = qty
		}

		for _, block := range operations {
			for _, item := range block.Items {
				if item.Figi == "" {
					continue
				}
				opDate := item.Date.UTC().Truncate(24 * time.Hour)
				if !opDate.Equal(currentDate) {
					continue
				}
				switch item.Type {
				case string(pkg.OperationTypeBuy):
					if !isInvestmentInstrument(item.InstrumentKind) {
						continue
					}
					current := positionsQuantity[yesterday][item.Figi]
					positionsQuantity[yesterday][item.Figi] = SubtractQuotations(current, domain.Quotation{Units: item.Quantity})
				case string(pkg.OperationTypeSell):
					if !isInvestmentInstrument(item.InstrumentKind) {
						continue
					}
					current := positionsQuantity[yesterday][item.Figi]
					positionsQuantity[yesterday][item.Figi] = AddQuotations(current, domain.Quotation{Units: item.Quantity})
				}
			}
		}

		currentDate = yesterday
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
) (map[string][]domain.Candle, error) {

	type candleResult struct {
		figi    string
		candles []domain.Candle
		err     error
	}

	resultCh := make(chan candleResult, len(figis))

	for _, figi := range figis {
		go func(f string) {
			candles, err := calc.GetCandles(token, f, from, to, candleInterval, pkg.CandleSourceUnspecified)
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

func extractUniqueFigis(holdings map[time.Time]map[string]domain.Quotation) []string {
	seen := make(map[string]struct{})
	for _, positions := range holdings {
		for figi, qty := range positions {
			if figi == "" {
				continue
			}
			if parseDecimal(qty.Units, qty.Nano).IsPositive() {
				seen[figi] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(seen))
	for figi := range seen {
		result = append(result, figi)
	}
	return result
}

// isInvestmentInstrument returns true for tradeable securities (shares, bonds, ETFs, etc.)
// Handles both lowercase (Position.InstrumentType) and uppercase (operation InstrumentKind) formats.
func isInvestmentInstrument(kind string) bool {
	switch kind {
	case "share", "bond", "etf", "sp", "clearing_certificate", "commodity":
		return true
	}
	switch pkg.InstrumentType(kind) {
	case pkg.InstrumentTypeBond,
		pkg.InstrumentTypeShare,
		pkg.InstrumentTypeETF,
		pkg.InstrumentTypeSP,
		pkg.InstrumentTypeClearingCertificate,
		pkg.InstrumentTypeCommodity:
		return true
	}
	return false
}

// isBond handles both lowercase and uppercase bond type strings.
func isBond(kind string) bool {
	return kind == "bond" || pkg.InstrumentType(kind) == pkg.InstrumentTypeBond
}
