package service

import (
	"context"
	"fmt"
	"time"

	"github.com/olemyk/ipquorum-platform/server/internal/agent"
	"github.com/olemyk/ipquorum-platform/server/internal/executor"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
	"go.uber.org/zap"
)

// InstanceService handles instance management operations
type InstanceService struct {
	db            *storage.Database
	executor      *executor.Executor
	agentRegistry *agent.Registry
	log           *zap.Logger
}

// NewInstanceService creates a new instance service
func NewInstanceService(db *storage.Database, exec *executor.Executor, agentRegistry *agent.Registry, log *zap.Logger) *InstanceService {
	return &InstanceService{
		db:            db,
		executor:      exec,
		agentRegistry: agentRegistry,
		log:           log,
	}
}

// CreateInstance creates a new IP Quorum instance
func (s *InstanceService) CreateInstance(ctx context.Context, instance *storage.Instance) error {
	s.log.Info("Creating instance",
		zap.String("name", instance.Name),
		zap.String("server_id", instance.ServerID),
		zap.String("api_endpoint", instance.APIEndpoint),
	)

	// Validate instance
	if err := s.validateInstance(instance); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if instance already exists
	existing, err := s.db.GetInstanceByName(instance.Name)
	if err == nil && existing != nil {
		return fmt.Errorf("instance with name '%s' already exists", instance.Name)
	}

	// Check if this is a remote or local instance
	if instance.ServerID != "" && instance.ServerID != "local" {
		// Remote instance - use agent
		s.log.Info("Creating instance on remote agent", zap.String("server_id", instance.ServerID))

		agentClient, err := s.agentRegistry.Get(instance.ServerID)
		if err != nil {
			return fmt.Errorf("agent not found for server %s: %w", instance.ServerID, err)
		}

		// Create instance request for agent
		req := &agent.CreateInstanceRequest{
			Name:              instance.Name,
			APIEndpoint:       instance.APIEndpoint,
			Username:          instance.Username,
			Password:          instance.Password,
			EnableDownload:    instance.EnableDownload,
			EnableMkquorumapp: instance.EnableMkquorumapp,
			Partnersystem:     instance.Partnersystem,
			IPQuorumName:      instance.IPQuorumName,
			IP6:               instance.IP6,
			PartnerIP6:        instance.PartnerIP6,
			NoMetadata:        instance.NoMetadata,
		}

		if err := agentClient.CreateInstance(ctx, req); err != nil {
			s.log.Error("Failed to create instance on agent",
				zap.String("name", instance.Name),
				zap.String("server_id", instance.ServerID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to create instance on agent: %w", err)
		}
	} else {
		// Local instance - use bash script
		s.log.Info("Creating instance locally")

		params := executor.CreateInstanceParams{
			InstanceName:      instance.Name,
			APIEndpoint:       instance.APIEndpoint,
			Username:          instance.Username,
			Password:          instance.Password,
			EnableDownload:    instance.EnableDownload,
			EnableMkquorumapp: instance.EnableMkquorumapp,
			Partnersystem:     instance.Partnersystem,
			IPQuorumName:      instance.IPQuorumName,
			IP6:               instance.IP6,
			PartnerIP6:        instance.PartnerIP6,
			NoMetadata:        instance.NoMetadata,
		}
		result := s.executor.CreateInstanceWithParams(ctx, params)

		if result.Error != nil || result.ExitCode != 0 {
			s.log.Error("Failed to create instance via script",
				zap.String("name", instance.Name),
				zap.Int("exit_code", result.ExitCode),
				zap.String("stderr", result.Stderr),
				zap.Error(result.Error),
			)
			return fmt.Errorf("script execution failed (exit code %d): %s", result.ExitCode, result.Stderr)
		}
	}

	// Set initial status
	instance.Status = "stopped"
	instance.Health = "unknown"
	instance.CreatedAt = time.Now()
	instance.UpdatedAt = time.Now()

	// Save to database
	if err := s.db.CreateInstance(instance); err != nil {
		s.log.Error("Failed to save instance to database",
			zap.String("name", instance.Name),
			zap.Error(err),
		)
		// Try to cleanup the created instance
		if instance.ServerID == "" || instance.ServerID == "local" {
			s.executor.DeleteInstance(ctx, instance.Name, true)
		} else {
			if agentClient, err := s.agentRegistry.Get(instance.ServerID); err == nil {
				agentClient.DeleteInstance(ctx, instance.Name)
			}
		}
		return fmt.Errorf("failed to save instance: %w", err)
	}

	s.log.Info("Instance created successfully",
		zap.String("name", instance.Name),
		zap.String("id", instance.ID),
		zap.String("server_id", instance.ServerID),
	)

	return nil
}

// StartInstance starts an IP Quorum instance
func (s *InstanceService) StartInstance(ctx context.Context, instanceID string) error {
	instance, err := s.db.GetInstance(instanceID)
	if err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	s.log.Info("Starting instance",
		zap.String("name", instance.Name),
		zap.String("id", instance.ID),
		zap.String("server_id", instance.ServerID),
	)

	// Check if remote or local
	if instance.ServerID != "" && instance.ServerID != "local" {
		// Remote instance - use agent
		agentClient, err := s.agentRegistry.Get(instance.ServerID)
		if err != nil {
			return fmt.Errorf("agent not found for server %s: %w", instance.ServerID, err)
		}

		if err := agentClient.StartInstance(ctx, instance.Name); err != nil {
			s.log.Error("Failed to start instance on agent",
				zap.String("name", instance.Name),
				zap.Error(err),
			)
			return fmt.Errorf("failed to start instance on agent: %w", err)
		}

		s.log.Info("Instance started successfully on remote agent",
			zap.String("name", instance.Name),
			zap.String("server_id", instance.ServerID),
		)

		// Update status in database
		instance.Status = "running"
		instance.UpdatedAt = time.Now()
		if err := s.db.UpdateInstance(instance); err != nil {
			s.log.Error("Failed to update instance status",
				zap.String("name", instance.Name),
				zap.Error(err),
			)
		}

		return nil
	}

	// Local instance - use executor
	result := s.executor.StartInstance(ctx, instance.Name)
	if result.Error != nil || result.ExitCode != 0 {
		s.log.Error("Failed to start instance",
			zap.String("name", instance.Name),
			zap.Int("exit_code", result.ExitCode),
			zap.String("stderr", result.Stderr),
			zap.Error(result.Error),
		)
		return fmt.Errorf("failed to start instance: %s", result.Stderr)
	}

	// Update status in database
	instance.Status = "running"
	instance.UpdatedAt = time.Now()
	if err := s.db.UpdateInstance(instance); err != nil {
		s.log.Error("Failed to update instance status",
			zap.String("name", instance.Name),
			zap.Error(err),
		)
	}

	s.log.Info("Instance started successfully",
		zap.String("name", instance.Name),
	)

	return nil
}

// StopInstance stops an IP Quorum instance
func (s *InstanceService) StopInstance(ctx context.Context, instanceID string) error {
	instance, err := s.db.GetInstance(instanceID)
	if err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	s.log.Info("Stopping instance",
		zap.String("name", instance.Name),
		zap.String("id", instance.ID),
		zap.String("server_id", instance.ServerID),
	)

	// Check if remote or local
	if instance.ServerID != "" && instance.ServerID != "local" {
		// Remote instance - use agent
		agentClient, err := s.agentRegistry.Get(instance.ServerID)
		if err != nil {
			return fmt.Errorf("agent not found for server %s: %w", instance.ServerID, err)
		}

		if err := agentClient.StopInstance(ctx, instance.Name); err != nil {
			s.log.Error("Failed to stop instance on agent",
				zap.String("name", instance.Name),
				zap.Error(err),
			)
			return fmt.Errorf("failed to stop instance on agent: %w", err)
		}

		s.log.Info("Instance stopped successfully on remote agent",
			zap.String("name", instance.Name),
			zap.String("server_id", instance.ServerID),
		)

		// Update status in database
		instance.Status = "stopped"
		instance.UpdatedAt = time.Now()
		if err := s.db.UpdateInstance(instance); err != nil {
			s.log.Error("Failed to update instance status",
				zap.String("name", instance.Name),
				zap.Error(err),
			)
		}

		return nil
	}

	// Local instance - use executor
	result := s.executor.StopInstance(ctx, instance.Name)
	if result.Error != nil || result.ExitCode != 0 {
		s.log.Error("Failed to stop instance",
			zap.String("name", instance.Name),
			zap.Int("exit_code", result.ExitCode),
			zap.String("stderr", result.Stderr),
			zap.Error(result.Error),
		)
		return fmt.Errorf("failed to stop instance: %s", result.Stderr)
	}

	// Update status in database
	instance.Status = "stopped"
	instance.UpdatedAt = time.Now()
	if err := s.db.UpdateInstance(instance); err != nil {
		s.log.Error("Failed to update instance status",
			zap.String("name", instance.Name),
			zap.Error(err),
		)
	}

	s.log.Info("Instance stopped successfully",
		zap.String("name", instance.Name),
	)

	return nil
}

// RestartInstance restarts an IP Quorum instance
func (s *InstanceService) RestartInstance(ctx context.Context, instanceID string) error {
	instance, err := s.db.GetInstance(instanceID)
	if err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	s.log.Info("Restarting instance",
		zap.String("name", instance.Name),
		zap.String("id", instance.ID),
		zap.String("server_id", instance.ServerID),
	)

	// Check if remote or local
	if instance.ServerID != "" && instance.ServerID != "local" {
		// Remote instance - use agent
		agentClient, err := s.agentRegistry.Get(instance.ServerID)
		if err != nil {
			return fmt.Errorf("agent not found for server %s: %w", instance.ServerID, err)
		}

		if err := agentClient.RestartInstance(ctx, instance.Name); err != nil {
			s.log.Error("Failed to restart instance on agent",
				zap.String("name", instance.Name),
				zap.Error(err),
			)
			return fmt.Errorf("failed to restart instance on agent: %w", err)
		}
	} else {
		// Local instance - use bash script
		result := s.executor.RestartInstance(ctx, instance.Name)
		if result.Error != nil || result.ExitCode != 0 {
			s.log.Error("Failed to restart instance",
				zap.String("name", instance.Name),
				zap.Int("exit_code", result.ExitCode),
				zap.String("stderr", result.Stderr),
				zap.Error(result.Error),
			)
			return fmt.Errorf("failed to restart instance: %s", result.Stderr)
		}
	}

	// Update status in database
	instance.Status = "running"
	instance.UpdatedAt = time.Now()
	if err := s.db.UpdateInstance(instance); err != nil {
		s.log.Error("Failed to update instance status",
			zap.String("name", instance.Name),
			zap.Error(err),
		)
	}

	s.log.Info("Instance restarted successfully",
		zap.String("name", instance.Name),
	)

	return nil
}

// DeleteInstance deletes an IP Quorum instance
func (s *InstanceService) DeleteInstance(ctx context.Context, instanceID string, force bool) error {
	instance, err := s.db.GetInstance(instanceID)
	if err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	s.log.Info("Deleting instance",
		zap.String("name", instance.Name),
		zap.String("id", instance.ID),
		zap.Bool("force", force),
	)

	// Stop instance first if running
	if instance.Status == "running" && !force {
		if err := s.StopInstance(ctx, instanceID); err != nil {
			return fmt.Errorf("failed to stop instance before deletion: %w", err)
		}
	}

	// Check if this is a remote instance
	if instance.ServerID != "" && instance.ServerID != "local" {
		// Remote instance: use agent API
		agentClient, err := s.agentRegistry.Get(instance.ServerID)
		if err != nil {
			s.log.Error("Failed to get agent client",
				zap.String("server_id", instance.ServerID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to get agent client: %w", err)
		}

		s.log.Info("Deleting remote instance via agent",
			zap.String("name", instance.Name),
			zap.String("server_id", instance.ServerID),
		)

		if err := agentClient.DeleteInstance(ctx, instance.Name); err != nil {
			s.log.Error("Failed to delete remote instance",
				zap.String("name", instance.Name),
				zap.String("server_id", instance.ServerID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to delete remote instance: %w", err)
		}
	} else {
		// Local instance: execute bash script
		result := s.executor.DeleteInstance(ctx, instance.Name, force)
		if result.Error != nil || result.ExitCode != 0 {
			s.log.Error("Failed to delete instance via script",
				zap.String("name", instance.Name),
				zap.Int("exit_code", result.ExitCode),
				zap.String("stderr", result.Stderr),
				zap.Error(result.Error),
			)
			return fmt.Errorf("script execution failed: %s", result.Stderr)
		}
	}

	// Delete from database
	if err := s.db.DeleteInstance(instanceID); err != nil {
		s.log.Error("Failed to delete instance from database",
			zap.String("name", instance.Name),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete from database: %w", err)
	}

	s.log.Info("Instance deleted successfully",
		zap.String("name", instance.Name),
	)

	return nil
}

// GetInstanceStatus gets the current status of an instance
func (s *InstanceService) GetInstanceStatus(ctx context.Context, instanceID string) (*storage.Instance, error) {
	instance, err := s.db.GetInstance(instanceID)
	if err != nil {
		return nil, fmt.Errorf("instance not found: %w", err)
	}

	// Check if this is a remote instance
	if instance.ServerID != "" && instance.ServerID != "local" {
		// Remote instance: get status from agent
		agentClient, err := s.agentRegistry.Get(instance.ServerID)
		if err != nil {
			s.log.Error("Failed to get agent client",
				zap.String("server_id", instance.ServerID),
				zap.Error(err),
			)
			instance.Status = "unknown"
			instance.Health = "unknown"
		} else {
			// Get status from agent
			agentStatus, err := agentClient.GetStatus(ctx, instance.Name)
			if err != nil {
				s.log.Error("Failed to get remote instance status",
					zap.String("name", instance.Name),
					zap.String("server_id", instance.ServerID),
					zap.Error(err),
				)
				instance.Status = "unknown"
				instance.Health = "unknown"
			} else {
				instance.Status = agentStatus.Status
				instance.Health = agentStatus.Health
			}
		}
	} else {
		// Local instance: get status from systemd via bash script
		result := s.executor.StatusInstance(ctx, instance.Name)
		if result.Error == nil && result.ExitCode == 0 {
			status, healthy := executor.ParseInstanceStatus(result.Stdout)
			instance.Status = status
			if healthy {
				instance.Health = "healthy"
			} else {
				instance.Health = "unhealthy"
			}
		} else {
			instance.Status = "unknown"
			instance.Health = "unknown"
		}
	}

	// Update database
	instance.UpdatedAt = time.Now()
	if err := s.db.UpdateInstance(instance); err != nil {
		s.log.Error("Failed to update instance status",
			zap.String("name", instance.Name),
			zap.Error(err),
		)
	}

	return instance, nil
}

// CheckInstanceNetwork checks network connectivity for an instance
func (s *InstanceService) CheckInstanceNetwork(ctx context.Context, instanceID string) (bool, string, error) {
	instance, err := s.db.GetInstance(instanceID)
	if err != nil {
		return false, "", fmt.Errorf("instance not found: %w", err)
	}

	result := s.executor.CheckNetwork(ctx, instance.Name)
	if result.Error != nil {
		return false, result.Stderr, result.Error
	}

	reachable, message := executor.ParseNetworkCheck(result.Stdout)
	return reachable, message, nil
}

// ValidateInstanceConfig validates the configuration of an instance
func (s *InstanceService) ValidateInstanceConfig(ctx context.Context, instanceID string) (bool, string, error) {
	instance, err := s.db.GetInstance(instanceID)
	if err != nil {
		return false, "", fmt.Errorf("instance not found: %w", err)
	}

	result := s.executor.ValidateConfig(ctx, instance.Name)
	if result.Error != nil || result.ExitCode != 0 {
		return false, result.Stderr, fmt.Errorf("validation failed: %s", result.Stderr)
	}

	return true, result.Stdout, nil
}

// validateInstance validates instance configuration
func (s *InstanceService) validateInstance(instance *storage.Instance) error {
	if instance.Name == "" {
		return fmt.Errorf("instance name is required")
	}
	if instance.APIEndpoint == "" {
		return fmt.Errorf("API endpoint is required")
	}
	if instance.Username == "" {
		return fmt.Errorf("username is required")
	}
	if instance.Password == "" {
		return fmt.Errorf("password is required")
	}
	if instance.EnableMkquorumapp && instance.Partnersystem == "" {
		return fmt.Errorf("partnersystem is required when mkquorumapp is enabled")
	}
	return nil
}

// ListInstances returns all instances
func (s *InstanceService) ListInstances() ([]*storage.Instance, error) {
	return s.db.ListInstances()
}

// GetInstance returns a single instance by ID
func (s *InstanceService) GetInstance(instanceID string) (*storage.Instance, error) {
	return s.db.GetInstance(instanceID)
}
