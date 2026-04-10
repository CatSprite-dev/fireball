package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/service"
)

type ChartHandler struct {
	portfolioService *service.Calculator
}

func NewChartHandler(calc *service.Calculator) *ChartHandler {
	return &ChartHandler{portfolioService: calc}
}

func (h *ChartHandler) HandlerChart(w http.ResponseWriter, r *http.Request) {
	token, err := getTokenFromHeader(r.Header)
	if err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

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

	log.Printf("Число запросов HandlerChart = %d", h.portfolioService.ApiClient.RequestCount())
	h.portfolioService.ApiClient.ResetRequestCount()
}
