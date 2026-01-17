package main

import (
	"net/http"
)

func (cfg *Config) handlerGetPortfolio(w http.ResponseWriter, req *http.Request) {
	token, err := getTokenFromHeader(req.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
	}
	userPortfolio, err := GetPositionsInfo(cfg, token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
	}
	respondWithJSON(w, http.StatusOK, userPortfolio)
}
