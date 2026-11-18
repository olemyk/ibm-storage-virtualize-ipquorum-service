package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
)

// IPQuorumService handles IPQuorum container operations
type IPQuorumService struct {
	db               *storage.Database
	containerService *ContainerService
}

// NewIPQuorumService creates a new IPQuorum service
func NewIPQuorumService(db *storage.Database, cs *ContainerService) *IPQuorumService {
	return &IPQuorumService{
		db:               db,
		containerService: cs,
	}
}

// CreateIPQuorumContainer creates a container configured for IPQuorum
func (s *IPQuorumService) CreateIPQuorumContainer(
	ctx context.Context,
	req *storage.CreateIPQuorumContainerRequest,
) (*storage.IPQuorumContainerResponse, error) {

	// Validate configuration
	if err := s.validateIPQuorumConfig(req.IPQuorumConfig); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Set defaults
	if req.ImageName == "" {
		req.ImageName = "localhost/ipquorum-runner"
	}
	if req.ImageTag == "" {
		req.ImageTag = "latest"
	}
	if req.RestartPolicy == "" {
		req.RestartPolicy = "unless-stopped"
	}
	if req.IPQuorumConfig.Port == 0 {
		req.IPQuorumConfig.Port = 1260
	}
	if req.IPQuorumConfig.JavaOpts == "" {
		req.IPQuorumConfig.JavaOpts = "-Xmx512m -Xms256m"
	}

	// Create Instance record first
	instance := &storage.Instance{
		ID:                uuid.New().String(),
		Name:              req.Name,
		ServerID:          req.AgentID, // Using AgentID as ServerID for now
		DeploymentType:    storage.DeploymentTypeContainer,
		APIEndpoint:       req.IPQuorumConfig.APIEndpoint,
		Username:          req.IPQuorumConfig.Username,
		StorageSystem:     req.IPQuorumConfig.PartnerSystem,
		Status:            "creating",
		Health:            "unknown",
		EnableDownload:    req.IPQuorumConfig.DownloadJAR,
		EnableMkquorumapp: req.IPQuorumConfig.CreateQuorum,
		Partnersystem:     req.IPQuorumConfig.PartnerSystem,
		IP6:               req.IPQuorumConfig.IP6,
		PartnerIP6:        req.IPQuorumConfig.PartnerIP6,
		NoMetadata:        req.IPQuorumConfig.NoMetadata,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Store instance
	if err := s.db.CreateInstance(instance); err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Build container configuration
	containerConfig := s.buildContainerConfig(req.IPQuorumConfig)

	// Create container request
	containerReq := &storage.CreateContainerInstanceRequest{
		InstanceID:    instance.ID,
		AgentID:       req.AgentID,
		ImageName:     req.ImageName,
		ImageTag:      req.ImageTag,
		RestartPolicy: req.RestartPolicy,
		Config:        containerConfig,
	}

	// Create container via container service
	container, err := s.containerService.Create(ctx, containerReq)
	if err != nil {
		// Update instance status to error
		instance.Status = "error"
		instance.Health = "unhealthy"
		instance.UpdatedAt = time.Now()
		s.db.UpdateInstance(instance)
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Update instance with container info
	instance.Status = container.Status
	instance.IPQuorumName = container.ContainerName
	instance.UpdatedAt = time.Now()
	if err := s.db.UpdateInstance(instance); err != nil {
		return nil, fmt.Errorf("failed to update instance: %w", err)
	}

	// Return response
	return &storage.IPQuorumContainerResponse{
		ContainerInstance: container,
		Instance:          instance,
		IPQuorumConfig:    req.IPQuorumConfig,
	}, nil
}

// buildContainerConfig builds container configuration from IPQuorum config
func (s *IPQuorumService) buildContainerConfig(config *storage.IPQuorumConfig) *storage.ContainerConfig {
	// Build environment variables
	env := map[string]string{
		"API_ENDPOINT":        config.APIEndpoint,
		"VIRTUALIZE_USERNAME": config.Username,
		"VIRTUALIZE_PASSWORD": config.Password, // TODO: Use secret management
		"PARTNERSYSTEM":       config.PartnerSystem,
		"IP6":                 fmt.Sprintf("%t", config.IP6),
		"PARTNERIP6":          fmt.Sprintf("%t", config.PartnerIP6),
		"NOMETADATA":          fmt.Sprintf("%t", config.NoMetadata),
		"INSECURE":            fmt.Sprintf("%t", config.Insecure),
		"DOWNLOAD_JAR":        fmt.Sprintf("%t", config.DownloadJAR),
		"CREATE_QUORUM_APP":   fmt.Sprintf("%t", config.CreateQuorum),
		"JAVA_OPTS":           config.JavaOpts,
	}

	if config.ForceDownload {
		env["FORCE_DOWNLOAD"] = "true"
	}

	// Build port mappings
	ports := []storage.PortMapping{
		{
			HostPort:      config.Port,
			ContainerPort: 1260,
			Protocol:      "tcp",
		},
	}

	// Build volume mounts
	// TODO: Implement proper volume management (create named volumes or use bind mounts)
	// For now, volumes are disabled to avoid "no such file or directory" errors
	volumes := []storage.VolumeMount{}

	// volumes := []storage.VolumeMount{
	// 	{
	// 		Source:      fmt.Sprintf("ipquorum-jar-%s", uuid.New().String()[:8]),
	// 		Destination: "/opt/ipquorum/jar",
	// 		ReadOnly:    false,
	// 	},
	// 	{
	// 		Source:      fmt.Sprintf("ipquorum-logs-%s", uuid.New().String()[:8]),
	// 		Destination: "/opt/ipquorum/logs",
	// 		ReadOnly:    false,
	// 	},
	// }

	// Build labels
	labels := map[string]string{
		"app":                    "ipquorum",
		"managed-by":             "ipquorum-management-platform",
		"ipquorum.api-endpoint":  config.APIEndpoint,
		"ipquorum.partner":       config.PartnerSystem,
		"ipquorum.create-quorum": fmt.Sprintf("%t", config.CreateQuorum),
	}

	return &storage.ContainerConfig{
		Ports:       ports,
		Volumes:     volumes,
		Environment: env,
		Labels:      labels,
		WorkingDir:  "/opt/ipquorum", // Set working directory for Podman
		Resources: &storage.ResourceConstraints{
			CPULimit:    2.0,        // 2 CPU cores
			MemoryLimit: 1073741824, // 1GB
		},
	}
}

// validateIPQuorumConfig validates IPQuorum configuration
func (s *IPQuorumService) validateIPQuorumConfig(config *storage.IPQuorumConfig) error {
	if config == nil {
		return fmt.Errorf("ipquorum_config is required")
	}

	if config.APIEndpoint == "" {
		return fmt.Errorf("api_endpoint is required")
	}

	if config.Username == "" {
		return fmt.Errorf("username is required")
	}

	if config.Password == "" {
		return fmt.Errorf("password is required")
	}

	if config.CreateQuorum && config.PartnerSystem == "" {
		return fmt.Errorf("partner_system is required when create_quorum is true")
	}

	if config.Port < 0 || config.Port > 65535 {
		return fmt.Errorf("port must be between 0 and 65535")
	}

	return nil
}

// GetIPQuorumContainer retrieves an IPQuorum container with full details
func (s *IPQuorumService) GetIPQuorumContainer(
	ctx context.Context,
	containerID string,
) (*storage.IPQuorumContainerResponse, error) {

	// Get container
	container, err := s.containerService.GetByID(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	// Get instance
	instance, err := s.db.GetInstance(container.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	// Reconstruct IPQuorum config from instance and container
	config := &storage.IPQuorumConfig{
		APIEndpoint:   instance.APIEndpoint,
		Username:      instance.Username,
		PartnerSystem: instance.Partnersystem,
		IP6:           instance.IP6,
		PartnerIP6:    instance.PartnerIP6,
		NoMetadata:    instance.NoMetadata,
		DownloadJAR:   instance.EnableDownload,
		CreateQuorum:  instance.EnableMkquorumapp,
		// Port and other details from container environment
	}

	return &storage.IPQuorumContainerResponse{
		ContainerInstance: container,
		Instance:          instance,
		IPQuorumConfig:    config,
	}, nil
}

// ListIPQuorumContainers lists all IPQuorum containers
func (s *IPQuorumService) ListIPQuorumContainers(
	ctx context.Context,
	agentID string,
) ([]*storage.IPQuorumContainerResponse, error) {

	// Get all containers
	containers, err := s.containerService.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// Filter and enrich with IPQuorum details
	// Initialize as empty slice to return [] instead of null when empty
	ipquorumContainers := make([]*storage.IPQuorumContainerResponse, 0)

	for _, container := range containers {
		// Filter by agent if specified
		if agentID != "" && container.AgentID != agentID {
			continue
		}

		// Get instance
		instance, err := s.db.GetInstance(container.InstanceID)
		if err != nil {
			continue // Skip if instance not found
		}

		// Only include container deployment types
		if instance.DeploymentType != storage.DeploymentTypeContainer {
			continue
		}

		// Reconstruct config
		config := &storage.IPQuorumConfig{
			APIEndpoint:   instance.APIEndpoint,
			Username:      instance.Username,
			PartnerSystem: instance.Partnersystem,
			IP6:           instance.IP6,
			PartnerIP6:    instance.PartnerIP6,
			NoMetadata:    instance.NoMetadata,
			DownloadJAR:   instance.EnableDownload,
			CreateQuorum:  instance.EnableMkquorumapp,
		}

		ipquorumContainers = append(ipquorumContainers, &storage.IPQuorumContainerResponse{
			ContainerInstance: container,
			Instance:          instance,
			IPQuorumConfig:    config,
		})
	}

	return ipquorumContainers, nil
}

// DeleteIPQuorumContainer deletes an IPQuorum container and its instance
func (s *IPQuorumService) DeleteIPQuorumContainer(
	ctx context.Context,
	containerID string,
) error {

	// Get container to find instance ID
	container, err := s.containerService.GetByID(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container: %w", err)
	}

	// Delete container
	if err := s.containerService.Remove(ctx, containerID, true); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	// Delete instance
	if err := s.db.DeleteInstance(container.InstanceID); err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	return nil
}

// StartIPQuorumContainer starts an IPQuorum container
func (s *IPQuorumService) StartIPQuorumContainer(
	ctx context.Context,
	containerID string,
) error {
	return s.containerService.Start(ctx, containerID)
}

// StopIPQuorumContainer stops an IPQuorum container
func (s *IPQuorumService) StopIPQuorumContainer(
	ctx context.Context,
	containerID string,
) error {
	return s.containerService.Stop(ctx, containerID, 30)
}

// RestartIPQuorumContainer restarts an IPQuorum container
func (s *IPQuorumService) RestartIPQuorumContainer(
	ctx context.Context,
	containerID string,
) error {
	return s.containerService.Restart(ctx, containerID, 30)
}

// GetIPQuorumContainerLogs retrieves logs from an IPQuorum container
func (s *IPQuorumService) GetIPQuorumContainerLogs(
	ctx context.Context,
	containerID string,
	tail int,
) (*storage.ContainerLogs, error) {
	return s.containerService.GetLogs(ctx, containerID, tail)
}

// GetIPQuorumContainerStats retrieves stats from an IPQuorum container
func (s *IPQuorumService) GetIPQuorumContainerStats(
	ctx context.Context,
	containerID string,
) (*storage.ContainerStats, error) {
	return s.containerService.GetStats(ctx, containerID)
}

// ExportIPQuorumConfig exports IPQuorum configuration as JSON
func (s *IPQuorumService) ExportIPQuorumConfig(
	ctx context.Context,
	containerID string,
) ([]byte, error) {

	response, err := s.GetIPQuorumContainer(ctx, containerID)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(response.IPQuorumConfig, "", "  ")
}

// Made with Bob
