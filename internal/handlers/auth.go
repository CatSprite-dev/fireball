package handlers

import (
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/service"
)

type AuthHandler struct {
	portfolioService *service.Calculator
}

func NewAuthHandler(calc *service.Calculator) *AuthHandler {
	return &AuthHandler{portfolioService: calc}
}

func (h *AuthHandler) HandlerAuth(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startAlloc := m.Alloc

	type returnVals struct {
		UserPortfolio domain.Portfolio `json:"user_portfolio"`
		ChartData     domain.ChartData `json:"chart_data"`
	}

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

	chartData, err := h.portfolioService.GetChartData(token, userPortfolio, "IMOEX", time.Now().AddDate(0, -6, 0), time.Now(), pkg.CandleIntervalDay, pkg.CandleSourceExchange)
	if err != nil {
		log.Printf("GetChartData error: %v", err)
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	pkg.RespondWithJSON(w, http.StatusOK, returnVals{
		UserPortfolio: userPortfolio,
		ChartData:     chartData,
	})

	runtime.ReadMemStats(&m)
	log.Printf("Память: использовано %d MB, всего %d MB, системой %d MB, сборок мусора %d",
		(m.Alloc-startAlloc)/1024/1024,
		m.Alloc/1024/1024,
		m.Sys/1024/1024,
		m.NumGC)

	log.Printf("Число запросов HandlerAuth = %d", h.portfolioService.ApiClient.RequestCount())
	h.portfolioService.ApiClient.ResetRequestCount()
}
