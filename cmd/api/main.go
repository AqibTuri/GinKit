// @title           Gin API Template
// @version         1.0
// @description     Modular Gin API: JWT auth, SQL migrations, per-module DTOs and presenters, unified JSON responses.
// @BasePath        /
// @schemes         http https
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Type "Bearer" followed by a space and your JWT (e.g. Bearer eyJhbGciOi...)
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "gin-api/docs"
	"gin-api/internal/app"
	"gin-api/internal/config"
)

// Main entry flow (read top to bottom):
//  1. Structured JSON logs to stdout.
//  2. Load env → Config (see internal/config).
//  3. app.New: open DB, wire repos/services/handlers, build Gin engine (see internal/app + router).
//  4. Blank import gin-api/docs runs docs.init() so /swagger can serve OpenAPI.
//  5. http.Server runs Gin; signal handler shuts down gracefully (in-flight requests get a deadline).
func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}

	a, err := app.New(cfg)
	if err != nil {
		slog.Error("app init failed", "err", err)
		os.Exit(1)
	}

	addr := ":" + cfg.HTTPPort
	srv := &http.Server{
		Addr:              addr,
		Handler:           a.Engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "err", err)
	}
	slog.Info("bye")
}
