package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func (c *Client) GetUserInfo() (UserInfo, error) {
	userUrl := c.baseURL + ".UsersService/GetInfo"
	err := godotenv.Load()
	if err != nil {
		return UserInfo{}, fmt.Errorf("error loading .env file: %s", err)
	}
	token := os.Getenv("token")

	value, ok := c.cache.Get(userUrl)
	if ok {
		var user UserInfo
		err := json.Unmarshal(value, &user)
		if err != nil {
			return UserInfo{}, fmt.Errorf("unmarshal error: %s", err)
		}
		return user, nil
	}

	data, err := c.DoRequest(userUrl, token, `{}`)
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

func (c *Client) GetBankAccounts() (UserAccounts, error) {
	userUrl := c.baseURL + ".UsersService/GetAccounts"
	err := godotenv.Load()
	if err != nil {
		return UserAccounts{}, fmt.Errorf("error loading .env file: %s", err)
	}
	token := os.Getenv("token")
	value, ok := c.cache.Get(userUrl)
	if ok {
		var accounts UserAccounts
		err := json.Unmarshal(value, &accounts)
		if err != nil {
			return UserAccounts{}, fmt.Errorf("unmarshal error: %s", err)
		}
		return accounts, nil
	}
	payload := `{"status": "ACCOUNT_STATUS_OPEN"}`
	data, err := c.DoRequest(userUrl, token, payload)
	if err != nil {
		return UserAccounts{}, fmt.Errorf("do request error: %s", err)
	}

	var accounts UserAccounts
	err = json.Unmarshal(data, &accounts)
	if err != nil {
		return UserAccounts{}, fmt.Errorf("unmarshal error: %s", err)
	}
	return accounts, nil
}

func (c *Client) GetUserOperations(accountId string, from time.Time, to time.Time) (UserOperation, error) {
	userUrl := c.baseURL + ".OperationsService/GetOperations"
	err := godotenv.Load()
	if err != nil {
		return UserOperation{}, fmt.Errorf("error loading .env file: %s", err)
	}
	token := os.Getenv("token")
	value, ok := c.cache.Get(userUrl)
	if ok {
		var operations UserOperation
		err := json.Unmarshal(value, &operations)
		if err != nil {
			return UserOperation{}, fmt.Errorf("unmarshal error: %s", err)
		}
		return operations, nil
	}
	payload := fmt.Sprintf(`{"accountId": "%s", "from": "%s", "to": "%s"}`, accountId, from.Format(time.RFC3339), to.Format(time.RFC3339))
	fmt.Println(payload)
	data, err := c.DoRequest(userUrl, token, payload)
	if err != nil {
		return UserOperation{}, fmt.Errorf("do request error: %s", err)
	}

	var operations UserOperation
	err = json.Unmarshal(data, &operations)
	if err != nil {
		return UserOperation{}, fmt.Errorf("unmarshal error: %s", err)
	}
	return operations, nil
}

func (c *Client) GetTotalDeposits() (int, error) {
	accounts, err := c.GetBankAccounts()
	if err != nil {
		return 0, fmt.Errorf("error fetching bank accounts: %s", err)
	}
	accountID := accounts.Accounts[0].ID
	openDate := accounts.Accounts[0].OpenedDate
	fmt.Println(accountID)
	fmt.Println(openDate)
	userOperations, err := c.GetUserOperations(accountID, openDate, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("error fetching user operations: %s", err)
	}
	totalDeposits := 0
	for _, operation := range userOperations.Operations {
		if operation.Type == "Пополнение брокерского счёта" {
			deposit, err := strconv.Atoi(operation.Payment.Units)
			if err != nil {
				return 0, err
			}
			totalDeposits += deposit
		}
	}
	return totalDeposits, nil
}
