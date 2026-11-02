package tinvest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Portfolio struct {
	TotalAmountShares     MoneyValue        `json:"totalAmountShares"`
	TotalAmountBonds      MoneyValue        `json:"totalAmountBonds"`
	TotalAmountEtf        MoneyValue        `json:"totalAmountEtf"`
	TotalAmountCurrencies MoneyValue        `json:"totalAmountCurrencies"`
	TotalAmountFutures    MoneyValue        `json:"totalAmountFutures"`
	ExpectedYield         Quotation         `json:"expectedYield"`
	Positions             []Position        `json:"positions"`
	AccountID             string            `json:"accountId"`
	TotalAmountOptions    MoneyValue        `json:"totalAmountOptions"`
	TotalAmountSp         MoneyValue        `json:"totalAmountSp"`
	TotalAmountPortfolio  MoneyValue        `json:"totalAmountPortfolio"`
	VirtualPositions      []VirtualPosition `json:"virtualPositions"`
	DailyYield            MoneyValue        `json:"dailyYield"`
	DailyYieldRelative    Quotation         `json:"dailyYieldRelative"`
}

type Position struct {
	Figi                     string     `json:"figi"`
	InstrumentType           string     `json:"instrumentType"`
	Quantity                 Quotation  `json:"quantity"`
	AveragePositionPrice     MoneyValue `json:"averagePositionPrice"`
	ExpectedYield            Quotation  `json:"expectedYield"`
	AveragePositionPricePt   Quotation  `json:"averagePositionPricePt"`
	CurrentPrice             MoneyValue `json:"currentPrice"`
	AveragePositionPriceFifo MoneyValue `json:"averagePositionPriceFifo"`
	QuantityLots             Quotation  `json:"quantityLots"`
	Blocked                  bool       `json:"blocked"`
	BlockedLots              Quotation  `json:"blockedLots"`
	PositionUID              string     `json:"positionUid"`
	InstrumentUID            string     `json:"instrumentUid"`
	VarMargin                MoneyValue `json:"varMargin"`
	ExpectedYieldFifo        Quotation  `json:"expectedYieldFifo"`
	DailyYield               MoneyValue `json:"dailyYield"`
	Ticker                   string     `json:"ticker"`
	ClassCode                string     `json:"classCode"`
	CurrentNkd               MoneyValue `json:"currentNkd"`
}

type VirtualPosition struct {
	ClassCode                string     `json:"classCode"`
	InstrumentType           string     `json:"instrumentType"`
	Ticker                   string     `json:"ticker"`
	Quantity                 Quotation  `json:"quantity"`
	ExpectedYieldFifo        Quotation  `json:"expectedYieldFifo"`
	AveragePositionPriceFifo MoneyValue `json:"averagePositionPriceFifo"`
	PositionUID              string     `json:"positionUid"`
	CurrentPrice             MoneyValue `json:"currentPrice"`
	Figi                     string     `json:"figi"`
	AveragePositionPrice     MoneyValue `json:"averagePositionPrice"`
	InstrumentUID            string     `json:"instrumentUid"`
	DailyYield               MoneyValue `json:"dailyYield"`
	ExpectedYield            Quotation  `json:"expectedYield"`
	ExpireDate               time.Time  `json:"expireDate"`
}

func (c *Client) GetPortfolio(ctx context.Context, accountID string, currency string) (*Portfolio, error) {
	type request struct {
		AccountID string `json:"accountId"`
		Currency  string `json:"currency"`
	}

	const method = "OperationsService/GetPortfolio"

	var response Portfolio
	res, err := c.DoRequest(ctx, method, request{
		AccountID: accountID,
		Currency:  currency,
	})
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(res, &response); err != nil {
		return nil, fmt.Errorf("%s: unmarshal response: %w", method, err)
	}

	return &response, nil
}
