package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	agentapp "github.com/kirbysss369/mini-heroku/agent/internal/agent"
	"github.com/kirbysss369/mini-heroku/agent/internal/config"
	"github.com/kirbysss369/mini-heroku/agent/internal/controlplane"
)

const requestTimeout = 5 * time.Second

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load agent config", "error", err)
		os.Exit(1)
	}

	client := controlplane.NewClient(cfg.ControlPlaneURL, requestTimeout)
	runner := agentapp.NewRunner(cfg, client, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("agent starting", "node", cfg.NodeName, "control_plane", cfg.ControlPlaneURL)
	if err := runner.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("agent stopped with error", "error", err)
		os.Exit(1)
	}
	logger.Info("agent stopped")
}
