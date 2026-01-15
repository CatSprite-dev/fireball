package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	client    Client
	accountID string
}

func NewConfig() Config {
	client := NewClient(5 * time.Second)
	account, err := client.GetBankAccount()
	if err != nil {
		log.Fatal(err)
	}

	cfg := Config{
		client:    *client,
		accountID: account.Accounts[0].ID,
	}

	cfg.saveAccount()
	return cfg
}

func (cfg *Config) GetUserInfo() (UserInfo, error) {
	userUrl := cfg.client.baseURL + ".UsersService/GetInfo"
	err := godotenv.Load()
	if err != nil {
		return UserInfo{}, fmt.Errorf("error loading .env file: %s", err)
	}
	token := os.Getenv("token")

	value, ok := cfg.client.cache.Get(userUrl)
	if ok {
		var user UserInfo
		err := json.Unmarshal(value, &user)
		if err != nil {
			return UserInfo{}, fmt.Errorf("unmarshal error: %s", err)
		}
		return user, nil
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
	err := godotenv.Load()
	if err != nil {
		return UserPortfolio{}, fmt.Errorf("error loading .env file: %s", err)
	}
	token := os.Getenv("token")

	value, ok := cfg.client.cache.Get(userUrl)
	if ok {
		var userPortfolio UserPortfolio
		err := json.Unmarshal(value, &userPortfolio)
		if err != nil {
			return UserPortfolio{}, fmt.Errorf("unmarshal error: %s", err)
		}
		return userPortfolio, nil
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
	err := godotenv.Load()
	if err != nil {
		return []UserOperations{}, fmt.Errorf("error loading .env file: %s", err)
	}
	token := os.Getenv("token")

	allOperations := []UserOperations{}
	cursor := ""
	limit := 1000
	for {
		payload := fmt.Sprintf(`{"accountId": "%s", 
			"from": "%s", 
			"to": "%s", 
			"cursor": "%s", 
			"limit": "%d"}, 
			"operationTypes": ["OPERATION_TYPE_INPUT", "OPERATION_TYPE_OUTPUT],
			"state":"OPERATION_STATE_EXECUTED"`,
			accountId, from.Format(time.RFC3339), to.Format(time.RFC3339), cursor, limit)

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
	accounts, err := cfg.client.GetBankAccount()
	if err != nil {
		return 0, fmt.Errorf("error fetching bank accounts: %s", err)
	}
	accountID := accounts.Accounts[0].ID
	openDate := accounts.Accounts[0].OpenedDate

	userOperations, err := cfg.GetUserOperations(accountID, openDate, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("error fetching user operations: %s", err)
	}

	totalUnits := 0
	totalNanos := 0
	var totalDeposits float64
	for _, operation := range userOperations {
		for _, item := range operation.Items {
			if item.Type == "OPERATION_TYPE_INPUT" || item.Type == "OPERATION_TYPE_OUTPUT" {
				unit, err := strconv.Atoi(item.Payment.Units)
				if err != nil {
					return 0, err
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
