package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kirbysss369/mini-heroku/agent/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load agent config", "error", err)
		os.Exit(1)
	}

	slog.Info(
		"agent starting",
		"agent_id", cfg.AgentID,
		"control_plane_url", cfg.ControlPlaneURL,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("agent stopped")
}

func shutdown(_ context.Context) error {
	// Future milestones will stop pollers and in-flight work here.
	return nil
}
