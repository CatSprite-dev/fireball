package handlers

import (
	"log"
	"net/http"

	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/service"
	"github.com/CatSprite-dev/fireball/internal/storage"
)

type PortfolioHandler struct {
	portfolioService *service.PortfolioService
	sessionManager   *storage.SessionManager
}

func NewPortfolioHandler(sm *storage.SessionManager, ps *service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioService: ps,
		sessionManager:   sm,
	}
}

func (h *PortfolioHandler) HandlerPing(w http.ResponseWriter, r *http.Request) {
	log.Println("serving ping handler")
	sessionID, err := getSessionFromCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = h.sessionManager.GetSession(r.Context(), sessionID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PortfolioHandler) HandlerPortfolio(w http.ResponseWriter, r *http.Request) {
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

	userPortfolio, err := h.portfolioService.GetOrFetchPortfolio(r.Context(), sessionID, request)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	pkg.RespondWithJSON(w, http.StatusOK, userPortfolio)
}
