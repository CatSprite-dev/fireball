package main

import (
	"time"
)

func LoadConfig() (Config, error) {
	var cfg Config

	client := NewClient(5 * time.Second)
	cfg.client = *client

	return cfg, nil
}
