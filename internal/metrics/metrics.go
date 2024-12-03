package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	RequestsTotal    prometheus.Counter
	ResponseTime     prometheus.Histogram
	ActiveConnections prometheus.Gauge
	BackendHealth    *prometheus.GaugeVec
	ErrorsTotal      prometheus.Counter
}

func New() *Metrics {
	m := &Metrics{
		RequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "loadbalancer_requests_total",
			Help: "The total number of processed requests",
		}),
		ResponseTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "loadbalancer_response_time_seconds",
			Help:    "Response time distribution",
			Buckets: prometheus.DefBuckets,
		}),
		ActiveConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "loadbalancer_active_connections",
			Help: "The current number of active connections",
		}),
		BackendHealth: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "loadbalancer_backend_health",
			Help: "Health status of backends (1 for healthy, 0 for unhealthy)",
		}, []string{"backend_url"}),
		ErrorsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "loadbalancer_errors_total",
			Help: "The total number of errors encountered",
		}),
	}

	return m
}
