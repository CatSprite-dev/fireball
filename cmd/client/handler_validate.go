package main

import (
	"errors"
	"net/http"
)

func getTokenFromHeader(headers http.Header) (string, error) {
	token := headers.Get("T-Token")
	if token == "" {
		return "", errors.New("token is not provided")
	}
	//check if only numerical? other validations on server before sending to T-API?
	return token, nil
}

func (cfg *Config) HandlerAuth(w http.ResponseWriter, req *http.Request) {
	token, err := getTokenFromHeader(req.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid token", err)
	}

	_, err = cfg.GetUserInfo(token)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid token", err)
	}

	respondWithJSON(w, 200, struct{}{})
}
