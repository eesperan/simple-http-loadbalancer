package balancer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"loadbalancer/internal/metrics"
)

func TestRollout(t *testing.T) {
	// Create test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server1"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server2"))
	}))
	defer server2.Close()

	server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server3"))
	}))
	defer server3.Close()

	// Initialize load balancer with initial backends
	lb := &LoadBalancer{
		metrics: metrics.New(),
	}
	initialBackends := []string{server1.URL}
	err := lb.updateBackends(initialBackends)
	if err != nil {
		t.Fatalf("Failed to initialize backends: %v", err)
	}

	// Test rollout configuration
	config := RolloutConfig{
		NewBackends: []string{server2.URL, server3.URL},
		BatchSize:   1,
		Interval:    100 * time.Millisecond,
	}

	// Perform rollout
	ctx := context.Background()
	err = lb.Rollout(ctx, config)
	if err != nil {
		t.Fatalf("Rollout failed: %v", err)
	}

	// Verify final backend count
	if len(lb.backends) != 2 {
		t.Errorf("Expected 2 backends after rollout, got %d", len(lb.backends))
	}

	// Test rollout with invalid backend
	invalidConfig := RolloutConfig{
		NewBackends: []string{"invalid-url"},
		BatchSize:   1,
		Interval:    100 * time.Millisecond,
	}

	err = lb.Rollout(ctx, invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid backend URL")
	}
}

func TestRollback(t *testing.T) {
	// Create test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server1"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server2"))
	}))
	defer server2.Close()

	// Initialize load balancer with initial backends
	lb := &LoadBalancer{
		metrics: metrics.New(),
	}
	initialBackends := []string{server2.URL}
	err := lb.updateBackends(initialBackends)
	if err != nil {
		t.Fatalf("Failed to initialize backends: %v", err)
	}

	// Test rollback configuration
	config := RollbackConfig{
		PreviousBackends: []string{server1.URL},
		BatchSize:        1,
		Interval:         100 * time.Millisecond,
	}

	// Perform rollback
	ctx := context.Background()
	err = lb.Rollback(ctx, config)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify final backend count
	if len(lb.backends) != 1 {
		t.Errorf("Expected 1 backend after rollback, got %d", len(lb.backends))
	}

	// Test rollback with invalid backend
	invalidConfig := RollbackConfig{
		PreviousBackends: []string{"invalid-url"},
		BatchSize:        1,
		Interval:         100 * time.Millisecond,
	}

	err = lb.Rollback(ctx, invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid backend URL")
	}
}

func TestRolloutState(t *testing.T) {
	state := &RolloutState{}

	// Test initial state
	phase, progress, err := state.getStatus()
	if phase != "" || progress != 0 || err != nil {
		t.Error("Unexpected initial state")
	}

	// Test state update
	testPhase := "testing"
	testProgress := 50.0
	testError := fmt.Errorf("test error")
	
	state.update(testPhase, testProgress, testError)
	
	phase, progress, err = state.getStatus()
	if phase != testPhase {
		t.Errorf("Expected phase %s, got %s", testPhase, phase)
	}
	if progress != testProgress {
		t.Errorf("Expected progress %f, got %f", testProgress, progress)
	}
	if err != testError {
		t.Error("Expected test error")
	}
}

func TestRolloutWithContext(t *testing.T) {
	lb := &LoadBalancer{
		metrics: metrics.New(),
	}

	// Create context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	}))
	defer server.Close()

	// Configure rollout
	config := RolloutConfig{
		NewBackends: []string{server.URL},
		BatchSize:   1,
		Interval:    1 * time.Second, // Long enough to cancel
	}

	// Start rollout in goroutine
	errChan := make(chan error)
	go func() {
		errChan <- lb.Rollout(ctx, config)
	}()

	// Cancel context
	cancel()

	// Check if rollout was cancelled
	select {
	case err := <-errChan:
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Rollout did not cancel in time")
	}
}
