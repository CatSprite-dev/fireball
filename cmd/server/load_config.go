package main

import (
	"encoding/json"
	"os"
	"time"
)

func LoadConfig() (Config, error) {
	var cfg Config

	client := NewClient(5 * time.Second)
	cfg.client = *client

	file, err := os.ReadFile("t-invest-api-account.json")
	if err != nil {
		if os.IsNotExist(err) {
			return NewConfig(), nil
		}
		return cfg, err
	}

	var savedData struct {
		AccountID  string    `json:"accountID"`
		OpenedDate time.Time `json:"openedDate"`
	}

	err = json.Unmarshal(file, &savedData)
	if err != nil {
		return cfg, err
	}

	cfg.accountID = savedData.AccountID
	cfg.openedDate = savedData.OpenedDate

	return cfg, nil
}
