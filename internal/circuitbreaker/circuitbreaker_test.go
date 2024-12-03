package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	cb := New(Config{
		Threshold:   3,
		Timeout:     100 * time.Millisecond,
		HalfOpenMax: 2,
	})

	// Test initial state
	if state := cb.GetState(); state != StateClosed {
		t.Errorf("Expected initial state to be Closed, got %v", state)
	}

	// Test successful operations
	for i := 0; i < 5; i++ {
		err := cb.Execute(func() error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected successful operation, got error: %v", err)
		}
	}

	// Test failure threshold
	failingOp := func() error {
		return errors.New("test error")
	}

	for i := 0; i < 3; i++ {
		_ = cb.Execute(failingOp)
	}

	if state := cb.GetState(); state != StateOpen {
		t.Errorf("Expected state to be Open after failures, got %v", state)
	}

	// Test circuit open rejection
	err := cb.Execute(func() error {
		return nil
	})
	if err == nil {
		t.Error("Expected error when circuit is open")
	}

	// Test timeout and half-open state
	time.Sleep(150 * time.Millisecond)

	if !cb.AllowRequest() {
		t.Error("Expected circuit to allow request after timeout")
	}

	if state := cb.GetState(); state != StateHalfOpen {
		t.Errorf("Expected state to be HalfOpen after timeout, got %v", state)
	}

	// Test successful recovery
	for i := 0; i < cb.halfOpenMax; i++ {
		err := cb.Execute(func() error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected successful operation in half-open state, got error: %v", err)
		}
	}

	if state := cb.GetState(); state != StateClosed {
		t.Errorf("Expected state to be Closed after recovery, got %v", state)
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := New(Config{
		Threshold:   2,
		Timeout:     100 * time.Millisecond,
		HalfOpenMax: 1,
	})

	// Force circuit to open
	failingOp := func() error {
		return errors.New("test error")
	}

	for i := 0; i < 2; i++ {
		_ = cb.Execute(failingOp)
	}

	if state := cb.GetState(); state != StateOpen {
		t.Errorf("Expected state to be Open, got %v", state)
	}

	// Reset circuit
	cb.Reset()

	if state := cb.GetState(); state != StateClosed {
		t.Errorf("Expected state to be Closed after reset, got %v", state)
	}

	// Verify circuit works after reset
	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected successful operation after reset, got error: %v", err)
	}
}

func TestCircuitBreakerConcurrency(t *testing.T) {
	cb := New(Config{
		Threshold:   5,
		Timeout:     100 * time.Millisecond,
		HalfOpenMax: 2,
	})

	// Test concurrent operations
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	const numGoroutines = 10
	errCh := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					errCh <- nil
					return
				default:
					err := cb.Execute(func() error {
						time.Sleep(10 * time.Millisecond)
						return nil
					})
					if err != nil {
						errCh <- err
						return
					}
				}
			}
		}()
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("Concurrent operation failed: %v", err)
		}
	}
}

func TestCircuitBreakerEdgeCases(t *testing.T) {
	// Test with zero threshold
	cb := New(Config{
		Threshold:   0,
		Timeout:     100 * time.Millisecond,
		HalfOpenMax: 1,
	})

	if cb.threshold <= 0 {
		t.Error("Expected positive threshold despite zero input")
	}

	// Test with zero timeout
	cb = New(Config{
		Threshold:   5,
		Timeout:     0,
		HalfOpenMax: 1,
	})

	if cb.timeout <= 0 {
		t.Error("Expected positive timeout despite zero input")
	}

	// Test with zero half-open max
	cb = New(Config{
		Threshold:   5,
		Timeout:     100 * time.Millisecond,
		HalfOpenMax: 0,
	})

	if cb.halfOpenMax <= 0 {
		t.Error("Expected positive half-open max despite zero input")
	}
}
