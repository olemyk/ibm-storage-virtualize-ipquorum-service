package metrics

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
)

// Collector periodically collects system and runtime metrics
type Collector struct {
	metrics  *Metrics
	logger   *zap.Logger
	interval time.Duration
	stopCh   chan struct{}
}

// NewCollector creates a new metrics collector
func NewCollector(metrics *Metrics, logger *zap.Logger, interval time.Duration) *Collector {
	if interval == 0 {
		interval = 30 * time.Second
	}

	return &Collector{
		metrics:  metrics,
		logger:   logger,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins collecting metrics
func (c *Collector) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Collect immediately on start
	c.collect()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Metrics collector stopped")
			return
		case <-c.stopCh:
			c.logger.Info("Metrics collector stopped")
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

// Stop stops the collector
func (c *Collector) Stop() {
	close(c.stopCh)
}

// collect gathers all metrics
func (c *Collector) collect() {
	// Update agent uptime
	c.metrics.UpdateAgentUptime()

	// Collect system metrics
	c.collectSystemMetrics()
}

// collectSystemMetrics collects system resource metrics
func (c *Collector) collectSystemMetrics() {
	// CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		c.logger.Warn("Failed to get CPU usage", zap.Error(err))
	} else if len(cpuPercent) > 0 {
		c.metrics.SystemCPUUsage.Set(cpuPercent[0])
	}

	// Memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		c.logger.Warn("Failed to get memory info", zap.Error(err))
	} else {
		c.metrics.SystemMemoryUsage.Set(float64(memInfo.Used))
		c.metrics.SystemMemoryTotal.Set(float64(memInfo.Total))
	}

	// Disk usage (root partition)
	diskInfo, err := disk.Usage("/")
	if err != nil {
		c.logger.Warn("Failed to get disk info", zap.Error(err))
	} else {
		c.metrics.SystemDiskUsage.Set(float64(diskInfo.Used))
		c.metrics.SystemDiskTotal.Set(float64(diskInfo.Total))
	}

	// Go runtime metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.logger.Debug("System metrics collected",
		zap.Float64("cpu_percent", cpuPercent[0]),
		zap.Uint64("memory_used", memInfo.Used),
		zap.Uint64("memory_total", memInfo.Total),
		zap.Uint64("disk_used", diskInfo.Used),
		zap.Uint64("disk_total", diskInfo.Total),
		zap.Uint64("go_alloc", m.Alloc),
		zap.Uint64("go_sys", m.Sys),
		zap.Uint32("go_goroutines", uint32(runtime.NumGoroutine())),
	)
}

// CollectRuntimeMetrics collects metrics from a container runtime
func (c *Collector) CollectRuntimeMetrics(ctx context.Context, rt interface{}) error {
	// This would be implemented to collect metrics from Docker/Podman
	// For now, it's a placeholder
	return nil
}

// Made with Bob
