package main

import (
	"encoding/json"
	"os"
)

func (cfg Config) saveAccount() error {
	data, err := json.MarshalIndent(cfg.accountID, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("t-invest-api-account.json", data, 0644)
}
