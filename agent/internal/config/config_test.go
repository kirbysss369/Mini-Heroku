package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("AGENT_ID", "")
	t.Setenv("CONTROL_PLANE_URL", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AgentID != "agent-local" {
		t.Fatalf("AgentID = %q, want %q", cfg.AgentID, "agent-local")
	}
	if cfg.ControlPlaneURL != "http://localhost:8000" {
		t.Fatalf("ControlPlaneURL = %q, want %q", cfg.ControlPlaneURL, "http://localhost:8000")
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want %s", cfg.ShutdownTimeout, 10*time.Second)
	}
}

func TestLoadShutdownTimeoutOverride(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT", "3s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ShutdownTimeout != 3*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want %s", cfg.ShutdownTimeout, 3*time.Second)
	}
}

func TestLoadRejectsInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}
