package service

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

func enrichFullPortfolio(calc *Calculator, portfolio domain.UserFullPortfolio, token string, accountID string, openedDate time.Time) (domain.UserFullPortfolio, error) {
	portfolio.OpenedDate = openedDate
	// 1. General portfolio metrics
	portfolio, err := enrichPortfolioMetrics(portfolio)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}

	// 2. Dividents for whole portfolio
	dividends, err := calc.GetDividends(token, accountID, "", openedDate, time.Now().UTC())
	if err != nil {
		log.Printf("failed to get dividends: %v", err)
	} else {
		portfolio.AllDividends = dividends
	}

	// 3. Metrics for position
	portfolio = enrichPositions(portfolio, calc, token)

	// 4. TotalReturn
	portfolio.TotalReturn, err = calc.GetTotalReturn(token, portfolio, accountID, openedDate)
	if err != nil {
		log.Printf("failed to calculate TotalReturn: %v", err)
	}
	return portfolio, nil
}

func enrichPortfolioMetrics(portfolio domain.UserFullPortfolio) (domain.UserFullPortfolio, error) {
	coeff, err := DivideMoneyValue(
		domain.MoneyValue{
			Units: portfolio.ExpectedYieldRelative.Units,
			Nano:  portfolio.ExpectedYieldRelative.Nano,
		},
		domain.MoneyValue{Units: "100", Nano: 0},
	)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}
	portfolio.ExpectedYield = MultiplyMoneyValue(portfolio.TotalAmountPortfolio, coeff)
	return portfolio, nil
}

func enrichPositions(portfolio domain.UserFullPortfolio, calc *Calculator, token string) domain.UserFullPortfolio {
	var wg sync.WaitGroup

	for i := range portfolio.Positions {
		wg.Add(2)
		pos := &portfolio.Positions[i]

		// Name and type for pos
		go getPositionInfo(&wg, pos, calc, token)

		// ExpectedYieldRelative, DailyYieldRelative, TotalYield, TotalYieldRelative and dividents for each pos
		go getPositionMetrics(&wg, &portfolio, pos)
	}

	wg.Wait()

	return portfolio
}

func getPositionInfo(wg *sync.WaitGroup, p *domain.Position, calc *Calculator, token string) {
	defer wg.Done()
	instrument, err := calc.GetInstrumentInfo(token, pkg.InstrumentIdTypePositionUid, "", p.PositionUID)
	if errors.Is(err, ErrNotFound) {
		instrument, err = calc.GetInstrumentInfo(token, pkg.InstrumentIdTypeFigi, "", p.Figi)
	}
	if err != nil {
		log.Printf("failed to get instrument info for position %s: %s: %v\n", p.PositionUID, p.Figi, err)
		return
	}
	p.Name = instrument.Name
	p.Type = instrument.InstrumentType
}

func getPositionMetrics(wg *sync.WaitGroup, portfolio *domain.UserFullPortfolio, p *domain.Position) {
	defer wg.Done()
	// ExpectedYieldRelative
	posAmount := MultiplyMoneyValue(p.AveragePositionPrice,
		domain.MoneyValue{
			Units: p.Quantity.Units,
			Nano:  p.Quantity.Nano,
		},
	)
	var err error
	p.ExpectedYieldRelative, err = DivideQuotation(
		domain.Quotation{
			Units: p.ExpectedYield.Units,
			Nano:  p.ExpectedYield.Nano,
		},
		domain.Quotation{
			Units: posAmount.Units,
			Nano:  posAmount.Nano,
		},
	)
	if err != nil {
		log.Printf("failed to calculate ExpectedYieldRelative for position %s: %v\n", p.PositionUID, err)
		return
	}
	p.ExpectedYieldRelative = MultiplyQuotation(p.ExpectedYieldRelative, domain.Quotation{Units: "100", Nano: 0})

	//DailyYieldRelative
	// p.DailyYieldRelative =

	// Dividends
	p.Dividends = portfolio.AllDividends[p.Ticker]

	// TotalYield
	p.TotalYield = AddMoneyValue(p.ExpectedYield, p.Dividends)

	// TotalYieldRelative
	p.TotalYieldRelative, err = DivideQuotation(
		domain.Quotation{
			Units: p.TotalYield.Units,
			Nano:  p.TotalYield.Nano,
		},
		domain.Quotation{
			Units: posAmount.Units,
			Nano:  posAmount.Nano,
		},
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

func extractUniqueFigis(holdings map[time.Time]map[string]domain.Quotation) []string {
	uniqueFigis := make(map[string]struct{})
	for _, positions := range holdings {
		for figi, qty := range positions {
			if figi == "" {
				continue
			}
			if parseDecimal(qty.Units, qty.Nano).IsPositive() {
				uniqueFigis[figi] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(uniqueFigis))
	for figi := range uniqueFigis {
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

func (calc *Calculator) getPaymentsByDay(
	token string,
	accountID string,
	from time.Time,
	to time.Time,
	candleInterval pkg.CandleInterval,
) (map[time.Time]domain.MoneyValue, error) {

	operations, err := calc.ApiClient.GetUserOperationsByCursor(
		token,
		accountID,
		"",
		&from,
		&to,
		[]pkg.OperationType{pkg.OperationTypeDividend, pkg.OperationTypeCoupon},
		pkg.OperationStateExecuted,
		false,
	)
	if err != nil {
		return nil, err
	}

	result := make(map[time.Time]domain.MoneyValue)
	for _, block := range operations {
		for _, item := range block.Items {
			interval := truncateToInterval(item.Date, candleInterval)
			current := result[interval]
			result[interval] = AddMoneyValue(current, domain.MoneyValue(item.Payment))
		}
	}
	return result, nil
}
