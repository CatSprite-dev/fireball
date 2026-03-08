package service

import (
	"testing"

	"github.com/CatSprite-dev/fireball/internal/api"
)

func TestConvertToFullPortfolio(t *testing.T) {
	rawPortfolio := api.UserPortfolio{
		TotalAmountShares: api.MoneyValue{
			Currency: "RUB",
			Units:    "100",
			Nano:     0,
		},
		TotalAmountBonds: api.MoneyValue{
			Currency: "RUB",
			Units:    "50",
			Nano:     0,
		},
		TotalAmountEtf: api.MoneyValue{
			Currency: "RUB",
			Units:    "30",
			Nano:     0,
		},
		TotalAmountCurrencies: api.MoneyValue{
			Currency: "RUB",
			Units:    "20",
			Nano:     0,
		},
		TotalAmountFutures: api.MoneyValue{
			Currency: "RUB",
			Units:    "10",
			Nano:     0,
		},
		ExpectedYield: api.Quotation{
			Units: "15",
			Nano:  0,
		},
		Positions: []struct {
			Figi                     string         `json:"figi"`
			InstrumentType           string         `json:"instrumentType"`
			Quantity                 api.Quotation  `json:"quantity"`
			AveragePositionPrice     api.MoneyValue `json:"averagePositionPrice"`
			ExpectedYield            api.Quotation  `json:"expectedYield"`
			AveragePositionPricePt   api.Quotation  `json:"averagePositionPricePt"`
			CurrentPrice             api.MoneyValue `json:"currentPrice"`
			AveragePositionPriceFifo api.MoneyValue `json:"averagePositionPriceFifo"`
			QuantityLots             api.Quotation  `json:"quantityLots"`
			Blocked                  bool           `json:"blocked"`
			BlockedLots              api.Quotation  `json:"blockedLots"`
			PositionUID              string         `json:"positionUid"`
			InstrumentUID            string         `json:"instrumentUid"`
			VarMargin                api.MoneyValue `json:"varMargin"`
			ExpectedYieldFifo        api.Quotation  `json:"expectedYieldFifo"`
			DailyYield               api.MoneyValue `json:"dailyYield"`
			Ticker                   string         `json:"ticker"`
			ClassCode                string         `json:"classCode"`
			CurrentNkd               api.MoneyValue `json:"currentNkd,omitempty"`
		}{
			{
				Figi:           "BBG000B9XRY4",
				InstrumentType: "Stock",
				Quantity: api.Quotation{
					Units: "10",
					Nano:  0,
				},
				AveragePositionPrice: api.MoneyValue{
					Currency: "RUB",
					Units:    "500",
					Nano:     0,
				},
				ExpectedYield: api.Quotation{
					Units: "50",
					Nano:  0,
				},
				AveragePositionPricePt: api.Quotation{
					Units: "500",
					Nano:  0,
				},
				CurrentPrice: api.MoneyValue{
					Currency: "RUB",
					Units:    "550",
					Nano:     0,
				},
				AveragePositionPriceFifo: api.MoneyValue{
					Currency: "RUB",
					Units:    "500",
					Nano:     0,
				},
				QuantityLots: api.Quotation{
					Units: "10",
					Nano:  0,
				},
				Blocked: false,
				BlockedLots: api.Quotation{
					Units: "0",
					Nano:  0,
				},
				PositionUID:   "test-position-uid",
				InstrumentUID: "test-instrument-uid",
				VarMargin: api.MoneyValue{
					Currency: "RUB",
					Units:    "0",
					Nano:     0,
				},
				ExpectedYieldFifo: api.Quotation{
					Units: "50",
					Nano:  0,
				},
				DailyYield: api.MoneyValue{
					Currency: "RUB",
					Units:    "50",
					Nano:     0,
				},
				Ticker:    "AAPL",
				ClassCode: "USD",
				CurrentNkd: api.MoneyValue{
					Currency: "RUB",
					Units:    "0",
					Nano:     0,
				},
			},
		},
		AccountID: "test-account-id",
		TotalAmountOptions: api.MoneyValue{
			Currency: "RUB",
			Units:    "0",
			Nano:     0,
		},
		TotalAmountSp: api.MoneyValue{
			Currency: "RUB",
			Units:    "0",
			Nano:     0,
		},
		TotalAmountPortfolio: api.MoneyValue{
			Currency: "RUB",
			Units:    "210",
			Nano:     0,
		},
		VirtualPositions: []any{},
		DailyYield: api.MoneyValue{
			Currency: "RUB",
			Units:    "50",
			Nano:     0,
		},
		DailyYieldRelative: api.Quotation{
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
