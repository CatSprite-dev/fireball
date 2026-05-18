package handlers

import (
	"net/http"
	"os"
	"time"
)

func setSessionCookie(w http.ResponseWriter, sessionID string, expireIn time.Duration, setToDelete bool) {
	expiration := time.Now().Add(time.Duration(expireIn) * time.Hour)
	maxAge := 0
	if setToDelete {
		maxAge = -1
		expiration = time.Unix(0, 0)
	}
	cookie := http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
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
