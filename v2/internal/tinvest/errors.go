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
	StatusCode int // HTTP статус
	GRPCCode   int
	Code       string // код T-Invest
	Message    string
	TrackingID string
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("tinvest: API error %s: %s (HTTP %d, gRPC %d)",
		e.Code, e.Message, e.StatusCode, e.GRPCCode)
	if e.TrackingID != "" {
		msg += ", tracking_id " + e.TrackingID
	}
	return msg
}

func newAPIError(res *http.Response, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: res.StatusCode,
		TrackingID: res.Header.Get(headerTrackingID),
	}

	var parsed struct {
		Code        int    `json:"code"`
		Message     string `json:"message"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		apiErr.Message = string(bytes.TrimSpace(body[:min(len(body), 200)]))
		return apiErr
	}

	apiErr.GRPCCode = parsed.Code
	apiErr.Code = parsed.Description
	apiErr.Message = parsed.Message

	return apiErr
}
