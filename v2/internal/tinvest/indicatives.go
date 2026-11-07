package tinvest

import (
	"context"
	"encoding/json"
	"fmt"
)

type Indicative struct {
	ClassCode         string             `json:"classCode"`
	UID               string             `json:"uid"`
	Ticker            string             `json:"ticker"`
	IndexComposition  []IndexComposition `json:"indexComposition"`
	Name              string             `json:"name"`
	SellAvailableFlag bool               `json:"sellAvailableFlag"`
	Figi              string             `json:"figi"`
	Currency          string             `json:"currency"`
	Exchange          string             `json:"exchange"`
	BuyAvailableFlag  bool               `json:"buyAvailableFlag"`
}

type IndexComposition struct {
	UID    string    `json:"uid"`
	Weight Quotation `json:"weight"`
}

type indicativesResponse struct {
	Indicatives []Indicative `json:"instruments"`
}

func (c *Client) GetIndicatives(ctx context.Context) ([]Indicative, error) {
	const method = "InstrumentsService/Indicatives"

	var response indicativesResponse
	res, err := c.DoRequest(ctx, method, struct{}{})
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(res, &response); err != nil {
		return nil, fmt.Errorf("%s: unmarshal response: %w", method, err)
	}

	return response.Indicatives, nil
}
