package handlers

import (
	"net/http"

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
	type returnVals struct {
		UserPortfolio domain.UserFullPortfolio `json:"user_portfolio"`
	}

	token, err := getTokenFromHeader(r.Header)
	if err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}
	userPortfolio, err := h.portfolioService.GetFullPortfolio(token)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	pkg.RespondWithJSON(w, http.StatusOK, returnVals{
		UserPortfolio: userPortfolio,
	})
}
