package domain

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
