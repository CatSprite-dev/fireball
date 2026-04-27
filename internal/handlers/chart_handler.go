package handlers

import (
	"log"
	"net/http"

	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/service"
	"github.com/CatSprite-dev/fireball/internal/storage"
)

type ChartHandler struct {
	portfolioService *service.PortfolioService
	sessionManager   *storage.SessionManager
}

func NewChartHandler(sm *storage.SessionManager, ps *service.PortfolioService) *ChartHandler {
	return &ChartHandler{
		portfolioService: ps,
		sessionManager:   sm,
	}
}

func (h *ChartHandler) HandlerChart(w http.ResponseWriter, r *http.Request) {
	log.Println("serving chart handler")
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

	period := r.URL.Query().Get("period")
	index := r.URL.Query().Get("index")
	log.Printf("%s %s", period, index)

	chartData, err := h.portfolioService.GetOrFetchChartData(r.Context(),
		sessionID,
		request.Token,
		userPortfolio,
		period,
		index,
		pkg.CandleSourceExchange)
	if err != nil {
		log.Printf("GetChartData error: %v", err)
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	pkg.RespondWithJSON(w, http.StatusOK, chartData)
}
