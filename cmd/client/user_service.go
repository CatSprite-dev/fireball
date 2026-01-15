package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

type Config struct {
	client     Client
	accountID  string
	openedDate time.Time
}

func NewConfig() Config {
	client := NewClient(5 * time.Second)
	account, err := client.GetBankAccount()
	if err != nil {
		log.Fatal(err)
	}

	cfg := Config{
		client:     *client,
		accountID:  account.Accounts[0].ID,
		openedDate: account.Accounts[0].OpenedDate,
	}

	cfg.saveAccount()
	return cfg
}

func (cfg *Config) GetUserInfo() (UserInfo, error) {
	userUrl := cfg.client.baseURL + ".UsersService/GetInfo"
	token, err := cfg.client.getToken()
	if err != nil {
		return UserInfo{}, err
	}

	payload := `{}`
	data, err := cfg.client.DoRequest(userUrl, token, payload)
	if err != nil {
		return UserInfo{}, fmt.Errorf("do request error: %s", err)
	}

	var user UserInfo
	err = json.Unmarshal(data, &user)
	if err != nil {
		return UserInfo{}, fmt.Errorf("unmarshal error: %s", err)
	}
	return user, nil
}

func (cfg *Config) GetPortfolio() (UserPortfolio, error) {
	userUrl := cfg.client.baseURL + ".OperationsService/GetPortfolio"
	token, err := cfg.client.getToken()
	if err != nil {
		return UserPortfolio{}, err
	}

	accountID := cfg.accountID
	payload := fmt.Sprintf(`{"accountId": "%s"}`, accountID)
	data, err := cfg.client.DoRequest(userUrl, token, payload)
	if err != nil {
		return UserPortfolio{}, fmt.Errorf("do request error: %s", err)
	}

	var userPortfolio UserPortfolio
	err = json.Unmarshal(data, &userPortfolio)
	if err != nil {
		return UserPortfolio{}, fmt.Errorf("unmarshal error: %s", err)
	}
	return userPortfolio, nil

}

func (cfg *Config) GetUserOperations(accountId string, from time.Time, to time.Time) ([]UserOperations, error) {
	userUrl := cfg.client.baseURL + ".OperationsService/GetOperationsByCursor"
	token, err := cfg.client.getToken()
	if err != nil {
		return []UserOperations{}, err
	}

	allOperations := []UserOperations{}
	cursor := ""
	limit := 1000
	for {
		payload := fmt.Sprintf(`{
			"accountId": "%s",
			"from": "%s",
			"to": "%s",
			"cursor": "%s",
			"limit": %d,
			"operationTypes": ["%s", "%s"],
			"state": "%s"
		}`,
			accountId,
			from.Format(time.RFC3339),
			to.Format(time.RFC3339),
			cursor,
			limit,
			OperationTypeInput,
			OperationTypeOutput,
			OperationStateExecuted)
		data, err := cfg.client.DoRequest(userUrl, token, payload)
		if err != nil {
			return []UserOperations{}, fmt.Errorf("do request error: %s", err)
		}

		var blockOfOperations UserOperations
		err = json.Unmarshal(data, &blockOfOperations)
		if err != nil {
			return []UserOperations{}, fmt.Errorf("unmarshal error: %s", err)
		}
		allOperations = append(allOperations, blockOfOperations)
		if blockOfOperations.HasNext {
			cursor = blockOfOperations.NextCursor
		} else {
			break
		}
	}
	return allOperations, nil
}

func (cfg *Config) GetTotalDeposits() (float64, error) {
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

func (cfg *Config) GetTotalReturn() (float64, error) {
	// Функция возвращает общую доходность за всё время существования портфеля
	// Формула расчета:
	// сумма всех вложений / акутальная стоимость портфеля
	// возвращает доходность от 0 до 1 в формате float64

	userPortfolio, err := cfg.GetPortfolio()
	if err != nil {
		return 0.0, err
	}
	units, err := strconv.Atoi(userPortfolio.TotalAmountPortfolio.Units)
	if err != nil {
		return 0.0, err
	}
	totalAmount := float64(units) + (float64(userPortfolio.TotalAmountPortfolio.Nano) / 1000000000)

	totalDeposits, err := cfg.GetTotalDeposits()
	if err != nil {
		return 0.0, err
	}

	totalReturn := ((totalAmount / totalDeposits) - 1)
	return totalReturn, nil
}
