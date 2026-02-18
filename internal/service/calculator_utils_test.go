package service

import (
	"testing"

	"github.com/CatSprite-dev/fireball/internal/api"
)

func TestConvertToFullPortfolio(t *testing.T) {
	rawPortfolio := api.UserPortfolio{
		TotalAmountShares: struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		}{
			Currency: "RUB",
			Units:    "100",
			Nano:     0,
		},
		TotalAmountBonds: struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		}{
			Currency: "RUB",
			Units:    "50",
			Nano:     0,
		},
		TotalAmountEtf: struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		}{
			Currency: "RUB",
			Units:    "30",
			Nano:     0,
		},
		TotalAmountCurrencies: struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		}{
			Currency: "RUB",
			Units:    "20",
			Nano:     0,
		},
		TotalAmountFutures: struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		}{
			Currency: "RUB",
			Units:    "10",
			Nano:     0,
		},
		ExpectedYield: struct {
			Units string `json:"units"`
			Nano  int    `json:"nano"`
		}{
			Units: "15",
			Nano:  0,
		},
		Positions: []struct {
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
		}{
			{
				Figi:           "BBG000B9XRY4",
				InstrumentType: "Stock",
				Quantity: struct {
					Units string `json:"units"`
					Nano  int    `json:"nano"`
				}{
					Units: "10",
					Nano:  0,
				},
				AveragePositionPrice: struct {
					Currency string `json:"currency"`
					Units    string `json:"units"`
					Nano     int    `json:"nano"`
				}{
					Currency: "RUB",
					Units:    "500",
					Nano:     0,
				},
				ExpectedYield: struct {
					Units string `json:"units"`
					Nano  int    `json:"nano"`
				}{
					Units: "50",
					Nano:  0,
				},
				AveragePositionPricePt: struct {
					Units string `json:"units"`
					Nano  int    `json:"nano"`
				}{
					Units: "500",
					Nano:  0,
				},
				CurrentPrice: struct {
					Currency string `json:"currency"`
					Units    string `json:"units"`
					Nano     int    `json:"nano"`
				}{
					Currency: "RUB",
					Units:    "550",
					Nano:     0,
				},
				AveragePositionPriceFifo: struct {
					Currency string `json:"currency"`
					Units    string `json:"units"`
					Nano     int    `json:"nano"`
				}{
					Currency: "RUB",
					Units:    "500",
					Nano:     0,
				},
				QuantityLots: struct {
					Units string `json:"units"`
					Nano  int    `json:"nano"`
				}{
					Units: "10",
					Nano:  0,
				},
				Blocked: false,
				BlockedLots: struct {
					Units string `json:"units"`
					Nano  int    `json:"nano"`
				}{
					Units: "0",
					Nano:  0,
				},
				PositionUID:   "test-position-uid",
				InstrumentUID: "test-instrument-uid",
				VarMargin: struct {
					Currency string `json:"currency"`
					Units    string `json:"units"`
					Nano     int    `json:"nano"`
				}{
					Currency: "RUB",
					Units:    "0",
					Nano:     0,
				},
				ExpectedYieldFifo: struct {
					Units string `json:"units"`
					Nano  int    `json:"nano"`
				}{
					Units: "50",
					Nano:  0,
				},
				DailyYield: struct {
					Currency string `json:"currency"`
					Units    string `json:"units"`
					Nano     int    `json:"nano"`
				}{
					Currency: "RUB",
					Units:    "50",
					Nano:     0,
				},
				Ticker:    "AAPL",
				ClassCode: "USD",
				CurrentNkd: struct {
					Currency string `json:"currency"`
					Units    string `json:"units"`
					Nano     int    `json:"nano"`
				}{
					Currency: "RUB",
					Units:    "0",
					Nano:     0,
				},
			},
		},
		AccountID: "test-account-id",
		TotalAmountOptions: struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		}{
			Currency: "RUB",
			Units:    "0",
			Nano:     0,
		},
		TotalAmountSp: struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		}{
			Currency: "RUB",
			Units:    "0",
			Nano:     0,
		},
		TotalAmountPortfolio: struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		}{
			Currency: "RUB",
			Units:    "210",
			Nano:     0,
		},
		VirtualPositions: []any{},
		DailyYield: struct {
			Currency string `json:"currency"`
			Units    string `json:"units"`
			Nano     int    `json:"nano"`
		}{
			Currency: "RUB",
			Units:    "50",
			Nano:     0,
		},
		DailyYieldRelative: struct {
			Units string `json:"units"`
			Nano  int    `json:"nano"`
		}{
			Units: "5",
			Nano:  0,
		},
	}

	result := convertFullPortfolio(rawPortfolio)

	if result.TotalAmountShares.Units != "100" {
		t.Errorf("Expected TotalAmountShares 100, got %s", result.TotalAmountShares.Units)
	}

	if len(result.Positions) != 1 {
		t.Errorf("Expected 1 position, got %d", len(result.Positions))
	}

	if result.Positions[0].Figi != "BBG000B9XRY4" {
		t.Errorf("Expected Figi BBG000B9XRY4, got %s", result.Positions[0].Figi)
	}
}
