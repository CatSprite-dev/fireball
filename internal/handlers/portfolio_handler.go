package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/CatSprite-dev/fireball/internal/pkg"
	"github.com/CatSprite-dev/fireball/internal/service"
)

type PortfolioHandler struct {
	portfolioService *service.PortfolioService
}

func NewPortfolioHandler(ps *service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioService: ps,
	}
}

func (h *PortfolioHandler) HandlerPing(w http.ResponseWriter, r *http.Request) {
	log.Println("serving ping handler")
	w.WriteHeader(http.StatusOK)
}

func (h *PortfolioHandler) HandlerPortfolio(w http.ResponseWriter, r *http.Request) {
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

	pkg.RespondWithJSON(w, http.StatusOK, userPortfolio)
}
