package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/service"
	"github.com/CatSprite-dev/fireball/internal/storage"
)

type ChartHandler struct {
	portfolioService *service.Calculator
	sessionManager   *storage.SessionManager
}

func NewChartHandler(sm *storage.SessionManager, calc *service.Calculator) *ChartHandler {
	return &ChartHandler{
		portfolioService: calc,
		sessionManager:   sm,
	}
}

func (h *ChartHandler) HandlerChart(w http.ResponseWriter, r *http.Request) {
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

	token := sessionData.Token

	var userPortfolio domain.Portfolio
	err = json.NewDecoder(r.Body).Decode(&userPortfolio)
	if err != nil {
		pkg.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	period := r.URL.Query().Get("period")
	index := r.URL.Query().Get("index")

	from, to, interval := PeriodToParams(period)

	chartData, err := h.portfolioService.GetChartData(token, userPortfolio, index, from, to, interval, pkg.CandleSourceExchange)
	if err != nil {
		log.Printf("GetChartData error: %v", err)
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	pkg.RespondWithJSON(w, http.StatusOK, chartData)

	log.Printf("Requrests number of HandlerChart = %d", h.portfolioService.ApiClient.RequestCount())
	h.portfolioService.ApiClient.ResetRequestCount()
}
