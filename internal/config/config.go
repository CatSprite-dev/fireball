package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL   string
	InvestURL string

	RedisURL string
	RedisTTL time.Duration

	ServerPort   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	sessionSecret string
}

func NewConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("failed loading .env file: %v\n", err)
		os.Exit(1)
	}
	investURL := os.Getenv("T_INVEST_URL")
	if investURL == "" {
		log.Println("T_INVEST_URL variable is not found in environment")
		os.Exit(1)
	}
	redisURL := os.Getenv("REDIS_URL")
	if investURL == "" {
		log.Println("REDIS_URL variable is not found in environment")
		os.Exit(1)
	}
	redisTTLStr := os.Getenv("REDIS_TTL")
	if investURL == "" {
		log.Println("REDIS_TTL variable is not found in environment\nSetting default 24h")
		redisTTLStr = "24"
	}
	secret := os.Getenv("SESSION_SECRET")
	if investURL == "" {
		log.Println("SESSION_SECRET variable is not found in environment")
		os.Exit(1)
	}
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		log.Println("PORT variable is not found in environment\nSetting default 8080")
		serverPort = "8080"
	}
	readTimeoutStr := os.Getenv("READ_TIMEOUT")
	if serverPort == "" {
		log.Println("READ_TIMEOUT variable is not found in environment\nSetting default 10s")
		readTimeoutStr = "10"
	}
	writeTimeoutStr := os.Getenv("WRITE_TIMEOUT")
	if serverPort == "" {
		log.Println("WRITE_TIMEOUT variable is not found in environment\nSetting default 10s")
		writeTimeoutStr = "10"
	}
	idleTimeoutStr := os.Getenv("IDLE_TIMEOUT")
	if serverPort == "" {
		log.Println("IDLE_TIMEOUT variable is not found in environment\nSetting default 30s")
		idleTimeoutStr = "30"
	}

	redisTTL, err := strconv.Atoi(redisTTLStr)
	if err != nil {
		log.Println("Wrong format of REDIS_TTL\nSetting default 24h")
		redisTTL = 24
	}
	readTimeout, err := strconv.Atoi(readTimeoutStr)
	if err != nil {
		log.Println("Wrong format of READ_TIMEOUT\nSetting default 10s")
		readTimeout = 10
	}
	writeTimeout, err := strconv.Atoi(writeTimeoutStr)
	if err != nil {
		log.Println("Wrong format of WRITE_TIMEOUT\nSetting default 10s")
		writeTimeout = 10
	}
	idleTimeout, err := strconv.Atoi(idleTimeoutStr)
	if err != nil {
		log.Println("Wrong format of IDLE_TIMEOUT\nSetting default 30s")
		idleTimeout = 30
	}

	baseURL := investURL

	return &Config{
		BaseURL:   baseURL,
		InvestURL: investURL,

		RedisURL: redisURL,
		RedisTTL: time.Duration(redisTTL) * time.Hour,

		ServerPort:   serverPort,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,

		sessionSecret: secret,
	}
}

func (c *Config) GetSecret() string {
	return c.sessionSecret
}
