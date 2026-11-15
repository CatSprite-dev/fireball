package tinvest

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestGetAccruedInterests(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			InstrumentID string    `json:"instrumentId"`
			From         time.Time `json:"from"`
			To           time.Time `json:"to"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.InstrumentID != "bond-uid" || req.From.IsZero() || req.To.IsZero() {
			t.Errorf("got request %+v", req)
		}
		w.Write([]byte(`{"accruedInterests":[
			{"date":"2026-11-14T00:00:00Z","value":{"units":"11","nano":380000000},"valuePercent":{"units":"1","nano":138000000},"nominal":{"units":"1000","nano":0}},
			{"date":"2026-11-15T00:00:00Z","value":{"units":"11","nano":380000000},"valuePercent":{"units":"1","nano":138000000},"nominal":{"units":"1000","nano":0}}
		]}`))
	})

	to := time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)
	list, err := c.GetAccruedInterests(context.Background(), "bond-uid", to.AddDate(0, 0, -2), to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("got %d records, want 2", len(list))
	}
	last := list[1]
	if !last.Date.Equal(time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)) ||
		last.Value.Units != "11" || last.Value.Nano != 380000000 || last.Nominal.Units != "1000" {
		t.Errorf("got %+v", last)
	}
}
