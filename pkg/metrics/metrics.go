package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "anarva_http_requests_total",
			Help: "Total number of HTTP requests processed by status code, method, and path",
		},
		[]string{"status", "method", "path"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "anarva_http_request_duration_seconds",
			Help:    "Histogram of HTTP request latencies in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Database Metrics
	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "anarva_db_queries_total",
			Help: "Total number of database queries executed by operation and status",
		},
		[]string{"operation", "status"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "anarva_db_query_duration_seconds",
			Help:    "Histogram of database query execution durations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "anarva_active_connections",
			Help: "Current number of active client connections",
		},
	)
)

// Handler returns an http.Handler for serving Prometheus metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}

// RecordHTTPRequest records status, method, path, and duration of an HTTP request.
func RecordHTTPRequest(status int, method, path string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(strconv.Itoa(status), method, path).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordDatabaseQuery records operation type, status ("success"/"error"), and duration.
func RecordDatabaseQuery(operation, status string, duration float64) {
	DatabaseQueriesTotal.WithLabelValues(operation, status).Inc()
	DatabaseQueryDuration.WithLabelValues(operation).Observe(duration)
}
