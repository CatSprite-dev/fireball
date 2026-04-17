package main

import (
	"log"
	"net/http"
	"os"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/config"
	"github.com/CatSprite-dev/fireball/internal/handlers"
	"github.com/CatSprite-dev/fireball/internal/service"
	"github.com/CatSprite-dev/fireball/internal/storage"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		os.Exit(1)
	}

	store, err := storage.NewRedisStore(cfg.RedisURL)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	sessionManager, err := storage.NewSessionManager(store, cfg.GetSecret(), cfg.RedisTTL)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	cacheManager := storage.NewCacheManager(store, cfg.CacheTTL)

	apiClient := api.NewClient(cfg.BaseURL)
	calculator := service.NewCalculator(apiClient)
	portfolioService := service.NewPortfolioService(calculator, cacheManager)

	loginHandler := handlers.NewLoginHandler(sessionManager, cacheManager, apiClient)

	loginRateLimiter := handlers.NewRateLimiter(10)
	authRateLimiter := handlers.NewRateLimiter(200)
	portfolioHandler := handlers.NewPortfolioHandler(sessionManager, portfolioService)
	chartHandler := handlers.NewChartHandler(sessionManager, portfolioService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/ping", authRateLimiter.Middleware(portfolioHandler.HandlerPing))
	mux.HandleFunc("POST /api/login", loginRateLimiter.Middleware(loginHandler.HandlerLogin))
	mux.HandleFunc("POST /api/logout", loginRateLimiter.Middleware(loginHandler.HandlerLogout))
	mux.HandleFunc("POST /api/portfolio", authRateLimiter.Middleware(portfolioHandler.HandlerPortfolio))
	mux.HandleFunc("GET /api/chart", authRateLimiter.Middleware(chartHandler.HandlerChart))

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Printf("Serving on: http://localhost:%s/\n", cfg.ServerPort)
	log.Fatal(srv.ListenAndServe())
}
