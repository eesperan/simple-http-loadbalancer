package balancer

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RolloutConfig defines the configuration for a rollout
type RolloutConfig struct {
	NewBackends []string
	BatchSize   int
	Interval    time.Duration
}

// RollbackConfig defines the configuration for a rollback
type RollbackConfig struct {
	PreviousBackends []string
	BatchSize        int
	Interval         time.Duration
}

// Rollout performs a gradual rollout of new backends
func (lb *LoadBalancer) Rollout(ctx context.Context, config RolloutConfig) error {
	if len(config.NewBackends) == 0 {
		return fmt.Errorf("no new backends provided for rollout")
	}

	if config.BatchSize <= 0 {
		config.BatchSize = 1
	}

	if config.Interval <= 0 {
		config.Interval = 30 * time.Second
	}

	// Store current backends for potential rollback
	lb.mu.RLock()
	oldBackends := make([]string, len(lb.backends))
	for i, b := range lb.backends {
		oldBackends[i] = b.URL.String()
	}
	lb.mu.RUnlock()

	// Perform rollout in batches
	for i := 0; i < len(config.NewBackends); i += config.BatchSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			end := i + config.BatchSize
			if end > len(config.NewBackends) {
				end = len(config.NewBackends)
			}

			// Replace backends with current batch
			batch := make([]string, end)
			copy(batch, config.NewBackends[:end])

			if err := lb.updateBackends(batch); err != nil {
				// Rollback on error
				_ = lb.updateBackends(oldBackends)
				return fmt.Errorf("rollout failed: %v", err)
			}

			// Wait for health checks to stabilize
			time.Sleep(config.Interval)
		}
	}

	return nil
}

// Rollback reverts to a previous backend configuration
func (lb *LoadBalancer) Rollback(ctx context.Context, config RollbackConfig) error {
	if len(config.PreviousBackends) == 0 {
		return fmt.Errorf("no previous backends provided for rollback")
	}

	if config.BatchSize <= 0 {
		config.BatchSize = 1
	}

	if config.Interval <= 0 {
		config.Interval = 30 * time.Second
	}

	// Store current backends in case rollback fails
	lb.mu.RLock()
	currentBackends := make([]string, len(lb.backends))
	for i, b := range lb.backends {
		currentBackends[i] = b.URL.String()
	}
	lb.mu.RUnlock()

	// Perform rollback in batches
	for i := 0; i < len(config.PreviousBackends); i += config.BatchSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			end := i + config.BatchSize
			if end > len(config.PreviousBackends) {
				end = len(config.PreviousBackends)
			}

			// Replace backends with current batch
			batch := make([]string, end)
			copy(batch, config.PreviousBackends[:end])

			if err := lb.updateBackends(batch); err != nil {
				// Attempt to restore current configuration
				_ = lb.updateBackends(currentBackends)
				return fmt.Errorf("rollback failed: %v", err)
			}

			// Wait for health checks to stabilize
			time.Sleep(config.Interval)
		}
	}

	return nil
}

// RolloutState tracks the state of ongoing rollouts
type RolloutState struct {
	InProgress bool
	Phase      string
	Progress   float64
	Error      error
	mu         sync.RWMutex
}

func (rs *RolloutState) update(phase string, progress float64, err error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.Phase = phase
	rs.Progress = progress
	rs.Error = err
}

func (rs *RolloutState) getStatus() (string, float64, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.Phase, rs.Progress, rs.Error
}
