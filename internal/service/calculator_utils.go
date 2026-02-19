package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

func enrichFullPortfolio(calc *Calculator, portfolio domain.UserFullPortfolio, token string, accountID string, openedDate time.Time) (domain.UserFullPortfolio, error) {
	// 1. Общие метрики портфеля
	portfolio, err := enrichPortfolioMetrics(portfolio)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}

	// 2. Дивиденды для всего портфеля
	portfolio.AllDividends, err = calc.GetDividends(token, accountID, "", openedDate, time.Now().UTC())
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}

	// 3. Метрики каждой позиции
	portfolio, err = enrichPositions(portfolio, calc, token, accountID, openedDate)
	return portfolio, nil
}

func enrichPortfolioMetrics(portfolio domain.UserFullPortfolio) (domain.UserFullPortfolio, error) {
	// Заполняем ExpectedYield всего портфеля
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

func enrichPositions(portfolio domain.UserFullPortfolio, calc *Calculator, token string, accountID string, openedDate time.Time) (domain.UserFullPortfolio, error) {
	var wg sync.WaitGroup

	errChan := make(chan error)
	for i := range portfolio.Positions {
		wg.Add(2)
		pos := &portfolio.Positions[i]

		// Заполняем имя и тип
		go getPositionInfo(&wg, pos, calc, token, errChan)

		// Заполняем ExpectedYieldRelative, DailyYieldRelative, TotalYield, TotalYieldRelative и дивиденды для каждой позиции
		go getPositionMetrics(&wg, &portfolio, pos, errChan)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	errorList := []error{}
	for err := range errChan {
		if err != nil {
			errorList = append(errorList, err)
		}
	}
	if len(errorList) > 0 {
		return domain.UserFullPortfolio{}, fmt.Errorf("%d errors occurred", len(errorList))
	}
	return portfolio, nil
}

func getPositionInfo(wg *sync.WaitGroup, p *domain.Position, calc *Calculator, token string, errChan chan<- error) {
	defer wg.Done()
	instrument, err := calc.GetInstrumentInfo(token, pkg.InstrumentIdTypePositionUid, p.PositionUID)
	if err != nil {
		errChan <- fmt.Errorf("failed to get instrument info for position %s: %v", p.PositionUID, err)
		return
	}
	p.Name = instrument.Name
	p.Type = instrument.InstrumentType
}

func getPositionMetrics(wg *sync.WaitGroup, portfolio *domain.UserFullPortfolio, p *domain.Position, errChan chan<- error) {
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
		errChan <- fmt.Errorf("failed to calculate ExpectedYieldRelative for position %s: %v", p.PositionUID, err)
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
		errChan <- fmt.Errorf("failed to calculate TotalYieldRelative for position %s: %v", p.PositionUID, err)
		return
	}
	p.TotalYieldRelative = MultiplyQuotation(p.TotalYieldRelative, domain.Quotation{Units: "100", Nano: 0})
}
