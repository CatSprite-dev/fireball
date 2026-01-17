package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func (cfg *Config) GetUserInfo(token string) (UserInfo, error) {
	userUrl := cfg.client.baseURL + ".UsersService/GetInfo"
	//temporary
	if token == "" {
		cfgToken, err := cfg.client.getToken()
		if err != nil {
			return UserInfo{}, err
		}
		token = cfgToken
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

func (cfg *Config) GetPortfolio(token string) (UserPortfolio, error) {
	userUrl := cfg.client.baseURL + ".OperationsService/GetPortfolio"

	//token, err := cfg.client.getToken()
	//if err != nil {
	//	return UserPortfolio{}, err
	//}

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
