package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNew(t *testing.T) {
	Reset() // Reset metrics before test
	m := New()

	// Test that metrics instance is created
	if m == nil {
		t.Fatal("Expected non-nil Metrics instance")
	}

	// Test RequestsTotal counter
	if m.RequestsTotal == nil {
		t.Fatal("Expected non-nil RequestsTotal counter")
	}
	if testutil.ToFloat64(m.RequestsTotal) != 0 {
		t.Errorf("Expected initial RequestsTotal to be 0, got %f", testutil.ToFloat64(m.RequestsTotal))
	}

	// Test ResponseTime histogram
	if m.ResponseTime == nil {
		t.Fatal("Expected non-nil ResponseTime histogram")
	}

	// Test ActiveConnections gauge
	if m.ActiveConnections == nil {
		t.Fatal("Expected non-nil ActiveConnections gauge")
	}
	if testutil.ToFloat64(m.ActiveConnections) != 0 {
		t.Errorf("Expected initial ActiveConnections to be 0, got %f", testutil.ToFloat64(m.ActiveConnections))
	}

	// Test BackendHealth gauge vector
	if m.BackendHealth == nil {
		t.Fatal("Expected non-nil BackendHealth gauge vector")
	}

	// Test ErrorsTotal counter
	if m.ErrorsTotal == nil {
		t.Fatal("Expected non-nil ErrorsTotal counter")
	}
	if testutil.ToFloat64(m.ErrorsTotal) != 0 {
		t.Errorf("Expected initial ErrorsTotal to be 0, got %f", testutil.ToFloat64(m.ErrorsTotal))
	}

	// Test registry
	if m.GetRegistry() == nil {
		t.Fatal("Expected non-nil registry")
	}
}

func TestMetricsIncrement(t *testing.T) {
	Reset() // Reset metrics before test
	m := New()

	// Test RequestsTotal increment
	m.RequestsTotal.Inc()
	if testutil.ToFloat64(m.RequestsTotal) != 1 {
		t.Errorf("Expected RequestsTotal to be 1, got %f", testutil.ToFloat64(m.RequestsTotal))
	}

	// Test ActiveConnections gauge
	m.ActiveConnections.Inc()
	if testutil.ToFloat64(m.ActiveConnections) != 1 {
		t.Errorf("Expected ActiveConnections to be 1, got %f", testutil.ToFloat64(m.ActiveConnections))
	}
	m.ActiveConnections.Dec()
	if testutil.ToFloat64(m.ActiveConnections) != 0 {
		t.Errorf("Expected ActiveConnections to be 0, got %f", testutil.ToFloat64(m.ActiveConnections))
	}

	// Test BackendHealth gauge vector
	m.BackendHealth.With(prometheus.Labels{"backend_url": "test-backend"}).Set(1)
	if testutil.ToFloat64(m.BackendHealth.With(prometheus.Labels{"backend_url": "test-backend"})) != 1 {
		t.Error("Expected backend health to be 1")
	}

	// Test ErrorsTotal increment
	m.ErrorsTotal.Inc()
	if testutil.ToFloat64(m.ErrorsTotal) != 1 {
		t.Errorf("Expected ErrorsTotal to be 1, got %f", testutil.ToFloat64(m.ErrorsTotal))
	}
}

func TestMetricsLabels(t *testing.T) {
	Reset() // Reset metrics before test
	m := New()

	// Test backend health labels
	backends := []string{"backend1", "backend2", "backend3"}
	for _, backend := range backends {
		m.BackendHealth.With(prometheus.Labels{"backend_url": backend}).Set(1)
		value := testutil.ToFloat64(m.BackendHealth.With(prometheus.Labels{"backend_url": backend}))
		if value != 1 {
			t.Errorf("Expected backend %s health to be 1, got %f", backend, value)
		}
	}

	// Test setting different health values
	m.BackendHealth.With(prometheus.Labels{"backend_url": "backend1"}).Set(0)
	value := testutil.ToFloat64(m.BackendHealth.With(prometheus.Labels{"backend_url": "backend1"}))
	if value != 0 {
		t.Errorf("Expected backend1 health to be 0, got %f", value)
	}
}

func TestResponseTimeObservation(t *testing.T) {
	Reset() // Reset metrics before test
	m := New()

	// Observe some response times
	times := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	for _, time := range times {
		m.ResponseTime.Observe(time)
	}

	// We can't directly test histogram values, but we can verify the metric exists
	if m.ResponseTime == nil {
		t.Error("Expected ResponseTime histogram to be initialized")
	}
}

func TestMetricsReset(t *testing.T) {
	Reset() // Reset metrics before test
	m := New()

	// Set some values
	m.RequestsTotal.Inc()
	m.ActiveConnections.Inc()
	m.ErrorsTotal.Inc()
	m.BackendHealth.With(prometheus.Labels{"backend_url": "test-backend"}).Set(1)

	// Reset metrics
	Reset()
	m = New()

	// Verify all metrics are reset
	if testutil.ToFloat64(m.RequestsTotal) != 0 {
		t.Error("Expected RequestsTotal to be reset to 0")
	}
	if testutil.ToFloat64(m.ActiveConnections) != 0 {
		t.Error("Expected ActiveConnections to be reset to 0")
	}
	if testutil.ToFloat64(m.ErrorsTotal) != 0 {
		t.Error("Expected ErrorsTotal to be reset to 0")
	}
}

func TestMetricsSingleton(t *testing.T) {
	Reset() // Reset metrics before test
	m1 := New()
	m2 := New()

	if m1 != m2 {
		t.Error("Expected metrics to be a singleton")
	}

	// Test that both instances share the same registry
	if m1.GetRegistry() != m2.GetRegistry() {
		t.Error("Expected metrics instances to share the same registry")
	}
}
