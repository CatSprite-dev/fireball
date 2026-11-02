package tinvest

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestGetOperationsByCursorPages(t *testing.T) {
	var calls atomic.Int32

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)

		var req struct {
			Cursor string `json:"cursor"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.Cursor == "" {
			w.Write([]byte(`{"hasNext":true,"nextCursor":"page2","items":[{"id":"1"},{"id":"2"}]}`))
			return
		}
		w.Write([]byte(`{"hasNext":false,"items":[{"id":"3"}]}`))
	})

	ops, err := c.GetOperationsByCursor(context.Background(), OperationsFilter{AccountID: "acc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ops) != 3 || ops[2].ID != "3" {
		t.Fatalf("got %+v, want 3 operations", ops)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("server got %d calls, want 2", got)
	}
}

func TestGetOperationsByCursorStuck(t *testing.T) {
	var calls atomic.Int32

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Write([]byte(`{"hasNext":true,"nextCursor":"same","items":[{"id":"1"}]}`))
	})

	_, err := c.GetOperationsByCursor(context.Background(), OperationsFilter{AccountID: "acc"})
	if err == nil {
		t.Fatal("want error for stuck cursor, got nil")
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("server got %d calls, want 2", got)
	}
}
