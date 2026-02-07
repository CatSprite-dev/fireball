package api

import (
	"time"
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
		Cursor            string    `json:"cursor"`
		BrokerAccountID   string    `json:"brokerAccountId"`
		ID                string    `json:"id"`
		ParentOperationID string    `json:"parentOperationId"`
		Name              string    `json:"name"`
		Date              time.Time `json:"date"`
		Type              string    `json:"type"`
		Description       string    `json:"description"`
		State             string    `json:"state"`
		InstrumentUID     string    `json:"instrumentUid"`
		Figi              string    `json:"figi"`
		InstrumentType    string    `json:"instrumentType"`
		InstrumentKind    string    `json:"instrumentKind"`
		PositionUID       string    `json:"positionUid"`
		Ticker            string    `json:"ticker"`
		ClassCode         string    `json:"classCode"`
		Payment           struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		} `json:"payment"`
		Price struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		} `json:"price"`
		Commission struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		} `json:"commission"`
		Yield struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		} `json:"yield"`
		YieldRelative struct {
			Units string `json:"units"`
			Nano  int    `json:"nano"`
		} `json:"yieldRelative"`
		AccruedInt struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		} `json:"accruedInt"`
		Quantity        string `json:"quantity"`
		QuantityRest    string `json:"quantityRest"`
		QuantityDone    string `json:"quantityDone"`
		CancelReason    string `json:"cancelReason"`
		AssetUID        string `json:"assetUid"`
		ChildOperations []struct {
			InstrumentUID string `json:"instrumentUid"`
			Payment       struct {
				Currency string `json:"currency"`
				Units    string `json:"units"`
				Nano     int    `json:"nano"`
			} `json:"payment"`
		} `json:"childOperations"`
	} `json:"items"`
}

type UserPortfolio struct {
	TotalAmountShares struct {
		Currency string `json:"currency"`
		Units    string `json:"units"`
		Nano     int    `json:"nano"`
	} `json:"totalAmountShares"`
	TotalAmountBonds struct {
		Currency string `json:"currency"`
		Units    string `json:"units"`
		Nano     int    `json:"nano"`
	} `json:"totalAmountBonds"`
	TotalAmountEtf struct {
		Currency string `json:"currency"`
		Units    string `json:"units"`
		Nano     int    `json:"nano"`
	} `json:"totalAmountEtf"`
	TotalAmountCurrencies struct {
		Currency string `json:"currency"`
		Units    string `json:"units"`
		Nano     int    `json:"nano"`
	} `json:"totalAmountCurrencies"`
	TotalAmountFutures struct {
		Currency string `json:"currency"`
		Units    string `json:"units"`
		Nano     int    `json:"nano"`
	} `json:"totalAmountFutures"`
	ExpectedYield struct {
		Units string `json:"units"`
		Nano  int    `json:"nano"`
	} `json:"expectedYield"`
	Positions []struct {
		Figi           string `json:"figi"`
		InstrumentType string `json:"instrumentType"`
		Quantity       struct {
			Units string `json:"units"`
			Nano  int    `json:"nano"`
		} `json:"quantity"`
		AveragePositionPrice struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		} `json:"averagePositionPrice"`
		ExpectedYield struct {
			Units string `json:"units"`
			Nano  int    `json:"nano"`
		} `json:"expectedYield"`
		AveragePositionPricePt struct {
			Units string `json:"units"`
			Nano  int    `json:"nano"`
		} `json:"averagePositionPricePt"`
		CurrentPrice struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		} `json:"currentPrice"`
		AveragePositionPriceFifo struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		} `json:"averagePositionPriceFifo"`
		QuantityLots struct {
			Units string `json:"units"`
			Nano  int    `json:"nano"`
		} `json:"quantityLots"`
		Blocked     bool `json:"blocked"`
		BlockedLots struct {
			Units string `json:"units"`
			Nano  int    `json:"nano"`
		} `json:"blockedLots"`
		PositionUID   string `json:"positionUid"`
		InstrumentUID string `json:"instrumentUid"`
		VarMargin     struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		} `json:"varMargin"`
		ExpectedYieldFifo struct {
			Units string `json:"units"`
			Nano  int    `json:"nano"`
		} `json:"expectedYieldFifo"`
		DailyYield struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		} `json:"dailyYield"`
		Ticker     string `json:"ticker"`
		ClassCode  string `json:"classCode"`
		CurrentNkd struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		} `json:"currentNkd,omitempty"`
	} `json:"positions"`
	AccountID          string `json:"accountId"`
	TotalAmountOptions struct {
		Currency string `json:"currency"`
		Units    string `json:"units"`
		Nano     int    `json:"nano"`
	} `json:"totalAmountOptions"`
	TotalAmountSp struct {
		Currency string `json:"currency"`
		Units    string `json:"units"`
		Nano     int    `json:"nano"`
	} `json:"totalAmountSp"`
	TotalAmountPortfolio struct {
		Currency string `json:"currency"`
		Units    string `json:"units"`
		Nano     int    `json:"nano"`
	} `json:"totalAmountPortfolio"`
	VirtualPositions []any `json:"virtualPositions"`
	DailyYield       struct {
		Currency string `json:"currency"`
		Units    string `json:"units"`
		Nano     int    `json:"nano"`
	} `json:"dailyYield"`
	DailyYieldRelative struct {
		Units string `json:"units"`
		Nano  int    `json:"nano"`
	} `json:"dailyYieldRelative"`
}
