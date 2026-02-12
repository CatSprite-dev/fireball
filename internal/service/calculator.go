package service

import (
	"log"
	"sync"
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

type Calculator struct {
	apiClient *api.Client
}

func NewCalculator(apiClient *api.Client) *Calculator {
	return &Calculator{apiClient: apiClient}
}

func (calc *Calculator) GetFullPortfolio(token string) (domain.UserFullPortfolio, error) {
	t := time.Now()

	userAccounts, err := calc.apiClient.GetAccounts(token, pkg.AccountStatusOpen)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}

	accountID := userAccounts.Accounts[0].ID
	openedDate := userAccounts.Accounts[0].OpenedDate

	rawPortfolio, err := calc.apiClient.GetPortfolio(token, accountID)
	fullPortfolio := convertToFullPortfolio(rawPortfolio)

	// Заполняем ExpectedYield всего портфеля
	coeff, err := DivideMoneyValue(
		domain.MoneyValue{
			Units: fullPortfolio.ExpectedYieldRelative.Units,
			Nano:  fullPortfolio.ExpectedYieldRelative.Nano,
		},
		domain.MoneyValue{Units: "100", Nano: 0},
	)
	if err != nil {
		return domain.UserFullPortfolio{}, err
	}
	fullPortfolio.ExpectedYield = MultiplyMoneyValue(fullPortfolio.TotalAmountPortfolio, coeff)

	// Получаем дивиденды для всего портфеля
	fullPortfolio.AllDividends, err = calc.GetDividends(token, accountID, "", openedDate, time.Now().UTC())
	log.Printf("AllDividends length: %d", len(fullPortfolio.AllDividends))

	var wg sync.WaitGroup
	for i := range fullPortfolio.Positions {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pos := &fullPortfolio.Positions[i]

			// Заполняем ExpectedYieldRelative для каждой позиции
			posAmount := MultiplyMoneyValue(pos.AveragePositionPrice,
				domain.MoneyValue{
					Units: pos.Quantity.Units,
					Nano:  pos.Quantity.Nano,
				},
			)
			pos.ExpectedYieldRelative, err = DivideQuotation(
				domain.Quotation{
					Units: pos.ExpectedYield.Units,
					Nano:  pos.ExpectedYield.Nano,
				},
				domain.Quotation{
					Units: posAmount.Units,
					Nano:  posAmount.Nano,
				},
			)
			if err != nil {
				log.Printf("Failed to calculate ExpectedYieldRelative for %s: %v", pos.Ticker, err)
				return
			}
			pos.ExpectedYieldRelative = MultiplyQuotation(pos.ExpectedYieldRelative, domain.Quotation{Units: "100", Nano: 0})

			// Заполняем DailyYieldRelative
			// pos.DailyYieldRelative =

			// Получаем дивиденды и заполняем TotalYield и TotalYieldRelative для каждой позиции
			divs, err := calc.GetDividends(token, accountID, pos.InstrumentUID, openedDate, time.Now().UTC())
			if err != nil {
				log.Printf("Failed to get dividends for %s: %v", pos.Ticker, err)
				return
			}
			pos.Dividends = divs[pos.Ticker]
			pos.TotalYield = AddMoneyValue(pos.ExpectedYield, pos.Dividends)

			pos.TotalYieldRelative, err = DivideQuotation(
				domain.Quotation{
					Units: pos.TotalYield.Units,
					Nano:  pos.TotalYield.Nano,
				},
				domain.Quotation{
					Units: posAmount.Units,
					Nano:  posAmount.Nano,
				},
			)
			if err != nil {
				log.Printf("Failed to calculate TotalYieldRelative for %s: %v", pos.Ticker, err)
				return
			}
			pos.TotalYieldRelative = MultiplyQuotation(pos.TotalYieldRelative, domain.Quotation{Units: "100", Nano: 0})
		}(i)
	}
	wg.Wait()

	log.Printf("Время выполнения GetFullPortfolio: %.2f сек\n", time.Since(t).Seconds())
	return fullPortfolio, nil
}

func (calc *Calculator) GetDividends(
	token string,
	accountID string,
	instrumentId string,
	from time.Time,
	to time.Time) (map[string]domain.MoneyValue, error) {

	operations, err := calc.apiClient.GetUserOperationsByCursor(
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
			current := result[item.Ticker]
			result[item.Ticker] = AddMoneyValue(current, item.Payment)
		}
	}
	return result, nil
}
