package balancer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"loadbalancer/internal/config"
	"loadbalancer/internal/metrics"
)

func setupTestBackends(t *testing.T, count int) ([]*httptest.Server, []string) {
	var servers []*httptest.Server
	var urls []string

	for i := 0; i < count; i++ {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		servers = append(servers, server)
		urls = append(urls, server.URL)
	}

	return servers, urls
}

func TestRollout(t *testing.T) {
	metrics.Reset() // Reset metrics before test

	// Setup initial backends
	servers, urls := setupTestBackends(t, 3)
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	// Create load balancer with initial backends
	lb, err := New(&config.Config{
		Backends: urls[:2], // Start with 2 backends
	}, metrics.New())
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Create new backend configuration
	newBackends := []string{urls[1], urls[2]} // Roll out to backend 2 and 3

	// Test rollout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = lb.Rollout(ctx, RolloutConfig{
		NewBackends: newBackends,
		BatchSize:   1,
		Interval:    100 * time.Millisecond,
	})

	if err != nil {
		t.Errorf("Rollout failed: %v", err)
	}

	// Verify new backend configuration
	if len(lb.backends) != len(newBackends) {
		t.Errorf("Expected %d backends after rollout, got %d", len(newBackends), len(lb.backends))
	}

	// Test rollback
	err = lb.Rollback(ctx, RollbackConfig{
		PreviousBackends: urls[:2],
		BatchSize:        1,
		Interval:         100 * time.Millisecond,
	})

	if err != nil {
		t.Errorf("Rollback failed: %v", err)
	}

	// Verify rolled back configuration
	if len(lb.backends) != 2 {
		t.Errorf("Expected 2 backends after rollback, got %d", len(lb.backends))
	}
}

func TestRolloutErrors(t *testing.T) {
	metrics.Reset() // Reset metrics before test

	// Setup test backend
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create load balancer
	lb, err := New(&config.Config{
		Backends: []string{server.URL},
	}, metrics.New())
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Test empty backends
	err = lb.Rollout(context.Background(), RolloutConfig{
		NewBackends: []string{},
		BatchSize:   1,
		Interval:    100 * time.Millisecond,
	})
	if err == nil {
		t.Error("Expected error for empty backends")
	}

	// Test invalid backend URL
	err = lb.Rollout(context.Background(), RolloutConfig{
		NewBackends: []string{"invalid-url"},
		BatchSize:   1,
		Interval:    100 * time.Millisecond,
	})
	if err == nil {
		t.Error("Expected error for invalid backend URL")
	}

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = lb.Rollout(ctx, RolloutConfig{
		NewBackends: []string{server.URL},
		BatchSize:   1,
		Interval:    100 * time.Millisecond,
	})
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

func TestRolloutConcurrency(t *testing.T) {
	metrics.Reset() // Reset metrics before test

	// Setup test backends
	servers, urls := setupTestBackends(t, 4)
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	// Create load balancer
	lb, err := New(&config.Config{
		Backends: urls[:2],
	}, metrics.New())
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Start serving requests
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				req := httptest.NewRequest("GET", "/", nil)
				w := httptest.NewRecorder()
				lb.ServeHTTP(w, req)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// Perform rollout while serving requests
	ctx := context.Background()
	err = lb.Rollout(ctx, RolloutConfig{
		NewBackends: urls[2:],
		BatchSize:   1,
		Interval:    100 * time.Millisecond,
	})

	close(done)

	if err != nil {
		t.Errorf("Rollout failed: %v", err)
	}

	// Verify final configuration
	if len(lb.backends) != 2 {
		t.Errorf("Expected 2 backends after rollout, got %d", len(lb.backends))
	}
}
