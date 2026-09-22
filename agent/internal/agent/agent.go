package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kirbysss369/mini-heroku/agent/internal/config"
	"github.com/kirbysss369/mini-heroku/agent/internal/controlplane"
)

const (
	defaultRegistrationAttempts = 5
	defaultRegistrationBackoff  = 250 * time.Millisecond
	maxRegistrationBackoff      = 2 * time.Second
)

type Runner struct {
	config               config.Config
	client               *controlplane.Client
	logger               *slog.Logger
	nodeID               string
	registrationAttempts int
	registrationBackoff  time.Duration
	maxBackoff           time.Duration
}

func NewRunner(cfg config.Config, client *controlplane.Client, logger *slog.Logger) *Runner {
	return &Runner{
		config:               cfg,
		client:               client,
		logger:               logger,
		registrationAttempts: defaultRegistrationAttempts,
		registrationBackoff:  defaultRegistrationBackoff,
		maxBackoff:           maxRegistrationBackoff,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	node, err := r.register(ctx)
	if err != nil {
		return fmt.Errorf("register node: %w", err)
	}

	r.nodeID = node.ID
	r.logger.Info("node registered", "node_id", r.nodeID, "name", r.config.NodeName)

	ticker := time.NewTicker(r.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := r.client.Heartbeat(ctx, r.nodeID); err != nil {
				if ctx.Err() != nil {
					return nil
				}
				r.logger.Warn("heartbeat failed", "error", err)
			}
		}
	}
}

func (r *Runner) register(ctx context.Context) (controlplane.Node, error) {
	backoff := r.registrationBackoff
	var lastErr error

	for attempt := 1; attempt <= r.registrationAttempts; attempt++ {
		node, err := r.client.Register(ctx, r.config.NodeName, r.config.NodePublicHost)
		if err == nil {
			return node, nil
		}
		lastErr = err

		if attempt == r.registrationAttempts {
			break
		}

		r.logger.Warn("registration failed; retrying", "attempt", attempt, "error", err)
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return controlplane.Node{}, ctx.Err()
		case <-timer.C:
		}

		backoff *= 2
		if backoff > r.maxBackoff {
			backoff = r.maxBackoff
		}
	}

	return controlplane.Node{}, lastErr
}
