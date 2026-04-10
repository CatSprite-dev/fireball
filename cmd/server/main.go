package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

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

	mux.HandleFunc("GET /api/ping", authRateLimiter.Middleware(authHandler.HandlerPing))
	mux.HandleFunc("POST /api/login", loginRateLimiter.Middleware(loginHandler.HandlerLogin))
	mux.HandleFunc("POST /api/logout", loginRateLimiter.Middleware(loginHandler.HandlerLogout))
	mux.HandleFunc("POST /api/auth", authRateLimiter.Middleware(authHandler.HandlerAuth))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Catch-all hit: %s %s", r.Method, r.URL.Path)
		path := filepath.Join("frontend/dist", r.URL.Path)
		log.Printf("Serving path: %s", path)
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			log.Printf("Not found, serving index.html")
			http.ServeFile(w, r, "frontend/dist/index.html")
			return
		}
		fileServer.ServeHTTP(w, r)
	}))

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
