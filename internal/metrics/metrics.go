package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	RequestsTotal     prometheus.Counter
	ResponseTime      prometheus.Histogram
	ActiveConnections prometheus.Gauge
	BackendHealth     *prometheus.GaugeVec
	ErrorsTotal       prometheus.Counter
	registry         *prometheus.Registry
}

var (
	once     sync.Once
	instance *Metrics
)

// New creates a new metrics instance or returns the existing one
func New() *Metrics {
	once.Do(func() {
		registry := prometheus.NewRegistry()
		factory := promauto.With(registry)

		instance = &Metrics{
			registry: registry,
			RequestsTotal: factory.NewCounter(prometheus.CounterOpts{
				Name: "loadbalancer_requests_total",
				Help: "The total number of processed requests",
			}),
			ResponseTime: factory.NewHistogram(prometheus.HistogramOpts{
				Name:    "loadbalancer_response_time_seconds",
				Help:    "Response time distribution",
				Buckets: prometheus.DefBuckets,
			}),
			ActiveConnections: factory.NewGauge(prometheus.GaugeOpts{
				Name: "loadbalancer_active_connections",
				Help: "The current number of active connections",
			}),
			BackendHealth: factory.NewGaugeVec(prometheus.GaugeOpts{
				Name: "loadbalancer_backend_health",
				Help: "Health status of backends (1 for healthy, 0 for unhealthy)",
			}, []string{"backend_url"}),
			ErrorsTotal: factory.NewCounter(prometheus.CounterOpts{
				Name: "loadbalancer_errors_total",
				Help: "The total number of errors encountered",
			}),
		}
	})
	return instance
}

// Reset resets all metrics (useful for testing)
func Reset() {
	once = sync.Once{}
	instance = nil
}

// GetRegistry returns the Prometheus registry
func (m *Metrics) GetRegistry() *prometheus.Registry {
	return m.registry
}
