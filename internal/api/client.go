package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	httpClient         http.Client
	baseURL            string
	usersLimiter       *rate.Limiter
	operationsLimiter  *rate.Limiter
	instrumentsLimiter *rate.Limiter
}

func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL:            baseURL,
		usersLimiter:       rate.NewLimiter(rate.Every(time.Minute), 100),
		operationsLimiter:  rate.NewLimiter(rate.Every(time.Minute), 200),
		instrumentsLimiter: rate.NewLimiter(rate.Every(time.Minute), 200),
	}
}

type RequestError struct {
	StatusCode int
	Message    string
}

func (e RequestError) Error() string {
	return e.Message
}

func (client *Client) DoRequest(url string, httpMethod string, token string, payload interface{}) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("payload marshal error: %w", err)
	}

	req, err := http.NewRequest(httpMethod, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, RequestError{StatusCode: res.StatusCode, Message: fmt.Sprintf("unexpected status code: %d", res.StatusCode)}
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("response read error: %w", err)
	}

	return data, nil
}
