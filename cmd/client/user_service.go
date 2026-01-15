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

func (c *Client) GetUserOperations(accountId string, from time.Time, to time.Time) ([]UserOperations, error) {
	userUrl := c.baseURL + ".OperationsService/GetOperationsByCursor"
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

		fmt.Println(payload)
		data, err := c.DoRequest(userUrl, token, payload)
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

func (c *Client) GetTotalDeposits() (float64, error) {
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
