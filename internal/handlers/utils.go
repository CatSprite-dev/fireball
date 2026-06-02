package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/storage"
)

type sessionContextKey struct{}

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

func SessionFromContext(ctx context.Context) storage.SessionData {
	session, ok := ctx.Value(sessionContextKey{}).(storage.SessionData)
	if !ok {
		return storage.SessionData{}
	}
	return session
}

func AuthMiddleware(sm *storage.SessionManager, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := getSessionFromCookie(r)
		if err != nil {
			pkg.RespondWithError(w, http.StatusUnauthorized, "missing session cookie", err)
			return
		}

		sessionData, err := sm.GetSession(r.Context(), sessionID)
		if err != nil {
			pkg.RespondWithError(w, http.StatusUnauthorized, "invalid session", err)
			return
		}

		ctx := context.WithValue(r.Context(), sessionContextKey{}, sessionData)
		next(w, r.WithContext(ctx))
	}
}
