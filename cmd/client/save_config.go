package main

import (
	"encoding/json"
	"os"
	"time"
)

func (cfg Config) saveAccount() error {
	data := struct {
		AccountID  string    `json:"accountID"`
		OpenedDate time.Time `json:"openedDate"`
	}{
		AccountID:  cfg.accountID,
		OpenedDate: cfg.openedDate,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("t-invest-api-account.json", jsonData, 0644)
}
