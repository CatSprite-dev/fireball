package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"
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
	type returnVals struct {
		UserInfo      UserAccounts       `json:"user_info"`
		UserPortfolio UserPortfolio      `json:"user_portfolio"`
		UserDividends map[string]float64 `json:"user_dividends"`
	}

	token, err := getTokenFromHeader(req.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid token", err)
	}

	defer func() {
		cfg.client.baseURL = baseUrlInvest
	}()

	account, err := cfg.GetBankAccount(token)
	if len(account.Accounts) == 0 {
		cfg.client.baseURL = baseUrlSandbox
		time.Sleep(5000)
		account, err = cfg.GetBankAccount(token)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error(), err)
		}
	}
	fmt.Println(account.Accounts[0].ID)

	accountID := account.Accounts[0].ID
	openedDate := account.Accounts[0].OpenedDate

	portfolio, err := cfg.GetPortfolio(token, accountID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
	}

	dividends, err := cfg.GetDividends(token, accountID, openedDate, time.Now().UTC())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
	}

	respondWithJSON(w, 200, returnVals{
		UserInfo:      account,
		UserPortfolio: portfolio,
		UserDividends: dividends,
	})
}
