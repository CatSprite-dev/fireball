package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL    string
	InvestURL  string
	SandboxUrl string

	ServerPort string
	Timeout    time.Duration
}

func NewConfig() *Config {
	godotenv.Load(".env")
	investUrl := os.Getenv("investURL")
	sandboxUrl := os.Getenv("sandboxUrl")
	port := os.Getenv("port")
	baseUrl := investUrl
	timeout := 5 * time.Second

	return &Config{
		BaseURL:    baseUrl,
		InvestURL:  investUrl,
		SandboxUrl: sandboxUrl,
		ServerPort: port,
		Timeout:    timeout,
	}
}
