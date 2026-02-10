package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/CatSprite-dev/fireball/internal/pkg"
)

func (client *Client) GetAccounts(token string, accountStatus pkg.AccountStatus) (UserAccounts, error) {
	type AccountsRequest struct {
		Status pkg.AccountStatus `json:"status,omitempty"`
	}

	url := client.baseURL + "/rest/tinkoff.public.invest.api.contract.v1.UsersService/GetAccounts"

	payload := AccountsRequest{Status: accountStatus}

	data, err := client.DoRequest(url, pkg.HTTPMethodPost, token, payload)
	if err != nil {
		return UserAccounts{}, fmt.Errorf("do request error (api.GetAccounts): %s", err)
	}

	var accounts UserAccounts
	err = json.Unmarshal(data, &accounts)
	if err != nil {
		return UserAccounts{}, fmt.Errorf("unmarshal error (api.GetAccounts): %s", err)
	}
	fmt.Println(len(accounts.Accounts))
	return accounts, nil
}

func (client *Client) GetInfo(token string) (UserInfo, error) {
	url := client.baseURL + "/rest/tinkoff.public.invest.api.contract.v1.UsersService/GetInfo"

	payload := `{}`
	data, err := client.DoRequest(url, pkg.HTTPMethodPost, token, payload)
	if err != nil {
		return UserInfo{}, err
	}

	var user UserInfo
	err = json.Unmarshal(data, &user)
	if err != nil {
		return UserInfo{}, fmt.Errorf("unmarshal error (api.GetInfo): %s", err)
	}
	return user, nil
}

func (client *Client) GetPortfolio(token string, accountID string) (UserPortfolio, error) {
	type PortfolioRequest struct {
		AccountID string `json:"accountId"`
	}

	url := client.baseURL + "/rest/tinkoff.public.invest.api.contract.v1.OperationsService/GetPortfolio"

	payload := PortfolioRequest{AccountID: accountID}
	data, err := client.DoRequest(url, pkg.HTTPMethodPost, token, payload)
	if err != nil {
		return UserPortfolio{}, fmt.Errorf("do request error (GetPortfolio): %s", err)
	}

	var userPortfolio UserPortfolio
	err = json.Unmarshal(data, &userPortfolio)
	if err != nil {
		return UserPortfolio{}, fmt.Errorf("unmarshal error (GetPortfolio): %s", err)
	}

	return userPortfolio, nil
}

func (client *Client) GetUserOperationsByCursor(
	token string,
	accountId string,
	instrumentId string,
	from *time.Time,
	to *time.Time,
	operationTypes []pkg.OperationType,
	operationState pkg.OperationState,
	WithoutCommissions bool) ([]UserOperations, error) {

	type OperationsRequest struct {
		AccountID          string              `json:"accountId"`
		InstrumentID       string              `json:"instrumentId,omitempty"`
		From               *time.Time          `json:"from,omitempty"`
		To                 *time.Time          `json:"to,omitempty"`
		Cursor             string              `json:"cursor,omitempty"`
		Limit              int32               `json:"limit,omitempty"`
		OperationTypes     []pkg.OperationType `json:"operationTypes,omitempty"`
		State              pkg.OperationState  `json:"state,omitempty"`
		WithoutCommissions bool                `json:"withoutCommissions,omitempty"` // True - если нужно исключить операции по списанию комиссий из тела ответа, false по умолчанию
		WithoutTrades      bool                `json:"withoutTrades,omitempty"`
		WithoutOvernights  bool                `json:"withoutOvernights,omitempty"`
	}

	url := client.baseURL + "/rest/tinkoff.public.invest.api.contract.v1.OperationsService/GetOperationsByCursor"

	cursor := ""
	allOperations := []UserOperations{}
	for {
		payload := OperationsRequest{
			AccountID:      accountId,
			InstrumentID:   instrumentId,
			From:           from,
			To:             to,
			Cursor:         cursor,
			Limit:          1000, // максимальный лимит для одного курсора
			OperationTypes: operationTypes,
			State:          operationState,
		}

		data, err := client.DoRequest(url, pkg.HTTPMethodPost, token, payload)
		if err != nil {
			return []UserOperations{}, fmt.Errorf("do request error (GetOperationsByCursor): %s", err)
		}

		var blockOfOperations UserOperations
		err = json.Unmarshal(data, &blockOfOperations)
		if err != nil {
			return []UserOperations{}, fmt.Errorf("unmarshal error (GetOperationsByCursor): %s", err)
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
