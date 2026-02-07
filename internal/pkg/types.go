// pkg/types.go
package pkg

import "time"

type UserFullPortfolio struct {
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
	Name                     string     `json:"name"`
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
	CurrentNkd               MoneyValue `json:"currentNkd,omitempty"`
	Dividends                Quotation  `json:"dividends"`
	TotalYield               Quotation  `json:"totalYield"`
}

type VirtualPosition struct {
	PositionUID              string     `json:"positionUid"`
	InstrumentUID            string     `json:"instrumentUid"`
	Figi                     string     `json:"figi"`
	InstrumentType           string     `json:"instrumentType"`
	Quantity                 Quotation  `json:"quantity"`
	AveragePositionPrice     MoneyValue `json:"averagePositionPrice"`
	ExpectedYield            Quotation  `json:"expectedYield"`
	ExpectedYieldFifo        Quotation  `json:"expectedYieldFifo"`
	ExpireDate               time.Time  `json:"expireDate"`
	CurrentPrice             MoneyValue `json:"currentPrice"`
	AveragePositionPriceFifo MoneyValue `json:"averagePositionPriceFifo"`
	DailyYield               MoneyValue `json:"dailyYield"`
	Ticker                   string     `json:"ticker"`
	ClassCode                string     `json:"classCode"`
}

type MoneyValue struct {
	Currency string `json:"currency"`
	Units    string `json:"units"`
	Nano     int    `json:"nano"`
}

type Quotation struct {
	Units string `json:"units"`
	Nano  int    `json:"nano"`
}
