package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/storage"
)

type LoginHandler struct {
	sessionManager *storage.SessionManager
	cacheManager   *storage.CacheManager
	apiClient      *api.Client
	secure         bool
}

func NewLoginHandler(sm *storage.SessionManager, cm *storage.CacheManager, client *api.Client, secure bool) *LoginHandler {
	return &LoginHandler{
		sessionManager: sm,
		cacheManager:   cm,
		apiClient:      client,
		secure:         secure,
	}
}

func (h *LoginHandler) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandlerLogin called, method: %s", r.Method)
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

	userAccounts, err := h.apiClient.GetAccounts(r.Context(), req.Token, pkg.AccountStatusOpen)
	if err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}
	if len(userAccounts.Accounts) == 0 {
		pkg.RespondWithError(w, http.StatusUnauthorized, "invalid token", nil)
		return
	}

	if len(userAccounts.Accounts) > 1 {
		log.Printf("user has %d accounts, using first one", len(userAccounts.Accounts))
	}
	accountID := userAccounts.Accounts[0].ID
	openedDate := userAccounts.Accounts[0].OpenedDate

	sessionID, err := h.sessionManager.CreateSession(r.Context(), req.Token, accountID, openedDate)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	setSessionCookie(w, sessionID, h.sessionManager.RedisTTL, false, h.secure)
	log.Printf("Cookie for %s is set, session: %s\n", accountID, sessionID)
	pkg.RespondWithJSON(w, http.StatusOK, struct{}{})
}

func (h *LoginHandler) HandlerLogout(w http.ResponseWriter, r *http.Request) {
	sessionID, err := getSessionFromCookie(r)
	if err != nil {
		pkg.RespondWithError(w, http.StatusBadRequest, "missing session cookie", err)
		return
	}

	err = h.cacheManager.DeletePortfolio(r.Context(), sessionID)
	if err != nil {
		log.Printf("couldn't delete cache for %s: %v", sessionID, err)
	}
	err = h.cacheManager.DeleteChartCache(r.Context(), sessionID)
	if err != nil {
		log.Printf("couldn't delete cache for %s: %v", sessionID, err)
	}

	err = h.sessionManager.DeleteSession(r.Context(), sessionID)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	setSessionCookie(w, sessionID, h.sessionManager.RedisTTL, true, h.secure)

	log.Printf("Cookie for %s is deleted\n", sessionID)

	pkg.RespondWithJSON(w, http.StatusOK, struct{}{})
}
