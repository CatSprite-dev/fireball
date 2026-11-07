package tinvest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Candles struct {
	PriceCurrency string   `json:"priceCurrency"`
	Candles       []Candle `json:"candles"`
}

type Candle struct {
	Volume     string    `json:"volume"`
	High       Quotation `json:"high"`
	Low        Quotation `json:"low"`
	VolumeBuy  string    `json:"volumeBuy"`
	VolumeSell string    `json:"volumeSell"`
	Time       time.Time `json:"time"`
	Close      Quotation `json:"close"`
	Open       Quotation `json:"open"`
	IsComplete bool      `json:"isComplete"`
}

type CandleInterval string

const (
	CandleIntervalDay   CandleInterval = "CANDLE_INTERVAL_DAY"
	CandleIntervalWeek  CandleInterval = "CANDLE_INTERVAL_WEEK"
	CandleIntervalMonth CandleInterval = "CANDLE_INTERVAL_MONTH"
)

func (c *Client) getCandlesOnce(ctx context.Context, instrumentUID string, from, to time.Time, interval CandleInterval) (*Candles, error) {
	type request struct {
		InstrumentID     string         `json:"instrumentId"`
		From             time.Time      `json:"from"`
		To               time.Time      `json:"to"`
		Interval         CandleInterval `json:"interval"`
		CandleSourceType string         `json:"candleSourceType"`
	}

	const method = "MarketDataService/GetCandles"

	data, err := c.DoRequest(ctx, method, request{
		InstrumentID:     instrumentUID,
		From:             from,
		To:               to,
		Interval:         interval,
		CandleSourceType: "CANDLE_SOURCE_EXCHANGE", // только биржевые торги
	})
	if err != nil {
		return nil, err
	}

	var candles Candles
	if err := json.Unmarshal(data, &candles); err != nil {
		return nil, fmt.Errorf("%s: unmarshal response: %w", method, err)
	}

	return &candles, nil
}

func (c *Client) GetCandles(ctx context.Context, instrumentUID string, from, to time.Time, interval CandleInterval) (*Candles, error) {
	result := &Candles{}
	var last time.Time

	for start := from; start.Before(to); start = start.AddDate(1, 0, 0) {
		end := start.AddDate(1, 0, 0)
		if end.After(to) {
			end = to
		}

		chunk, err := c.getCandlesOnce(ctx, instrumentUID, start, end, interval)
		if err != nil {
			return nil, err
		}

		if chunk.PriceCurrency != "" {
			result.PriceCurrency = chunk.PriceCurrency
		}
		for _, candle := range chunk.Candles {
			if !candle.Time.After(last) {
				continue
			}
			result.Candles = append(result.Candles, candle)
			last = candle.Time
		}
	}
	return result, nil
}
