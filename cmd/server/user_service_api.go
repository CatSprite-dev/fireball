package main

import (
	"encoding/json"
	"fmt"
	"strconv"
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

func (cfg *Config) GetUserOperations(
	token string,
	accountId string,
	instrumentId string,
	from time.Time,
	to time.Time,
	operationTypes []OperationType,
	operationState OperationState) ([]UserOperations, error) {

	userUrl := cfg.client.baseURL + ".OperationsService/GetOperationsByCursor"

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

func (cfg *Config) GetDividends(token string) (map[string]float64, error) {

	operations, err := cfg.GetUserOperations(
		token,
		cfg.accountID,
		"",
		cfg.openedDate,
		time.Now().UTC(),
		[]OperationType{OperationTypeDividend, OperationTypeCoupon},
		OperationStateExecuted,
	)
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64)

	for _, block := range operations {
		for _, item := range block.Items {
			key := item.Figi
			if key == "" {
				continue
			}
			units, _ := strconv.ParseFloat(item.Payment.Units, 64)
			payment := units + (float64(item.Payment.Nano) / 1000000000)
			result[item.Ticker] += payment
		}
	}

	return result, nil
}
