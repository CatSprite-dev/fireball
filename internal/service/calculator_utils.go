package service

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

func enrichFullPortfolio(calc *Calculator, portfolio domain.Portfolio, token string, accountID string, openedDate time.Time) (domain.Portfolio, error) {
	portfolio.OpenedDate = openedDate

	var err error
	portfolio, err = enrichPortfolioMetrics(portfolio)
	if err != nil {
		return domain.Portfolio{}, err
	}

	dividends, err := calc.GetDividends(token, accountID, "", openedDate, time.Now().UTC())
	if err != nil {
		log.Printf("failed to get dividends: %v", err)
	} else {
		portfolio.AllDividends = dividends
	}

	portfolio = enrichPositions(portfolio, calc, token)

	portfolio.TotalReturn, err = calc.GetTotalReturn(token, portfolio, accountID, openedDate)
	if err != nil {
		log.Printf("failed to calculate TotalReturn: %v", err)
	}
	return portfolio, nil
}

func enrichPortfolioMetrics(portfolio domain.Portfolio) (domain.Portfolio, error) {
	coeff, err := DivideMoneyValue(
		domain.MoneyValue{Units: portfolio.ExpectedYieldRelative.Units, Nano: portfolio.ExpectedYieldRelative.Nano},
		domain.MoneyValue{Units: "100", Nano: 0},
	)
	if err != nil {
		return domain.Portfolio{}, err
	}
	portfolio.ExpectedYield = MultiplyMoneyValue(portfolio.TotalAmountPortfolio, coeff)
	return portfolio, nil
}

func enrichPositions(portfolio domain.Portfolio, calc *Calculator, token string) domain.Portfolio {
	var wg sync.WaitGroup
	for i := range portfolio.Positions {
		wg.Add(2)
		pos := &portfolio.Positions[i]
		go getPositionInfo(&wg, pos, calc, token)
		go getPositionMetrics(&wg, &portfolio, pos)
	}
	wg.Wait()
	return portfolio
}

func getPositionInfo(wg *sync.WaitGroup, p *domain.Position, calc *Calculator, token string) {
	defer wg.Done()
	instrument, err := calc.GetInstrument(token, pkg.InstrumentIdTypePositionUid, "", p.PositionUID)
	if errors.Is(err, ErrNotFound) {
		instrument, err = calc.GetInstrument(token, pkg.InstrumentIdTypeFigi, "", p.Figi)
	}
	if err != nil {
		log.Printf("failed to get instrument info for position %s: %s: %v\n", p.PositionUID, p.Figi, err)
		return
	}
	p.Name = instrument.Name
	p.Type = instrument.InstrumentType
}

func getPositionMetrics(wg *sync.WaitGroup, portfolio *domain.Portfolio, p *domain.Position) {
	defer wg.Done()

	posAmount := MultiplyMoneyValue(p.AveragePositionPrice, domain.MoneyValue{Units: p.Quantity.Units, Nano: p.Quantity.Nano})

	var err error
	p.ExpectedYieldRelative, err = DivideQuotation(
		domain.Quotation{Units: p.ExpectedYield.Units, Nano: p.ExpectedYield.Nano},
		domain.Quotation{Units: posAmount.Units, Nano: posAmount.Nano},
	)
	if err != nil {
		log.Printf("failed to calculate ExpectedYieldRelative for position %s: %v\n", p.PositionUID, err)
		return
	}
	p.ExpectedYieldRelative = MultiplyQuotation(p.ExpectedYieldRelative, domain.Quotation{Units: "100", Nano: 0})

	p.Dividends = portfolio.AllDividends[p.Ticker]
	p.TotalYield = AddMoneyValue(p.ExpectedYield, p.Dividends)

	p.TotalYieldRelative, err = DivideQuotation(
		domain.Quotation{Units: p.TotalYield.Units, Nano: p.TotalYield.Nano},
		domain.Quotation{Units: posAmount.Units, Nano: posAmount.Nano},
	)
	if err != nil {
		log.Printf("failed to calculate TotalYieldRelative for position %s: %v", p.PositionUID, err)
		return
	}
	p.TotalYieldRelative = MultiplyQuotation(p.TotalYieldRelative, domain.Quotation{Units: "100", Nano: 0})
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

// extractUniqueFigis returns all figis that had a positive quantity at any point in time.
func extractUniqueFigis(holdings map[time.Time]map[string]domain.Quotation) []string {
	uniqueFigis := make(map[string]struct{})
	for _, positions := range holdings {
		for figi := range positions {
			uniqueFigis[figi] = struct{}{}
		}
	}
	result := make([]string, 0, len(uniqueFigis))
	for figi := range uniqueFigis {
		result = append(result, figi)
	}
	return result
}

// isInvestmentInstrument returns true for tradeable securities.
// Handles both lowercase (Position.InstrumentType) and uppercase (operation InstrumentKind) formats.
func isInvestmentInstrument(kind string) bool {
	switch kind {
	case "share", "bond", "etf", "sp", "clearing_certificate", "commodity", "currency":
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

// getPaymentsByInterval sums dividends and coupons received per interval.
func getPaymentsByInterval(
	operations domain.UserOperations,
	candleInterval pkg.CandleInterval,
) (map[time.Time]domain.MoneyValue, error) {
	result := make(map[time.Time]domain.MoneyValue)
	for _, item := range operations.Items {
		switch pkg.OperationType(item.Type) {
		case pkg.OperationTypeDividend, pkg.OperationTypeCoupon:
			interval := truncateToInterval(item.Date, candleInterval)
			result[interval] = AddMoneyValue(result[interval], domain.MoneyValue(item.Payment))
		}
	}
	return result, nil
}

// BuildIndexPortfolioCandles simulates portfolio performance if all buy/sell
// operations were executed in the index instead.
// qty_change = |payment| / index_price at operation interval
func buildIndexPortfolioCandles(
	operations domain.UserOperations,
	indexCandles []domain.Candle,
	portfolioCandles []domain.Candle,
	candleInterval pkg.CandleInterval,
) ([]domain.Candle, error) {

	if len(indexCandles) == 0 || len(portfolioCandles) == 0 {
		return nil, fmt.Errorf("index candle not found")
	}

	candleIndex := make(map[time.Time]domain.Candle)
	for _, c := range indexCandles {
		candleIndex[truncateToInterval(c.Time, candleInterval)] = c
	}

	opsByInterval := make(map[time.Time][]domain.Item)
	for _, item := range operations.Items {
		switch pkg.OperationType(item.Type) {
		case pkg.OperationTypeBuy, pkg.OperationTypeSell:
			if !isInvestmentInstrument(item.InstrumentType) {
				continue
			}
			interval := truncateToInterval(item.Date, candleInterval)
			opsByInterval[interval] = append(opsByInterval[interval], item)
		}
	}

	var currentQty domain.Quotation
	var lastIndexCandle domain.Candle

	// Seed initial position: buy index for the value of the first portfolio candle
	firstPortfolioClose := portfolioCandles[0].Close
	for _, portfolioCandle := range portfolioCandles {
		firstInterval := truncateToInterval(portfolioCandle.Time, candleInterval)
		firstIndexCandle, ok := candleIndex[firstInterval]
		if ok {
			initialQty, err := DivideQuotation(firstPortfolioClose, firstIndexCandle.Close)
			if err == nil {
				currentQty = initialQty
				lastIndexCandle = firstIndexCandle
			}
			break
		}
	}

	result := make([]domain.Candle, 0, len(portfolioCandles))

	for _, portfolioCandle := range portfolioCandles {
		interval := truncateToInterval(portfolioCandle.Time, candleInterval)

		if c, ok := candleIndex[interval]; ok {
			lastIndexCandle = c
		}
		if lastIndexCandle.Close.Units == "" && lastIndexCandle.Close.Nano == 0 {
			continue
		}

		opAmount := domain.MoneyValue{}
		for _, item := range opsByInterval[interval] {
			itemCost := MultiplyQuotation(
				domain.Quotation{Units: item.InstrumentPrice.Units, Nano: item.InstrumentPrice.Nano},
				domain.Quotation{Units: item.Quantity},
			)
			opAmount = AddMoneyValue(opAmount, domain.MoneyValue{Units: itemCost.Units, Nano: itemCost.Nano})
			qtyChange, err := DivideQuotation(itemCost, lastIndexCandle.Close)
			if err != nil {
				continue
			}
			switch pkg.OperationType(item.Type) {
			case pkg.OperationTypeBuy:
				currentQty = AddQuotations(currentQty, qtyChange)
			case pkg.OperationTypeSell:
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
