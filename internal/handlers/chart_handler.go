package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/service"
)

type ChartHandler struct {
	portfolioService *service.PortfolioService
}

func NewChartHandler(ps *service.PortfolioService) *ChartHandler {
	return &ChartHandler{
		portfolioService: ps,
	}
}

func (h *ChartHandler) HandlerChart(w http.ResponseWriter, r *http.Request) {
	log.Println("serving chart handler")
	force, err := strconv.ParseBool(r.URL.Query().Get("force"))
	if err != nil {
		log.Println("'force' param incorrect, the default false param is set")
		force = false
	}

	sessionData := SessionFromContext(r.Context())

	request := service.PortfolioRequest{
		Token:      sessionData.Token,
		AccountID:  sessionData.AccountID,
		OpenedDate: sessionData.OpenedDate,
	}

	userPortfolio, err := h.portfolioService.GetOrFetchPortfolio(r.Context(), force, sessionData.ID, request)
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	period := r.URL.Query().Get("period")
	index := r.URL.Query().Get("index")
	log.Printf("%s %s", period, index)

	chartData, err := h.portfolioService.GetOrFetchChartData(r.Context(),
		force,
		request.ID,
		request.Token,
		userPortfolio,
		period,
		index,
		pkg.CandleSourceExchange)
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		log.Printf("GetChartData error: %v", err)
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	pkg.RespondWithJSON(w, http.StatusOK, chartData)
}
