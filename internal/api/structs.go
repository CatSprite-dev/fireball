package api

import (
	"time"

	"github.com/CatSprite-dev/fireball/internal/pkg"
)

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
		Cursor            string         `json:"cursor"`
		BrokerAccountID   string         `json:"brokerAccountId"`
		ID                string         `json:"id"`
		ParentOperationID string         `json:"parentOperationId"`
		Name              string         `json:"name"`
		Date              time.Time      `json:"date"`
		Type              string         `json:"type"`
		Description       string         `json:"description"`
		State             string         `json:"state"`
		InstrumentUID     string         `json:"instrumentUid"`
		Figi              string         `json:"figi"`
		InstrumentType    string         `json:"instrumentType"`
		InstrumentKind    string         `json:"instrumentKind"`
		PositionUID       string         `json:"positionUid"`
		Ticker            string         `json:"ticker"`
		ClassCode         string         `json:"classCode"`
		Payment           pkg.MoneyValue `json:"payment"`
		Price             pkg.MoneyValue `json:"price"`
		Commission        pkg.MoneyValue `json:"commission"`
		Yield             pkg.MoneyValue `json:"yield"`
		YieldRelative     pkg.Quotation  `json:"yieldRelative"`
		AccruedInt        pkg.MoneyValue `json:"accruedInt"`
		Quantity          string         `json:"quantity"`
		QuantityRest      string         `json:"quantityRest"`
		QuantityDone      string         `json:"quantityDone"`
		CancelReason      string         `json:"cancelReason"`
		AssetUID          string         `json:"assetUid"`
		ChildOperations   []struct {
			InstrumentUID string         `json:"instrumentUid"`
			Payment       pkg.MoneyValue `json:"payment"`
		} `json:"childOperations"`
	} `json:"items"`
}

type UserPortfolio struct {
	TotalAmountShares     pkg.MoneyValue `json:"totalAmountShares"`
	TotalAmountBonds      pkg.MoneyValue `json:"totalAmountBonds"`
	TotalAmountEtf        pkg.MoneyValue `json:"totalAmountEtf"`
	TotalAmountCurrencies pkg.MoneyValue `json:"totalAmountCurrencies"`
	TotalAmountFutures    pkg.MoneyValue `json:"totalAmountFutures"`
	ExpectedYield         pkg.Quotation  `json:"expectedYield"`
	Positions             []struct {
		Figi                     string         `json:"figi"`
		InstrumentType           string         `json:"instrumentType"`
		Quantity                 pkg.Quotation  `json:"quantity"`
		AveragePositionPrice     pkg.MoneyValue `json:"averagePositionPrice"`
		ExpectedYield            pkg.Quotation  `json:"expectedYield"`
		AveragePositionPricePt   pkg.Quotation  `json:"averagePositionPricePt"`
		CurrentPrice             pkg.MoneyValue `json:"currentPrice"`
		AveragePositionPriceFifo pkg.MoneyValue `json:"averagePositionPriceFifo"`
		QuantityLots             pkg.Quotation  `json:"quantityLots"`
		Blocked                  bool           `json:"blocked"`
		BlockedLots              pkg.Quotation  `json:"blockedLots"`
		PositionUID              string         `json:"positionUid"`
		InstrumentUID            string         `json:"instrumentUid"`
		VarMargin                pkg.MoneyValue `json:"varMargin"`
		ExpectedYieldFifo        pkg.Quotation  `json:"expectedYieldFifo"`
		DailyYield               pkg.MoneyValue `json:"dailyYield"`
		Ticker                   string         `json:"ticker"`
		ClassCode                string         `json:"classCode"`
		CurrentNkd               pkg.MoneyValue `json:"currentNkd,omitempty"`
	} `json:"positions"`
	AccountID            string         `json:"accountId"`
	TotalAmountOptions   pkg.MoneyValue `json:"totalAmountOptions"`
	TotalAmountSp        pkg.MoneyValue `json:"totalAmountSp"`
	TotalAmountPortfolio pkg.MoneyValue `json:"totalAmountPortfolio"`
	VirtualPositions     []struct {
		PositionUID              string         `json:"positionUid"`
		InstrumentUID            string         `json:"instrumentUid"`
		Figi                     string         `json:"figi"`
		InstrumentType           string         `json:"instrumentType"`
		Quantity                 pkg.Quotation  `json:"quantity"`
		AveragePositionPrice     pkg.MoneyValue `json:"averagePositionPrice"`
		ExpectedYield            pkg.Quotation  `json:"expectedYield"`
		ExpectedYieldFifo        pkg.Quotation  `json:"expectedYieldFifo"`
		ExpireDate               time.Time      `json:"expireDate"`
		CurrentPrice             pkg.MoneyValue `json:"currentPrice"`
		AveragePositionPriceFifo pkg.MoneyValue `json:"averagePositionPriceFifo"`
		DailyYield               pkg.MoneyValue `json:"dailyYield"`
		Ticker                   string         `json:"ticker"`
		ClassCode                string         `json:"classCode"`
	} `json:"virtualPositions"`
	DailyYield         pkg.MoneyValue `json:"dailyYield"`
	DailyYieldRelative pkg.Quotation  `json:"dailyYieldRelative"`
}
