package tinvest

import (
	"context"
	"encoding/json"
	"net/http"
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
