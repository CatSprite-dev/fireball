package handlers

import (
	"context"
	"encoding/json"
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
	type loginRequest struct {
		Token string `json:"token"`
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.RespondWithError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if req.Token == "" {
		pkg.RespondWithError(w, http.StatusBadRequest, "token is required", nil)
		return
	}

	userAccounts, err := h.apiClient.GetAccounts(req.Token, pkg.AccountStatusOpen)
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

	sessionID, err := h.sessionManager.CreateSession(context.Background(), req.Token, accountID, openedDate)
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
