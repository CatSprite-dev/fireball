package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/config"
	"github.com/CatSprite-dev/fireball/internal/handlers"
	"github.com/CatSprite-dev/fireball/internal/service"
	"github.com/CatSprite-dev/fireball/internal/storage"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewConfig()
	if err != nil {
		os.Exit(1)
	}

	store, err := storage.NewRedisStore(cfg.RedisURL)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	pool, err := storage.NewPostgresPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v\n", err)
	}

	sessionManager, err := storage.NewSessionManager(store, cfg.GetSecret(), cfg.RedisTTL)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	cacheManager := storage.NewCacheManager(store, cfg.PortfolioCacheTTL, cfg.ChartDataCacheTTL)

	apiClient := api.NewClient(cfg.BaseURL)
	calculator := service.NewCalculator(apiClient)

	portfolioService := service.NewPortfolioService(calculator, cacheManager)

	loginHandler := handlers.NewLoginHandler(sessionManager, cacheManager, apiClient)

	loginRateLimiter := handlers.NewRateLimiter(10)
	authRateLimiter := handlers.NewRateLimiter(200)
	portfolioHandler := handlers.NewPortfolioHandler(sessionManager, portfolioService)
	chartHandler := handlers.NewChartHandler(sessionManager, portfolioService)

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("frontend/dist"))

	mux.HandleFunc("GET /api/ping", authRateLimiter.Middleware(portfolioHandler.HandlerPing))
	mux.HandleFunc("POST /api/login", loginRateLimiter.Middleware(loginHandler.HandlerLogin))
	mux.HandleFunc("POST /api/logout", loginRateLimiter.Middleware(loginHandler.HandlerLogout))
	mux.HandleFunc("POST /api/portfolio", authRateLimiter.Middleware(portfolioHandler.HandlerPortfolio))
	mux.HandleFunc("GET /api/chart", authRateLimiter.Middleware(chartHandler.HandlerChart))

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

	go func() {
		log.Printf("Serving on: http://localhost:%s/\n", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	pool.Close()

	if err := store.Close(); err != nil {
		log.Printf("Redis close error: %v", err)
	}

	os.Exit(0)
}
