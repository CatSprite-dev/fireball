package tinvest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	limiter    *Limiter
	logger     *slog.Logger
	semaphore  chan struct{}
}

func New(token string, logger *slog.Logger) *Client {
	return &Client{
		baseURL:    "https://invest-public-api.tbank.ru/rest/tinkoff.public.invest.api.contract.v1",
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		limiter:    NewLimiter(),
		logger:     logger.With("component", "tinvest"),
		semaphore:  make(chan struct{}, 10), // не больше 10 запросов одновременно
	}
}

func (client *Client) DoRequest(ctx context.Context, method string, payload any) ([]byte, error) {
	waitStart := time.Now()
	if err := client.limiter.Wait(ctx, method); err != nil {
		return nil, fmt.Errorf("%s: rate limit: %w", method, err)
	}
	if waited := time.Since(waitStart); waited > 100*time.Millisecond {
		client.logger.Debug("rate limit wait", "method", method, "waited", waited)
	}

	select {
	case client.semaphore <- struct{}{}:
		defer func() { <-client.semaphore }()
	case <-ctx.Done():
		return nil, fmt.Errorf("%s: wait for slot: %w", method, ctx.Err())
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: marshal payload: %w", method, err)
	}

	url := client.baseURL + "." + method
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", method, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+client.token)

	start := time.Now()
	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: send request: %w", method, err)
	}
	defer res.Body.Close()

	client.logger.Debug("request",
		"method", method,
		"status", res.StatusCode,
		"duration", time.Since(start),
		"tracking_id", res.Header.Get("x-tracking-id"),
	)

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: read response: %w", method, err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, newAPIError(res, data)
	}

	return data, nil
}
