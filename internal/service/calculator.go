package service

import (
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/pkg"
)

type Calculator struct {
	apiClient *api.Client
}

func NewCalculator(apiClient *api.Client) *Calculator {
	return &Calculator{apiClient: apiClient}
}

func (calc *Calculator) GetFullPortfolio(token string) (pkg.UserFullPortfolio, error) {
	userAccounts, err := calc.apiClient.GetBankAccount(token)
	if err != nil {
		return pkg.UserFullPortfolio{}, err
	}
	accountID := userAccounts.Accounts[0].ID
	openedDate := userAccounts.Accounts[0].OpenedDate
	rawPortfolio, err := calc.apiClient.GetPortfolio(token, accountID)
	fullPortfolio := convertToFullPortfolio(rawPortfolio)

	for i := range fullPortfolio.Positions {
		pos := &fullPortfolio.Positions[i]
		divs, err := calc.GetDividends(token, accountID, pos.InstrumentUID, openedDate, time.Now().UTC())
		if err != nil {
			return pkg.UserFullPortfolio{}, err
		}
		pos.Dividends = divs[pos.Ticker]
		pos.TotalYield = AddQuotations(pos.ExpectedYield, pos.Dividends)
	}

	return fullPortfolio, nil
}

func (calc *Calculator) GetDividends(
	token string,
	accountID string,
	instrumentId string,
	from time.Time,
	to time.Time) (map[string]pkg.Quotation, error) {

	operations, err := calc.apiClient.GetUserOperations(
		token,
		accountID,
		instrumentId,
		from,
		to,
		[]pkg.OperationType{pkg.OperationTypeDividend, pkg.OperationTypeCoupon},
		pkg.OperationStateExecuted,
	)
	if err != nil {
		return nil, err
	}

	result := make(map[string]pkg.Quotation)
	for _, block := range operations {
		for _, item := range block.Items {
			key := item.Ticker
			if key == "" {
				continue
			}
			current := result[item.Ticker]
			result[item.Ticker] = AddQuotations(current, pkg.Quotation{Units: item.Payment.Units, Nano: item.Payment.Nano})
		}
	}
	return result, nil
}
