package tinvest

import (
	"errors"
	"fmt"
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("tinvest: API error %s: %s (HTTP %d)", e.Code, e.Message, e.StatusCode)
}

var ErrUnknownMethod = errors.New("tinvest: unknown method")
