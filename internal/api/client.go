package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	httpClient http.Client
	baseURL    string
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		httpClient: http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

func (client *Client) DoRequest(url string, token string, payload string) ([]byte, error) {
	body := strings.NewReader(payload)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("request error: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("response error: %s", err)
	}

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read error: %s", err)
	}

	return data, nil
}
