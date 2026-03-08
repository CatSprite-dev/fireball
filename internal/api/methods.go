package api

import (
	"context"
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

	client.usersLimiter.Wait(context.Background())

	data, err := client.DoRequest(url, pkg.HTTPMethodPost, token, payload)
	if err != nil {
		return UserAccounts{}, fmt.Errorf("do request error (api.GetAccounts): %w", err)
	}

	var accounts UserAccounts
	err = json.Unmarshal(data, &accounts)
	if err != nil {
		return UserAccounts{}, fmt.Errorf("unmarshal error (api.GetAccounts): %w", err)
	}
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
		return UserInfo{}, fmt.Errorf("unmarshal error (api.GetInfo): %w", err)
	}
	return user, nil
}

func (client *Client) GetPortfolio(token string, accountID string) (UserPortfolio, error) {
	type PortfolioRequest struct {
		AccountID string `json:"accountId"`
	}

	url := client.baseURL + "/rest/tinkoff.public.invest.api.contract.v1.OperationsService/GetPortfolio"

	payload := PortfolioRequest{AccountID: accountID}

	client.operationsLimiter.Wait(context.Background())

	data, err := client.DoRequest(url, pkg.HTTPMethodPost, token, payload)
	if err != nil {
		return UserPortfolio{}, fmt.Errorf("do request error (GetPortfolio): %w", err)
	}

	var userPortfolio UserPortfolio
	err = json.Unmarshal(data, &userPortfolio)
	if err != nil {
		return UserPortfolio{}, fmt.Errorf("unmarshal error (GetPortfolio): %w", err)
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

		client.operationsLimiter.Wait(context.Background())

		data, err := client.DoRequest(url, pkg.HTTPMethodPost, token, payload)
		if err != nil {
			return []UserOperations{}, fmt.Errorf("do request error (GetOperationsByCursor): %w", err)
		}

		var blockOfOperations UserOperations
		err = json.Unmarshal(data, &blockOfOperations)
		if err != nil {
			return []UserOperations{}, fmt.Errorf("unmarshal error (GetOperationsByCursor): %w", err)
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

func (client *Client) GetInstrumentBy(token string, idType pkg.InstrumentIdType, classCode pkg.ClassCode, id string) (Instrument, error) {
	type InstrumentRequest struct {
		IDType    pkg.InstrumentIdType `json:"idType"`
		ClassCode pkg.ClassCode        `json:"classCode,omitempty"`
		ID        string               `json:"id"`
	}

	if idType == pkg.InstrumentIdTypeTicker && classCode == pkg.ClassCodeUnspecified {
		return Instrument{}, fmt.Errorf("classCode is required when idType is TICKER")
	}

	url := client.baseURL + "/rest/tinkoff.public.invest.api.contract.v1.InstrumentsService/GetInstrumentBy"

	payload := InstrumentRequest{
		IDType:    idType,
		ClassCode: classCode,
		ID:        id,
	}

	client.instrumentsLimiter.Wait(context.Background())

	data, err := client.DoRequest(url, pkg.HTTPMethodPost, token, payload)
	if err != nil {
		return Instrument{}, fmt.Errorf("do request error (GetInstrumentBy): %w", err)
	}

	var instrument Instrument
	err = json.Unmarshal(data, &instrument)
	if err != nil {
		return Instrument{}, fmt.Errorf("unmarshal error (GetInstrumentBy): %w", err)
	}

	return instrument, nil
}

func (client *Client) Indicatives(token string) (IndicativeInstruments, error) {
	url := client.baseURL + "/rest/tinkoff.public.invest.api.contract.v1.InstrumentsService/Indicatives"

	payload := `{}`
	data, err := client.DoRequest(url, pkg.HTTPMethodPost, token, payload)
	if err != nil {
		return IndicativeInstruments{}, err
	}
	var indicativeInstruments IndicativeInstruments
	err = json.Unmarshal(data, &indicativeInstruments)
	if err != nil {
		return IndicativeInstruments{}, fmt.Errorf("unmarshal error (api.Indicatives): %w", err)
	}
	return indicativeInstruments, nil
}

func (client *Client) GetCandles(token string,
	from *time.Time,
	to *time.Time,
	interval pkg.CandleInterval,
	instrumentId string,
	candleSourceType pkg.CandleSource,
	limit int) (Candles, error) {

	type GetCandlesRequest struct {
		From             *time.Time         `json:"from"`
		To               *time.Time         `json:"to"`
		Interval         pkg.CandleInterval `json:"interval"`
		InstrumentID     string             `json:"instrumentId"`
		CandleSourceType pkg.CandleSource   `json:"candleSourceType,omitempty"`
		Limit            int                `json:"limit,omitempty"`
	}

	url := client.baseURL + "/rest/tinkoff.public.invest.api.contract.v1.MarketDataService/GetCandles"

	payload := GetCandlesRequest{
		From:             from,
		To:               to,
		Interval:         interval,
		InstrumentID:     instrumentId,
		CandleSourceType: candleSourceType,
		Limit:            limit,
	}

	data, err := client.DoRequest(url, pkg.HTTPMethodPost, token, payload)
	if err != nil {
		return Candles{}, fmt.Errorf("do request error (GetCandles): %s", err)
	}

	var candles Candles
	err = json.Unmarshal(data, &candles)
	if err != nil {
		return Candles{}, fmt.Errorf("unmarshal error (GetCandles): %s", err)
	}

	return candles, nil
}
