package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/service"
)

const demoAccountID = "demo-account"
const demoOpenedDate = "2023-01-01"

type DemoHandler struct {
	portfolioService *service.PortfolioService
}

func NewDemoHandler(ps *service.PortfolioService) *DemoHandler {
	return &DemoHandler{portfolioService: ps}
}

func (h *DemoHandler) HandlerDemoPortfolio(w http.ResponseWriter, r *http.Request) {
	force, err := strconv.ParseBool(r.URL.Query().Get("force"))
	if err != nil {
		force = false
	}

	openedDate, _ := time.Parse("2006-01-02", demoOpenedDate)

	request := service.PortfolioRequest{
		Token:      "",
		AccountID:  demoAccountID,
		OpenedDate: openedDate,
	}

	portfolio, err := h.portfolioService.GetOrFetchPortfolio(r.Context(), force, demoAccountID, request)
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		log.Printf("demo portfolio error: %v", err)
		pkg.RespondWithError(w, http.StatusInternalServerError, "failed to load demo portfolio", err)
		return
	}

	pkg.RespondWithJSON(w, http.StatusOK, portfolio)
}

func (h *DemoHandler) HandlerDemoChart(w http.ResponseWriter, r *http.Request) {
	force, err := strconv.ParseBool(r.URL.Query().Get("force"))
	if err != nil {
		force = false
	}

	openedDate, _ := time.Parse("2006-01-02", demoOpenedDate)

	request := service.PortfolioRequest{
		Token:      "",
		AccountID:  demoAccountID,
		OpenedDate: openedDate,
	}

	portfolio, err := h.portfolioService.GetOrFetchPortfolio(r.Context(), force, demoAccountID, request)
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		log.Printf("demo portfolio error: %v", err)
		pkg.RespondWithError(w, http.StatusInternalServerError, "failed to load demo portfolio", err)
		return
	}

	period := r.URL.Query().Get("period")
	index := r.URL.Query().Get("index")

	chartData, err := h.portfolioService.GetOrFetchChartData(
		r.Context(),
		force,
		demoAccountID,
		"",
		portfolio,
		period,
		index,
		pkg.CandleSourceExchange)
	if err != nil {
		if r.Context().Err() != nil {
			return
		}
		log.Printf("demo chart error: %v", err)
		pkg.RespondWithError(w, http.StatusInternalServerError, "failed to load demo chart", err)
		return
	}

	pkg.RespondWithJSON(w, http.StatusOK, chartData)
}
