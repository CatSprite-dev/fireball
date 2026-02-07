package handlers

import (
	"errors"
	"net/http"
	"strings"
)

func getTokenFromHeader(headers http.Header) (string, error) {
	token := headers.Get("T-Token")
	if token == "" {
		return "", errors.New("token is not provided")
	}
	if len(token) < 10 || !strings.HasPrefix(token, "t.") {
		return "", errors.New("invalid token")
	}
	return token, nil
}
