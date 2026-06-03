package domain

import (
	"time"
)

type Instrument struct {
	Figi           string `json:"figi"`
	Ticker         string `json:"ticker"`
	UID            string `json:"uid"`
	InstrumentType string `json:"instrumentType"`
	Name           string `json:"name"`
}

type Bond struct {
	PositionUID    string     `json:"positionUid"`
	Name           string     `json:"name"`
	Figi           string     `json:"figi"`
	UID            string     `json:"uid"`
	Nominal        MoneyValue `json:"nominal"`
	InitialNominal MoneyValue `json:"initialNominal"`
	Currency       string     `json:"currency"`
	AciValue       MoneyValue `json:"aciValue"`
	ClassCode      string     `json:"classCode"`
	Ticker         string     `json:"ticker"`
}

type IndicativeInstruments struct {
	Instruments []Instrument
}

type Candle struct {
	Time  time.Time `json:"time"`
	Close Quotation `json:"close"`
	Open  Quotation `json:"open"`
}

type ChartData struct {
	Times     []time.Time `json:"times"`
	Benchmark []Quotation `json:"benchmark"`
	Portfolio []Quotation `json:"portfolio"`
}

type UserOperations struct {
	Items []Item
}

type Item struct {
	BrokerAccountID string     `json:"brokerAccountId"`
	ID              string     `json:"id"`
	InstrumentName  string     `json:"name"`
	Date            time.Time  `json:"date"`
	Type            string     `json:"type"`
	Description     string     `json:"description"`
	State           string     `json:"state"`
	InstrumentUID   string     `json:"instrumentUid"`
	Figi            string     `json:"figi"`
	InstrumentType  string     `json:"instrumentType"`
	PositionUID     string     `json:"positionUid"`
	Ticker          string     `json:"ticker"`
	ClassCode       string     `json:"classCode"`
	Payment         MoneyValue `json:"payment"`
	InstrumentPrice MoneyValue `json:"price"`
	Commission      MoneyValue `json:"commission"`
	Yield           MoneyValue `json:"yield"`
	YieldRelative   Quotation  `json:"yieldRelative"`
	AccruedInt      MoneyValue `json:"accruedInt"`
	Quantity        string     `json:"quantity"`
}

type ClosePrice struct {
	Figi          string    `json:"figi"`
	InstrumentUID string    `json:"instrumentUid"`
	Ticker        string    `json:"ticker"`
	ClassCode     string    `json:"classCode"`
	ClosePrice    Quotation `json:"price"`
	Time          time.Time `json:"time"`
}

type Currency struct {
	Figi            string     `json:"figi"`
	Ticker          string     `json:"ticker"`
	ClassCode       string     `json:"classCode"`
	Isin            string     `json:"isin"`
	Lot             int        `json:"lot"`
	Currency        string     `json:"currency"`
	Name            string     `json:"name"`
	Exchange        string     `json:"exchange"`
	Nominal         MoneyValue `json:"nominal"`
	IsoCurrencyName string     `json:"isoCurrencyName"`
	UID             string     `json:"uid"`
	RealExchange    string     `json:"realExchange"`
	PositionUID     string     `json:"positionUid"`
	AssetUID        string     `json:"assetUid"`
}
