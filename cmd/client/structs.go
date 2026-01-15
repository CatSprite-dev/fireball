package main

import "time"

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
