package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CatSprite-dev/fireball/internal/api"
	"github.com/CatSprite-dev/fireball/internal/config"
	"github.com/CatSprite-dev/fireball/internal/demo"
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

	candleRepository := storage.NewCandleRepository(pool)
	operationsRepository := storage.NewOperationsRepository(pool)

	sessionManager, err := storage.NewSessionManager(store, cfg.GetSecret(), cfg.RedisTTL)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	cacheManager := storage.NewCacheManager(store, cfg.PortfolioCacheTTL, cfg.ChartDataCacheTTL)

	apiClient := api.NewClient(cfg.BaseURL)
	calculator := service.NewCalculator(apiClient, candleRepository, operationsRepository)

	portfolioService := service.NewPortfolioService(calculator, cacheManager)

	loginRateLimiter := handlers.NewRateLimiter(10)
	authRateLimiter := handlers.NewRateLimiter(200)

	loginHandler := handlers.NewLoginHandler(sessionManager, cacheManager, apiClient, cfg.IsProduction)
	portfolioHandler := handlers.NewPortfolioHandler(portfolioService)
	chartHandler := handlers.NewChartHandler(portfolioService)

	mux := http.NewServeMux()

	if cfg.DemoToken != "" {
		demoClient := demo.NewDemoClient(apiClient, cfg.DemoToken)
		demoCalculator := service.NewCalculator(demoClient, candleRepository, nil)
		demoCacheManager := storage.NewCacheManager(store, 24*time.Hour, 24*time.Hour)
		demoService := service.NewPortfolioService(demoCalculator, demoCacheManager)
		demoHandler := handlers.NewDemoHandler(demoService)

		mux.HandleFunc("GET /api/demo/portfolio", authRateLimiter.LimiterMiddleware(demoHandler.HandlerDemoPortfolio))
		mux.HandleFunc("GET /api/demo/chart", authRateLimiter.LimiterMiddleware(demoHandler.HandlerDemoChart))
	}

	mux.HandleFunc("GET /api/ping", authRateLimiter.LimiterMiddleware(handlers.AuthMiddleware(sessionManager, portfolioHandler.HandlerPing)))
	mux.HandleFunc("POST /api/login", loginRateLimiter.LimiterMiddleware(loginHandler.HandlerLogin))
	mux.HandleFunc("POST /api/logout", loginRateLimiter.LimiterMiddleware(loginHandler.HandlerLogout))
	mux.HandleFunc("GET /api/portfolio", authRateLimiter.LimiterMiddleware(handlers.AuthMiddleware(sessionManager, portfolioHandler.HandlerPortfolio)))
	mux.HandleFunc("GET /api/chart", authRateLimiter.LimiterMiddleware(handlers.AuthMiddleware(sessionManager, chartHandler.HandlerChart)))

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
