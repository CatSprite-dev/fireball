package main

import (
	"log"
	"time"
)

type Config struct {
	client     Client
	accountID  string
	openedDate time.Time
}

func NewConfig() Config {
	client := NewClient(5 * time.Second)
	account, err := client.GetBankAccount()
	if err != nil {
		log.Fatal(err)
	}

	cfg := Config{
		client:     *client,
		accountID:  account.Accounts[0].ID,
		openedDate: account.Accounts[0].OpenedDate,
	}

	cfg.saveAccount()
	return cfg
}
