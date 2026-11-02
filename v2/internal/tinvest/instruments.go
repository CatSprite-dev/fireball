package tinvest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Instrument struct {
	AssetUID              string    `json:"assetUid"`
	Figi                  string    `json:"figi"`
	DshortMin             Quotation `json:"dshortMin"`
	CountryOfRisk         string    `json:"countryOfRisk"`
	Lot                   int       `json:"lot"`
	UID                   string    `json:"uid"`
	RequiredTests         []string  `json:"requiredTests"`
	BlockedTcaFlag        bool      `json:"blockedTcaFlag"`
	Dlong                 Quotation `json:"dlong"`
	DlongClient           Quotation `json:"dlongClient"`
	SellAvailableFlag     bool      `json:"sellAvailableFlag"`
	Currency              string    `json:"currency"`
	First1DayCandleDate   time.Time `json:"first1dayCandleDate"`
	Brand                 Brand     `json:"brand"`
	BuyAvailableFlag      bool      `json:"buyAvailableFlag"`
	WeekendFlag           bool      `json:"weekendFlag"`
	ClassCode             string    `json:"classCode"`
	Ticker                string    `json:"ticker"`
	InstrumentType        string    `json:"instrumentType"`
	ForQualInvestorFlag   bool      `json:"forQualInvestorFlag"`
	ForIisFlag            bool      `json:"forIisFlag"`
	PositionUID           string    `json:"positionUid"`
	APITradeAvailableFlag bool      `json:"apiTradeAvailableFlag"`
	DlongMin              Quotation `json:"dlongMin"`
	ShortEnabledFlag      bool      `json:"shortEnabledFlag"`
	Kshort                Quotation `json:"kshort"`
	First1MinCandleDate   time.Time `json:"first1minCandleDate"`
	MinPriceIncrement     Quotation `json:"minPriceIncrement"`
	OtcFlag               bool      `json:"otcFlag"`
	DshortClient          Quotation `json:"dshortClient"`
	Klong                 Quotation `json:"klong"`
	Dshort                Quotation `json:"dshort"`
	Name                  string    `json:"name"`
	Exchange              string    `json:"exchange"`
	CountryOfRiskName     string    `json:"countryOfRiskName"`
	Isin                  string    `json:"isin"`
}

type Brand struct {
	LogoName      string `json:"logoName"`
	LogoBaseColor string `json:"logoBaseColor"`
	TextColor     string `json:"textColor"`
}

type getInstrumentByResponse struct {
	Instrument Instrument `json:"instrument"`
}

type InstrumentIDType string

const (
	InstrumentIDUID    InstrumentIDType = "INSTRUMENT_ID_TYPE_UID"
	InstrumentIDFigi   InstrumentIDType = "INSTRUMENT_ID_TYPE_FIGI"
	InstrumentIDTicker InstrumentIDType = "INSTRUMENT_ID_TYPE_TICKER"
)

func (c *Client) GetInstrumentBy(ctx context.Context, idType InstrumentIDType, classCode, id string) (*Instrument, error) {
	const method = "InstrumentsService/GetInstrumentBy"

	type request struct {
		IDType    InstrumentIDType `json:"idType"`
		ClassCode string           `json:"classCode,omitempty"`
		ID        string           `json:"id"`
	}

	var response getInstrumentByResponse
	res, err := c.DoRequest(ctx, method, request{
		IDType:    idType,
		ClassCode: classCode,
		ID:        id,
	})
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(res, &response); err != nil {
		return nil, fmt.Errorf("%s: unmarshal response: %w", method, err)
	}

	return &response.Instrument, nil
}
