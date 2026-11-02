package tinvest

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := New("test-token", logger)
	c.baseURL = server.URL + "/rest/test.v1"
	return c
}

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

func TestRetryOn503(t *testing.T) {
	var calls atomic.Int32

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte(`{"accounts":[]}`))
	})

	_, err := c.GetAccounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("server got %d calls, want 3", got)
	}
}

func TestNoRetryOn401(t *testing.T) {
	var calls atomic.Int32

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code":16,"message":"Authentication token is missing or invalid","description":"40003"}`))
	})

	_, err := c.GetAccounts(context.Background())

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized || apiErr.Code != "40003" {
		t.Fatalf("got %+v", apiErr)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("server got %d calls, want 1", got)
	}
}
