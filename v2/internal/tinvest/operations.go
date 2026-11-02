package tinvest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type operations struct {
	HasNext    bool        `json:"hasNext"`
	NextCursor string      `json:"nextCursor"`
	Operations []Operation `json:"items"`
}

type Operation struct {
	Cursor            string           `json:"cursor"`
	BrokerAccountID   string           `json:"brokerAccountId"`
	ID                string           `json:"id"`
	ParentOperationID string           `json:"parentOperationId"`
	Name              string           `json:"name"`
	Date              time.Time        `json:"date"`
	Type              string           `json:"type"`
	Description       string           `json:"description"`
	State             string           `json:"state"`
	InstrumentUID     string           `json:"instrumentUid"`
	Figi              string           `json:"figi"`
	InstrumentType    string           `json:"instrumentType"`
	InstrumentKind    string           `json:"instrumentKind"`
	PositionUID       string           `json:"positionUid"`
	Ticker            string           `json:"ticker"`
	ClassCode         string           `json:"classCode"`
	Payment           MoneyValue       `json:"payment"`
	Price             MoneyValue       `json:"price"`
	Commission        MoneyValue       `json:"commission"`
	Yield             MoneyValue       `json:"yield"`
	YieldRelative     Quotation        `json:"yieldRelative"`
	AccruedInt        MoneyValue       `json:"accruedInt"`
	Quantity          string           `json:"quantity"`
	QuantityRest      string           `json:"quantityRest"`
	QuantityDone      string           `json:"quantityDone"`
	CancelReason      string           `json:"cancelReason"`
	AssetUID          string           `json:"assetUid"`
	ChildOperations   []ChildOperation `json:"childOperations"`
}

type ChildOperation struct {
	InstrumentUID string     `json:"instrumentUid"`
	Payment       MoneyValue `json:"payment"`
}

type OperationsFilter struct {
	AccountID          string
	InstrumentID       string
	From, To           *time.Time
	OperationTypes     []string
	State              string
	WithoutCommissions bool
}

const operationsPageLimit = 1000

func (c *Client) GetOperationsByCursor(ctx context.Context, filter OperationsFilter) ([]Operation, error) {
	type request struct {
		AccountID          string     `json:"accountId"`
		InstrumentID       string     `json:"instrumentId,omitempty"`
		From               *time.Time `json:"from,omitempty"`
		To                 *time.Time `json:"to,omitempty"`
		Cursor             string     `json:"cursor,omitempty"`
		Limit              int32      `json:"limit"`
		OperationTypes     []string   `json:"operationTypes,omitempty"`
		State              string     `json:"state,omitempty"`
		WithoutCommissions bool       `json:"withoutCommissions,omitempty"`
	}

	const method = "OperationsService/GetOperationsByCursor"

	var all []Operation
	cursor := ""
	for {
		data, err := c.DoRequest(ctx, method, request{
			AccountID:          filter.AccountID,
			InstrumentID:       filter.InstrumentID,
			From:               filter.From,
			To:                 filter.To,
			Cursor:             cursor,
			Limit:              operationsPageLimit,
			OperationTypes:     filter.OperationTypes,
			State:              filter.State,
			WithoutCommissions: filter.WithoutCommissions,
		})
		if err != nil {
			return nil, err
		}

		var page operations
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, fmt.Errorf("%s: unmarshal response: %w", method, err)
		}
		all = append(all, page.Operations...)

		if !page.HasNext {
			return all, nil
		}
		if page.NextCursor == "" || page.NextCursor == cursor {
			return nil, fmt.Errorf("%s: hasNext=true but cursor did not advance", method)
		}
		cursor = page.NextCursor
	}
}
