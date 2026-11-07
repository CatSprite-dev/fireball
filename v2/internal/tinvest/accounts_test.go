package tinvest

import (
	"context"
	"net/http"
	"testing"
)

func TestGetAccounts(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"accounts":[{"id":"123","name":"Тест"}]}`))
	})

	accounts, err := c.GetAccounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 1 || accounts[0].ID != "123" {
		t.Fatalf("got %+v, want one account with id 123", accounts)
	}
}
