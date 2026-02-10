package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL    string
	InvestURL  string
	SandboxURL string

	ServerPort   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func NewConfig() *Config {
	godotenv.Load(".env")
	investURL := os.Getenv("investURL")
	sandboxURL := os.Getenv("sandboxUrl")

	serverPort := os.Getenv("PORT")
	readTimeout := 10 * time.Second
	writeTimeout := 10 * time.Second
	idleTimeout := 30 * time.Second

	baseURL := investURL

	return &Config{
		BaseURL:    baseURL,
		InvestURL:  investURL,
		SandboxURL: sandboxURL,

		ServerPort:   serverPort,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}
