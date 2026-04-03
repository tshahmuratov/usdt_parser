package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	// RED metrics (gRPC)
	GRPCRequestsTotal   *prometheus.CounterVec
	GRPCRequestDuration *prometheus.HistogramVec

	// Exchange dependency
	GrinexFetchTotal    *prometheus.CounterVec
	GrinexFetchDuration prometheus.Histogram

	// DB persistence dependency
	DBPersistTotal    *prometheus.CounterVec
	DBPersistDuration prometheus.Histogram

	// Fallback
	FallbackTotal prometheus.Counter

	// Singleflight
	SingleflightTotal  prometheus.Counter
	SingleflightShared prometheus.Counter
}

func NewMetrics() *Metrics {
	return &Metrics{
		GRPCRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests by method and status code.",
		}, []string{"method", "code"}),

		GRPCRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds.",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 5, 10},
		}, []string{"method"}),

		GrinexFetchTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "grinex_fetch_total",
			Help: "Total number of exchange API fetch calls by status.",
		}, []string{"status"}),

		GrinexFetchDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "grinex_fetch_duration_seconds",
			Help:    "Exchange API fetch duration in seconds.",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),

		DBPersistTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_persist_total",
			Help: "Total number of DB persistence operations by status.",
		}, []string{"status"}),

		DBPersistDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "db_persist_duration_seconds",
			Help:    "DB persistence write duration in seconds.",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		}),

		FallbackTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "grinex_fallback_total",
			Help: "Total number of times fallback to cached data was used.",
		}),

		SingleflightTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "singleflight_requests_total",
			Help: "Total number of singleflight calls.",
		}),

		SingleflightShared: promauto.NewCounter(prometheus.CounterOpts{
			Name: "singleflight_shared_total",
			Help: "Total number of singleflight calls that shared a result.",
		}),
	}
}
