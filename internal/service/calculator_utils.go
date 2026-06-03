package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

func enrichFullPortfolio(
	ctx context.Context,
	calc *Calculator,
	portfolio domain.Portfolio,
	token string,
	accountID string,
	openedDate time.Time,
) (domain.Portfolio, error) {
	portfolio.OpenedDate = openedDate

	var err error
	portfolio, err = enrichPortfolioMetrics(portfolio)
	if err != nil {
		return domain.Portfolio{}, err
	}

	now := time.Now().UTC()
	dividends, err := calc.GetDividends(ctx, token, accountID, "", openedDate, now)
	if err != nil {
		log.Printf("failed to get dividends: %v", err)
	} else {
		portfolio.AllDividends = dividends
	}

	portfolio = enrichPositions(ctx, portfolio, calc, token)

	portfolio.TotalReturn, portfolio.TotalReturnRelative, portfolio.TotalInvested, err = calc.GetTotalReturn(ctx, token, portfolio, accountID, openedDate)
	if err != nil {
		log.Printf("failed to calculate TotalReturn: %v", err)
	}
	return portfolio, nil
}

func enrichPortfolioMetrics(portfolio domain.Portfolio) (domain.Portfolio, error) {
	coeff, err := DivideQuotation(
		portfolio.ExpectedYieldRelative,
		domain.Quotation{Units: "100", Nano: 0},
	)
	if err != nil {
		return domain.Portfolio{}, err
	}
	portfolio.ExpectedYield = MultiplyMoneyValue(portfolio.TotalAmountPortfolio, domain.MoneyValue{Units: coeff.Units, Nano: coeff.Nano, Currency: portfolio.TotalAmountPortfolio.Currency})
	return portfolio, nil
}

func enrichPositions(ctx context.Context, portfolio domain.Portfolio, calc *Calculator, token string) domain.Portfolio {
	var wg sync.WaitGroup
	for i := range portfolio.Positions {
		pos := &portfolio.Positions[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			getPositionInfo(ctx, token, pos, calc)
			getPositionMetrics(ctx, token, portfolio.AllDividends, pos, calc)
		}()
	}
	wg.Wait()
	return portfolio
}

func getPositionInfo(ctx context.Context, token string, p *domain.Position, calc *Calculator) {
	instrument, err := calc.GetInstrument(ctx, token, pkg.InstrumentIdTypePositionUid, "", p.PositionUID)
	if errors.Is(err, ErrNotFound) {
		instrument, err = calc.GetInstrument(ctx, token, pkg.InstrumentIdTypeFigi, "", p.Figi)
	}
	if err != nil {
		log.Printf("failed to get instrument info for position %s: %s: %v\n", p.PositionUID, p.Figi, err)
		return
	}
	p.Name = instrument.Name
	p.Type = instrument.InstrumentType
}

func getPositionMetrics(ctx context.Context, token string, allDividends map[string]domain.MoneyValue, pos *domain.Position, calc *Calculator) {
	posAmount := MultiplyMoneyValue(pos.AveragePositionPrice, domain.MoneyValue{Units: pos.Quantity.Units, Nano: pos.Quantity.Nano, Currency: pos.AveragePositionPrice.Currency})

	var err error
	pos.ExpectedYieldRelative, err = DivideQuotation(
		domain.Quotation{Units: pos.ExpectedYield.Units, Nano: pos.ExpectedYield.Nano},
		domain.Quotation{Units: posAmount.Units, Nano: posAmount.Nano},
	)
	if err != nil {
		log.Printf("failed to calculate ExpectedYieldRelative for position %s: %v\n", pos.PositionUID, err)
		return
	}
	pos.ExpectedYieldRelative = MultiplyQuotation(pos.ExpectedYieldRelative, domain.Quotation{Units: "100", Nano: 0})

	/*candles := []domain.Candle{}
	candles = append(candles, domain.Candle{
		Time: time.Now(),
		Close: domain.Quotation{
			Units: "80", Nano: 0,
		},
		Open: domain.Quotation{
			Units: "80", Nano: 0,
		},
	})*/
	pos.Dividends = allDividends[pos.Ticker]
	if pos.Dividends.Currency != "rub" {
		pos.Dividends = calc.ConvertLastPriceToRUB(ctx, token, pos.Dividends)
	}
	pos.TotalYield = AddMoneyValue(pos.ExpectedYield, pos.Dividends)

	pos.TotalYieldRelative, err = DivideQuotation(
		domain.Quotation{Units: pos.TotalYield.Units, Nano: pos.TotalYield.Nano},
		domain.Quotation{Units: posAmount.Units, Nano: posAmount.Nano},
	)
	if err != nil {
		log.Printf("failed to calculate TotalYieldRelative for position %s: %v", pos.PositionUID, err)
	}
	pos.TotalYieldRelative = MultiplyQuotation(pos.TotalYieldRelative, domain.Quotation{Units: "100", Nano: 0})
}

func maxIntervalRange(interval pkg.CandleInterval) time.Duration {
	switch interval {
	case pkg.CandleInterval5Sec:
		return 200 * time.Minute
	case pkg.CandleInterval10Sec:
		return 200 * time.Minute
	case pkg.CandleInterval30Sec:
		return 20 * time.Hour
	case pkg.CandleInterval1Min:
		return 24 * time.Hour
	case pkg.CandleInterval2Min:
		return 24 * time.Hour
	case pkg.CandleInterval3Min:
		return 24 * time.Hour
	case pkg.CandleInterval5Min:
		return 7 * 24 * time.Hour
	case pkg.CandleInterval10Min:
		return 7 * 24 * time.Hour
	case pkg.CandleInterval15Min:
		return 21 * 24 * time.Hour
	case pkg.CandleInterval30Min:
		return 21 * 24 * time.Hour
	case pkg.CandleIntervalHour:
		return 90 * 24 * time.Hour
	case pkg.CandleInterval2Hour:
		return 90 * 24 * time.Hour
	case pkg.CandleInterval4Hour:
		return 90 * 24 * time.Hour
	case pkg.CandleIntervalDay:
		return 6 * 365 * 24 * time.Hour
	case pkg.CandleIntervalWeek:
		return 5 * 365 * 24 * time.Hour
	case pkg.CandleIntervalMonth:
		return 10 * 365 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

func truncateToInterval(t time.Time, interval pkg.CandleInterval) time.Time {
	t = t.UTC()
	switch interval {
	case pkg.CandleIntervalWeek:
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return t.Truncate(24*time.Hour).AddDate(0, 0, -(weekday - 1))
	case pkg.CandleIntervalMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	case pkg.CandleIntervalDay:
		return t.Truncate(24 * time.Hour)
	default:
		return t.Truncate(candleIntervalDuration(interval))
	}
}

func candleIntervalDuration(interval pkg.CandleInterval) time.Duration {
	switch interval {
	case pkg.CandleInterval5Sec:
		return 5 * time.Second
	case pkg.CandleInterval10Sec:
		return 10 * time.Second
	case pkg.CandleInterval30Sec:
		return 30 * time.Second
	case pkg.CandleInterval1Min:
		return time.Minute
	case pkg.CandleInterval2Min:
		return 2 * time.Minute
	case pkg.CandleInterval3Min:
		return 3 * time.Minute
	case pkg.CandleInterval5Min:
		return 5 * time.Minute
	case pkg.CandleInterval10Min:
		return 10 * time.Minute
	case pkg.CandleInterval15Min:
		return 15 * time.Minute
	case pkg.CandleInterval30Min:
		return 30 * time.Minute
	case pkg.CandleIntervalHour:
		return time.Hour
	case pkg.CandleInterval2Hour:
		return 2 * time.Hour
	case pkg.CandleInterval4Hour:
		return 4 * time.Hour
	case pkg.CandleIntervalDay:
		return 24 * time.Hour
	case pkg.CandleIntervalWeek:
		return 7 * 24 * time.Hour
	case pkg.CandleIntervalMonth:
		return 30 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

// pattern can be "figi", "ticker", "currency", "classCode", "instrumentType" or "positionUid"
func extractUniqueElements(operations domain.UserOperations, pattern string) map[string]struct{} {
	getValue := func(item domain.Item) string {
		switch pattern {
		case "figi":
			return item.Figi
		case "ticker":
			return item.Ticker
		case "currency":
			return item.Payment.Currency
		case "classCode":
			return item.ClassCode
		case "instrumentType":
			return item.InstrumentType
		case "positionUid":
			return item.PositionUID
		default:
			return ""
		}
	}

	unique := make(map[string]struct{})
	for _, item := range operations.Items {
		if val := getValue(item); val != "" {
			unique[val] = struct{}{}
		}
	}
	return unique
}

// isInvestmentInstrument returns true for tradeable securities.
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
func isBond(instrumentType string) bool {
	return instrumentType == "bond" || pkg.InstrumentType(instrumentType) == pkg.InstrumentTypeBond
}

// BuildIndexPortfolioCandles simulates portfolio performance if all buy/sell
// operations were executed in the index instead.
// qty_change = |payment| / index_price at operation interval
func buildBenchmarkCandles(
	opsByInterval map[time.Time][]domain.Item,
	indexCandles []domain.Candle,
	portfolioCandles []domain.Candle,
	paymentsByInterval map[time.Time]domain.MoneyValue,
	candleInterval pkg.CandleInterval,
) ([]domain.Candle, error) {
	if len(indexCandles) == 0 || len(portfolioCandles) == 0 {
		return nil, fmt.Errorf("index candle not found")
	}

	candlesByTime := makeCandlesByTime(indexCandles, candleInterval)

	var currentQty domain.Quotation
	var lastIndexCandle domain.Candle

	// Seed initial position: buy index for the value of the first portfolio candle
	firstInterval := truncateToInterval(portfolioCandles[0].Time, candleInterval)
	firstIndexCandle, ok := candlesByTime[firstInterval]
	if !ok {
		firstIndexCandle = indexCandles[0]
	}
	currentQty, _ = DivideQuotation(portfolioCandles[0].Close, firstIndexCandle.Close)
	lastIndexCandle = firstIndexCandle
	accumulatedDividends := domain.Quotation{}

	result := make([]domain.Candle, 0, len(portfolioCandles))

	for _, portfolioCandle := range portfolioCandles {
		interval := truncateToInterval(portfolioCandle.Time, candleInterval)
		if c, ok := candlesByTime[interval]; ok {
			lastIndexCandle = c
		}

		// Accumulate dividends/coupons for this interval
		if payment, ok := paymentsByInterval[interval]; ok {
			accumulatedDividends = AddQuotations(accumulatedDividends, domain.Quotation{Units: payment.Units, Nano: payment.Nano})
		}

		for _, item := range opsByInterval[interval] {
			itemCost := MultiplyQuotation(
				domain.Quotation{Units: item.InstrumentPrice.Units, Nano: item.InstrumentPrice.Nano},
				domain.Quotation{Units: item.Quantity},
			)

			switch pkg.OperationType(item.Type) {
			case pkg.OperationTypeBuy:
				// Use accumulated dividends to offset the cost
				effectiveCost := SubtractQuotations(itemCost, accumulatedDividends)
				qtyChange, err := DivideQuotation(effectiveCost, lastIndexCandle.Close)
				if err != nil {
					log.Println(err.Error())
					continue
				}
				currentQty = AddQuotations(currentQty, qtyChange)
				accumulatedDividends = domain.Quotation{} // Reset after use
			case pkg.OperationTypeSell:
				qtyChange, err := DivideQuotation(itemCost, lastIndexCandle.Close)
				if err != nil {
					log.Println(err.Error())
					continue
				}
				currentQty = SubtractQuotations(currentQty, qtyChange)
			case pkg.OperationTypeBondRepayment, pkg.OperationTypeBondRepaymentFull:
				itemCost = domain.Quotation{Units: item.Payment.Units, Nano: item.Payment.Nano}
				qtyChange, err := DivideQuotation(itemCost, lastIndexCandle.Close)
				if err != nil {
					log.Println(err.Error())
					continue
				}
				currentQty = SubtractQuotations(currentQty, qtyChange)
			}
		}

		closeVal := MultiplyQuotation(currentQty, lastIndexCandle.Close)

		result = append(result, domain.Candle{
			Time:  portfolioCandle.Time,
			Close: closeVal,
		})
	}

	return result, nil
}

// CalculateHistoricalHoldings reconstructs portfolio positions for each interval
// by walking backwards from `to` and reversing buy/sell operations.
func calculateHistoricalHoldings(
	operations map[time.Time][]domain.Item,
	positions []domain.Position,
	bondNominals map[string]domain.MoneyValue,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
) (map[time.Time]map[string]domain.Quotation, error) {
	start := truncateToInterval(from, candleInterval)
	end := truncateToInterval(to, candleInterval)

	holdings := make(map[time.Time]map[string]domain.Quotation)
	holdings[end] = make(map[string]domain.Quotation)
	positionUIDtoFigi := make(map[string]string)

	for _, pos := range positions {
		if !isInvestmentInstrument(pos.InstrumentType) {
			continue
		}
		holdings[end][pos.Figi] = pos.Quantity
		positionUIDtoFigi[pos.PositionUID] = pos.Figi
	}

	currentTime := end
	for currentTime.After(start) {
		var prevTime time.Time
		switch candleInterval {
		case pkg.CandleIntervalWeek:
			prevTime = currentTime.AddDate(0, 0, -7)
		case pkg.CandleIntervalMonth:
			prevTime = currentTime.AddDate(0, -1, 0)
		default:
			prevTime = currentTime.Add(-candleIntervalDuration(candleInterval))
		}

		holdings[prevTime] = make(map[string]domain.Quotation)
		for figi, qty := range holdings[currentTime] {
			holdings[prevTime][figi] = qty
		}

		for _, item := range operations[currentTime] {
			// splits handler
			if split, ok := pkg.Splits[item.Ticker]; ok && item.Date.Before(split.Date) {
				qtyFloat, err := strconv.ParseFloat(item.Quantity, 64)
				if err != nil {
					log.Printf("failed to parse quantity %s: %v", item.Quantity, err)
					continue
				}
				newQtyFloat := qtyFloat * split.Coef
				item.Quantity = strconv.FormatFloat(newQtyFloat, 'f', -1, 64)
			}

			if figi, ok := positionUIDtoFigi[item.PositionUID]; ok {
				item.Figi = figi
			} else {
				positionUIDtoFigi[item.PositionUID] = item.Figi
			}

			switch pkg.OperationType(item.Type) {
			case pkg.OperationTypeBuy:
				holdings[prevTime][item.Figi] = SubtractQuotations(
					holdings[prevTime][item.Figi],
					domain.Quotation{Units: item.Quantity},
				)
			case pkg.OperationTypeSell:
				holdings[prevTime][item.Figi] = AddQuotations(
					holdings[prevTime][item.Figi],
					domain.Quotation{Units: item.Quantity},
				)
			case pkg.OperationTypeBondRepaymentFull:
				if nominal, ok := bondNominals[item.Figi]; ok {
					qty, err := DivideMoneyValueToQuotation(item.Payment, nominal)
					if err != nil {
						log.Printf("failed to calculate quantity for bond repayment of %s: %v", item.Figi, err)
						continue
					}
					holdings[prevTime][item.Figi] = AddQuotations(
						holdings[prevTime][item.Figi],
						qty,
					)
				}

			}
			if holdings[prevTime][item.Figi].Units == "0" && holdings[prevTime][item.Figi].Nano == 0 {
				delete(holdings[prevTime], item.Figi)
			}
		}
		currentTime = prevTime
	}

	return holdings, nil
}

func opsByInterval(operations domain.UserOperations, candleInterval pkg.CandleInterval) map[time.Time][]domain.Item {
	opsByInterval := make(map[time.Time][]domain.Item)
	for _, item := range operations.Items {
		if !isInvestmentInstrument(item.InstrumentType) {
			continue
		}
		interval := truncateToInterval(item.Date, candleInterval)
		opsByInterval[interval] = append(opsByInterval[interval], item)
	}
	return opsByInterval
}

func sortedIntervals(holdings map[time.Time]map[string]domain.Quotation) []time.Time {
	intervals := make([]time.Time, 0, len(holdings))
	for interval := range holdings {
		intervals = append(intervals, interval)
	}
	slices.SortFunc(intervals, time.Time.Compare)
	return intervals
}

func makeCandlesByTime(candles []domain.Candle, candleInterval pkg.CandleInterval) map[time.Time]domain.Candle {
	candleMap := make(map[time.Time]domain.Candle)
	for _, c := range candles {
		candleMap[truncateToInterval(c.Time, candleInterval)] = c
	}
	return candleMap
}

func buildLastPriceCache(historicalCandles map[string][]domain.Candle, from time.Time) map[string]domain.Candle {
	lastPrice := make(map[string]domain.Candle)
	for figi, candles := range historicalCandles {
		for _, c := range candles {
			if c.Time.Before(from) {
				lastPrice[figi] = c
			}
		}
	}
	return lastPrice
}

func convertToRUB(amount domain.MoneyValue, date time.Time, currencyRates []domain.Candle) (domain.MoneyValue, error) {
	if amount.Currency == "rub" {
		return amount, nil
	}

	if len(currencyRates) == 0 {
		return amount, fmt.Errorf("no currency rates found for %s", amount.Currency)
	}

	slices.SortFunc(currencyRates, func(a, b domain.Candle) int {
		return a.Time.Compare(b.Time)
	})

	var closestCandle *domain.Candle
	for i := range currencyRates {
		if !currencyRates[i].Time.After(date) {
			closestCandle = &currencyRates[i]
		} else {
			break
		}
	}

	if closestCandle == nil {
		return amount, fmt.Errorf("no exchange rate found for %s on or before %v", amount.Currency, date)
	}

	exchangeRate := domain.MoneyValue{
		Units:    closestCandle.Close.Units,
		Nano:     closestCandle.Close.Nano,
		Currency: amount.Currency,
	}

	result := MultiplyMoneyValue(amount, exchangeRate)
	result.Currency = "rub"

	return result, nil
}
