package main

import (
	"time"
)

type Config struct {
	client Client
}

func NewConfig() Config {
	client := NewClient(5 * time.Second)

	cfg := Config{
		client: *client,
	}

	return cfg
}
