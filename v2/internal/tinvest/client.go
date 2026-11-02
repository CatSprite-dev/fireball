package tinvest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	headerTrackingID     = "Grpc-Metadata-X-Tracking-Id"
	headerRateLimitReset = "X-Ratelimit-Reset"
	maxAttempts          = 5
	baseBackoff          = 500 * time.Millisecond
	maxBackoff           = 10 * time.Second
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

func (c *Client) DoRequest(ctx context.Context, method string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: marshal request: %w", method, err)
	}

	for attempt := 1; ; attempt++ {
		data, retryAfter, err := c.doOnce(ctx, method, body)
		if err == nil {
			return data, nil
		}
		if attempt >= maxAttempts || !c.retryable(ctx, err) {
			return nil, err
		}

		wait := retryAfter
		if wait == 0 {
			wait = backoff(attempt)
		}
		c.logger.Warn("request failed, retrying",
			"method", method, "attempt", attempt, "wait", wait, "error", err)

		if err := sleep(ctx, wait); err != nil {
			return nil, fmt.Errorf("%s: wait before retry: %w", method, err)
		}
	}
}

func (c *Client) doOnce(ctx context.Context, method string, body []byte) ([]byte, time.Duration, error) {
	waitStart := time.Now()
	if err := c.limiter.Wait(ctx, method); err != nil {
		return nil, 0, fmt.Errorf("%s: rate limit: %w", method, err)
	}
	if waited := time.Since(waitStart); waited > 100*time.Millisecond {
		c.logger.Debug("rate limit wait", "method", method, "waited", waited)
	}

	select {
	case c.semaphore <- struct{}{}:
		defer func() { <-c.semaphore }()
	case <-ctx.Done():
		return nil, 0, fmt.Errorf("%s: wait for slot: %w", method, ctx.Err())
	}

	url := c.baseURL + "." + method
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("%s: create request: %w", method, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.token)

	start := time.Now()
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: send request: %w", method, err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: read response: %w", method, err)
	}

	c.logger.Debug("request",
		"method", method,
		"status", res.StatusCode,
		"duration", time.Since(start),
		"bytes", len(data),
		"tracking_id", res.Header.Get(headerTrackingID),
	)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var retryAfter time.Duration
		if res.StatusCode == http.StatusTooManyRequests {
			retryAfter = parseResetSeconds(res.Header.Get(headerRateLimitReset))
		}
		return nil, retryAfter, newAPIError(res, data)
	}

	return data, 0, nil
}

func (c *Client) retryable(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}
	if errors.Is(err, ErrUnknownMethod) {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500
	}
	return true
}

func backoff(attempt int) time.Duration {
	d := baseBackoff << (attempt - 1)
	if d > maxBackoff {
		d = maxBackoff
	}
	return rand.N(d)
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func parseResetSeconds(s string) time.Duration {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return 0
	}
	return time.Duration(n) * time.Second
}
