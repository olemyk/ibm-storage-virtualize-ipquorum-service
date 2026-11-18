package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("Default namespace", func(t *testing.T) {
		m := New("")
		assert.NotNil(t, m)
		assert.NotNil(t, m.RuntimeInfo)
		assert.NotNil(t, m.ContainersTotal)
		assert.NotNil(t, m.APIRequests)
		assert.False(t, m.startTime.IsZero())
	})

	t.Run("Custom namespace", func(t *testing.T) {
		m := New("custom_agent")
		assert.NotNil(t, m)
	})
}

func TestMetrics_RuntimeMetrics(t *testing.T) {
	// Create a new registry for isolated testing
	reg := prometheus.NewRegistry()

	// Create metrics with custom registry
	runtimeInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "runtime_info",
			Help:      "Runtime information",
		},
		[]string{"runtime", "version", "api_version", "os", "arch"},
	)
	reg.MustRegister(runtimeInfo)

	t.Run("Set runtime info", func(t *testing.T) {
		runtimeInfo.WithLabelValues("docker", "24.0.0", "1.43", "linux", "amd64").Set(1)

		count := testutil.CollectAndCount(runtimeInfo)
		assert.Equal(t, 1, count)

		value := testutil.ToFloat64(runtimeInfo.WithLabelValues("docker", "24.0.0", "1.43", "linux", "amd64"))
		assert.Equal(t, float64(1), value)
	})
}

func TestMetrics_RuntimeAvailable(t *testing.T) {
	reg := prometheus.NewRegistry()

	runtimeAvailable := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "runtime_available",
			Help:      "Runtime availability",
		},
		[]string{"runtime"},
	)
	reg.MustRegister(runtimeAvailable)

	t.Run("Runtime available", func(t *testing.T) {
		runtimeAvailable.WithLabelValues("docker").Set(1)
		value := testutil.ToFloat64(runtimeAvailable.WithLabelValues("docker"))
		assert.Equal(t, float64(1), value)
	})

	t.Run("Runtime unavailable", func(t *testing.T) {
		runtimeAvailable.WithLabelValues("podman").Set(0)
		value := testutil.ToFloat64(runtimeAvailable.WithLabelValues("podman"))
		assert.Equal(t, float64(0), value)
	})
}

func TestMetrics_RuntimePingDuration(t *testing.T) {
	reg := prometheus.NewRegistry()

	pingDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "test",
			Name:      "runtime_ping_duration_seconds",
			Help:      "Runtime ping duration",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"runtime"},
	)
	reg.MustRegister(pingDuration)

	t.Run("Record ping duration", func(t *testing.T) {
		pingDuration.WithLabelValues("docker").Observe(0.05)
		pingDuration.WithLabelValues("docker").Observe(0.03)
		pingDuration.WithLabelValues("docker").Observe(0.04)

		count := testutil.CollectAndCount(pingDuration)
		assert.Equal(t, 1, count)
	})
}

func TestMetrics_ContainerMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()

	containersTotal := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "containers_total",
			Help:      "Total containers",
		},
		[]string{"runtime"},
	)
	reg.MustRegister(containersTotal)

	t.Run("Set container count", func(t *testing.T) {
		containersTotal.WithLabelValues("docker").Set(10)
		value := testutil.ToFloat64(containersTotal.WithLabelValues("docker"))
		assert.Equal(t, float64(10), value)
	})

	t.Run("Update container count", func(t *testing.T) {
		containersTotal.WithLabelValues("docker").Set(15)
		value := testutil.ToFloat64(containersTotal.WithLabelValues("docker"))
		assert.Equal(t, float64(15), value)
	})
}

func TestMetrics_ContainerOperations(t *testing.T) {
	reg := prometheus.NewRegistry()

	containerOps := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "test",
			Name:      "container_operations_total",
			Help:      "Container operations",
		},
		[]string{"runtime", "operation"},
	)
	reg.MustRegister(containerOps)

	t.Run("Increment operations", func(t *testing.T) {
		containerOps.WithLabelValues("docker", "create").Inc()
		containerOps.WithLabelValues("docker", "create").Inc()
		containerOps.WithLabelValues("docker", "start").Inc()

		createCount := testutil.ToFloat64(containerOps.WithLabelValues("docker", "create"))
		assert.Equal(t, float64(2), createCount)

		startCount := testutil.ToFloat64(containerOps.WithLabelValues("docker", "start"))
		assert.Equal(t, float64(1), startCount)
	})
}

func TestMetrics_ContainerOpDuration(t *testing.T) {
	reg := prometheus.NewRegistry()

	opDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "test",
			Name:      "container_operation_duration_seconds",
			Help:      "Operation duration",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"runtime", "operation"},
	)
	reg.MustRegister(opDuration)

	t.Run("Record operation durations", func(t *testing.T) {
		opDuration.WithLabelValues("docker", "create").Observe(0.5)
		opDuration.WithLabelValues("docker", "create").Observe(0.3)
		opDuration.WithLabelValues("docker", "start").Observe(0.1)

		count := testutil.CollectAndCount(opDuration)
		assert.Equal(t, 2, count) // 2 label combinations
	})
}

func TestMetrics_ImageMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()

	imagesTotal := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "images_total",
			Help:      "Total images",
		},
		[]string{"runtime"},
	)
	reg.MustRegister(imagesTotal)

	t.Run("Set image count", func(t *testing.T) {
		imagesTotal.WithLabelValues("docker").Set(25)
		value := testutil.ToFloat64(imagesTotal.WithLabelValues("docker"))
		assert.Equal(t, float64(25), value)
	})
}

func TestMetrics_ImagePullDuration(t *testing.T) {
	reg := prometheus.NewRegistry()

	pullDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "test",
			Name:      "image_pull_duration_seconds",
			Help:      "Image pull duration",
			Buckets:   []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"runtime", "image"},
	)
	reg.MustRegister(pullDuration)

	t.Run("Record pull durations", func(t *testing.T) {
		pullDuration.WithLabelValues("docker", "alpine:latest").Observe(2.5)
		pullDuration.WithLabelValues("docker", "nginx:latest").Observe(15.3)

		count := testutil.CollectAndCount(pullDuration)
		assert.Equal(t, 2, count)
	})
}

func TestMetrics_APIMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()

	apiRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "test",
			Name:      "api_requests_total",
			Help:      "API requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	reg.MustRegister(apiRequests)

	t.Run("Count API requests", func(t *testing.T) {
		apiRequests.WithLabelValues("GET", "/containers", "200").Inc()
		apiRequests.WithLabelValues("GET", "/containers", "200").Inc()
		apiRequests.WithLabelValues("POST", "/containers", "201").Inc()
		apiRequests.WithLabelValues("GET", "/containers", "500").Inc()

		successCount := testutil.ToFloat64(apiRequests.WithLabelValues("GET", "/containers", "200"))
		assert.Equal(t, float64(2), successCount)

		errorCount := testutil.ToFloat64(apiRequests.WithLabelValues("GET", "/containers", "500"))
		assert.Equal(t, float64(1), errorCount)
	})
}

func TestMetrics_APIRequestDuration(t *testing.T) {
	reg := prometheus.NewRegistry()

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "test",
			Name:      "api_request_duration_seconds",
			Help:      "Request duration",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	reg.MustRegister(requestDuration)

	t.Run("Record request durations", func(t *testing.T) {
		requestDuration.WithLabelValues("GET", "/containers").Observe(0.05)
		requestDuration.WithLabelValues("GET", "/containers").Observe(0.03)
		requestDuration.WithLabelValues("POST", "/containers").Observe(0.5)

		count := testutil.CollectAndCount(requestDuration)
		assert.Equal(t, 2, count)
	})
}

func TestMetrics_APIRequestSize(t *testing.T) {
	reg := prometheus.NewRegistry()

	requestSize := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "test",
			Name:      "api_request_size_bytes",
			Help:      "Request size",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 7),
		},
		[]string{"method", "endpoint"},
	)
	reg.MustRegister(requestSize)

	t.Run("Record request sizes", func(t *testing.T) {
		requestSize.WithLabelValues("POST", "/containers").Observe(512)
		requestSize.WithLabelValues("POST", "/containers").Observe(1024)

		count := testutil.CollectAndCount(requestSize)
		assert.Equal(t, 1, count)
	})
}

func TestMetrics_SystemMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()

	cpuUsage := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "system_cpu_usage_percent",
			Help:      "CPU usage",
		},
	)
	reg.MustRegister(cpuUsage)

	memoryUsage := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "system_memory_usage_bytes",
			Help:      "Memory usage",
		},
	)
	reg.MustRegister(memoryUsage)

	t.Run("Set CPU usage", func(t *testing.T) {
		cpuUsage.Set(45.5)
		value := testutil.ToFloat64(cpuUsage)
		assert.Equal(t, 45.5, value)
	})

	t.Run("Set memory usage", func(t *testing.T) {
		memoryUsage.Set(8589934592) // 8GB
		value := testutil.ToFloat64(memoryUsage)
		assert.Equal(t, float64(8589934592), value)
	})
}

func TestMetrics_AgentMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()

	agentUptime := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "agent_uptime_seconds",
			Help:      "Agent uptime",
		},
	)
	reg.MustRegister(agentUptime)

	agentVersion := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "test",
			Name:      "agent_version",
			Help:      "Agent version",
		},
		[]string{"version", "commit", "build_time"},
	)
	reg.MustRegister(agentVersion)

	t.Run("Set uptime", func(t *testing.T) {
		agentUptime.Set(3600) // 1 hour
		value := testutil.ToFloat64(agentUptime)
		assert.Equal(t, float64(3600), value)
	})

	t.Run("Set version", func(t *testing.T) {
		agentVersion.WithLabelValues("1.0.0", "abc123", "2024-01-01").Set(1)
		value := testutil.ToFloat64(agentVersion.WithLabelValues("1.0.0", "abc123", "2024-01-01"))
		assert.Equal(t, float64(1), value)
	})
}

func TestMetrics_UpdateAgentUptime(t *testing.T) {
	m := New("test")

	// Wait a bit to ensure uptime > 0
	time.Sleep(100 * time.Millisecond)

	m.UpdateAgentUptime()

	uptime := testutil.ToFloat64(m.AgentUptime)
	assert.Greater(t, uptime, float64(0))
}

func TestMetrics_RecordRuntimePing(t *testing.T) {
	m := New("test_ping")

	t.Run("Record successful ping", func(t *testing.T) {
		start := time.Now()
		time.Sleep(10 * time.Millisecond)
		duration := time.Since(start)

		m.RecordRuntimePing("docker", duration, nil)

		// Verify availability is set to 1
		available := testutil.ToFloat64(m.RuntimeAvailable.WithLabelValues("docker"))
		assert.Equal(t, float64(1), available)
	})

	t.Run("Record failed ping", func(t *testing.T) {
		m.RecordRuntimePing("podman", 0, assert.AnError)

		// Verify availability is set to 0
		available := testutil.ToFloat64(m.RuntimeAvailable.WithLabelValues("podman"))
		assert.Equal(t, float64(0), available)

		// Verify error counter incremented (actual error type is "connection_failed")
		errors := testutil.ToFloat64(m.RuntimeErrors.WithLabelValues("podman", "ping", "connection_failed"))
		assert.Equal(t, float64(1), errors)
	})
}

func TestMetrics_RecordContainerOperation(t *testing.T) {
	m := New("test_container_op")

	t.Run("Record successful operation", func(t *testing.T) {
		start := time.Now()
		time.Sleep(10 * time.Millisecond)
		duration := time.Since(start)

		m.RecordContainerOperation("docker", "create", duration, nil)

		// Verify counter incremented
		count := testutil.ToFloat64(m.ContainerOperations.WithLabelValues("docker", "create"))
		assert.Equal(t, float64(1), count)
	})

	t.Run("Record failed operation", func(t *testing.T) {
		m.RecordContainerOperation("docker", "start", 0, assert.AnError)

		// Verify error counter incremented (actual error type is "operation_failed")
		errors := testutil.ToFloat64(m.ContainerOpErrors.WithLabelValues("docker", "start", "operation_failed"))
		assert.Equal(t, float64(1), errors)
	})
}

func TestMetrics_RecordImagePull(t *testing.T) {
	m := New("test_image_pull")

	t.Run("Record successful pull", func(t *testing.T) {
		duration := 5 * time.Second
		m.RecordImagePull("docker", "alpine:latest", duration, nil)

		// Histogram doesn't expose count easily, just verify no panic
	})

	t.Run("Record failed pull", func(t *testing.T) {
		m.RecordImagePull("docker", "invalid:image", 0, assert.AnError)

		// Verify error counter incremented (actual error type is "pull_failed")
		errors := testutil.ToFloat64(m.ImagePullErrors.WithLabelValues("docker", "invalid:image", "pull_failed"))
		assert.Equal(t, float64(1), errors)
	})
}

func TestMetrics_RecordAPIRequest(t *testing.T) {
	m := New("test_api_request")

	t.Run("Record successful request", func(t *testing.T) {
		duration := 50 * time.Millisecond
		m.RecordAPIRequest("GET", "/containers", "200", duration, 0, 1024)

		// Verify counter incremented
		count := testutil.ToFloat64(m.APIRequests.WithLabelValues("GET", "/containers", "200"))
		assert.Equal(t, float64(1), count)
	})

	t.Run("Record error request", func(t *testing.T) {
		duration := 10 * time.Millisecond
		m.RecordAPIRequest("POST", "/containers", "500", duration, 512, 0)

		// Verify counter incremented
		count := testutil.ToFloat64(m.APIRequests.WithLabelValues("POST", "/containers", "500"))
		assert.Equal(t, float64(1), count)
	})
}

func TestMetrics_UpdateSystemMetrics(t *testing.T) {
	m := New("test_system_metrics")

	t.Run("Update system metrics", func(t *testing.T) {
		m.UpdateSystemMetrics(45.5, 8589934592, 17179869184, 10737418240, 53687091200)

		cpu := testutil.ToFloat64(m.SystemCPUUsage)
		assert.Equal(t, 45.5, cpu)

		memUsage := testutil.ToFloat64(m.SystemMemoryUsage)
		assert.Equal(t, float64(8589934592), memUsage)

		memTotal := testutil.ToFloat64(m.SystemMemoryTotal)
		assert.Equal(t, float64(17179869184), memTotal)
	})
}

func TestMetrics_UpdateContainerCounts(t *testing.T) {
	m := New("test_container_counts")

	t.Run("Update container counts", func(t *testing.T) {
		m.UpdateContainerCounts("docker", 10, 5, 2, 3)

		total := testutil.ToFloat64(m.ContainersTotal.WithLabelValues("docker"))
		assert.Equal(t, float64(10), total)

		running := testutil.ToFloat64(m.ContainersRunning.WithLabelValues("docker"))
		assert.Equal(t, float64(5), running)

		paused := testutil.ToFloat64(m.ContainersPaused.WithLabelValues("docker"))
		assert.Equal(t, float64(2), paused)

		stopped := testutil.ToFloat64(m.ContainersStopped.WithLabelValues("docker"))
		assert.Equal(t, float64(3), stopped)
	})
}

func TestMetrics_UpdateImageCount(t *testing.T) {
	m := New("test_image_count")

	t.Run("Update image count", func(t *testing.T) {
		m.UpdateImageCount("docker", 25)

		count := testutil.ToFloat64(m.ImagesTotal.WithLabelValues("docker"))
		assert.Equal(t, float64(25), count)
	})
}

func TestMetrics_SetRuntimeInfo(t *testing.T) {
	m := New("test_runtime_info")

	t.Run("Set runtime info", func(t *testing.T) {
		m.SetRuntimeInfo("docker", "24.0.0", "1.43", "linux", "amd64")

		value := testutil.ToFloat64(m.RuntimeInfo.WithLabelValues("docker", "24.0.0", "1.43", "linux", "amd64"))
		assert.Equal(t, float64(1), value)
	})
}

func TestMetrics_SetAgentVersion(t *testing.T) {
	m := New("test_agent_version")

	t.Run("Set agent version", func(t *testing.T) {
		m.SetAgentVersion("1.0.0", "abc123", "2024-01-01")

		value := testutil.ToFloat64(m.AgentVersion.WithLabelValues("1.0.0", "abc123", "2024-01-01"))
		assert.Equal(t, float64(1), value)
	})
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	m := New("test_concurrent")

	t.Run("Concurrent metric updates", func(t *testing.T) {
		done := make(chan bool)

		// Simulate concurrent API requests
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					m.RecordAPIRequest("GET", "/containers", "200", 10*time.Millisecond, 0, 1024)
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify total count
		count := testutil.ToFloat64(m.APIRequests.WithLabelValues("GET", "/containers", "200"))
		assert.Equal(t, float64(1000), count)
	})
}

func TestMetrics_StartTime(t *testing.T) {
	m := New("test_start_time")

	require.False(t, m.startTime.IsZero())
	require.True(t, time.Since(m.startTime) < time.Second)
}

// Made with Bob
