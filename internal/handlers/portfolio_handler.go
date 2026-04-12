package handlers

import (
	"log"
	"net/http"

	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/service"
)

type PortfolioHandler struct {
	portfolioService *service.Calculator
}

func NewPortfolioHandler(calc *service.Calculator) *PortfolioHandler {
	return &PortfolioHandler{portfolioService: calc}
}

func (h *PortfolioHandler) HandlerPortfolio(w http.ResponseWriter, r *http.Request) {
	token, err := getTokenFromHeader(r.Header)
	if err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}
	userPortfolio, err := h.portfolioService.GetPortfolio(token)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	pkg.RespondWithJSON(w, http.StatusOK, userPortfolio)

	log.Printf("Число запросов HandlerPortfolio = %d", h.portfolioService.ApiClient.RequestCount())
	h.portfolioService.ApiClient.ResetRequestCount()
}
