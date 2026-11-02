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

const headerTrackingID = "Grpc-Metadata-X-Tracking-Id"

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

func (c *Client) DoRequest(ctx context.Context, method string, payload any) ([]byte, error) {
	waitStart := time.Now()
	if err := c.limiter.Wait(ctx, method); err != nil {
		return nil, fmt.Errorf("%s: rate limit: %w", method, err)
	}
	if waited := time.Since(waitStart); waited > 100*time.Millisecond {
		c.logger.Debug("rate limit wait", "method", method, "waited", waited)
	}

	select {
	case c.semaphore <- struct{}{}:
		defer func() { <-c.semaphore }()
	case <-ctx.Done():
		return nil, fmt.Errorf("%s: wait for slot: %w", method, ctx.Err())
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: marshal payload: %w", method, err)
	}

	url := c.baseURL + "." + method
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%s: create request: %w", method, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.token)

	start := time.Now()
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: send request: %w", method, err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: read response: %w", method, err)
	}

	c.logger.Debug("request",
		"method", method,
		"status", res.StatusCode,
		"duration", time.Since(start),
		"bytes", len(data),
		"tracking_id", res.Header.Get(headerTrackingID),
	)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, newAPIError(res, data)
	}

	return data, nil
}
