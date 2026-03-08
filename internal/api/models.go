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
	Items      []struct {
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
	} `json:"items"`
}

type UserPortfolio struct {
	TotalAmountShares     MoneyValue `json:"totalAmountShares"`
	TotalAmountBonds      MoneyValue `json:"totalAmountBonds"`
	TotalAmountEtf        MoneyValue `json:"totalAmountEtf"`
	TotalAmountCurrencies MoneyValue `json:"totalAmountCurrencies"`
	TotalAmountFutures    MoneyValue `json:"totalAmountFutures"`
	ExpectedYield         Quotation  `json:"expectedYield"`
	Positions             []struct {
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
	} `json:"positions"`
	AccountID            string     `json:"accountId"`
	TotalAmountOptions   MoneyValue `json:"totalAmountOptions"`
	TotalAmountSp        MoneyValue `json:"totalAmountSp"`
	TotalAmountPortfolio MoneyValue `json:"totalAmountPortfolio"`
	VirtualPositions     []any      `json:"virtualPositions"`
	DailyYield           MoneyValue `json:"dailyYield"`
	DailyYieldRelative   Quotation  `json:"dailyYieldRelative"`
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

type Instrument struct {
	Instrument struct {
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
	} `json:"instrument"`
}

type Candles struct {
	Candles []struct {
		Volume string `json:"volume"`
		High   struct {
			Nano  int    `json:"nano"`
			Units string `json:"units"`
		} `json:"high"`
		Low struct {
			Nano  int    `json:"nano"`
			Units string `json:"units"`
		} `json:"low"`
		VolumeBuy  string    `json:"volumeBuy"`
		VolumeSell string    `json:"volumeSell"`
		Time       time.Time `json:"time"`
		Close      struct {
			Nano  int    `json:"nano"`
			Units string `json:"units"`
		} `json:"close"`
		Open struct {
			Nano  int    `json:"nano"`
			Units string `json:"units"`
		} `json:"open"`
		IsComplete bool `json:"isComplete"`
	} `json:"candles"`
}
