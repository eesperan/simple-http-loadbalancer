package integration

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"loadbalancer/internal/balancer"
	"loadbalancer/internal/config"
	"loadbalancer/internal/metrics"
)

func setupTestBackend(t *testing.T, port int, id string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Response from backend %s", id)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "healthy")
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Slow response from backend %s", id)
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error from backend %s", id)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("Test backend %s failed: %v", id, err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	return server
}

func TestLoadBalancerIntegration(t *testing.T) {
	// Setup test backends
	backend1 := setupTestBackend(t, 9001, "1")
	defer backend1.Shutdown(context.Background())
	backend2 := setupTestBackend(t, 9002, "2")
	defer backend2.Shutdown(context.Background())

	// Create temporary config file
	configContent := `
frontends:
- port: 8080
backends:
- "http://localhost:9001"
- "http://localhost:9002"
healthcheck:
  interval: "1s"
  timeout: "500ms"
  path: "/health"
metrics:
  enabled: true
  port: 9090
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpfile.Close()

	// Load config and start load balancer
	cfg, err := config.Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	lb, err := balancer.New(cfg, metrics.New())
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := lb.Start(ctx); err != nil {
			t.Errorf("Load balancer failed: %v", err)
		}
	}()

	// Wait for load balancer to start
	time.Sleep(1 * time.Second)

	// Test round-robin distribution
	client := &http.Client{Timeout: 5 * time.Second}
	responses := make(map[string]int)
	var mu sync.Mutex

	for i := 0; i < 4; i++ {
		resp, err := client.Get("http://localhost:8080")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		mu.Lock()
		responses[string(body)]++
		mu.Unlock()
	}

	// Test circuit breaker
	for i := 0; i < 10; i++ {
		_, err := client.Get("http://localhost:8080/error")
		if err != nil {
			continue
		}
	}

	// Test rate limiting
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = client.Get("http://localhost:8080")
		}()
	}
	wg.Wait()

	// Test timeouts
	resp, err := client.Get("http://localhost:8080/slow")
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode != http.StatusGatewayTimeout {
			t.Errorf("Expected timeout status, got %d", resp.StatusCode)
		}
	}

	// Test metrics endpoint
	resp, err = client.Get("http://localhost:9090/metrics")
	if err != nil {
		t.Fatalf("Metrics endpoint failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected metrics status 200, got %d", resp.StatusCode)
	}

	// Verify distribution
	if len(responses) != 2 {
		t.Errorf("Expected responses from 2 backends, got %d", len(responses))
	}

	for backend, count := range responses {
		if count != 2 {
			t.Errorf("Expected 2 requests to backend, got %d for %s", count, backend)
		}
	}
}

func TestLoadBalancerSSL(t *testing.T) {
	// Skip if SSL certificates are not available
	if _, err := os.Stat("test-cert.pem"); os.IsNotExist(err) {
		t.Skip("SSL certificates not available")
	}

	configContent := `
frontends:
- port: 8443
backends:
- "http://localhost:9001"
healthcheck:
  interval: "1s"
  timeout: "500ms"
  path: "/health"
ssl:
  certFile: "test-cert.pem"
  keyFile: "test-key.pem"
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpfile.Close()

	backend := setupTestBackend(t, 9001, "1")
	defer backend.Shutdown(context.Background())

	cfg, err := config.Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	lb, err := balancer.New(cfg, metrics.New())
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := lb.Start(ctx); err != nil {
			t.Errorf("Load balancer failed: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Get("https://localhost:8443")
	if err != nil {
		t.Fatalf("HTTPS request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
