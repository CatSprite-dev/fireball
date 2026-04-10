package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/session"
)

type LoginHandler struct {
	sessionManager *session.SessionManager
	apiClient      *api.Client
}

func NewLoginHandler(sm *session.SessionManager, client *api.Client) *LoginHandler {
	return &LoginHandler{
		sessionManager: sm,
		apiClient:      client,
	}
}

func (h *LoginHandler) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	token, err := getTokenFromHeader(r.Header)
	if err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	userAccounts, err := h.apiClient.GetAccounts(token, pkg.AccountStatusOpen)
	if err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}
	if len(userAccounts.Accounts) == 0 {
		pkg.RespondWithError(w, http.StatusBadRequest, "found no accounts", errors.New("found no accounts"))
		return
	}
	accountID := userAccounts.Accounts[0].ID
	openedDate := userAccounts.Accounts[0].OpenedDate

	sessionID, err := h.sessionManager.CreateSession(context.Background(), token, accountID, openedDate)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	setSessionCookie(w, sessionID, false)
	log.Printf("Cookie for %s is set, session: %s\n", accountID, sessionID)
	pkg.RespondWithJSON(w, http.StatusOK, struct{}{})
}

func (h *LoginHandler) HandlerLogout(w http.ResponseWriter, r *http.Request) {
	sessionID, err := getSessionFromCookie(r)
	if err != nil {
		pkg.RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = h.sessionManager.DeleteSession(context.Background(), sessionID)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	setSessionCookie(w, sessionID, true)
	log.Printf("Cookie for %s is deleted\n", sessionID)
	pkg.RespondWithJSON(w, http.StatusOK, struct{}{})
}
