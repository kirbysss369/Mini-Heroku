package config

import (
	"fmt"
	"os"
	"time"
)

const defaultHeartbeatInterval = 10 * time.Second

type Config struct {
	ControlPlaneURL   string
	NodeName          string
	NodePublicHost    string
	HeartbeatInterval time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		ControlPlaneURL:   getenv("CONTROL_PLANE_URL", "http://localhost:8000"),
		NodeName:          getenv("NODE_NAME", "worker-1"),
		NodePublicHost:    getenv("NODE_PUBLIC_HOST", "127.0.0.1"),
		HeartbeatInterval: defaultHeartbeatInterval,
	}

	if raw := os.Getenv("HEARTBEAT_INTERVAL"); raw != "" {
		interval, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("parse HEARTBEAT_INTERVAL: %w", err)
		}
		if interval <= 0 {
			return Config{}, fmt.Errorf("HEARTBEAT_INTERVAL must be positive")
		}
		cfg.HeartbeatInterval = interval
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
