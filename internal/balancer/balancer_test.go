package balancer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"loadbalancer/internal/balancer/algorithm"
	"loadbalancer/internal/config"
	"loadbalancer/internal/metrics"
)

func TestNew(t *testing.T) {
	metrics.Reset() // Reset metrics before test
	cfg := &config.Config{
		Backends: []string{"http://localhost:8001", "http://localhost:8002"},
	}
	m := metrics.New()
	lb, err := New(cfg, m)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	if lb == nil {
		t.Fatal("Expected non-nil LoadBalancer")
	}

	if len(lb.backends) != 2 {
		t.Errorf("Expected 2 backends, got %d", len(lb.backends))
	}
}

func TestUpdateBackends(t *testing.T) {
	metrics.Reset() // Reset metrics before test
	lb := &LoadBalancer{
		metrics: metrics.New(),
		wrr:     algorithm.NewWeightedRoundRobin(),
	}

	backends := []string{"http://localhost:8001", "http://localhost:8002"}
	err := lb.updateBackends(backends)
	if err != nil {
		t.Fatalf("Failed to update backends: %v", err)
	}

	if len(lb.backends) != len(backends) {
		t.Errorf("Expected %d backends, got %d", len(backends), len(lb.backends))
	}

	// Test invalid backend URL
	err = lb.updateBackends([]string{"not-a-valid-url"})
	if err == nil {
		t.Error("Expected error for invalid backend URL")
	}
}

func TestServeHTTP(t *testing.T) {
	metrics.Reset() // Reset metrics before test
	// Create test backend servers
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	// Create load balancer
	cfg := &config.Config{
		Backends: []string{backend1.URL, backend2.URL},
	}
	lb, err := New(cfg, metrics.New())
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Create test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Test first request
	lb.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Test second request (should go to different backend)
	w2 := httptest.NewRecorder()
	lb.ServeHTTP(w2, req)
	if w2.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w2.Code)
	}

	if w.Body.String() == w2.Body.String() {
		t.Error("Expected different responses from different backends")
	}
}

func TestGracefulShutdown(t *testing.T) {
	metrics.Reset() // Reset metrics before test
	cfg := &config.Config{
		Frontends: []config.Frontend{{Port: 18080}}, // Use high port number to avoid conflicts
	}
	lb, err := New(cfg, metrics.New())
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- lb.Start(ctx)
	}()

	// Wait a bit then cancel
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Expected no error on shutdown, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for graceful shutdown")
	}
}
