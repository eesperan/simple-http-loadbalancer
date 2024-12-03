package algorithm

import (
	"sync"
	"sync/atomic"
)

// WeightedBackend represents a backend with an assigned weight
type WeightedBackend struct {
	ID            string
	Weight        int
	CurrentWeight int64
	EffectiveWeight int64
}

// WeightedRoundRobin implements a weighted round-robin algorithm
type WeightedRoundRobin struct {
	backends []*WeightedBackend
	mu       sync.RWMutex
}

// New creates a new WeightedRoundRobin instance
func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{
		backends: make([]*WeightedBackend, 0),
	}
}

// Add adds a new backend with a specified weight
func (wrr *WeightedRoundRobin) Add(id string, weight int) {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	if weight <= 0 {
		weight = 1
	}

	backend := &WeightedBackend{
		ID:              id,
		Weight:          weight,
		EffectiveWeight: int64(weight),
	}

	wrr.backends = append(wrr.backends, backend)
}

// Remove removes a backend by ID
func (wrr *WeightedRoundRobin) Remove(id string) {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	for i, backend := range wrr.backends {
		if backend.ID == id {
			wrr.backends = append(wrr.backends[:i], wrr.backends[i+1:]...)
			return
		}
	}
}

// Next selects the next backend using the weighted round-robin algorithm
func (wrr *WeightedRoundRobin) Next() *WeightedBackend {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	if len(wrr.backends) == 0 {
		return nil
	}

	var totalWeight int64
	var maxWeightBackend *WeightedBackend

	// Update weights and find the backend with maximum current weight
	for _, backend := range wrr.backends {
		atomic.AddInt64(&backend.CurrentWeight, backend.EffectiveWeight)
		totalWeight += backend.EffectiveWeight

		if maxWeightBackend == nil || 
			atomic.LoadInt64(&backend.CurrentWeight) > atomic.LoadInt64(&maxWeightBackend.CurrentWeight) {
			maxWeightBackend = backend
		}
	}

	if maxWeightBackend == nil {
		return nil
	}

	// Decrease the current weight by the total weight of all servers
	atomic.AddInt64(&maxWeightBackend.CurrentWeight, -totalWeight)

	return maxWeightBackend
}

// UpdateWeight updates the weight of a specific backend
func (wrr *WeightedRoundRobin) UpdateWeight(id string, weight int) bool {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	for _, backend := range wrr.backends {
		if backend.ID == id {
			if weight <= 0 {
				weight = 1
			}
			backend.Weight = weight
			atomic.StoreInt64(&backend.EffectiveWeight, int64(weight))
			return true
		}
	}
	return false
}

// AdjustWeight temporarily adjusts the effective weight of a backend
// This can be used for dynamic load balancing based on backend performance
func (wrr *WeightedRoundRobin) AdjustWeight(id string, delta int) bool {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	for _, backend := range wrr.backends {
		if backend.ID == id {
			newWeight := atomic.LoadInt64(&backend.EffectiveWeight) + int64(delta)
			if newWeight <= 0 {
				newWeight = 1
			}
			if newWeight > int64(backend.Weight*2) {
				newWeight = int64(backend.Weight * 2)
			}
			atomic.StoreInt64(&backend.EffectiveWeight, newWeight)
			return true
		}
	}
	return false
}

// Reset resets all current weights to their original values
func (wrr *WeightedRoundRobin) Reset() {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	for _, backend := range wrr.backends {
		atomic.StoreInt64(&backend.CurrentWeight, 0)
		atomic.StoreInt64(&backend.EffectiveWeight, int64(backend.Weight))
	}
}

// GetBackends returns a copy of the current backend list
func (wrr *WeightedRoundRobin) GetBackends() []WeightedBackend {
	wrr.mu.RLock()
	defer wrr.mu.RUnlock()

	backends := make([]WeightedBackend, len(wrr.backends))
	for i, backend := range wrr.backends {
		backends[i] = WeightedBackend{
			ID:              backend.ID,
			Weight:          backend.Weight,
			CurrentWeight:   atomic.LoadInt64(&backend.CurrentWeight),
			EffectiveWeight: atomic.LoadInt64(&backend.EffectiveWeight),
		}
	}
	return backends
}
