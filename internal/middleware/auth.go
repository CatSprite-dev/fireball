package middleware

import (
	"context"
	"net/http"

	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/storage"
)

func AuthMiddleware(sm *storage.SessionManager, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := GetSessionFromCookie(r)
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
