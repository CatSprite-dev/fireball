package tinvest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Currency struct {
	AssetUID              string     `json:"assetUid"`
	IsoCurrencyName       string     `json:"isoCurrencyName"`
	Figi                  string     `json:"figi"`
	DshortMin             Quotation  `json:"dshortMin"`
	CountryOfRisk         string     `json:"countryOfRisk"`
	Lot                   int        `json:"lot"`
	UID                   string     `json:"uid"`
	RequiredTests         []string   `json:"requiredTests"`
	BlockedTcaFlag        bool       `json:"blockedTcaFlag"`
	Dlong                 Quotation  `json:"dlong"`
	DlongClient           Quotation  `json:"dlongClient"`
	Nominal               MoneyValue `json:"nominal"`
	SellAvailableFlag     bool       `json:"sellAvailableFlag"`
	Currency              string     `json:"currency"`
	First1DayCandleDate   time.Time  `json:"first1dayCandleDate"`
	Brand                 Brand      `json:"brand"`
	BuyAvailableFlag      bool       `json:"buyAvailableFlag"`
	WeekendFlag           bool       `json:"weekendFlag"`
	ClassCode             string     `json:"classCode"`
	Ticker                string     `json:"ticker"`
	ForQualInvestorFlag   bool       `json:"forQualInvestorFlag"`
	ForIisFlag            bool       `json:"forIisFlag"`
	PositionUID           string     `json:"positionUid"`
	APITradeAvailableFlag bool       `json:"apiTradeAvailableFlag"`
	DlongMin              Quotation  `json:"dlongMin"`
	ShortEnabledFlag      bool       `json:"shortEnabledFlag"`
	Kshort                Quotation  `json:"kshort"`
	First1MinCandleDate   time.Time  `json:"first1minCandleDate"`
	MinPriceIncrement     Quotation  `json:"minPriceIncrement"`
	OtcFlag               bool       `json:"otcFlag"`
	DshortClient          Quotation  `json:"dshortClient"`
	Klong                 Quotation  `json:"klong"`
	Dshort                Quotation  `json:"dshort"`
	Name                  string     `json:"name"`
	Exchange              string     `json:"exchange"`
	CountryOfRiskName     string     `json:"countryOfRiskName"`
	Isin                  string     `json:"isin"`
}

type getCurrenciesResponse struct {
	Instruments []Currency `json:"instruments"`
}

type InstrumentStatus string

const (
	InstrumentStatusBase InstrumentStatus = "INSTRUMENT_STATUS_BASE"
	InstrumentStatusAll  InstrumentStatus = "INSTRUMENT_STATUS_ALL"
)

func (c *Client) GetCurrencies(ctx context.Context, status InstrumentStatus) ([]Currency, error) {
	const method = "InstrumentsService/Currencies"

	type request struct {
		InstrumentStatus InstrumentStatus `json:"instrumentStatus"`
	}

	data, err := c.DoRequest(ctx, method, request{InstrumentStatus: status})
	if err != nil {
		return nil, err
	}

	var response getCurrenciesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("%s: unmarshal response: %w", method, err)
	}

	return response.Instruments, nil
}
