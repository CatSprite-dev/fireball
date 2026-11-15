package tinvest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// AccruedInterest — НКД и номинал облигации на один день.
type AccruedInterest struct {
	Date         time.Time `json:"date"`
	Value        Quotation `json:"value"`
	ValuePercent Quotation `json:"valuePercent"`
	Nominal      Quotation `json:"nominal"`
}

type getAccruedInterestsResponse struct {
	AccruedInterests []AccruedInterest `json:"accruedInterests"`
}

func (c *Client) GetAccruedInterests(ctx context.Context, instrumentUID string, from, to time.Time) ([]AccruedInterest, error) {
	type request struct {
		InstrumentID string    `json:"instrumentId"`
		From         time.Time `json:"from"`
		To           time.Time `json:"to"`
	}

	const method = "InstrumentsService/GetAccruedInterests"

	data, err := c.DoRequest(ctx, method, request{InstrumentID: instrumentUID, From: from, To: to})
	if err != nil {
		return nil, err
	}

	var response getAccruedInterestsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("%s: unmarshal response: %w", method, err)
	}
	return response.AccruedInterests, nil
}
