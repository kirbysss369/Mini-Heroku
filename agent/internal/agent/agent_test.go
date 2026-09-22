package agent

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kirbysss369/mini-heroku/agent/internal/config"
	"github.com/kirbysss369/mini-heroku/agent/internal/controlplane"
)

func TestRegistrationRetry(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/nodes/register" {
			http.NotFound(w, r)
			return
		}

		attempt := attempts.Add(1)
		if attempt < 3 {
			http.Error(w, "try again", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"node-1","name":"worker-1","public_host":"127.0.0.1","status":"ONLINE"}`))
	}))
	defer server.Close()

	runner := newTestRunner(server.URL)
	runner.registrationBackoff = time.Millisecond
	runner.maxBackoff = 2 * time.Millisecond

	node, err := runner.register(context.Background())
	if err != nil {
		t.Fatalf("register() error = %v", err)
	}
	if node.ID != "node-1" {
		t.Fatalf("node ID = %q", node.ID)
	}
	if attempts.Load() != 3 {
		t.Fatalf("registration attempts = %d, want 3", attempts.Load())
	}
}

func TestHeartbeatFailureDoesNotStopAgent(t *testing.T) {
	var heartbeatAttempts atomic.Int32
	secondHeartbeat := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/nodes/register":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"node-1","name":"worker-1","public_host":"127.0.0.1","status":"ONLINE"}`))
		case "/internal/nodes/node-1/heartbeat":
			attempt := heartbeatAttempts.Add(1)
			if attempt == 1 {
				http.Error(w, "temporary failure", http.StatusInternalServerError)
				return
			}
			if attempt == 2 {
				close(secondHeartbeat)
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	runner := newTestRunner(server.URL)
	runner.config.HeartbeatInterval = 5 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- runner.Run(ctx) }()

	select {
	case <-secondHeartbeat:
		cancel()
	case <-time.After(time.Second):
		cancel()
		t.Fatal("agent stopped heartbeating after the first failure")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not stop after cancellation")
	}
}

func newTestRunner(baseURL string) *Runner {
	cfg := config.Config{
		ControlPlaneURL:   baseURL,
		NodeName:          "worker-1",
		NodePublicHost:    "127.0.0.1",
		HeartbeatInterval: time.Hour,
	}
	client := controlplane.NewClient(baseURL, time.Second)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRunner(cfg, client, logger)
}
