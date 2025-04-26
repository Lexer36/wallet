package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := godotenv.Load("config.env")
	if err != nil {
		panic("failed to load env")
	}
	logger := setupLogger(os.Getenv("ENV"))

	db, err := postgres.Init(ctx, os.Getenv("DB_DSN"))
	if err != nil {
		logger.Error("Failed to connect to database", err)
	}
	defer db.Close()

	repo := postgres.New(db)
	cache := cache.New()

	walletService := services.NewWalletService(repo, cache, logger)

	walletHandler := rest.NewWalletHandler(walletService)

	mux := http.NewServeMux()

	initMetrics(mux)

	mux.HandleFunc("/api/v1/wallet", walletHandler.WalletOperation)
	mux.HandleFunc("/api/v1/wallets/", walletHandler.GetBalance)

	server := &http.Server{
		Addr:    os.Getenv("SERVER_ADDRESS"),
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("Starting server")
		if err := server.ListenAndServe(); err != nil {
			logger.Error("failed to start server", err)
		}
	}()

	<-stop

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed: %v", err)
	}

	fmt.Println("Server exited properly")
}

func initMetrics(mux *http.ServeMux) {
	metrics.Register()

	mux.Handle("/metrics", metrics.Handler())
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
