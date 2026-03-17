package domain

import "time"

type Instrument struct {
	Figi           string `json:"figi"`
	Ticker         string `json:"ticker"`
	UID            string `json:"uid"`
	InstrumentType string `json:"instrumentType"`
	Name           string `json:"name"`
}

type IndicativeInstruments struct {
	Instruments []Instrument
}

type Candle struct {
	Time       time.Time `json:"time"`
	Close      Quotation `json:"close"`
	Open       Quotation `json:"open"`
	High       Quotation `json:"high"`
	Low        Quotation `json:"low"`
	IsComplete bool      `json:"isComplete"`
}

type ChartData struct {
	IndexCandles     []Candle
	PortfolioCandles []Candle
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
