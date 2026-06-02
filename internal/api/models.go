package api

import (
	"time"
)

type MoneyValue struct {
	Currency string `json:"currency"`
	Units    string `json:"units"`
	Nano     int    `json:"nano"`
}

type Quotation struct {
	Units string `json:"units"`
	Nano  int    `json:"nano"`
}

type UserInfo struct {
	QualifiedForWorkWith []string `json:"qualifiedForWorkWith"`
	RiskLevelCode        string   `json:"riskLevelCode"`
	QualStatus           bool     `json:"qualStatus"`
	PremStatus           bool     `json:"premStatus"`
	Tariff               string   `json:"tariff"`
	UserID               string   `json:"userId"`
}

type UserAccounts struct {
	Accounts []struct {
		ID          string    `json:"id"`
		Type        string    `json:"type"`
		Name        string    `json:"name"`
		Status      string    `json:"status"`
		OpenedDate  time.Time `json:"openedDate"`
		ClosedDate  time.Time `json:"closedDate"`
		AccessLevel string    `json:"accessLevel"`
	} `json:"accounts"`
}

type UserOperations struct {
	HasNext    bool   `json:"hasNext"`
	NextCursor string `json:"nextCursor"`
	Items      []Item `json:"items"`
}

type Item struct {
	Cursor            string     `json:"cursor"`
	BrokerAccountID   string     `json:"brokerAccountId"`
	ID                string     `json:"id"`
	ParentOperationID string     `json:"parentOperationId"`
	Name              string     `json:"name"`
	Date              time.Time  `json:"date"`
	Type              string     `json:"type"`
	Description       string     `json:"description"`
	State             string     `json:"state"`
	InstrumentUID     string     `json:"instrumentUid"`
	Figi              string     `json:"figi"`
	InstrumentType    string     `json:"instrumentType"`
	InstrumentKind    string     `json:"instrumentKind"`
	PositionUID       string     `json:"positionUid"`
	Ticker            string     `json:"ticker"`
	ClassCode         string     `json:"classCode"`
	Payment           MoneyValue `json:"payment"`
	Price             MoneyValue `json:"price"`
	Commission        MoneyValue `json:"commission"`
	Yield             MoneyValue `json:"yield"`
	YieldRelative     Quotation  `json:"yieldRelative"`
	AccruedInt        MoneyValue `json:"accruedInt"`
	Quantity          string     `json:"quantity"`
	QuantityRest      string     `json:"quantityRest"`
	QuantityDone      string     `json:"quantityDone"`
	CancelReason      string     `json:"cancelReason"`
	AssetUID          string     `json:"assetUid"`
	ChildOperations   []struct {
		InstrumentUID string     `json:"instrumentUid"`
		Payment       MoneyValue `json:"payment"`
	} `json:"childOperations"`
}

type UserPortfolio struct {
	TotalAmountShares     MoneyValue `json:"totalAmountShares"`
	TotalAmountBonds      MoneyValue `json:"totalAmountBonds"`
	TotalAmountEtf        MoneyValue `json:"totalAmountEtf"`
	TotalAmountCurrencies MoneyValue `json:"totalAmountCurrencies"`
	TotalAmountFutures    MoneyValue `json:"totalAmountFutures"`
	ExpectedYield         Quotation  `json:"expectedYield"`
	Positions             []Position `json:"positions"`
	AccountID             string     `json:"accountId"`
	TotalAmountOptions    MoneyValue `json:"totalAmountOptions"`
	TotalAmountSp         MoneyValue `json:"totalAmountSp"`
	TotalAmountPortfolio  MoneyValue `json:"totalAmountPortfolio"`
	VirtualPositions      []any      `json:"virtualPositions"`
	DailyYield            MoneyValue `json:"dailyYield"`
	DailyYieldRelative    Quotation  `json:"dailyYieldRelative"`
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
	CurrentNkd               MoneyValue `json:"currentNkd,omitempty"`
}

type IndicativeInstruments struct {
	Instruments []struct {
		Figi              string `json:"figi"`
		Ticker            string `json:"ticker"`
		ClassCode         string `json:"classCode"`
		Currency          string `json:"currency"`
		InstrumentKind    string `json:"instrumentKind"`
		Name              string `json:"name"`
		Exchange          string `json:"exchange"`
		UID               string `json:"uid"`
		BuyAvailableFlag  bool   `json:"buyAvailableFlag"`
		SellAvailableFlag bool   `json:"sellAvailableFlag"`
	} `json:"instruments"`
}

type InstrumentResponse struct {
	Instrument Instrument `json:"instrument"`
}

type Instrument struct {
	AssetUID            string    `json:"assetUid"`
	Figi                string    `json:"figi"`
	DshortMin           Quotation `json:"dshortMin"`
	CountryOfRisk       string    `json:"countryOfRisk"`
	Lot                 int       `json:"lot"`
	UID                 string    `json:"uid"`
	RequiredTests       []string  `json:"requiredTests"`
	BlockedTcaFlag      bool      `json:"blockedTcaFlag"`
	Dlong               Quotation `json:"dlong"`
	DlongClient         Quotation `json:"dlongClient"`
	SellAvailableFlag   bool      `json:"sellAvailableFlag"`
	Currency            string    `json:"currency"`
	First1DayCandleDate time.Time `json:"first1dayCandleDate"`
	Brand               struct {
		LogoName      string `json:"logoName"`
		LogoBaseColor string `json:"logoBaseColor"`
		TextColor     string `json:"textColor"`
	} `json:"brand"`
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

type InstrumentBond struct {
	Bond `json:"instrument"`
}

type Bond struct {
	AssetUID            string     `json:"assetUid"`
	CallDate            time.Time  `json:"callDate"`
	CountryOfRisk       string     `json:"countryOfRisk"`
	BlockedTcaFlag      bool       `json:"blockedTcaFlag"`
	DlongClient         Quotation  `json:"dlongClient"`
	MaturityDate        time.Time  `json:"maturityDate"`
	SellAvailableFlag   bool       `json:"sellAvailableFlag"`
	First1DayCandleDate time.Time  `json:"first1dayCandleDate"`
	PlacementPrice      MoneyValue `json:"placementPrice"`
	Sector              string     `json:"sector"`
	Brand               struct {
		LogoName      string `json:"logoName"`
		LogoBaseColor string `json:"logoBaseColor"`
		TextColor     string `json:"textColor"`
	} `json:"brand"`
	LiquidityFlag         bool       `json:"liquidityFlag"`
	ForIisFlag            bool       `json:"forIisFlag"`
	PositionUID           string     `json:"positionUid"`
	ShortEnabledFlag      bool       `json:"shortEnabledFlag"`
	DshortClient          Quotation  `json:"dshortClient"`
	Dshort                Quotation  `json:"dshort"`
	Name                  string     `json:"name"`
	Exchange              string     `json:"exchange"`
	SubordinatedFlag      bool       `json:"subordinatedFlag"`
	FloatingCouponFlag    bool       `json:"floatingCouponFlag"`
	Figi                  string     `json:"figi"`
	DshortMin             Quotation  `json:"dshortMin"`
	Lot                   int        `json:"lot"`
	UID                   string     `json:"uid"`
	RequiredTests         []string   `json:"requiredTests"`
	Dlong                 Quotation  `json:"dlong"`
	Nominal               MoneyValue `json:"nominal"`
	Currency              string     `json:"currency"`
	AciValue              MoneyValue `json:"aciValue"`
	BuyAvailableFlag      bool       `json:"buyAvailableFlag"`
	WeekendFlag           bool       `json:"weekendFlag"`
	ClassCode             string     `json:"classCode"`
	Ticker                string     `json:"ticker"`
	CouponQuantityPerYear int        `json:"couponQuantityPerYear"`
	ForQualInvestorFlag   bool       `json:"forQualInvestorFlag"`
	InitialNominal        MoneyValue `json:"initialNominal"`
	APITradeAvailableFlag bool       `json:"apiTradeAvailableFlag"`
	DlongMin              Quotation  `json:"dlongMin"`
	Kshort                Quotation  `json:"kshort"`
	First1MinCandleDate   time.Time  `json:"first1minCandleDate"`
	StateRegDate          time.Time  `json:"stateRegDate"`
	IssueSizePlan         string     `json:"issueSizePlan"`
	MinPriceIncrement     Quotation  `json:"minPriceIncrement"`
	OtcFlag               bool       `json:"otcFlag"`
	Klong                 Quotation  `json:"klong"`
	IssueKind             string     `json:"issueKind"`
	PlacementDate         time.Time  `json:"placementDate"`
	AmortizationFlag      bool       `json:"amortizationFlag"`
	PerpetualFlag         bool       `json:"perpetualFlag"`
	IssueSize             string     `json:"issueSize"`
	CountryOfRiskName     string     `json:"countryOfRiskName"`
	Isin                  string     `json:"isin"`
}

type Candles struct {
	Candles []struct {
		Volume     string    `json:"volume"`
		High       Quotation `json:"high"`
		Low        Quotation `json:"low"`
		VolumeBuy  string    `json:"volumeBuy"`
		VolumeSell string    `json:"volumeSell"`
		Time       time.Time `json:"time"`
		Close      Quotation `json:"close"`
		Open       Quotation `json:"open"`
		IsComplete bool      `json:"isComplete"`
	} `json:"candles"`
}

type GenerateBrokerReportResponse struct {
	TaskID string `json:"taskId"`
}

type GetBrokerReportResponse struct {
	BrokerReport []struct {
		ExchangeClearingCommission MoneyValue `json:"exchangeClearingCommission"`
		SeparateAgreementDate      string     `json:"separateAgreementDate"`
		OrderID                    string     `json:"orderId"`
		Figi                       string     `json:"figi"`
		ExecuteSign                string     `json:"executeSign"`
		BrokerCommission           MoneyValue `json:"brokerCommission"`
		RepoRate                   Quotation  `json:"repoRate"`
		OrderAmount                MoneyValue `json:"orderAmount"`
		Price                      MoneyValue `json:"price"`
		AciValue                   Quotation  `json:"aciValue"`
		SecValueDate               time.Time  `json:"secValueDate"`
		Direction                  string     `json:"direction"`
		ClassCode                  string     `json:"classCode"`
		Ticker                     string     `json:"ticker"`
		Quantity                   string     `json:"quantity"`
		DeliveryType               string     `json:"deliveryType"`
		TradeDatetime              time.Time  `json:"tradeDatetime"`
		ExchangeCommission         MoneyValue `json:"exchangeCommission"`
		BrokerStatus               string     `json:"brokerStatus"`
		TotalOrderAmount           MoneyValue `json:"totalOrderAmount"`
		SeparateAgreementNumber    string     `json:"separateAgreementNumber"`
		ClearValueDate             time.Time  `json:"clearValueDate"`
		Name                       string     `json:"name"`
		Exchange                   string     `json:"exchange"`
		SeparateAgreementType      string     `json:"separateAgreementType"`
		TradeID                    string     `json:"tradeId"`
		Party                      string     `json:"party"`
	} `json:"brokerReport"`
	PagesCount int    `json:"pagesCount"`
	Page       int    `json:"page"`
	ItemsCount int    `json:"itemsCount"`
	TaskID     string `json:"taskId"`
}

type ClosePrices struct {
	ClosePrices []struct {
		Figi          string `json:"figi"`
		InstrumentUID string `json:"instrumentUid"`
		Ticker        string `json:"ticker"`
		ClassCode     string `json:"classCode"`
		Price         struct {
			Units string `json:"units"`
			Nano  int    `json:"nano"`
		} `json:"price"`
		EveningSessionPrice struct {
			Units string `json:"units"`
			Nano  int    `json:"nano"`
		} `json:"eveningSessionPrice"`
		Time                    time.Time `json:"time"`
		EveningSessionPriceTime time.Time `json:"eveningSessionPriceTime"`
	} `json:"closePrices"`
}

type Currencies struct {
	Instruments []struct {
		Figi                  string        `json:"figi"`
		Ticker                string        `json:"ticker"`
		ClassCode             string        `json:"classCode"`
		Isin                  string        `json:"isin"`
		Lot                   int           `json:"lot"`
		Currency              string        `json:"currency"`
		ShortEnabledFlag      bool          `json:"shortEnabledFlag"`
		Name                  string        `json:"name"`
		Exchange              string        `json:"exchange"`
		Nominal               MoneyValue    `json:"nominal"`
		CountryOfRisk         string        `json:"countryOfRisk"`
		CountryOfRiskName     string        `json:"countryOfRiskName"`
		TradingStatus         string        `json:"tradingStatus"`
		OtcFlag               bool          `json:"otcFlag"`
		BuyAvailableFlag      bool          `json:"buyAvailableFlag"`
		SellAvailableFlag     bool          `json:"sellAvailableFlag"`
		IsoCurrencyName       string        `json:"isoCurrencyName"`
		MinPriceIncrement     Quotation     `json:"minPriceIncrement"`
		APITradeAvailableFlag bool          `json:"apiTradeAvailableFlag"`
		UID                   string        `json:"uid"`
		RealExchange          string        `json:"realExchange"`
		PositionUID           string        `json:"positionUid"`
		RequiredTests         []interface{} `json:"requiredTests"`
		AssetUID              string        `json:"assetUid"`
		ForIisFlag            bool          `json:"forIisFlag"`
		ForQualInvestorFlag   bool          `json:"forQualInvestorFlag"`
		WeekendFlag           bool          `json:"weekendFlag"`
		BlockedTcaFlag        bool          `json:"blockedTcaFlag"`
		First1MinCandleDate   time.Time     `json:"first1minCandleDate,omitempty"`
		First1DayCandleDate   time.Time     `json:"first1dayCandleDate,omitempty"`
		Brand                 struct {
			LogoName      string `json:"logoName"`
			LogoBaseColor string `json:"logoBaseColor"`
			TextColor     string `json:"textColor"`
		} `json:"brand"`
		Dlong        Quotation `json:"dlong,omitempty"`
		Dshort       Quotation `json:"dshort,omitempty"`
		DlongMin     Quotation `json:"dlongMin,omitempty"`
		DshortMin    Quotation `json:"dshortMin,omitempty"`
		DlongClient  Quotation `json:"dlongClient,omitempty"`
		DshortClient Quotation `json:"dshortClient,omitempty"`
	} `json:"instruments"`
}
