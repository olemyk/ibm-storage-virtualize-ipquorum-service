package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the container runtime agent
type Metrics struct {
	// Runtime metrics
	RuntimeInfo         *prometheus.GaugeVec
	RuntimeAvailable    *prometheus.GaugeVec
	RuntimePingDuration *prometheus.HistogramVec
	RuntimeErrors       *prometheus.CounterVec

	// Container metrics
	ContainersTotal     *prometheus.GaugeVec
	ContainersRunning   *prometheus.GaugeVec
	ContainersPaused    *prometheus.GaugeVec
	ContainersStopped   *prometheus.GaugeVec
	ContainerOperations *prometheus.CounterVec
	ContainerOpDuration *prometheus.HistogramVec
	ContainerOpErrors   *prometheus.CounterVec

	// Image metrics
	ImagesTotal       *prometheus.GaugeVec
	ImagePullDuration *prometheus.HistogramVec
	ImagePullErrors   *prometheus.CounterVec

	// API metrics
	APIRequests        *prometheus.CounterVec
	APIRequestDuration *prometheus.HistogramVec
	APIRequestSize     *prometheus.HistogramVec
	APIResponseSize    *prometheus.HistogramVec
	APIErrors          *prometheus.CounterVec

	// System metrics
	SystemCPUUsage    prometheus.Gauge
	SystemMemoryUsage prometheus.Gauge
	SystemMemoryTotal prometheus.Gauge
	SystemDiskUsage   prometheus.Gauge
	SystemDiskTotal   prometheus.Gauge

	// Agent metrics
	AgentUptime      prometheus.Gauge
	AgentVersion     *prometheus.GaugeVec
	AgentConnections prometheus.Gauge

	mu        sync.RWMutex
	startTime time.Time
}

// New creates a new Metrics instance with all Prometheus metrics registered
func New(namespace string) *Metrics {
	if namespace == "" {
		namespace = "ipquorum_agent"
	}

	m := &Metrics{
		startTime: time.Now(),

		// Runtime metrics
		RuntimeInfo: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "runtime_info",
				Help:      "Runtime information (version, type)",
			},
			[]string{"runtime", "version", "api_version", "os", "arch"},
		),
		RuntimeAvailable: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "runtime_available",
				Help:      "Runtime availability (1=available, 0=unavailable)",
			},
			[]string{"runtime"},
		),
		RuntimePingDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "runtime_ping_duration_seconds",
				Help:      "Runtime ping duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"runtime"},
		),
		RuntimeErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "runtime_errors_total",
				Help:      "Total number of runtime errors",
			},
			[]string{"runtime", "operation", "error_type"},
		),

		// Container metrics
		ContainersTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "containers_total",
				Help:      "Total number of containers",
			},
			[]string{"runtime"},
		),
		ContainersRunning: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "containers_running",
				Help:      "Number of running containers",
			},
			[]string{"runtime"},
		),
		ContainersPaused: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "containers_paused",
				Help:      "Number of paused containers",
			},
			[]string{"runtime"},
		),
		ContainersStopped: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "containers_stopped",
				Help:      "Number of stopped containers",
			},
			[]string{"runtime"},
		),
		ContainerOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "container_operations_total",
				Help:      "Total number of container operations",
			},
			[]string{"runtime", "operation"},
		),
		ContainerOpDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "container_operation_duration_seconds",
				Help:      "Container operation duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"runtime", "operation"},
		),
		ContainerOpErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "container_operation_errors_total",
				Help:      "Total number of container operation errors",
			},
			[]string{"runtime", "operation", "error_type"},
		),

		// Image metrics
		ImagesTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "images_total",
				Help:      "Total number of images",
			},
			[]string{"runtime"},
		),
		ImagePullDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "image_pull_duration_seconds",
				Help:      "Image pull duration in seconds",
				Buckets:   []float64{1, 5, 10, 30, 60, 120, 300},
			},
			[]string{"runtime", "image"},
		),
		ImagePullErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "image_pull_errors_total",
				Help:      "Total number of image pull errors",
			},
			[]string{"runtime", "image", "error_type"},
		),

		// API metrics
		APIRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "api_requests_total",
				Help:      "Total number of API requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		APIRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "api_request_duration_seconds",
				Help:      "API request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		APIRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "api_request_size_bytes",
				Help:      "API request size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 7),
			},
			[]string{"method", "endpoint"},
		),
		APIResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "api_response_size_bytes",
				Help:      "API response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 7),
			},
			[]string{"method", "endpoint"},
		),
		APIErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "api_errors_total",
				Help:      "Total number of API errors",
			},
			[]string{"method", "endpoint", "error_type"},
		),

		// System metrics
		SystemCPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "system_cpu_usage_percent",
				Help:      "System CPU usage percentage",
			},
		),
		SystemMemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "system_memory_usage_bytes",
				Help:      "System memory usage in bytes",
			},
		),
		SystemMemoryTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "system_memory_total_bytes",
				Help:      "System total memory in bytes",
			},
		),
		SystemDiskUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "system_disk_usage_bytes",
				Help:      "System disk usage in bytes",
			},
		),
		SystemDiskTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "system_disk_total_bytes",
				Help:      "System total disk space in bytes",
			},
		),

		// Agent metrics
		AgentUptime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "agent_uptime_seconds",
				Help:      "Agent uptime in seconds",
			},
		),
		AgentVersion: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "agent_version_info",
				Help:      "Agent version information",
			},
			[]string{"version", "commit", "build_time"},
		),
		AgentConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "agent_connections",
				Help:      "Number of active agent connections",
			},
		),
	}

	return m
}

// RecordRuntimePing records a runtime ping operation
func (m *Metrics) RecordRuntimePing(runtime string, duration time.Duration, err error) {
	m.RuntimePingDuration.WithLabelValues(runtime).Observe(duration.Seconds())

	if err != nil {
		m.RuntimeAvailable.WithLabelValues(runtime).Set(0)
		m.RuntimeErrors.WithLabelValues(runtime, "ping", "connection_failed").Inc()
	} else {
		m.RuntimeAvailable.WithLabelValues(runtime).Set(1)
	}
}

// RecordContainerOperation records a container operation
func (m *Metrics) RecordContainerOperation(runtime, operation string, duration time.Duration, err error) {
	m.ContainerOperations.WithLabelValues(runtime, operation).Inc()
	m.ContainerOpDuration.WithLabelValues(runtime, operation).Observe(duration.Seconds())

	if err != nil {
		errorType := "unknown"
		// Categorize error types
		if err.Error() != "" {
			errorType = "operation_failed"
		}
		m.ContainerOpErrors.WithLabelValues(runtime, operation, errorType).Inc()
	}
}

// UpdateContainerCounts updates container count metrics
func (m *Metrics) UpdateContainerCounts(runtime string, total, running, paused, stopped int) {
	m.ContainersTotal.WithLabelValues(runtime).Set(float64(total))
	m.ContainersRunning.WithLabelValues(runtime).Set(float64(running))
	m.ContainersPaused.WithLabelValues(runtime).Set(float64(paused))
	m.ContainersStopped.WithLabelValues(runtime).Set(float64(stopped))
}

// RecordImagePull records an image pull operation
func (m *Metrics) RecordImagePull(runtime, image string, duration time.Duration, err error) {
	m.ImagePullDuration.WithLabelValues(runtime, image).Observe(duration.Seconds())

	if err != nil {
		errorType := "pull_failed"
		m.ImagePullErrors.WithLabelValues(runtime, image, errorType).Inc()
	}
}

// UpdateImageCount updates image count metrics
func (m *Metrics) UpdateImageCount(runtime string, count int) {
	m.ImagesTotal.WithLabelValues(runtime).Set(float64(count))
}

// RecordAPIRequest records an API request
func (m *Metrics) RecordAPIRequest(method, endpoint, status string, duration time.Duration, requestSize, responseSize int64) {
	m.APIRequests.WithLabelValues(method, endpoint, status).Inc()
	m.APIRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

	if requestSize > 0 {
		m.APIRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
	}
	if responseSize > 0 {
		m.APIResponseSize.WithLabelValues(method, endpoint).Observe(float64(responseSize))
	}
}

// RecordAPIError records an API error
func (m *Metrics) RecordAPIError(method, endpoint, errorType string) {
	m.APIErrors.WithLabelValues(method, endpoint, errorType).Inc()
}

// UpdateSystemMetrics updates system resource metrics
func (m *Metrics) UpdateSystemMetrics(cpuPercent float64, memUsage, memTotal, diskUsage, diskTotal uint64) {
	m.SystemCPUUsage.Set(cpuPercent)
	m.SystemMemoryUsage.Set(float64(memUsage))
	m.SystemMemoryTotal.Set(float64(memTotal))
	m.SystemDiskUsage.Set(float64(diskUsage))
	m.SystemDiskTotal.Set(float64(diskTotal))
}

// UpdateAgentUptime updates the agent uptime metric
func (m *Metrics) UpdateAgentUptime() {
	m.AgentUptime.Set(time.Since(m.startTime).Seconds())
}

// SetAgentVersion sets the agent version information
func (m *Metrics) SetAgentVersion(version, commit, buildTime string) {
	m.AgentVersion.WithLabelValues(version, commit, buildTime).Set(1)
}

// UpdateAgentConnections updates the active connections count
func (m *Metrics) UpdateAgentConnections(count int) {
	m.AgentConnections.Set(float64(count))
}

// SetRuntimeInfo sets runtime information
func (m *Metrics) SetRuntimeInfo(runtime, version, apiVersion, os, arch string) {
	m.RuntimeInfo.WithLabelValues(runtime, version, apiVersion, os, arch).Set(1)
}

// Made with Bob
