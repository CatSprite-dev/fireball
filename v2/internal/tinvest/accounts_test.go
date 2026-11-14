package tinvest

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetAccounts(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"accounts":[{"id":"123","name":"Тест"}]}`))
	})

	accounts, err := c.GetAccounts(context.Background(), AccountStatusAll)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 1 || accounts[0].ID != "123" {
		t.Fatalf("got %+v, want one account with id 123", accounts)
	}
}

func TestGetAccountsSendsStatus(t *testing.T) {
	tests := []struct {
		name   string
		status AccountStatus
	}{
		{"open", AccountStatusOpen},
		{"closed", AccountStatusClosed},
		{"all", AccountStatusAll},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got struct {
				Status string `json:"status"`
			}
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Errorf("decode request body: %v", err)
				}
				w.Write([]byte(`{"accounts":[]}`))
			})

			if _, err := c.GetAccounts(context.Background(), tt.status); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Status != string(tt.status) {
				t.Errorf("request status = %q, want %q", got.Status, tt.status)
			}
		})
	}
}
