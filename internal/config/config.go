package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL   string
	InvestURL string

	ServerPort   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func NewConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("failed loading .env file")
		os.Exit(1)
	}
	investURL := os.Getenv("T_INVEST_URL")
	if investURL == "" {
		log.Println("failed loading investURL")
		os.Exit(1)
	}
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		log.Println("PORT variable is not found in environment")
		os.Exit(1)
	}
	readTimeout := 10 * time.Second
	writeTimeout := 10 * time.Second
	idleTimeout := 30 * time.Second

	baseURL := investURL

	return &Config{
		BaseURL:   baseURL,
		InvestURL: investURL,

		ServerPort:   serverPort,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}
