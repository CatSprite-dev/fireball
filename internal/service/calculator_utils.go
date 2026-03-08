package service

import (
	"log"
	"sync"
	"time"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

func enrichFullPortfolio(calc *Calculator, portfolio domain.UserFullPortfolio, token string, accountID string, openedDate time.Time) (domain.UserFullPortfolio, error) {
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

	// 3. Each position metric
	portfolio = enrichPositions(portfolio, calc, token)
	return portfolio, nil
}

func enrichPortfolioMetrics(portfolio domain.UserFullPortfolio) (domain.UserFullPortfolio, error) {
	// Filling ExpectedYield for whole portfolio
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
	instrument, err := calc.GetInstrumentInfo(token, pkg.InstrumentIdTypePositionUid, p.PositionUID)
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
