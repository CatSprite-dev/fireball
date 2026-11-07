package tinvest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Account struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	OpenedDate  time.Time `json:"openedDate"`
	ClosedDate  time.Time `json:"closedDate"`
	AccessLevel string    `json:"accessLevel"`
}

type getAccountsResponse struct {
	Accounts []Account `json:"accounts"`
}

func (c *Client) GetAccounts(ctx context.Context) ([]Account, error) {
	type request struct {
		Status string `json:"status"`
	}

	const method = "UsersService/GetAccounts"

	var response getAccountsResponse
	res, err := c.DoRequest(ctx, method, request{Status: "ACCOUNT_STATUS_OPEN"})
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(res, &response); err != nil {
		return nil, fmt.Errorf("%s: unmarshal response: %w", method, err)
	}

	return response.Accounts, nil

}
