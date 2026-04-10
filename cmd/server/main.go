package main

import (
	"log"
	"net/http"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/config"
	"github.com/CatSprite-dev/fireball/internal/handlers"
	"github.com/CatSprite-dev/fireball/internal/service"
	"github.com/CatSprite-dev/fireball/internal/session"
)

func main() {
	cfg := config.NewConfig()

	store, err := session.NewRedisStore(cfg.RedisURL, cfg.RedisTTL)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	sessionManager, err := session.NewManager(store, cfg.GetSecret())
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	apiClient := api.NewClient(cfg.BaseURL)
	calculator := service.NewCalculator(apiClient)

	loginHandler := handlers.NewLoginHandler(sessionManager, apiClient)
	authHandler := handlers.NewAuthHandler(sessionManager, calculator)

	loginRateLimiter := handlers.NewRateLimiter(10)
	authRateLimiter := handlers.NewRateLimiter(200)

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("frontend/dist"))

	mux.HandleFunc("GET /ping", authRateLimiter.Middleware(authHandler.HandlerPing))
	mux.HandleFunc("POST /login", loginRateLimiter.Middleware(loginHandler.HandlerLogin))
	mux.HandleFunc("POST /logout", loginRateLimiter.Middleware(loginHandler.HandlerLogout))
	mux.HandleFunc("POST /auth", authRateLimiter.Middleware(authHandler.HandlerAuth))

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
