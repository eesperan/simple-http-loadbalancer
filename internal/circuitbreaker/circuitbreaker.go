package circuitbreaker

import (
	"sync"
	"time"

	"loadbalancer/internal/errors"
)

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

type CircuitBreaker struct {
	mu sync.RWMutex

	failures     int
	threshold    int
	timeout      time.Duration
	lastFailure  time.Time
	state        State
	halfOpenMax  int
	successCount int
}

type Config struct {
	Threshold   int
	Timeout     time.Duration
	HalfOpenMax int
}

func New(config Config) *CircuitBreaker {
	if config.Threshold <= 0 {
		config.Threshold = 5
	}
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	if config.HalfOpenMax <= 0 {
		config.HalfOpenMax = 3
	}

	return &CircuitBreaker{
		threshold:   config.Threshold,
		timeout:     config.Timeout,
		halfOpenMax: config.HalfOpenMax,
		state:      StateClosed,
	}
}

func (cb *CircuitBreaker) Execute(operation func() error) error {
	if !cb.AllowRequest() {
		return errors.New(errors.ErrCircuitOpen, "circuit breaker is open", nil)
	}

	err := operation()
	cb.RecordResult(err)
	return err
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.successCount = 0
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) RecordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.state == StateClosed && cb.failures >= cb.threshold {
			cb.state = StateOpen
		} else if cb.state == StateHalfOpen {
			cb.state = StateOpen
		}
	} else {
		switch cb.state {
		case StateHalfOpen:
			cb.successCount++
			if cb.successCount >= cb.halfOpenMax {
				cb.state = StateClosed
				cb.failures = 0
			}
		case StateClosed:
			cb.failures = 0
		}
	}
}

func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failures = 0
	cb.state = StateClosed
	cb.successCount = 0
}
