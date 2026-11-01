package tinvest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var ErrUnknownMethod = errors.New("tinvest: unknown method")

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	TrackingID string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("tinvest: API error %s: %s (HTTP %d, tracking_id %s)",
		e.Code, e.Message, e.StatusCode, e.TrackingID)
}

func newAPIError(res *http.Response, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: res.StatusCode,
		TrackingID: res.Header.Get("x-tracking-id"),
	}

	var parsed struct {
		Code        json.Number `json:"code"`
		Message     string      `json:"message"`
		Description string      `json:"description"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		apiErr.Message = string(bytes.TrimSpace(body[:min(len(body), 200)]))
		return apiErr
	}

	apiErr.Code = parsed.Code.String()
	apiErr.Message = parsed.Description
	if apiErr.Message == "" {
		apiErr.Message = parsed.Message
	}
	return apiErr
}
