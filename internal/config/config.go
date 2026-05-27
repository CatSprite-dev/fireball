package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL string

	RedisURL          string
	RedisTTL          time.Duration
	PortfolioCacheTTL time.Duration
	ChartDataCacheTTL time.Duration

	PostgresURL string

	ServerPort   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	sessionSecret string
}

func NewConfig() (*Config, error) {
	_, err := os.Stat(".env")
	if err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file exists but failed to load: %v", err)
		}
	}

	investURL := os.Getenv("T_INVEST_URL")
	if investURL == "" {
		log.Println("T_INVEST_URL variable is not found in environment")
		return nil, err
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Println("REDIS_URL variable is not found in environment")
		return nil, err
	}
	redisTTLStr := os.Getenv("REDIS_TTL")
	if redisTTLStr == "" {
		log.Println("REDIS_TTL variable is not found in environment\nSetting default 24h")
		redisTTLStr = "24"
	}
	portfolioCacheTTLStr := os.Getenv("PORTFOLIO_CACHE_TTL")
	if portfolioCacheTTLStr == "" {
		log.Println("PORTFOLIO_CACHE_TTL variable is not found in environment\nSetting default 3m")
		portfolioCacheTTLStr = "3"
	}
	chartDataCacheTTLStr := os.Getenv("CHART_DATA_CACHE_TTL")
	if chartDataCacheTTLStr == "" {
		log.Println("CHART_DATA_CACHE_TTL variable is not found in environment\nSetting default 30m")
		chartDataCacheTTLStr = "30"
	}
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		log.Println("SESSION_SECRET variable is not found in environment")
		return nil, err
	}
	db_user := os.Getenv("DB_USER")
	if db_user == "" {
		log.Println("DB_USER variable is not found in environment")
		return nil, err
	}

	db_password := os.Getenv("DB_PASSWORD")
	if db_password == "" {
		log.Println("DB_PASSWORD variable is not found in environment")
		return nil, err
	}

	db_name := os.Getenv("DB_NAME")
	if db_name == "" {
		log.Println("DB_NAME variable is not found in environment")
		return nil, err
	}

	db_host := "postgres"

	postgresURL := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable",
		db_user,
		db_password,
		db_host,
		db_name,
	)

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		log.Println("PORT variable is not found in environment\nSetting default 8080")
		serverPort = "8080"
	}
	readTimeoutStr := os.Getenv("READ_TIMEOUT")
	if readTimeoutStr == "" {
		log.Println("READ_TIMEOUT variable is not found in environment\nSetting default 10s")
		readTimeoutStr = "10"
	}
	writeTimeoutStr := os.Getenv("WRITE_TIMEOUT")
	if writeTimeoutStr == "" {
		log.Println("WRITE_TIMEOUT variable is not found in environment\nSetting default 10s")
		writeTimeoutStr = "10"
	}
	idleTimeoutStr := os.Getenv("IDLE_TIMEOUT")
	if idleTimeoutStr == "" {
		log.Println("IDLE_TIMEOUT variable is not found in environment\nSetting default 30s")
		idleTimeoutStr = "30"
	}

	redisTTL, err := strconv.Atoi(redisTTLStr)
	if err != nil {
		log.Println("Wrong format of REDIS_TTL\nSetting default 24h")
		redisTTL = 24
	}
	portfolioCacheTTL, err := strconv.Atoi(portfolioCacheTTLStr)
	if err != nil {
		log.Println("Wrong format of PORTFOLIO_CACHE_TTL\nSetting default 3m")
		portfolioCacheTTL = 3
	}
	chartDataCacheTTL, err := strconv.Atoi(chartDataCacheTTLStr)
	if err != nil {
		log.Println("Wrong format of CHART_DATA_CACHE_TTL\nSetting default 3m")
		chartDataCacheTTL = 3
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

	return &Config{
		BaseURL: investURL,

		RedisURL:          redisURL,
		RedisTTL:          time.Duration(redisTTL) * time.Hour,
		PortfolioCacheTTL: time.Duration(portfolioCacheTTL) * time.Minute,
		ChartDataCacheTTL: time.Duration(chartDataCacheTTL) * time.Minute,

		PostgresURL: postgresURL,

		ServerPort:   serverPort,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,

		sessionSecret: secret,
	}, nil
}

func (c *Config) GetSecret() string {
	return c.sessionSecret
}
