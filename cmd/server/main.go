package main

import (
	"log"
	"net/http"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/config"
	"github.com/CatSprite-dev/fireball/internal/handlers"
	"github.com/CatSprite-dev/fireball/internal/service"
)

func main() {
	cfg := config.NewConfig()
	apiClient := api.NewClient(cfg.BaseURL)
	calculator := service.NewCalculator(apiClient)
	authHandler := handlers.NewPortfolioHandler(calculator)
	chartHandler := handlers.NewChartHandler(calculator)

	rateLimiter := handlers.NewRateLimiter(200)
	limitedPortfolioHandler := rateLimiter.Middleware(authHandler.HandlerPortfolio)
	limitedChartHandler := rateLimiter.Middleware(chartHandler.HandlerChart)

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("frontend/dist"))

	mux.HandleFunc("POST /portfolio", limitedPortfolioHandler)
	mux.HandleFunc("POST /chart", limitedChartHandler)

	mux.Handle("/", fileServer)

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
