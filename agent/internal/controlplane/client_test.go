package controlplane

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegisterSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/nodes/register" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var request struct {
			Name       string `json:"name"`
			PublicHost string `json:"public_host"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.Name != "worker-1" || request.PublicHost != "127.0.0.1" {
			t.Fatalf("unexpected registration body: %+v", request)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"node-1","name":"worker-1","public_host":"127.0.0.1","status":"ONLINE"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, time.Second)
	node, err := client.Register(context.Background(), "worker-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if node.ID != "node-1" {
		t.Fatalf("node ID = %q", node.ID)
	}
}

func TestHeartbeatSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/internal/nodes/node-1/heartbeat" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, time.Second)
	if err := client.Heartbeat(context.Background(), "node-1"); err != nil {
		t.Fatalf("Heartbeat() error = %v", err)
	}
}

func TestHTTP500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "temporary failure", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, time.Second)
	if _, err := client.Register(context.Background(), "worker-1", "127.0.0.1"); err == nil {
		t.Fatal("Register() error = nil, want HTTP 500 error")
	}
}

func TestRegisterMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":`))
	}))
	defer server.Close()

	client := NewClient(server.URL, time.Second)
	if _, err := client.Register(context.Background(), "worker-1", "127.0.0.1"); err == nil {
		t.Fatal("Register() error = nil, want JSON decode error")
	}
}

func TestRegisterContextCancellation(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := client.Register(ctx, "worker-1", "127.0.0.1")
		done <- err
	}()

	<-started
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Register() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		close(release)
		t.Fatal("Register() did not return after context cancellation")
	}
	close(release)
}
