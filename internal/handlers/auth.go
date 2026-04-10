package handlers

import (
	"net/http"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/service"
	"github.com/CatSprite-dev/fireball/internal/session"
)

type AuthHandler struct {
	portfolioService *service.Calculator
	sessionManager   *session.SessionManager
}

func NewAuthHandler(sm *session.SessionManager, calc *service.Calculator) *AuthHandler {
	return &AuthHandler{
		portfolioService: calc,
		sessionManager:   sm,
	}
}

func (h *AuthHandler) HandlerAuth(w http.ResponseWriter, r *http.Request) {
	type returnVals struct {
		UserPortfolio domain.UserFullPortfolio `json:"user_portfolio"`
	}

	sessionID, err := getSessionFromCookie(r)
	if err != nil {
		pkg.RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	sessionData, err := h.sessionManager.GetSession(r.Context(), sessionID)
	if err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, "invalid session", err)
		return
	}

	request := service.PortfolioRequest{
		Token:      sessionData.Token,
		AccountID:  sessionData.AccountID,
		OpenedDate: sessionData.OpenedDate,
	}

	userPortfolio, err := h.portfolioService.GetFullPortfolio(request)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	pkg.RespondWithJSON(w, http.StatusOK, returnVals{
		UserPortfolio: userPortfolio,
	})
}
