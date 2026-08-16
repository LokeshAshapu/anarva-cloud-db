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

	// Operations & Reliability Metrics
	OperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "anarva_operations_total",
			Help: "Total control-plane operations by operation type and status",
		},
		[]string{"type", "status"},
	)

	ActiveOperationsGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "anarva_active_operations",
			Help: "Current count of control-plane operations by status",
		},
		[]string{"status"},
	)

	ActiveResourceLocksGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "anarva_active_resource_locks",
			Help: "Current count of active distributed resource lock leases",
		},
	)

	OperationTimeoutTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "anarva_operation_timeouts_total",
			Help: "Total control-plane operation timeouts detected",
		},
	)

	OperationRecoveryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "anarva_operation_recoveries_total",
			Help: "Total operation recoveries executed by result (success/failure)",
		},
		[]string{"result"},
	)

	IdempotencyConflictsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "anarva_idempotency_conflicts_total",
			Help: "Total idempotency key reuse conflicts",
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

// RecordOperationEvent records operation metric transitions.
func RecordOperationEvent(opType, status string) {
	OperationsTotal.WithLabelValues(opType, status).Inc()
}

// RecordOperationRecovery records operation recovery result.
func RecordOperationRecovery(result string, duration float64) {
	OperationRecoveryTotal.WithLabelValues(result).Inc()
}

// RecordLockConflict records a distributed resource lock conflict.
func RecordLockConflict(resourceID string) {
	IdempotencyConflictsTotal.Inc()
}

// RecordLockExpiration records an expired lock lease event.
func RecordLockExpiration(resourceID string) {
	OperationTimeoutTotal.Inc()
}
