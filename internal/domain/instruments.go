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

type Candles struct {
	Candles []Candle `json:"candles"`
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
	IndexCandles     Candles
	PortfolioCandles Candles
}
