package balancer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"loadbalancer/internal/balancer/algorithm"
	"loadbalancer/internal/circuitbreaker"
	"loadbalancer/internal/config"
	"loadbalancer/internal/errors"
	"loadbalancer/internal/metrics"
	"loadbalancer/internal/ratelimit"
	"loadbalancer/internal/ssl"
)

type Backend struct {
	URL           *url.URL
	Proxy         *httputil.ReverseProxy
	Healthy       atomic.Bool
	ActiveConns   atomic.Int64
	TotalRequests atomic.Uint64
	CircuitBreaker *circuitbreaker.CircuitBreaker
	RateLimiter    *ratelimit.TokenBucket
}

type LoadBalancer struct {
	backends []*Backend
	mu       sync.RWMutex
	metrics  *metrics.Metrics
	config   *config.Config
	ssl      *ssl.Manager
	wrr      *algorithm.WeightedRoundRobin
}

func New(cfg *config.Config, metrics *metrics.Metrics) (*LoadBalancer, error) {
	lb := &LoadBalancer{
		metrics: metrics,
		config:  cfg,
		wrr:     algorithm.NewWeightedRoundRobin(),
	}

	// Initialize SSL if configured
	if cfg.SSL != nil {
		sslManager, err := ssl.New(&ssl.Config{
			CertFile:   cfg.SSL.CertFile,
			KeyFile:    cfg.SSL.KeyFile,
			CAFile:     cfg.SSL.CAFile,
			ClientAuth: cfg.SSL.ClientAuth,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize SSL: %v", err)
		}
		lb.ssl = sslManager
	}

	if err := lb.updateBackends(cfg.Backends); err != nil {
		return nil, err
	}

	return lb, nil
}

func (lb *LoadBalancer) updateBackends(backends []string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	var newBackends []*Backend
	for i, backend := range backends {
		url, err := url.Parse(backend)
		if err != nil {
			return errors.New(errors.ErrConfigInvalid, fmt.Sprintf("invalid backend URL %s", backend), err)
		}

		proxy := httputil.NewSingleHostReverseProxy(url)
		b := &Backend{
			URL:   url,
			Proxy: proxy,
			CircuitBreaker: circuitbreaker.New(circuitbreaker.Config{
				Threshold:   5,
				Timeout:     10 * time.Second,
				HalfOpenMax: 2,
			}),
			RateLimiter: ratelimit.New(ratelimit.Config{
				Rate:     100,
				Capacity: 100,
			}),
		}
		b.Healthy.Store(true)
		newBackends = append(newBackends, b)

		// Add to weighted round-robin with default weight of 1
		lb.wrr.Add(fmt.Sprintf("backend-%d", i), 1)
	}

	lb.backends = newBackends
	return nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.nextBackend()
	if backend == nil {
		http.Error(w, "No available backends", http.StatusServiceUnavailable)
		lb.metrics.ErrorsTotal.Inc()
		return
	}

	// Check circuit breaker
	if err := backend.CircuitBreaker.Execute(func() error {
		// Check rate limiter
		if err := backend.RateLimiter.Allow(); err != nil {
			return err
		}

		backend.ActiveConns.Add(1)
		defer backend.ActiveConns.Add(-1)
		backend.TotalRequests.Add(1)

		start := time.Now()
		lb.metrics.RequestsTotal.Inc()
		
		// Create error channel for proxy errors
		errChan := make(chan error, 1)
		
		// Wrap the response writer to capture status
		wrapped := &responseWriter{ResponseWriter: w}
		
		// Proxy the request
		go func() {
			backend.Proxy.ServeHTTP(wrapped, r)
			if wrapped.status >= 500 {
				errChan <- fmt.Errorf("backend error: %d", wrapped.status)
			} else {
				errChan <- nil
			}
		}()

		// Wait for response or timeout
		select {
		case err := <-errChan:
			if err != nil {
				lb.metrics.ErrorsTotal.Inc()
				return err
			}
		case <-time.After(30 * time.Second):
			lb.metrics.ErrorsTotal.Inc()
			return errors.New(errors.ErrTimeout, "request timeout", nil)
		}

		lb.metrics.ResponseTime.Observe(time.Since(start).Seconds())
		return nil
	}); err != nil {
		var lbErr *errors.LoadBalancerError
		if errors.As(err, &lbErr) {
			switch lbErr.Code {
			case errors.ErrCircuitOpen:
				http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
			case errors.ErrRateLimitExceeded:
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
			default:
				http.Error(w, "Backend error", http.StatusBadGateway)
			}
		} else {
			http.Error(w, "Backend error", http.StatusBadGateway)
		}
		lb.metrics.ErrorsTotal.Inc()
		return
	}
}

func (lb *LoadBalancer) nextBackend() *Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.backends) == 0 {
		return nil
	}

	// Use weighted round-robin to select backend
	selected := lb.wrr.Next()
	if selected == nil {
		return nil
	}

	// Convert backend ID to index
	var index int
	fmt.Sscanf(selected.ID, "backend-%d", &index)
	
	if index >= 0 && index < len(lb.backends) {
		return lb.backends[index]
	}

	return nil
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (lb *LoadBalancer) Start(ctx context.Context) error {
	// Start admin server
	go lb.startAdminServer()

	// Start frontend servers
	errChan := make(chan error, len(lb.config.Frontends))
	var wg sync.WaitGroup

	for _, frontend := range lb.config.Frontends {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()

			var handler http.Handler = lb
			if lb.ssl != nil {
				server := &http.Server{
					Addr:      fmt.Sprintf(":%d", port),
					Handler:   handler,
					TLSConfig: lb.ssl.GetTLSConfig(),
				}

				go func() {
					<-ctx.Done()
					server.Shutdown(context.Background())
				}()

				if err := server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
					errChan <- fmt.Errorf("frontend server error: %v", err)
				}
			} else {
				server := &http.Server{
					Addr:    fmt.Sprintf(":%d", port),
					Handler: handler,
				}

				go func() {
					<-ctx.Done()
					server.Shutdown(context.Background())
				}()

				if err := server.ListenAndServe(); err != http.ErrServerClosed {
					errChan <- fmt.Errorf("frontend server error: %v", err)
				}
			}
		}(frontend.Port)
	}

	// Wait for shutdown or error
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Return first error if any
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (lb *LoadBalancer) startAdminServer() {
	// Implementation of admin server
	// TODO: Add admin endpoints for configuration and monitoring
}
