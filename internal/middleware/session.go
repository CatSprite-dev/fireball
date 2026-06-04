package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/CatSprite-dev/fireball/internal/storage"
)

type sessionContextKey struct{}

func SessionFromContext(ctx context.Context) storage.SessionData {
	session, ok := ctx.Value(sessionContextKey{}).(storage.SessionData)
	if !ok {
		return storage.SessionData{}
	}
	return session
}

func GetSessionFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func SetSessionCookie(w http.ResponseWriter, sessionID string, expireIn time.Duration, setToDelete bool, secure bool) {
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
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	}
	http.SetCookie(w, &cookie)
}
