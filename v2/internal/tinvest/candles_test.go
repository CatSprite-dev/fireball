package tinvest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetCandlesRequest(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			InstrumentID     string `json:"instrumentId"`
			Interval         string `json:"interval"`
			CandleSourceType string `json:"candleSourceType"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.CandleSourceType != "CANDLE_SOURCE_EXCHANGE" {
			t.Errorf("candleSourceType = %q, want CANDLE_SOURCE_EXCHANGE", req.CandleSourceType)
		}
		if req.InstrumentID != "uid-1" || req.Interval != "CANDLE_INTERVAL_DAY" {
			t.Errorf("got %+v", req)
		}
		w.Write([]byte(`{"priceCurrency":"rub","candles":[{"time":"2026-11-01T00:00:00Z","close":{"units":"100","nano":0}}]}`))
	})

	to := time.Now()
	candles, err := c.GetCandles(context.Background(), "uid-1", to.AddDate(0, 0, -7), to, CandleIntervalDay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if candles.PriceCurrency != "rub" || len(candles.Candles) != 1 {
		t.Fatalf("got %+v", candles)
	}
}

func TestGetCandlesSplitsByYear(t *testing.T) {
	var calls atomic.Int32

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)

		var req struct {
			From time.Time `json:"from"`
			To   time.Time `json:"to"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		// Две свечи на кусок: на его начале и на его конце.
		fmt.Fprintf(w, `{"priceCurrency":"rub","candles":[{"time":%q},{"time":%q}]}`,
			req.From.Format(time.RFC3339), req.To.Format(time.RFC3339))
	})

	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)

	candles, err := c.GetCandles(context.Background(), "uid-1", from, to, CandleIntervalDay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := calls.Load(); got != 3 {
		t.Fatalf("server got %d calls, want 3", got)
	}

	want := []time.Time{
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	if len(candles.Candles) != len(want) {
		t.Fatalf("got %d candles, want %d", len(candles.Candles), len(want))
	}
	for i, w := range want {
		if !candles.Candles[i].Time.Equal(w) {
			t.Errorf("candle %d: time %v, want %v", i, candles.Candles[i].Time, w)
		}
	}
}
