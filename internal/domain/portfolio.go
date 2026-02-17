package domain

type UserFullPortfolio struct {
	TotalAmountShares     MoneyValue            `json:"totalAmountShares"`
	TotalAmountBonds      MoneyValue            `json:"totalAmountBonds"`
	TotalAmountEtf        MoneyValue            `json:"totalAmountEtf"`
	TotalAmountCurrencies MoneyValue            `json:"totalAmountCurrencies"`
	TotalAmountFutures    MoneyValue            `json:"totalAmountFutures"`
	ExpectedYield         MoneyValue            `json:"expectedYield"`
	ExpectedYieldRelative Quotation             `json:"expectedYieldRelative"`
	Positions             []Position            `json:"positions"`
	AccountID             string                `json:"accountId"`
	TotalAmountOptions    MoneyValue            `json:"totalAmountOptions"`
	TotalAmountSp         MoneyValue            `json:"totalAmountSp"`
	TotalAmountPortfolio  MoneyValue            `json:"totalAmountPortfolio"`
	DailyYield            MoneyValue            `json:"dailyYield"`
	DailyYieldRelative    Quotation             `json:"dailyYieldRelative"`
	AllDividends          map[string]MoneyValue `json:"allDividends"`
}

type Position struct {
	Name                     string     `json:"name"`
	Type                     string     `json:"type"`
	Figi                     string     `json:"figi"`
	InstrumentType           string     `json:"instrumentType"`
	Quantity                 Quotation  `json:"quantity"`
	AveragePositionPrice     MoneyValue `json:"averagePositionPrice"`
	ExpectedYield            MoneyValue `json:"expectedYield"`
	ExpectedYieldRelative    Quotation  `json:"expectedYieldRelative"`
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
	DailyYieldRelative       Quotation  `json:"dailyYieldRelative"`
	Ticker                   string     `json:"ticker"`
	ClassCode                string     `json:"classCode"`
	CurrentNkd               MoneyValue `json:"currentNkd,omitempty"`
	Dividends                MoneyValue `json:"dividends"`
	TotalYield               MoneyValue `json:"totalYield"`
	TotalYieldRelative       Quotation  `json:"totalYieldRelative"`
}
