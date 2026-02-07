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

	apiClient := api.NewClient(cfg.BaseURL, cfg.Timeout)
	calculator := service.NewCalculator(apiClient)
	authHandler := handlers.NewAuthHandler(calculator)

	mux := http.NewServeMux()

	mux.HandleFunc("/auth", authHandler.HandlerAuth)

	mux.Handle("/", http.FileServer(http.Dir("web")))

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	log.Printf("Serving on: http://localhost:%s/\n", cfg.ServerPort)
	log.Fatal(srv.ListenAndServe())
}
