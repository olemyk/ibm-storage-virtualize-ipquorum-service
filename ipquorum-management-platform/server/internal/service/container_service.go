package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/olemyk/ipquorum-platform/server/internal/agent"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
	"go.uber.org/zap"
)

// ContainerService handles container instance operations
type ContainerService struct {
	db            *sql.DB
	agentRegistry *agent.Registry
	logger        *zap.Logger
}

// NewContainerService creates a new container service
func NewContainerService(db *sql.DB, agentRegistry *agent.Registry, logger *zap.Logger) *ContainerService {
	return &ContainerService{
		db:            db,
		agentRegistry: agentRegistry,
		logger:        logger,
	}
}

// Create creates a new container instance
func (s *ContainerService) Create(ctx context.Context, req *storage.CreateContainerInstanceRequest) (*storage.ContainerInstance, error) {
	s.logger.Info("Creating container instance",
		zap.String("instance_id", req.InstanceID),
		zap.String("agent_id", req.AgentID),
		zap.String("image", req.ImageName))

	// Validate instance exists
	var instanceExists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM instances WHERE id = ?)", req.InstanceID).Scan(&instanceExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check instance existence: %w", err)
	}
	if !instanceExists {
		return nil, fmt.Errorf("instance not found: %s", req.InstanceID)
	}

	// Get agent client (skip if registry is nil for testing)
	var agentClient *agent.Client
	if s.agentRegistry != nil {
		var err error
		agentClient, err = s.agentRegistry.Get(req.AgentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get agent: %w", err)
		}
	}

	// Prepare container configuration for agent
	// Runtime will be "docker" by default, can be overridden via config
	runtime := "docker"
	agentConfig := &agent.CreateContainerRequest{
		Name:          fmt.Sprintf("ipquorum-%s", req.InstanceID[:8]),
		Image:         req.ImageName,
		Runtime:       runtime,
		RestartPolicy: req.RestartPolicy,
	}

	if req.ImageTag != "" {
		agentConfig.Image = fmt.Sprintf("%s:%s", req.ImageName, req.ImageTag)
	} else {
		agentConfig.Image = fmt.Sprintf("%s:latest", req.ImageName)
		req.ImageTag = "latest"
	}

	// Add configuration if provided
	if req.Config != nil {
		if len(req.Config.Ports) > 0 {
			agentConfig.Ports = make([]agent.PortMapping, len(req.Config.Ports))
			for i, p := range req.Config.Ports {
				agentConfig.Ports[i] = agent.PortMapping{
					HostPort:      p.HostPort,
					ContainerPort: p.ContainerPort,
					Protocol:      p.Protocol,
				}
			}
		}

		if len(req.Config.Volumes) > 0 {
			agentConfig.Volumes = make([]agent.VolumeMount, len(req.Config.Volumes))
			for i, v := range req.Config.Volumes {
				agentConfig.Volumes[i] = agent.VolumeMount{
					HostPath:      v.Source,
					ContainerPath: v.Destination,
					ReadOnly:      v.ReadOnly,
				}
			}
		}

		if len(req.Config.Environment) > 0 {
			agentConfig.Env = req.Config.Environment
		}

		if len(req.Config.Labels) > 0 {
			agentConfig.Labels = req.Config.Labels
		}

		if req.Config.WorkingDir != "" {
			agentConfig.WorkingDir = req.Config.WorkingDir
		}
	}

	// Create container via agent (skip if no agent client for testing)
	var containerResp map[string]interface{}
	if agentClient != nil {
		var err error
		containerResp, err = agentClient.CreateContainer(ctx, agentConfig)
		if err != nil {
			// Parse error to provide better messages
			errMsg := err.Error()

			// Check for image not found errors
			if strings.Contains(errMsg, "image not known") ||
				strings.Contains(errMsg, "no such image") ||
				strings.Contains(errMsg, "404 Not Found") {
				// Try to list available images to help user
				images, listErr := agentClient.ListImages(ctx)
				if listErr == nil && len(images) > 0 {
					var imageNames []string
					for _, img := range images {
						if repoTags, ok := img["repo_tags"].([]interface{}); ok {
							for _, tag := range repoTags {
								if tagStr, ok := tag.(string); ok && tagStr != "<none>:<none>" {
									imageNames = append(imageNames, tagStr)
								}
							}
						}
					}
					if len(imageNames) > 0 {
						return nil, fmt.Errorf("image '%s' not found. Available images: %s",
							agentConfig.Image, strings.Join(imageNames, ", "))
					}
				}
				return nil, fmt.Errorf("image '%s' not found. Please build or pull the image first", agentConfig.Image)
			}

			return nil, fmt.Errorf("failed to create container via agent: %w", err)
		}
	} else {
		// For testing without agent, create mock response
		containerResp = map[string]interface{}{
			"id":      "test-container-id-" + req.InstanceID,
			"name":    req.ImageName + "-container",
			"runtime": "docker",
		}
	}

	// Extract container ID and name from response
	containerID, _ := containerResp["id"].(string)
	containerName, _ := containerResp["name"].(string)
	runtimeType, _ := containerResp["runtime"].(string)

	if containerID == "" {
		return nil, fmt.Errorf("agent did not return container ID")
	}

	// Default to "podman" if runtime not returned by agent
	if runtimeType == "" {
		runtimeType = "podman"
	}

	// Create database record
	container := &storage.ContainerInstance{
		ID:            uuid.New().String(),
		InstanceID:    req.InstanceID,
		AgentID:       req.AgentID,
		ContainerID:   containerID,
		ContainerName: containerName,
		ImageName:     req.ImageName,
		ImageTag:      req.ImageTag,
		Runtime:       runtimeType,
		Status:        "created",
		State:         "created",
		RestartPolicy: req.RestartPolicy,
		RestartCount:  0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if container.RestartPolicy == "" {
		container.RestartPolicy = "no"
	}

	// Insert into database
	query := `
		INSERT INTO container_instances (
			id, instance_id, agent_id, container_id, container_name,
			image_name, image_tag, runtime, status, state,
			restart_policy, restart_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query,
		container.ID, container.InstanceID, container.AgentID,
		container.ContainerID, container.ContainerName,
		container.ImageName, container.ImageTag, container.Runtime,
		container.Status, container.State, container.RestartPolicy,
		container.RestartCount, container.CreatedAt, container.UpdatedAt,
	)
	if err != nil {
		// Try to cleanup container on agent if DB insert fails
		_ = agentClient.RemoveContainer(ctx, containerID, true)
		return nil, fmt.Errorf("failed to insert container record: %w", err)
	}

	// Store configuration in container_configs table
	if req.Config != nil {
		if err := s.storeContainerConfig(ctx, container.ID, req.Config); err != nil {
			s.logger.Warn("Failed to store container config", zap.Error(err))
		}
	}

	s.logger.Info("Container instance created successfully",
		zap.String("id", container.ID),
		zap.String("container_id", containerID))

	return container, nil
}

// storeContainerConfig stores container configuration in the database
func (s *ContainerService) storeContainerConfig(ctx context.Context, containerID string, config *storage.ContainerConfig) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertConfig := func(configType, key, value, extra string) error {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO container_configs (container_instance_id, config_type, key, value, extra) VALUES (?, ?, ?, ?, ?)",
			containerID, configType, key, value, extra)
		return err
	}

	// Store ports
	for _, port := range config.Ports {
		extra, _ := json.Marshal(map[string]interface{}{"protocol": port.Protocol})
		if err := insertConfig("port", fmt.Sprintf("%d", port.ContainerPort), fmt.Sprintf("%d", port.HostPort), string(extra)); err != nil {
			return err
		}
	}

	// Store volumes
	for _, vol := range config.Volumes {
		extra, _ := json.Marshal(map[string]interface{}{"read_only": vol.ReadOnly})
		if err := insertConfig("volume", vol.Destination, vol.Source, string(extra)); err != nil {
			return err
		}
	}

	// Store environment variables
	for key, value := range config.Environment {
		if err := insertConfig("env", key, value, ""); err != nil {
			return err
		}
	}

	// Store labels
	for key, value := range config.Labels {
		if err := insertConfig("label", key, value, ""); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Start starts a container instance
func (s *ContainerService) Start(ctx context.Context, id string) error {
	s.logger.Info("Starting container instance", zap.String("id", id))

	container, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	agentClient, err := s.agentRegistry.Get(container.AgentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// Start container via agent
	if err := agentClient.StartContainer(ctx, container.ContainerID); err != nil {
		// Update status to error
		s.updateStatus(ctx, id, "error", err.Error())
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Update status in database
	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		"UPDATE container_instances SET status = ?, state = ?, started_at = ?, error = NULL, updated_at = ? WHERE id = ?",
		"running", "running", now, now, id)

	if err != nil {
		return fmt.Errorf("failed to update container status: %w", err)
	}

	s.logger.Info("Container instance started successfully", zap.String("id", id))
	return nil
}

// Stop stops a container instance
func (s *ContainerService) Stop(ctx context.Context, id string, timeout int) error {
	s.logger.Info("Stopping container instance", zap.String("id", id), zap.Int("timeout", timeout))

	container, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	agentClient, err := s.agentRegistry.Get(container.AgentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// Stop container via agent
	if err := agentClient.StopContainer(ctx, container.ContainerID, timeout); err != nil {
		s.updateStatus(ctx, id, "error", err.Error())
		return fmt.Errorf("failed to stop container: %w", err)
	}

	// Update status in database
	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		"UPDATE container_instances SET status = ?, state = ?, finished_at = ?, error = NULL, updated_at = ? WHERE id = ?",
		"stopped", "stopped", now, now, id)

	if err != nil {
		return fmt.Errorf("failed to update container status: %w", err)
	}

	s.logger.Info("Container instance stopped successfully", zap.String("id", id))
	return nil
}

// Restart restarts a container instance
func (s *ContainerService) Restart(ctx context.Context, id string, timeout int) error {
	s.logger.Info("Restarting container instance", zap.String("id", id))

	container, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	agentClient, err := s.agentRegistry.Get(container.AgentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// Restart container via agent
	if err := agentClient.RestartContainer(ctx, container.ContainerID, timeout); err != nil {
		s.updateStatus(ctx, id, "error", err.Error())
		return fmt.Errorf("failed to restart container: %w", err)
	}

	// Increment restart count and update status
	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		"UPDATE container_instances SET status = ?, state = ?, restart_count = restart_count + 1, started_at = ?, error = NULL, updated_at = ? WHERE id = ?",
		"running", "running", now, now, id)

	if err != nil {
		return fmt.Errorf("failed to update container status: %w", err)
	}

	s.logger.Info("Container instance restarted successfully", zap.String("id", id))
	return nil
}

// Remove removes a container instance
func (s *ContainerService) Remove(ctx context.Context, id string, force bool) error {
	s.logger.Info("Removing container instance", zap.String("id", id), zap.Bool("force", force))

	container, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Try to get agent client
	agentClient, err := s.agentRegistry.Get(container.AgentID)
	if err != nil {
		// Agent not found - this is an orphaned container
		s.logger.Warn("Agent not found for container, removing orphaned record",
			zap.String("container_id", id),
			zap.String("agent_id", container.AgentID),
			zap.Error(err))

		// Delete orphaned container record from database
		_, err = s.db.ExecContext(ctx, "DELETE FROM container_instances WHERE id = ?", id)
		if err != nil {
			return fmt.Errorf("failed to delete orphaned container record: %w", err)
		}

		s.logger.Info("Orphaned container record removed successfully", zap.String("id", id))
		return nil
	}

	// Try to remove container via agent
	if err := agentClient.RemoveContainer(ctx, container.ContainerID, force); err != nil {
		s.logger.Warn("Failed to remove container from agent, removing database record anyway",
			zap.String("container_id", id),
			zap.Error(err))
		// Continue to delete from database even if agent removal fails
	}

	// Delete from database
	_, err = s.db.ExecContext(ctx, "DELETE FROM container_instances WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete container record: %w", err)
	}

	s.logger.Info("Container instance removed successfully", zap.String("id", id))
	return nil
}

// GetByID retrieves a container instance by ID
func (s *ContainerService) GetByID(ctx context.Context, id string) (*storage.ContainerInstance, error) {
	container := &storage.ContainerInstance{}

	query := `
		SELECT id, instance_id, agent_id, container_id, container_name,
		       image_name, image_tag, runtime, status, state,
		       restart_policy, restart_count, exit_code, error,
		       created_at, started_at, finished_at, updated_at
		FROM container_instances
		WHERE id = ?
	`

	var exitCode sql.NullInt64
	var errorMsg sql.NullString
	var startedAt, finishedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&container.ID, &container.InstanceID, &container.AgentID,
		&container.ContainerID, &container.ContainerName,
		&container.ImageName, &container.ImageTag, &container.Runtime,
		&container.Status, &container.State, &container.RestartPolicy,
		&container.RestartCount, &exitCode, &errorMsg,
		&container.CreatedAt, &startedAt, &finishedAt, &container.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("container instance not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query container instance: %w", err)
	}

	if exitCode.Valid {
		code := int(exitCode.Int64)
		container.ExitCode = &code
	}
	if errorMsg.Valid {
		container.Error = errorMsg.String
	}
	if startedAt.Valid {
		container.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		container.FinishedAt = &finishedAt.Time
	}

	return container, nil
}

// GetByInstanceID retrieves container instance by instance ID
func (s *ContainerService) GetByInstanceID(ctx context.Context, instanceID string) (*storage.ContainerInstance, error) {
	container := &storage.ContainerInstance{}

	query := `
		SELECT id, instance_id, agent_id, container_id, container_name,
		       image_name, image_tag, runtime, status, state,
		       restart_policy, restart_count, exit_code, error,
		       created_at, started_at, finished_at, updated_at
		FROM container_instances
		WHERE instance_id = ?
	`

	var exitCode sql.NullInt64
	var errorMsg sql.NullString
	var startedAt, finishedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, instanceID).Scan(
		&container.ID, &container.InstanceID, &container.AgentID,
		&container.ContainerID, &container.ContainerName,
		&container.ImageName, &container.ImageTag, &container.Runtime,
		&container.Status, &container.State, &container.RestartPolicy,
		&container.RestartCount, &exitCode, &errorMsg,
		&container.CreatedAt, &startedAt, &finishedAt, &container.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("container instance not found for instance: %s", instanceID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query container instance: %w", err)
	}

	if exitCode.Valid {
		code := int(exitCode.Int64)
		container.ExitCode = &code
	}
	if errorMsg.Valid {
		container.Error = errorMsg.String
	}
	if startedAt.Valid {
		container.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		container.FinishedAt = &finishedAt.Time
	}

	return container, nil
}

// List retrieves all container instances
func (s *ContainerService) List(ctx context.Context) ([]*storage.ContainerInstance, error) {
	query := `
		SELECT id, instance_id, agent_id, container_id, container_name,
		       image_name, image_tag, runtime, status, state,
		       restart_policy, restart_count, exit_code, error,
		       created_at, started_at, finished_at, updated_at
		FROM container_instances
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query container instances: %w", err)
	}
	defer rows.Close()

	// Initialize as empty slice to return [] instead of null when empty
	containers := make([]*storage.ContainerInstance, 0)
	for rows.Next() {
		container := &storage.ContainerInstance{}
		var exitCode sql.NullInt64
		var errorMsg sql.NullString
		var startedAt, finishedAt sql.NullTime

		err := rows.Scan(
			&container.ID, &container.InstanceID, &container.AgentID,
			&container.ContainerID, &container.ContainerName,
			&container.ImageName, &container.ImageTag, &container.Runtime,
			&container.Status, &container.State, &container.RestartPolicy,
			&container.RestartCount, &exitCode, &errorMsg,
			&container.CreatedAt, &startedAt, &finishedAt, &container.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan container instance: %w", err)
		}

		if exitCode.Valid {
			code := int(exitCode.Int64)
			container.ExitCode = &code
		}
		if errorMsg.Valid {
			container.Error = errorMsg.String
		}
		if startedAt.Valid {
			container.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			container.FinishedAt = &finishedAt.Time
		}

		containers = append(containers, container)
	}

	return containers, nil
}

// GetLogs retrieves container logs
func (s *ContainerService) GetLogs(ctx context.Context, id string, tail int) (*storage.ContainerLogs, error) {
	container, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	agentClient, err := s.agentRegistry.Get(container.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Get logs from agent (returns plain text)
	logsText, err := agentClient.GetContainerLogs(ctx, container.ContainerID, tail)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	// Split logs into lines
	logLines := []string{}
	if logsText != "" {
		logLines = strings.Split(strings.TrimSpace(logsText), "\n")
	}

	logs := &storage.ContainerLogs{
		ContainerID: container.ContainerID,
		Timestamp:   time.Now(),
		Logs:        logLines,
	}

	return logs, nil
}

// GetStats retrieves container statistics
func (s *ContainerService) GetStats(ctx context.Context, id string) (*storage.ContainerStats, error) {
	container, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	agentClient, err := s.agentRegistry.Get(container.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// Get stats from agent
	statsResp, err := agentClient.GetContainerStats(ctx, container.ContainerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}

	stats := &storage.ContainerStats{
		ContainerID: container.ContainerID,
		Timestamp:   time.Now(),
	}

	// Parse stats response
	if cpu, ok := statsResp["cpu_percent"].(float64); ok {
		stats.CPUPercent = cpu
	}
	if mem, ok := statsResp["memory_usage"].(float64); ok {
		stats.MemoryUsage = int64(mem)
	}
	if limit, ok := statsResp["memory_limit"].(float64); ok {
		stats.MemoryLimit = int64(limit)
	}
	if pct, ok := statsResp["memory_percent"].(float64); ok {
		stats.MemoryPercent = pct
	}
	if rx, ok := statsResp["network_rx"].(float64); ok {
		stats.NetworkRxBytes = int64(rx)
	}
	if tx, ok := statsResp["network_tx"].(float64); ok {
		stats.NetworkTxBytes = int64(tx)
	}
	if read, ok := statsResp["block_read"].(float64); ok {
		stats.BlockRead = int64(read)
	}
	if write, ok := statsResp["block_write"].(float64); ok {
		stats.BlockWrite = int64(write)
	}
	if pids, ok := statsResp["pids"].(float64); ok {
		stats.PIDs = int(pids)
	}

	// Store stats in database for historical tracking
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO container_stats (
			container_instance_id, cpu_percent, memory_usage, memory_limit,
			memory_percent, network_rx_bytes, network_tx_bytes,
			block_read, block_write, pids, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, stats.CPUPercent, stats.MemoryUsage, stats.MemoryLimit,
		stats.MemoryPercent, stats.NetworkRxBytes, stats.NetworkTxBytes,
		stats.BlockRead, stats.BlockWrite, stats.PIDs, stats.Timestamp)

	if err != nil {
		s.logger.Warn("Failed to store container stats", zap.Error(err))
	}

	return stats, nil
}

// updateStatus updates container status and error message
func (s *ContainerService) updateStatus(ctx context.Context, id, status, errorMsg string) {
	_, err := s.db.ExecContext(ctx,
		"UPDATE container_instances SET status = ?, error = ?, updated_at = ? WHERE id = ?",
		status, errorMsg, time.Now(), id)
	if err != nil {
		s.logger.Error("Failed to update container status", zap.Error(err))
	}
}

// Made with Bob
