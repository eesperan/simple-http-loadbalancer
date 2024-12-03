package algorithm

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestWeightedRoundRobin(t *testing.T) {
	wrr := NewWeightedRoundRobin()

	// Test adding backends with different weights
	wrr.Add("backend1", 5)
	wrr.Add("backend2", 3)
	wrr.Add("backend3", 2)

	// Count selections to verify distribution
	selections := make(map[string]int)
	totalRequests := 100

	for i := 0; i < totalRequests; i++ {
		backend := wrr.Next()
		if backend == nil {
			t.Fatal("Expected non-nil backend")
		}
		selections[backend.ID]++
	}

	// Verify distribution roughly matches weights
	expectedRatio1 := float64(selections["backend1"]) / float64(totalRequests)
	expectedRatio2 := float64(selections["backend2"]) / float64(totalRequests)
	expectedRatio3 := float64(selections["backend3"]) / float64(totalRequests)

	if expectedRatio1 < 0.45 || expectedRatio1 > 0.55 { // ~0.5 (weight 5/10)
		t.Errorf("Backend1 ratio %f not within expected range", expectedRatio1)
	}
	if expectedRatio2 < 0.25 || expectedRatio2 > 0.35 { // ~0.3 (weight 3/10)
		t.Errorf("Backend2 ratio %f not within expected range", expectedRatio2)
	}
	if expectedRatio3 < 0.15 || expectedRatio3 > 0.25 { // ~0.2 (weight 2/10)
		t.Errorf("Backend3 ratio %f not within expected range", expectedRatio3)
	}
}

func TestWeightedRoundRobinEdgeCases(t *testing.T) {
	wrr := NewWeightedRoundRobin()

	// Test with no backends
	if backend := wrr.Next(); backend != nil {
		t.Error("Expected nil backend when no backends available")
	}

	// Test with zero weight
	wrr.Add("backend1", 0)
	backend := wrr.Next()
	if backend == nil || backend.Weight != 1 {
		t.Error("Expected minimum weight of 1 for zero weight input")
	}

	// Test removing backend
	wrr.Remove("backend1")
	if backend := wrr.Next(); backend != nil {
		t.Error("Expected nil backend after removing only backend")
	}

	// Test updating weight
	wrr.Add("backend1", 5)
	wrr.UpdateWeight("backend1", 10)
	backend = wrr.Next()
	if backend == nil || backend.Weight != 10 {
		t.Error("Expected weight to be updated to 10")
	}

	// Test updating non-existent backend
	if wrr.UpdateWeight("nonexistent", 5) {
		t.Error("Expected update of non-existent backend to return false")
	}
}

func TestWeightedRoundRobinConcurrency(t *testing.T) {
	wrr := NewWeightedRoundRobin()
	wrr.Add("backend1", 5)
	wrr.Add("backend2", 5)

	var wg sync.WaitGroup
	numGoroutines := 100
	numRequests := 1000
	mutex := sync.Mutex{}
	selections := make(map[string]int)

	// Concurrent backend selection
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numRequests; j++ {
				backend := wrr.Next()
				if backend != nil {
					mutex.Lock()
					selections[backend.ID]++
					mutex.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Verify total requests
	totalSelections := 0
	for _, count := range selections {
		totalSelections += count
	}

	expectedTotal := numGoroutines * numRequests
	if totalSelections != expectedTotal {
		t.Errorf("Expected %d total selections, got %d", expectedTotal, totalSelections)
	}

	// Verify roughly even distribution (since weights are equal)
	for _, count := range selections {
		ratio := float64(count) / float64(totalSelections)
		if ratio < 0.45 || ratio > 0.55 {
			t.Errorf("Distribution ratio %f outside acceptable range", ratio)
		}
	}
}

func TestWeightedRoundRobinDynamicAdjustment(t *testing.T) {
	wrr := NewWeightedRoundRobin()
	wrr.Add("backend1", 5)

	// Test weight adjustment
	if !wrr.AdjustWeight("backend1", 2) {
		t.Error("Expected successful weight adjustment")
	}

	backend := wrr.Next()
	if backend == nil || atomic.LoadInt64(&backend.EffectiveWeight) != 7 {
		t.Error("Expected effective weight to be adjusted")
	}

	// Test maximum weight limit
	wrr.AdjustWeight("backend1", 100)
	backend = wrr.Next()
	if backend == nil || atomic.LoadInt64(&backend.EffectiveWeight) > int64(backend.Weight*2) {
		t.Error("Expected effective weight to be capped at double the original weight")
	}

	// Test minimum weight limit
	wrr.AdjustWeight("backend1", -100)
	backend = wrr.Next()
	if backend == nil || atomic.LoadInt64(&backend.EffectiveWeight) < 1 {
		t.Error("Expected effective weight to be minimum 1")
	}

	// Test reset
	wrr.Reset()
	backend = wrr.Next()
	if backend == nil || atomic.LoadInt64(&backend.EffectiveWeight) != int64(backend.Weight) {
		t.Error("Expected weight to be reset to original value")
	}
}

func TestWeightedRoundRobinGetBackends(t *testing.T) {
	wrr := NewWeightedRoundRobin()
	wrr.Add("backend1", 5)
	wrr.Add("backend2", 3)

	backends := wrr.GetBackends()
	if len(backends) != 2 {
		t.Errorf("Expected 2 backends, got %d", len(backends))
	}

	// Verify backend properties
	for _, backend := range backends {
		switch backend.ID {
		case "backend1":
			if backend.Weight != 5 {
				t.Errorf("Expected weight 5 for backend1, got %d", backend.Weight)
			}
		case "backend2":
			if backend.Weight != 3 {
				t.Errorf("Expected weight 3 for backend2, got %d", backend.Weight)
			}
		default:
			t.Errorf("Unexpected backend ID: %s", backend.ID)
		}
	}
}
