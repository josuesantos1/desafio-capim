package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/josuesantos1/desafio/docs"
	"github.com/josuesantos1/desafio/internal/clinic"
	"github.com/josuesantos1/desafio/internal/config"
	"github.com/josuesantos1/desafio/internal/dentist"
	"github.com/josuesantos1/desafio/internal/payment"
	"github.com/josuesantos1/desafio/internal/server"
	"github.com/josuesantos1/desafio/pkg/pix"
)

const shutdownTimeout = 10 * time.Second

// @title Clinic API
// @version 1.0
// @description API de gestão de clínica odontológica — clínicas e dentistas.
// @BasePath /api
func main() {
	cfg := config.Load()
	setupLogger(cfg.LogLevel)

	router := server.NewRouter()

	clinicRepo := clinic.NewMemoryRepository()
	dentistRepo := dentist.NewMemoryRepository(clinicRepo)
	paymentRepo := payment.NewMemoryRepository(clinicRepo, dentistRepo)
	pixClient := pix.NewClient(pix.Config{APIKey: "simulated"})

	if cfg.SeedData {
		seedData(context.Background(), clinicRepo, dentistRepo, paymentRepo)
	}

	router.Route("/api", func(api chi.Router) {
		clinic.RegisterRoutes(api, clinic.NewService(clinicRepo))
		dentist.RegisterRoutes(api, dentist.NewService(dentistRepo, clinicRepo))
		payment.RegisterRoutes(api, payment.NewService(paymentRepo, clinicRepo, dentistRepo, pixClient))
	})

	router.Get("/swagger/*", httpSwagger.WrapHandler)

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
