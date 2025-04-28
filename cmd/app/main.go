package main

import (
	"context"
	"fmt"
	lru2 "github.com/hashicorp/golang-lru"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wallet/internal/metrics"
	"wallet/internal/repository/cache"
	"wallet/internal/repository/postgres"
	"wallet/internal/rest"
	"wallet/internal/services"

	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	err := godotenv.Load("config.env")
	if err != nil {
		panic("failed to load env")
	}
	logger := setupLogger(os.Getenv("ENV"))

	pool, err := postgres.Init(os.Getenv("DB_DSN"))
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	initMigrations(pool)
	repo := postgres.New(pool)

	lru, err := lru2.New(1000) // can be replaced by other cache implementations
	if err != nil {
		panic(err)
	}
	cache := cache.New(lru)

	walletService := services.NewWalletService(repo, cache, logger)
	walletHandler := rest.NewWalletHandler(walletService)

	mux := http.NewServeMux()

	initMetrics(mux)

	mux.Handle("/api/v1/wallet", metrics.MetricsMiddleware(http.HandlerFunc(walletHandler.WalletOperation), "WalletOperation"))
	mux.Handle("/api/v1/wallets/", metrics.MetricsMiddleware(http.HandlerFunc(walletHandler.GetBalance), "GetBalance"))

	server := &http.Server{
		Addr:    os.Getenv("SERVER_ADDRESS"),
		Handler: mux,
	}

	// listen to OS signals and gracefully shutdown HTTP server
	stopped := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("HTTP Server Shutdown Error", "error", err)
		}
		close(stopped)
	}()

	logger.Info("Starting server")
	if err := server.ListenAndServe(); err != nil {
		logger.Error("failed to start server", "error", err)
	}

	<-stopped

	fmt.Println("Server exited properly")
}

func initMetrics(mux *http.ServeMux) {
	metrics.Register()

	mux.Handle("/metrics", metrics.Handler())
}

func initMigrations(pool *pgxpool.Pool) {
	// we can change this to local only usage
	// run migrations
	db := stdlib.OpenDBFromPool(pool)
	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}
	if err := goose.Up(db, "./migrations"); err != nil {
		panic(err)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}

	switch env {
	case envLocal:
		fallthrough
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
