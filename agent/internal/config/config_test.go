package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("CONTROL_PLANE_URL", "")
	t.Setenv("NODE_NAME", "")
	t.Setenv("NODE_PUBLIC_HOST", "")
	t.Setenv("HEARTBEAT_INTERVAL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ControlPlaneURL != "http://localhost:8000" {
		t.Fatalf("ControlPlaneURL = %q", cfg.ControlPlaneURL)
	}
	if cfg.NodeName != "worker-1" {
		t.Fatalf("NodeName = %q", cfg.NodeName)
	}
	if cfg.NodePublicHost != "127.0.0.1" {
		t.Fatalf("NodePublicHost = %q", cfg.NodePublicHost)
	}
	if cfg.HeartbeatInterval != 10*time.Second {
		t.Fatalf("HeartbeatInterval = %s", cfg.HeartbeatInterval)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("CONTROL_PLANE_URL", "http://control-plane:8000")
	t.Setenv("NODE_NAME", "worker-9")
	t.Setenv("NODE_PUBLIC_HOST", "10.0.0.9")
	t.Setenv("HEARTBEAT_INTERVAL", "3s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ControlPlaneURL != "http://control-plane:8000" {
		t.Fatalf("ControlPlaneURL = %q", cfg.ControlPlaneURL)
	}
	if cfg.NodeName != "worker-9" {
		t.Fatalf("NodeName = %q", cfg.NodeName)
	}
	if cfg.NodePublicHost != "10.0.0.9" {
		t.Fatalf("NodePublicHost = %q", cfg.NodePublicHost)
	}
	if cfg.HeartbeatInterval != 3*time.Second {
		t.Fatalf("HeartbeatInterval = %s", cfg.HeartbeatInterval)
	}
}

func TestLoadRejectsInvalidHeartbeatInterval(t *testing.T) {
	t.Setenv("HEARTBEAT_INTERVAL", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}
