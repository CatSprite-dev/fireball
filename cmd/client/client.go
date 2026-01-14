package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/CatSprite-dev/fireball/internal/cache"
)

type Client struct {
	baseURL        string
	baseUrlSandbox string
	httpClient     http.Client
	cache          *cache.Cache
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		baseURL:        baseUrl,
		baseUrlSandbox: baseUrlSandbox,
		httpClient: http.Client{
			Timeout: timeout,
		},
		cache: cache.NewCache(5 * time.Second),
	}
}

func (c *Client) GetBaseURL() *string {
	return &c.baseURL
}

func (c *Client) DoRequest(url string, token string, payload string) ([]byte, error) {
	body := strings.NewReader(payload)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("request error: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("response error: %s", err)
	}

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read error: %s", err)
	}

	c.cache.Add(url, data)
	return data, nil
}
