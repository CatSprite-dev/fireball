package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"
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

func setSessionCookie(w http.ResponseWriter, sessionID string, setToDelete bool) {
	maxAge := 0
	if setToDelete {
		maxAge = -1
	}
	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	}
	http.SetCookie(w, &cookie)
}

func getSessionFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
