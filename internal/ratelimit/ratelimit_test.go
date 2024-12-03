package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	// Test with 100 requests per second limit
	limiter := New(Config{
		Rate:     100,
		Capacity: 100,
	})

	// Test initial state
	if err := limiter.Allow(); err != nil {
		t.Error("First request should be allowed")
	}

	// Test burst handling
	for i := 0; i < 98; i++ {
		if err := limiter.Allow(); err != nil {
			t.Errorf("Request %d should be allowed within burst capacity", i)
		}
	}

	// Test rate limiting
	if err := limiter.Allow(); err == nil {
		t.Error("Expected rate limit to be exceeded")
	}

	// Test refill
	time.Sleep(time.Second)
	if err := limiter.Allow(); err != nil {
		t.Error("Request should be allowed after refill")
	}
}

func TestTokenBucketConcurrency(t *testing.T) {
	limiter := New(Config{
		Rate:     100,
		Capacity: 100,
	})

	var wg sync.WaitGroup
	numGoroutines := 10
	requestsPerGoroutine := 20

	// Track allowed and rejected requests
	var (
		allowed  int32
		rejected int32
		mu       sync.Mutex
	)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				err := limiter.Allow()
				mu.Lock()
				if err == nil {
					allowed++
				} else {
					rejected++
				}
				mu.Unlock()
				time.Sleep(time.Millisecond) // Small delay to simulate real requests
			}
		}()
	}

	wg.Wait()

	totalRequests := numGoroutines * requestsPerGoroutine
	t.Logf("Total: %d, Allowed: %d, Rejected: %d", totalRequests, allowed, rejected)

	if allowed+rejected != int32(totalRequests) {
		t.Errorf("Expected %d total requests, got %d", totalRequests, allowed+rejected)
	}
}

func TestWindowRateLimiter(t *testing.T) {
	limiter := NewWindow(WindowConfig{
		Window:      time.Second,
		Limit:       10,
		CleanupTime: time.Second,
	})
	defer limiter.Stop()

	// Test initial state
	for i := 0; i < 10; i++ {
		if err := limiter.Allow(); err != nil {
			t.Errorf("Request %d should be allowed within limit", i)
		}
	}

	// Test limit exceeded
	if err := limiter.Allow(); err == nil {
		t.Error("Expected rate limit to be exceeded")
	}

	// Test window sliding
	time.Sleep(time.Second)
	if err := limiter.Allow(); err != nil {
		t.Error("Request should be allowed after window slides")
	}
}

func TestWindowRateLimiterConcurrency(t *testing.T) {
	limiter := NewWindow(WindowConfig{
		Window:      time.Second,
		Limit:       50,
		CleanupTime: time.Second,
	})
	defer limiter.Stop()

	var wg sync.WaitGroup
	numGoroutines := 10
	requestsPerGoroutine := 10

	var (
		allowed  int32
		rejected int32
		mu       sync.Mutex
	)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				err := limiter.Allow()
				mu.Lock()
				if err == nil {
					allowed++
				} else {
					rejected++
				}
				mu.Unlock()
				time.Sleep(time.Millisecond)
			}
		}()
	}

	wg.Wait()

	totalRequests := numGoroutines * requestsPerGoroutine
	t.Logf("Total: %d, Allowed: %d, Rejected: %d", totalRequests, allowed, rejected)

	if allowed+rejected != int32(totalRequests) {
		t.Errorf("Expected %d total requests, got %d", totalRequests, allowed+rejected)
	}
}

func TestRateLimiterEdgeCases(t *testing.T) {
	// Test zero rate
	limiter := New(Config{
		Rate:     0,
		Capacity: 100,
	})
	if limiter.rate <= 0 {
		t.Error("Expected positive rate despite zero input")
	}

	// Test zero capacity
	limiter = New(Config{
		Rate:     100,
		Capacity: 0,
	})
	if limiter.capacity <= 0 {
		t.Error("Expected positive capacity despite zero input")
	}

	// Test window limiter with zero values
	windowLimiter := NewWindow(WindowConfig{
		Window:      0,
		Limit:       0,
		CleanupTime: 0,
	})
	defer windowLimiter.Stop()

	if windowLimiter.window <= 0 {
		t.Error("Expected positive window duration despite zero input")
	}
	if windowLimiter.limit <= 0 {
		t.Error("Expected positive limit despite zero input")
	}
	if windowLimiter.cleanupTime <= 0 {
		t.Error("Expected positive cleanup time despite zero input")
	}
}

func TestRateLimiterBurstHandling(t *testing.T) {
	limiter := New(Config{
		Rate:     10,
		Capacity: 20, // Allow bursts up to 20
	})

	// Test burst capacity
	for i := 0; i < 20; i++ {
		if err := limiter.Allow(); err != nil {
			t.Errorf("Request %d should be allowed within burst capacity", i)
		}
	}

	// Verify burst limit
	if err := limiter.Allow(); err == nil {
		t.Error("Expected burst limit to be exceeded")
	}

	// Test recovery
	time.Sleep(2 * time.Second)
	if err := limiter.Allow(); err != nil {
		t.Error("Expected request to be allowed after recovery period")
	}
}
