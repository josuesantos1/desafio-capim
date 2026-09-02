package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/internal/config"
	"github.com/josuesantos1/desafio/internal/server"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg := config.Load()
	setupLogger(cfg.LogLevel)

	db := connectDB(cfg.DB)
	defer db.Close()

	router := server.NewRouter()
	clinic.RegisterRoutes(router, clinic.NewService(clinic.NewPostgresRepository(db)))

	srv := server.New(cfg, router)

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	case sig := <-sigCh:
		slog.Info("shutdown signal received", "signal", sig.String())
		shutdown(srv)
	}
}

func connectDB(cfg config.DBConfig) *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	return db
}

func shutdown(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("shutdown complete")
}

func setupLogger(resolvedLevel string) {
	level := parseLevel(resolvedLevel)
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))

	if raw, ok := os.LookupEnv("LOG_LEVEL"); ok && !config.IsValidLogLevel(raw) {
		slog.Warn("invalid LOG_LEVEL, falling back to info", "value", raw)
	}
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
