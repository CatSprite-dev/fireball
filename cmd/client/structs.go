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

type UserOperation struct {
	Operations []struct {
		ID                string `json:"id"`
		ParentOperationID string `json:"parentOperationId"`
		Currency          string `json:"currency"`
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
		State           string        `json:"state"`
		Quantity        string        `json:"quantity"`
		QuantityRest    string        `json:"quantityRest"`
		Figi            string        `json:"figi"`
		InstrumentType  string        `json:"instrumentType"`
		Date            time.Time     `json:"date"`
		Type            string        `json:"type"`
		OperationType   string        `json:"operationType"`
		Trades          []interface{} `json:"trades"`
		AssetUID        string        `json:"assetUid"`
		PositionUID     string        `json:"positionUid"`
		InstrumentUID   string        `json:"instrumentUid"`
		ChildOperations []struct {
			InstrumentUID string `json:"instrumentUid"`
			Payment       struct {
				Currency string `json:"currency"`
				Units    string `json:"units"`
				Nano     int    `json:"nano"`
			} `json:"payment"`
		} `json:"childOperations"`
	} `json:"operations"`
}
