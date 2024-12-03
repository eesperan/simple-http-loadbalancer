package balancer

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"loadbalancer/internal/balancer/algorithm"
	"loadbalancer/internal/circuitbreaker"
	"loadbalancer/internal/config"
	"loadbalancer/internal/errors"
	"loadbalancer/internal/metrics"
	"loadbalancer/internal/ratelimit"
)

// BenchmarkLoadBalancer runs a series of benchmarks to measure load balancer performance
// under different scenarios and configurations.
func BenchmarkLoadBalancer(b *testing.B) {
	scenarios := []struct {
		name    string
		ssl     bool
		backend func(w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "SimpleForward",
			ssl:  false,
			backend: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("OK"))
			},
		},
		{
			name: "LargeResponse",
			ssl:  false,
			backend: func(w http.ResponseWriter, r *http.Request) {
				response := make([]byte, 1024*1024) // 1MB response
				w.Write(response)
			},
		},
		{
			name: "SlowBackend",
			ssl:  false,
			backend: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.Write([]byte("OK"))
			},
		},
		{
			name: "WithSSL",
			ssl:  true,
			backend: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("OK"))
			},
		},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			// Setup test backend
			backend := httptest.NewServer(http.HandlerFunc(scenario.backend))
			defer backend.Close()

			// Configure load balancer
			cfg := &config.Config{
				Frontends: []config.Frontend{{Port: 0}}, // Use random port
				Backends:  []string{backend.URL},
			}

			if scenario.ssl {
				cfg.SSL = &config.SSL{
					CertFile: "../ssl/test-cert.pem",
					KeyFile:  "../ssl/test-key.pem",
				}
			}

			// Create load balancer
			lb, err := New(cfg, metrics.New())
			if err != nil {
				b.Fatalf("Failed to create load balancer: %v", err)
			}

			// Start load balancer
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				if err := lb.Start(ctx); err != nil && err != http.ErrServerClosed {
					b.Errorf("Load balancer error: %v", err)
				}
			}()

			// Create test client
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			// Run benchmark
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					resp, err := client.Get(fmt.Sprintf("http://localhost:%d", cfg.Frontends[0].Port))
					if err != nil {
						b.Errorf("Request failed: %v", err)
						continue
					}
					resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						b.Errorf("Expected status 200, got %d", resp.StatusCode)
					}
				}
			})
		})
	}
}

// BenchmarkWeightedRoundRobin measures the performance of the weighted round-robin algorithm
func BenchmarkWeightedRoundRobin(b *testing.B) {
	scenarios := []struct {
		name         string
		numBackends  int
		updateWeight bool
	}{
		{"Small-Static", 3, false},
		{"Medium-Static", 10, false},
		{"Large-Static", 100, false},
		{"Small-Dynamic", 3, true},
		{"Medium-Dynamic", 10, true},
		{"Large-Dynamic", 100, true},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			// Setup weighted round robin
			wrr := algorithm.NewWeightedRoundRobin()
			for i := 0; i < scenario.numBackends; i++ {
				wrr.Add(fmt.Sprintf("backend-%d", i), i+1)
			}

			if scenario.updateWeight {
				go func() {
					for i := 0; i < scenario.numBackends; i++ {
						wrr.UpdateWeight(fmt.Sprintf("backend-%d", i), i%5+1)
						time.Sleep(time.Millisecond)
					}
				}()
			}

			// Run benchmark
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					backend := wrr.Next()
					if backend == nil {
						b.Error("Failed to get backend")
					}
				}
			})
		})
	}
}

// BenchmarkCircuitBreaker measures the performance impact of the circuit breaker
func BenchmarkCircuitBreaker(b *testing.B) {
	scenarios := []struct {
		name          string
		failureRate   float64
		numGoroutines int
	}{
		{"NoFailures", 0.0, 10},
		{"LowFailures", 0.1, 10},
		{"HighFailures", 0.5, 10},
		{"Concurrent-Low", 0.1, 100},
		{"Concurrent-High", 0.5, 100},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			cb := circuitbreaker.New(circuitbreaker.Config{
				Threshold:   5,
				Timeout:    time.Second,
				HalfOpenMax: 2,
			})

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					err := cb.Execute(func() error {
						if scenario.failureRate > 0 && rand.Float64() < scenario.failureRate {
							return fmt.Errorf("simulated failure")
						}
						return nil
					})
					if err != nil {
						var lbErr *errors.LoadBalancerError
						if errors.As(err, &lbErr) && lbErr.Code == errors.ErrCircuitOpen {
							continue
						}
						b.Error(err)
					}
				}
			})
		})
	}
}

// BenchmarkRateLimiter measures the performance of the rate limiter
func BenchmarkRateLimiter(b *testing.B) {
	scenarios := []struct {
		name      string
		rate      float64
		burst     float64
		parallel  int
	}{
		{"Low-Rate", 100.0, 10.0, 10},
		{"Medium-Rate", 1000.0, 100.0, 50},
		{"High-Rate", 10000.0, 1000.0, 100},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			rl := ratelimit.New(ratelimit.Config{
				Rate:     scenario.rate,
				Capacity: scenario.burst,
			})

			b.ResetTimer()
			b.SetParallelism(scenario.parallel)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					err := rl.Allow()
					if err != nil {
						var lbErr *errors.LoadBalancerError
						if errors.As(err, &lbErr) && lbErr.Code == errors.ErrRateLimitExceeded {
							continue
						}
						b.Error(err)
					}
				}
			})
		})
	}
}
