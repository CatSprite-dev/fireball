package main

import (
	"encoding/json"
	"os"
	"time"
)

func LoadConfig() (Config, error) {
	var cfg Config

	// Создаем клиент
	client := NewClient(5 * time.Second)
	cfg.client = *client

	// Пытаемся загрузить accountID из файла
	file, err := os.ReadFile("t-invest-api-account.json")
	if err != nil {
		// Если файла нет, создаем новую конфигурацию
		if os.IsNotExist(err) {
			return NewConfig(), nil
		}
		return cfg, err
	}

	// Загружаем accountID из файла
	err = json.Unmarshal(file, &cfg.accountID)
	return cfg, err
}
