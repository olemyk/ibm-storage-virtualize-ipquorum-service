package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.SummaryVec
	HTTPResponseSize    *prometheus.SummaryVec

	// Instance metrics
	InstancesTotal            prometheus.Gauge
	InstancesHealthy          prometheus.Gauge
	InstancesUnhealthy        prometheus.Gauge
	InstancesDegraded         prometheus.Gauge
	InstanceOperations        *prometheus.CounterVec
	InstanceOperationDuration *prometheus.HistogramVec

	// Authentication metrics
	AuthAttempts     *prometheus.CounterVec
	AuthTokensIssued prometheus.Counter
	AuthTokensActive prometheus.Gauge

	// Database metrics
	DBQueriesTotal      *prometheus.CounterVec
	DBQueryDuration     *prometheus.HistogramVec
	DBConnectionsActive prometheus.Gauge

	// Health check metrics
	HealthChecksTotal   *prometheus.CounterVec
	HealthCheckDuration *prometheus.HistogramVec

	// Script execution metrics
	ScriptExecutionsTotal   *prometheus.CounterVec
	ScriptExecutionDuration *prometheus.HistogramVec
}

// New creates and registers all Prometheus metrics
func New(namespace string) *Metrics {
	m := &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
			},
			[]string{"method", "path"},
		),

		// Instance metrics
		InstancesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "instances_total",
				Help:      "Total number of IP Quorum instances",
			},
		),
		InstancesHealthy: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "instances_healthy",
				Help:      "Number of healthy IP Quorum instances",
			},
		),
		InstancesUnhealthy: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "instances_unhealthy",
				Help:      "Number of unhealthy IP Quorum instances",
			},
		),
		InstancesDegraded: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "instances_degraded",
				Help:      "Number of degraded IP Quorum instances",
			},
		),
		InstanceOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "instance_operations_total",
				Help:      "Total number of instance operations",
			},
			[]string{"operation", "status"},
		),
		InstanceOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "instance_operation_duration_seconds",
				Help:      "Instance operation duration in seconds",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
			},
			[]string{"operation"},
		),

		// Authentication metrics
		AuthAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_attempts_total",
				Help:      "Total number of authentication attempts",
			},
			[]string{"status"},
		),
		AuthTokensIssued: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_tokens_issued_total",
				Help:      "Total number of authentication tokens issued",
			},
		),
		AuthTokensActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "auth_tokens_active",
				Help:      "Number of active authentication tokens",
			},
		),

		// Database metrics
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "status"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
			},
			[]string{"operation"},
		),
		DBConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_active",
				Help:      "Number of active database connections",
			},
		),

		// Health check metrics
		HealthChecksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "health_checks_total",
				Help:      "Total number of health checks performed",
			},
			[]string{"instance_id", "status"},
		),
		HealthCheckDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "health_check_duration_seconds",
				Help:      "Health check duration in seconds",
				Buckets:   []float64{0.5, 1, 2, 5, 10, 30},
			},
			[]string{"instance_id"},
		),

		// Script execution metrics
		ScriptExecutionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "script_executions_total",
				Help:      "Total number of script executions",
			},
			[]string{"script", "status"},
		),
		ScriptExecutionDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "script_execution_duration_seconds",
				Help:      "Script execution duration in seconds",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120},
			},
			[]string{"script"},
		),
	}

	return m
}

// UpdateInstanceMetrics updates instance-related metrics based on current state
func (m *Metrics) UpdateInstanceMetrics(total, healthy, unhealthy, degraded int) {
	m.InstancesTotal.Set(float64(total))
	m.InstancesHealthy.Set(float64(healthy))
	m.InstancesUnhealthy.Set(float64(unhealthy))
	m.InstancesDegraded.Set(float64(degraded))
}
