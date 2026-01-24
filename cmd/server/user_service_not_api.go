package main

import (
	"fmt"
	"strings"
)

/*
func GetTotalDeposits(cfg *Config) (float64, error) {
	accountID := cfg.accountID
	openDate := cfg.openedDate

	userOperations, err := cfg.GetUserOperations(accountID, openDate, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("error fetching user operations: %s", err)
	}

	totalUnits := 0
	totalNanos := 0
	var totalDeposits float64
	for _, operation := range userOperations {
		for _, item := range operation.Items {
			if item.Type == string(OperationTypeInput) || item.Type == string(OperationTypeOutput) {
				unit, err := strconv.Atoi(item.Payment.Units)
				if err != nil {
					return 0, fmt.Errorf("error converting units to int: %s", err)
				}
				totalUnits += unit
				totalNanos += item.Payment.Nano
			}
		}
	}

	totalDeposits = float64(totalUnits) + (float64(totalNanos) / 1000000000)
	return totalDeposits, nil
}


func GetTotalReturn(cfg *Config, token string) (float64, error) {
	// Функция возвращает общую доходность за всё время существования портфеля
	// Формула расчета:
	// сумма всех вложений / акутальная стоимость портфеля
	// возвращает доходность от 0 до 1 в формате float64

	userPortfolio, err := cfg.GetPortfolio(token)
	if err != nil {
		return 0.0, err
	}
	units, err := strconv.Atoi(userPortfolio.TotalAmountPortfolio.Units)
	if err != nil {
		return 0.0, err
	}
	totalAmount := float64(units) + (float64(userPortfolio.TotalAmountPortfolio.Nano) / 1000000000)

	totalDeposits, err := GetTotalDeposits(cfg)
	if err != nil {
		return 0.0, err
	}

	totalReturn := ((totalAmount / totalDeposits) - 1)
	return totalReturn, nil
}
*/

func GetPositionsInfo(cfg *Config, token string) (string, error) {
	userPortfolio, err := cfg.GetPortfolio(token)
	if err != nil {
		return "", err
	}
	var results strings.Builder
	for _, position := range userPortfolio.Positions {
		results.WriteString(
			fmt.Sprintf("%s, Quantity: %s, Average Price: %s %s\n",
				position.Ticker,
				position.Quantity.Units,
				position.AveragePositionPrice.Units,
				position.AveragePositionPrice.Currency,
			),
		)
	}
	portfolioString := results.String()
	return portfolioString, nil
}
