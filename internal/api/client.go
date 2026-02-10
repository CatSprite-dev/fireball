package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient http.Client
	baseURL    string
}

func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (client *Client) DoRequest(url string, httpMethod string, token string, payload interface{}) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("payload marshal error:  %v", err)
	}

	req, err := http.NewRequest(httpMethod, url, bytes.NewBuffer(body))
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
