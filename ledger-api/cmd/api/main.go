package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/suphakin-th/ledger-service/ledger-api/internal/adapters/postgres"
	redisbus "github.com/suphakin-th/ledger-service/ledger-api/internal/adapters/redis"
	httpadapter "github.com/suphakin-th/ledger-service/ledger-api/internal/adapters/http"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/usecases"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	dbURL := envOrDefault("DATABASE_URL", "postgres://ledger:ledger@localhost:5432/ledger?sslmode=disable")
	redisAddr := envOrDefault("REDIS_ADDR", "localhost:6379")
	httpPort := envOrDefault("HTTP_PORT", "8080")

	pool, err := postgres.Connect(ctx, dbURL)
	if err != nil {
		slog.Error("db connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := postgres.Migrate(ctx, pool); err != nil {
		slog.Error("db migrate", "error", err)
		os.Exit(1)
	}

	accountRepo := postgres.NewAccountRepo(pool)
	txRepo := postgres.NewTransactionRepo(pool)
	bus := redisbus.NewEventBus(redisAddr)
	defer bus.Close()

	handler := httpadapter.NewHandler(
		usecases.NewCreateAccount(accountRepo),
		usecases.NewCreateTransaction(accountRepo, txRepo, bus),
		usecases.NewGetSummary(accountRepo, txRepo),
	)

	router := httpadapter.NewRouter(handler)

	slog.Info("ledger-api starting", "port", httpPort)
	if err := router.Run(":" + httpPort); err != nil {
		slog.Error("server error", "error", err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
