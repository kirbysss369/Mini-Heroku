package config

import (
	"fmt"
	"os"
	"time"
)

const defaultShutdownTimeout = 10 * time.Second

type Config struct {
	AgentID         string
	ControlPlaneURL string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AgentID:         getenv("AGENT_ID", "agent-local"),
		ControlPlaneURL: getenv("CONTROL_PLANE_URL", "http://localhost:8000"),
		ShutdownTimeout: defaultShutdownTimeout,
	}

	if raw := os.Getenv("SHUTDOWN_TIMEOUT"); raw != "" {
		timeout, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("parse SHUTDOWN_TIMEOUT: %w", err)
		}
		if timeout <= 0 {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT must be positive")
		}
		cfg.ShutdownTimeout = timeout
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
