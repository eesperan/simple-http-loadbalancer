package ratelimit

import (
	"sync"
	"time"

	"loadbalancer/internal/errors"
)

// TokenBucket implements the token bucket algorithm for rate limiting
type TokenBucket struct {
	rate       float64    // tokens per second
	capacity   float64    // maximum burst size
	tokens     float64    // current number of tokens
	lastRefill time.Time  // last time tokens were added
	mu         sync.Mutex // protects concurrent access
}

// Config holds configuration for the rate limiter
type Config struct {
	Rate     float64 // tokens per second
	Capacity float64 // maximum burst size
}

// New creates a new token bucket rate limiter
func New(config Config) *TokenBucket {
	if config.Rate <= 0 {
		config.Rate = 100 // default to 100 requests per second
	}
	if config.Capacity <= 0 {
		config.Capacity = config.Rate // default capacity to rate
	}

	return &TokenBucket{
		rate:       config.Rate,
		capacity:   config.Capacity,
		tokens:     config.Capacity,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request should be allowed and consumes a token if available
func (tb *TokenBucket) Allow() error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	tb.refill(now)

	if tb.tokens >= 1 {
		tb.tokens--
		return nil
	}

	return errors.New(errors.ErrRateLimitExceeded, "rate limit exceeded", nil)
}

// refill adds tokens based on elapsed time
func (tb *TokenBucket) refill(now time.Time) {
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.rate

	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = now
}

// WindowRateLimiter implements a sliding window rate limiter
type WindowRateLimiter struct {
	mu          sync.Mutex
	window      time.Duration
	limit       int
	requests    map[int64]int
	cleanupTime time.Duration
}

// WindowConfig holds configuration for the sliding window rate limiter
type WindowConfig struct {
	Window      time.Duration
	Limit       int
	CleanupTime time.Duration
}

// NewWindow creates a new sliding window rate limiter
func NewWindow(config WindowConfig) *WindowRateLimiter {
	if config.Window <= 0 {
		config.Window = time.Second
	}
	if config.Limit <= 0 {
		config.Limit = 100
	}
	if config.CleanupTime <= 0 {
		config.CleanupTime = time.Minute
	}

	limiter := &WindowRateLimiter{
		window:      config.Window,
		limit:       config.Limit,
		requests:    make(map[int64]int),
		cleanupTime: config.CleanupTime,
	}

	go limiter.cleanup()
	return limiter
}

// Allow checks if a request should be allowed under the sliding window
func (wrl *WindowRateLimiter) Allow() error {
	wrl.mu.Lock()
	defer wrl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-wrl.window).UnixNano()

	// Count requests in current window
	var count int
	for timestamp, reqs := range wrl.requests {
		if timestamp >= windowStart {
			count += reqs
		}
	}

	if count >= wrl.limit {
		return errors.New(errors.ErrRateLimitExceeded, "rate limit exceeded", nil)
	}

	// Record new request
	bucket := now.UnixNano()
	wrl.requests[bucket]++

	return nil
}

// cleanup periodically removes old entries
func (wrl *WindowRateLimiter) cleanup() {
	ticker := time.NewTicker(wrl.cleanupTime)
	for range ticker.C {
		wrl.mu.Lock()
		threshold := time.Now().Add(-wrl.window).UnixNano()
		for timestamp := range wrl.requests {
			if timestamp < threshold {
				delete(wrl.requests, timestamp)
			}
		}
		wrl.mu.Unlock()
	}
}

// Stop stops the cleanup goroutine
func (wrl *WindowRateLimiter) Stop() {
	wrl.mu.Lock()
	defer wrl.mu.Unlock()
	wrl.requests = nil
}
