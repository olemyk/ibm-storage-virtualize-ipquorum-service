package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/olemyk/ipquorum-platform/server/internal/agent"
	"github.com/olemyk/ipquorum-platform/server/internal/executor"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
	"go.uber.org/zap"
)

// HealthChecker performs periodic health checks on instances
type HealthChecker struct {
	db               *storage.Database
	executor         *executor.Executor
	agentRegistry    *agent.Registry
	containerService ContainerService
	log              *zap.Logger
	interval         time.Duration
	stopCh           chan struct{}
	wg               sync.WaitGroup
}

// ContainerService interface for container operations
type ContainerService interface {
	List(ctx context.Context) ([]*storage.ContainerInstance, error)
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *storage.Database, exec *executor.Executor, agentRegistry *agent.Registry, containerService ContainerService, log *zap.Logger, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		db:               db,
		executor:         exec,
		agentRegistry:    agentRegistry,
		containerService: containerService,
		log:              log,
		interval:         interval,
		stopCh:           make(chan struct{}),
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

// checkAllInstances checks health of all instances, container instances, and servers
func (h *HealthChecker) checkAllInstances() {
	// Check servers/agents first
	h.checkAllServers()

	// Check legacy instances
	instances, err := h.db.ListInstances()
	if err != nil {
		h.log.Error("Failed to list instances for health check", zap.Error(err))
	} else {
		h.log.Debug("Running health checks for legacy instances", zap.Int("instance_count", len(instances)))
		for _, instance := range instances {
			h.checkInstance(instance)
		}
	}

	// Check container instances
	if h.containerService != nil {
		ctx := context.Background()
		containers, err := h.containerService.List(ctx)
		if err != nil {
			h.log.Error("Failed to list container instances for health check", zap.Error(err))
		} else {
			h.log.Debug("Running health checks for container instances", zap.Int("container_count", len(containers)))
			for _, container := range containers {
				h.checkContainerInstance(container)
			}
		}
	}
}

// checkAllServers checks health of all registered servers/agents
func (h *HealthChecker) checkAllServers() {
	servers, err := h.db.ListServers()
	if err != nil {
		h.log.Error("Failed to list servers for health check", zap.Error(err))
		return
	}

	h.log.Debug("Running health checks for servers", zap.Int("server_count", len(servers)))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, server := range servers {
		h.checkServer(ctx, server)
	}
}

// checkServer checks health of a single server/agent
func (h *HealthChecker) checkServer(ctx context.Context, server *storage.Server) {
	startTime := time.Now()

	// Try to get agent client
	client, err := h.agentRegistry.Get(server.ID)
	if err != nil {
		// Agent not registered - mark as offline
		h.log.Debug("Agent not found in registry",
			zap.String("server_id", server.ID),
			zap.String("hostname", server.Hostname),
		)
		h.updateServerStatus(server, "offline", nil)
		return
	}

	// Check agent health
	err = client.Health(ctx)
	responseTime := time.Since(startTime)

	if err != nil {
		// Agent health check failed - mark as offline
		h.log.Debug("Agent health check failed",
			zap.String("server_id", server.ID),
			zap.String("hostname", server.Hostname),
			zap.Error(err),
		)
		h.updateServerStatus(server, "offline", nil)
	} else {
		// Agent is healthy - mark as online
		h.log.Debug("Agent health check passed",
			zap.String("server_id", server.ID),
			zap.String("hostname", server.Hostname),
			zap.Duration("response_time", responseTime),
		)
		now := time.Now()
		h.updateServerStatus(server, "online", &now)
	}
}

// updateServerStatus updates server status in database
func (h *HealthChecker) updateServerStatus(server *storage.Server, status string, lastSeen *time.Time) {
	server.Status = status
	server.LastSeen = lastSeen

	if err := h.db.UpdateServer(server); err != nil {
		h.log.Error("Failed to update server status",
			zap.String("server_id", server.ID),
			zap.String("hostname", server.Hostname),
			zap.Error(err),
		)
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
			// Get container status from agent using container name
			// Container name format: ipquorum-{first 8 chars of instance ID}
			containerName := fmt.Sprintf("ipquorum-%s", instance.ID[:8])
			h.log.Debug("Calling agent InspectContainer",
				zap.String("name", instance.Name),
				zap.String("container_name", containerName),
				zap.String("server_id", instance.ServerID),
			)
			containerInfo, err := agentClient.InspectContainer(ctx, containerName)
			if err != nil {
				h.log.Warn("Failed to get container status from agent",
					zap.String("name", instance.Name),
					zap.String("container_name", containerName),
					zap.String("server_id", instance.ServerID),
					zap.Error(err),
				)
				status = "unknown"
				healthy = false
				networkReachable = false
				networkMsg = "Failed to query agent"
				startedAt = nil
			} else {
				// Extract status from container info
				// Podman returns State as a nested object with Status field
				if stateObj, ok := containerInfo["State"].(map[string]interface{}); ok {
					if statusStr, ok := stateObj["Status"].(string); ok {
						status = statusStr
						healthy = (status == "running")
					} else {
						status = "unknown"
						healthy = false
					}
				} else if state, ok := containerInfo["state"].(string); ok {
					// Fallback for flat structure
					status = state
					healthy = (status == "running")
				} else {
					status = "unknown"
					healthy = false
				}
				// For remote instances, we assume network is reachable if we got a response
				networkReachable = true
				networkMsg = "Agent reachable"

				// Debug: Log container info
				h.log.Debug("Container info from agent",
					zap.String("name", instance.Name),
					zap.String("container_name", containerName),
					zap.String("status", status),
					zap.Any("container_info", containerInfo),
				)

				// Try to extract started_at from container info
				// Podman returns StartedAt in the State object
				var startedAtStr string
				if stateObj, ok := containerInfo["State"].(map[string]interface{}); ok {
					if str, ok := stateObj["StartedAt"].(string); ok {
						startedAtStr = str
					}
				}
				// Fallback to flat structure
				if startedAtStr == "" {
					if str, ok := containerInfo["started_at"].(string); ok {
						startedAtStr = str
					}
				}

				if startedAtStr != "" {
					if t, err := time.Parse(time.RFC3339, startedAtStr); err == nil {
						startedAt = &t
					} else {
						h.log.Warn("Failed to parse started_at",
							zap.String("name", instance.Name),
							zap.String("started_at", startedAtStr),
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

// checkContainerInstance checks health of a single container instance
func (h *HealthChecker) checkContainerInstance(container *storage.ContainerInstance) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	startTime := time.Now()
	var status string
	var healthy bool
	var startedAt *time.Time
	responseTime := int64(0)

	// Get agent client for this container
	agentClient, err := h.agentRegistry.Get(container.AgentID)
	if err != nil {
		h.log.Warn("Agent not found for container health check",
			zap.String("container_id", container.ID),
			zap.String("container_name", container.ContainerName),
			zap.String("agent_id", container.AgentID),
			zap.Error(fmt.Errorf("agent not found: %s", container.AgentID)),
		)
		// Mark as unhealthy if agent not found
		status = "error"
		healthy = false
	} else {
		// Inspect container via agent
		containerInfo, err := agentClient.InspectContainer(ctx, container.ContainerID)
		if err != nil {
			h.log.Warn("Failed to get container status from agent",
				zap.String("container_id", container.ID),
				zap.String("container_name", container.ContainerName),
				zap.String("agent_id", container.AgentID),
				zap.Error(err),
			)
			status = "error"
			healthy = false
		} else {
			// Debug: Log the raw container info structure
			h.log.Info("Container inspect response received",
				zap.String("container_name", container.ContainerName),
				zap.Any("container_info", containerInfo),
			)

			// Extract status from container info
			// Podman returns State as a nested object with Status field
			if stateObj, ok := containerInfo["State"].(map[string]interface{}); ok {
				h.log.Info("Found State object",
					zap.String("container_name", container.ContainerName),
					zap.Any("state_obj", stateObj),
				)
				if statusStr, ok := stateObj["Status"].(string); ok {
					status = statusStr
					healthy = (status == "running")
					h.log.Info("Extracted status from State.Status",
						zap.String("container_name", container.ContainerName),
						zap.String("status", status),
					)
				} else {
					h.log.Warn("State.Status is not a string",
						zap.String("container_name", container.ContainerName),
						zap.Any("Status_value", stateObj["Status"]),
					)
				}
			} else if state, ok := containerInfo["state"].(string); ok {
				// Fallback for flat structure
				status = state
				healthy = (status == "running")
				h.log.Info("Extracted status from flat state",
					zap.String("container_name", container.ContainerName),
					zap.String("status", status),
				)
			} else {
				h.log.Warn("Could not extract status - no State object or state field",
					zap.String("container_name", container.ContainerName),
				)
			}

			// Extract started_at from State.StartedAt
			var startedAtStr string
			if stateObj, ok := containerInfo["State"].(map[string]interface{}); ok {
				if str, ok := stateObj["StartedAt"].(string); ok {
					startedAtStr = str
				}
			}
			// Fallback to flat structure
			if startedAtStr == "" {
				if str, ok := containerInfo["started_at"].(string); ok {
					startedAtStr = str
				}
			}

			if startedAtStr != "" {
				if t, err := time.Parse(time.RFC3339, startedAtStr); err == nil {
					startedAt = &t
				} else {
					h.log.Warn("Failed to parse started_at for container",
						zap.String("container_name", container.ContainerName),
						zap.String("started_at", startedAtStr),
						zap.Error(err),
					)
				}
			}
		}
		responseTime = time.Since(startTime).Milliseconds()
	}

	// Update container instance in database
	now := time.Now()
	container.Status = status
	container.StartedAt = startedAt
	container.UpdatedAt = now

	// Debug: Log the status value before database update
	h.log.Info("Updating container instance",
		zap.String("container_name", container.ContainerName),
		zap.String("status", status),
		zap.String("instance_id", container.ID),
	)

	// Update via database (need to add UpdateContainerInstance method)
	if err := h.db.UpdateContainerInstance(container); err != nil {
		h.log.Error("Failed to update container instance health",
			zap.String("container_name", container.ContainerName),
			zap.Error(err),
		)
		return
	}

	h.log.Debug("Container health check completed",
		zap.String("container_name", container.ContainerName),
		zap.String("container_id", container.ContainerID),
		zap.String("status", status),
		zap.Bool("healthy", healthy),
		zap.Int64("response_time_ms", responseTime),
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
