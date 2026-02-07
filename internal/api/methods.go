package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/CatSprite-dev/fireball/internal/pkg"
)

func (client *Client) GetBankAccount(token string) (UserAccounts, error) {
	userUrl := client.baseURL + ".UsersService/GetAccounts"

	payload := `{"status": "ACCOUNT_STATUS_OPEN"}`
	data, err := client.DoRequest(userUrl, token, payload)
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

func (client *Client) GetUserInfo(token string) (UserInfo, error) {
	userUrl := client.baseURL + ".UsersService/GetInfo"

	payload := `{}`
	data, err := client.DoRequest(userUrl, token, payload)
	if err != nil {
		return UserInfo{}, err
	}

	var user UserInfo
	err = json.Unmarshal(data, &user)
	if err != nil {
		return UserInfo{}, fmt.Errorf("unmarshal error: %s", err)
	}
	return user, nil
}

func (client *Client) GetPortfolio(token string, accountID string) (UserPortfolio, error) {
	userUrl := client.baseURL + ".OperationsService/GetPortfolio"

	payload := fmt.Sprintf(`{"accountId": "%s"}`, accountID)
	data, err := client.DoRequest(userUrl, token, payload)
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

func (client *Client) GetUserOperations(
	token string,
	accountId string,
	instrumentId string,
	from time.Time,
	to time.Time,
	operationTypes []pkg.OperationType,
	operationState pkg.OperationState) ([]UserOperations, error) {

	userUrl := client.baseURL + ".OperationsService/GetOperationsByCursor"
	if strings.Contains(client.baseURL, "sandbox") {
		userUrl = client.baseURL + ".OperationsService/GetSandboxOperationsByCursor"
	}

	allOperations := []UserOperations{}
	cursor := ""
	limit := 1000

	operationTypesJSON, err := json.Marshal(operationTypes)
	if err != nil {
		return []UserOperations{}, err
	}
	operationTypesStr := string(operationTypesJSON)
	for {
		payload := fmt.Sprintf(`{
			"accountId": "%s",
			"instrumentId": "%s",
			"from": "%s",
			"to": "%s",
			"cursor": "%s",
			"limit": %d,
			"operationTypes": %s,
			"state": "%s"
		}`,
			accountId,
			instrumentId,
			from.Format(time.RFC3339),
			to.Format(time.RFC3339),
			cursor,
			limit,
			operationTypesStr,
			operationState)
		data, err := client.DoRequest(userUrl, token, payload)
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
