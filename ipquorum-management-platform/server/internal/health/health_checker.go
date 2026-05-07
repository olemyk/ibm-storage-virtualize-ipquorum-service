package health

import (
	"context"
	"sync"
	"time"

	"github.com/olemyk/ipquorum-platform/server/internal/agent"
	"github.com/olemyk/ipquorum-platform/server/internal/executor"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
	"go.uber.org/zap"
)

// HealthChecker performs periodic health checks on instances
type HealthChecker struct {
	db            *storage.Database
	executor      *executor.Executor
	agentRegistry *agent.Registry
	log           *zap.Logger
	interval      time.Duration
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *storage.Database, exec *executor.Executor, agentRegistry *agent.Registry, log *zap.Logger, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		db:            db,
		executor:      exec,
		agentRegistry: agentRegistry,
		log:           log,
		interval:      interval,
		stopCh:        make(chan struct{}),
	}
}

// Start begins periodic health checks
func (h *HealthChecker) Start() {
	h.wg.Add(1)
	go h.run()
	h.log.Info("Health checker started", zap.Duration("interval", h.interval))
}

// Stop stops the health checker
func (h *HealthChecker) Stop() {
	close(h.stopCh)
	h.wg.Wait()
	h.log.Info("Health checker stopped")
}

// run is the main health check loop
func (h *HealthChecker) run() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	// Run initial check immediately
	h.checkAllInstances()

	for {
		select {
		case <-ticker.C:
			h.checkAllInstances()
		case <-h.stopCh:
			return
		}
	}
}

// checkAllInstances checks health of all instances
func (h *HealthChecker) checkAllInstances() {
	instances, err := h.db.ListInstances()
	if err != nil {
		h.log.Error("Failed to list instances for health check", zap.Error(err))
		return
	}

	h.log.Debug("Running health checks", zap.Int("instance_count", len(instances)))

	for _, instance := range instances {
		h.checkInstance(instance)
	}
}

// checkInstance checks health of a single instance
func (h *HealthChecker) checkInstance(instance *storage.Instance) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	h.log.Debug("Checking instance health",
		zap.String("name", instance.Name),
		zap.String("id", instance.ID),
		zap.String("server_id", instance.ServerID),
		zap.Bool("is_remote", instance.ServerID != "" && instance.ServerID != "local"),
	)

	var status string
	var healthy bool
	var networkReachable bool
	var networkMsg string
	var responseTime int64
	var startedAt *time.Time
	startTime := time.Now()

	// Check if remote or local
	h.log.Debug("Evaluating remote check",
		zap.String("server_id", instance.ServerID),
		zap.Bool("not_empty", instance.ServerID != ""),
		zap.Bool("not_local", instance.ServerID != "local"),
		zap.Bool("condition_result", instance.ServerID != "" && instance.ServerID != "local"),
	)
	if instance.ServerID != "" && instance.ServerID != "local" {
		// Remote instance - use agent
		agentClient, err := h.agentRegistry.Get(instance.ServerID)
		if err != nil {
			h.log.Warn("Agent not found for health check",
				zap.String("name", instance.Name),
				zap.String("server_id", instance.ServerID),
				zap.Error(err),
			)
			// Mark as unknown if agent unavailable
			status = "unknown"
			healthy = false
			networkReachable = false
			networkMsg = "Agent unavailable"
			startedAt = nil
		} else {
			// Get status from agent
			h.log.Debug("Calling agent GetStatus",
				zap.String("name", instance.Name),
				zap.String("server_id", instance.ServerID),
			)
			agentStatus, err := agentClient.GetStatus(ctx, instance.Name)
			if err != nil {
				h.log.Warn("Failed to get status from agent",
					zap.String("name", instance.Name),
					zap.String("server_id", instance.ServerID),
					zap.Error(err),
				)
				status = "unknown"
				healthy = false
				networkReachable = false
				networkMsg = "Failed to query agent"
				startedAt = nil
			} else {
				status = agentStatus.Status
				healthy = (status == "running" || status == "active")
				// For remote instances, we assume network is reachable if we got a response
				networkReachable = true
				networkMsg = "Agent reachable"

				// Debug: Log full agent response
				h.log.Debug("Agent status response",
					zap.String("name", instance.Name),
					zap.String("status", agentStatus.Status),
					zap.String("started_at", agentStatus.StartedAt),
					zap.Int64("uptime", agentStatus.Uptime),
				)

				// Parse StartedAt from agent response
				if agentStatus.StartedAt != "" {
					if t, err := time.Parse(time.RFC3339, agentStatus.StartedAt); err == nil {
						startedAt = &t
					} else {
						h.log.Warn("Failed to parse started_at",
							zap.String("name", instance.Name),
							zap.String("started_at", agentStatus.StartedAt),
							zap.Error(err),
						)
					}
				}
			}
		}
		responseTime = time.Since(startTime).Milliseconds()
	} else {
		// Local instance - use executor
		statusResult := h.executor.StatusInstance(ctx, instance.Name)
		status, healthy = executor.ParseInstanceStatus(statusResult.Stdout)

		// Check network connectivity
		networkResult := h.executor.CheckNetwork(ctx, instance.Name)
		networkReachable, networkMsg = executor.ParseNetworkCheck(networkResult.Stdout)
		responseTime = statusResult.Duration.Milliseconds()

		// For local instances, we don't have StartedAt yet (would need to parse from systemctl)
		startedAt = nil
	}

	// Determine overall health
	var health string
	if !healthy {
		health = "unhealthy"
	} else if !networkReachable {
		health = "degraded"
	} else {
		health = "healthy"
	}

	// Update instance in database
	now := time.Now()
	instance.Status = status
	instance.Health = health
	instance.StartedAt = startedAt
	instance.LastHealthCheck = &now
	instance.UpdatedAt = now

	// Calculate uptime for JSON response (not stored in DB)
	if startedAt != nil {
		instance.Uptime = int64(time.Since(*startedAt).Seconds())
	} else {
		instance.Uptime = 0
	}

	if err := h.db.UpdateInstance(instance); err != nil {
		h.log.Error("Failed to update instance health",
			zap.String("name", instance.Name),
			zap.Error(err),
		)
		return
	}

	// Create health check record
	healthCheck := &storage.HealthCheck{
		InstanceID:       instance.ID,
		Status:           status,
		Health:           health,
		NetworkReachable: networkReachable,
		NetworkMessage:   networkMsg,
		ResponseTime:     responseTime,
		CheckedAt:        time.Now(),
	}

	if err := h.db.CreateHealthCheck(healthCheck); err != nil {
		h.log.Error("Failed to create health check record",
			zap.String("name", instance.Name),
			zap.Error(err),
		)
	}

	h.log.Debug("Health check completed",
		zap.String("name", instance.Name),
		zap.String("status", status),
		zap.String("health", health),
		zap.Bool("network_reachable", networkReachable),
	)
}

// CheckInstanceNow performs an immediate health check on a specific instance
func (h *HealthChecker) CheckInstanceNow(instanceID string) (*storage.HealthCheck, error) {
	instance, err := h.db.GetInstance(instanceID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var status string
	var healthy bool
	var networkReachable bool
	var networkMsg string
	var responseTime int64
	startTime := time.Now()

	// Check if remote or local
	if instance.ServerID != "" && instance.ServerID != "local" {
		// Remote instance - use agent
		agentClient, err := h.agentRegistry.Get(instance.ServerID)
		if err != nil {
			status = "unknown"
			healthy = false
			networkReachable = false
			networkMsg = "Agent unavailable"
		} else {
			// Get status from agent
			agentStatus, err := agentClient.GetStatus(ctx, instance.Name)
			if err != nil {
				status = "unknown"
				healthy = false
				networkReachable = false
				networkMsg = "Failed to query agent"
			} else {
				status = agentStatus.Status
				healthy = (status == "running" || status == "active")
				networkReachable = true
				networkMsg = "Agent reachable"
			}
		}
		responseTime = time.Since(startTime).Milliseconds()
	} else {
		// Local instance - use executor
		statusResult := h.executor.StatusInstance(ctx, instance.Name)
		status, healthy = executor.ParseInstanceStatus(statusResult.Stdout)

		// Check network connectivity
		networkResult := h.executor.CheckNetwork(ctx, instance.Name)
		networkReachable, networkMsg = executor.ParseNetworkCheck(networkResult.Stdout)
		responseTime = statusResult.Duration.Milliseconds()
	}

	// Determine overall health
	var health string
	if !healthy {
		health = "unhealthy"
	} else if !networkReachable {
		health = "degraded"
	} else {
		health = "healthy"
	}

	// Update instance
	instance.Status = status
	instance.Health = health
	instance.UpdatedAt = time.Now()

	if err := h.db.UpdateInstance(instance); err != nil {
		return nil, err
	}

	// Create health check record
	healthCheck := &storage.HealthCheck{
		InstanceID:       instance.ID,
		Status:           status,
		Health:           health,
		NetworkReachable: networkReachable,
		NetworkMessage:   networkMsg,
		ResponseTime:     responseTime,
		CheckedAt:        time.Now(),
	}

	if err := h.db.CreateHealthCheck(healthCheck); err != nil {
		return nil, err
	}

	return healthCheck, nil
}

// GetInstanceHealthHistory returns health check history for an instance
func (h *HealthChecker) GetInstanceHealthHistory(instanceID string, limit int) ([]*storage.HealthCheck, error) {
	return h.db.GetHealthChecksByInstance(instanceID, limit)
}
