package domain

import (
	"time"
)

type Instrument struct {
	AssetUID              string     `json:"assetUid"`
	Figi                  string     `json:"figi"`
	DshortMin             Quotation  `json:"dshortMin"`
	CountryOfRisk         string     `json:"countryOfRisk"`
	Lot                   int        `json:"lot"`
	UID                   string     `json:"uid"`
	RequiredTests         []string   `json:"requiredTests"`
	BlockedTcaFlag        bool       `json:"blockedTcaFlag"`
	Dlong                 Quotation  `json:"dlong"`
	DlongClient           Quotation  `json:"dlongClient"`
	SellAvailableFlag     bool       `json:"sellAvailableFlag"`
	Currency              string     `json:"currency"`
	First1DayCandleDate   time.Time  `json:"first1dayCandleDate"`
	Brand                 MoneyValue `json:"brand"`
	BuyAvailableFlag      bool       `json:"buyAvailableFlag"`
	WeekendFlag           bool       `json:"weekendFlag"`
	ClassCode             string     `json:"classCode"`
	Ticker                string     `json:"ticker"`
	InstrumentType        string     `json:"instrumentType"`
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
