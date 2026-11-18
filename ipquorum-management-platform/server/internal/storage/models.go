package storage

import (
	"time"
)

// Instance represents an IP Quorum instance
type Instance struct {
	ID                string         `db:"id" json:"id"`
	Name              string         `db:"name" json:"name"`
	ServerID          string         `db:"server_id" json:"server_id"`
	DeploymentType    DeploymentType `db:"deployment_type" json:"deployment_type"` // systemd or container
	APIEndpoint       string         `db:"api_endpoint" json:"api_endpoint"`
	Username          string         `db:"username" json:"username"`
	Password          string         `db:"-" json:"-"` // Never stored in DB, only used for creation
	StorageSystem     string         `db:"storage_system" json:"storage_system"`
	Description       string         `db:"description" json:"description"`
	Location          string         `db:"location" json:"location"`
	Status            string         `db:"status" json:"status"`                                 // running, stopped, error
	Health            string         `db:"health" json:"health"`                                 // healthy, degraded, unhealthy
	StartedAt         *time.Time     `db:"started_at" json:"started_at,omitempty"`               // When instance was started (NULL if not running)
	Uptime            int64          `db:"-" json:"uptime,omitempty"`                            // Calculated uptime in seconds (not stored in DB)
	LastHealthCheck   *time.Time     `db:"last_health_check" json:"last_health_check,omitempty"` // Last health check timestamp
	EnableDownload    bool           `db:"enable_download" json:"enable_download"`
	EnableMkquorumapp bool           `db:"enable_mkquorumapp" json:"enable_mkquorumapp"`
	Partnersystem     string         `db:"partnersystem" json:"partnersystem"`
	IPQuorumName      string         `db:"ipquorum_name" json:"ipquorum_name"`
	IP6               bool           `db:"ip6" json:"ip6"`
	PartnerIP6        bool           `db:"partnerip6" json:"partnerip6"`
	NoMetadata        bool           `db:"nometadata" json:"nometadata"`
	CreatedAt         time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at" json:"updated_at"`
}

// Server represents a managed server
type Server struct {
	ID        string     `db:"id" json:"id"`
	Hostname  string     `db:"hostname" json:"hostname"`
	IPAddress string     `db:"ip_address" json:"ip_address"`
	AgentPort int        `db:"agent_port" json:"agent_port"`
	Status    string     `db:"status" json:"status"` // online, offline, error
	Version   string     `db:"version" json:"version,omitempty"`
	LastSeen  *time.Time `db:"last_seen" json:"last_seen,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

// User represents a platform user
type User struct {
	ID           string     `db:"id" json:"id"`
	Username     string     `db:"username" json:"username"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"` // Never expose in JSON
	Role         string     `db:"role" json:"role"`       // admin, operator, viewer
	Active       bool       `db:"active" json:"active"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	LastLogin    *time.Time `db:"last_login" json:"last_login,omitempty"` // Pointer to handle NULL
}

// HealthCheck represents a health check result
type HealthCheck struct {
	ID               int64     `db:"id" json:"id"`
	InstanceID       string    `db:"instance_id" json:"instance_id"`
	Timestamp        time.Time `db:"timestamp" json:"timestamp"`
	Status           string    `db:"status" json:"status"` // running, stopped, error
	Health           string    `db:"health" json:"health"` // healthy, degraded, unhealthy
	NetworkReachable bool      `db:"network_reachable" json:"network_reachable"`
	NetworkMessage   string    `db:"network_message" json:"network_message"`
	ResponseTime     int64     `db:"response_time_ms" json:"response_time_ms"`
	CheckedAt        time.Time `db:"checked_at" json:"checked_at"`
	ErrorMessage     string    `db:"error_message" json:"error_message,omitempty"`
}

// Metrics represents instance metrics
type Metrics struct {
	ID             int64     `db:"id" json:"id"`
	InstanceID     string    `db:"instance_id" json:"instance_id"`
	Timestamp      time.Time `db:"timestamp" json:"timestamp"`
	CPUUsage       float64   `db:"cpu_usage" json:"cpu_usage"`
	MemoryUsage    int64     `db:"memory_usage_bytes" json:"memory_usage_bytes"`
	NetworkRxBytes int64     `db:"network_rx_bytes" json:"network_rx_bytes"`
	NetworkTxBytes int64     `db:"network_tx_bytes" json:"network_tx_bytes"`
	UptimeSeconds  int64     `db:"uptime_seconds" json:"uptime_seconds"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           int64     `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"user_id"`
	Action       string    `db:"action" json:"action"`
	ResourceType string    `db:"resource_type" json:"resource_type"`
	ResourceID   string    `db:"resource_id" json:"resource_id"`
	Details      string    `db:"details" json:"details"`
	Timestamp    time.Time `db:"timestamp" json:"timestamp"`
}

// CreateInstanceRequest represents a request to create an instance
type CreateInstanceRequest struct {
	Name              string         `json:"name" binding:"required"`
	ServerID          string         `json:"server_id"`
	DeploymentType    DeploymentType `json:"deployment_type" binding:"required,oneof=systemd container"`
	APIEndpoint       string         `json:"api_endpoint" binding:"required"`
	Username          string         `json:"username" binding:"required"`
	Password          string         `json:"password" binding:"required"`
	StorageSystem     string         `json:"storage_system"`
	Description       string         `json:"description"`
	Location          string         `json:"location"`
	EnableDownload    *bool          `json:"enable_download"`
	EnableMkquorumapp *bool          `json:"enable_mkquorumapp"`
	Partnersystem     string         `json:"partnersystem"`
	IPQuorumName      string         `json:"ipquorum_name"`
	IP6               *bool          `json:"ip6"`
	PartnerIP6        *bool          `json:"partnerip6"`
	NoMetadata        *bool          `json:"nometadata"`
	// Container-specific fields (only used when deployment_type=container)
	ContainerConfig *CreateContainerInstanceRequest `json:"container_config,omitempty"`
}

// UpdateInstanceRequest represents a request to update an instance
type UpdateInstanceRequest struct {
	APIEndpoint       string `json:"api_endpoint"`
	Username          string `json:"username"`
	Password          string `json:"password,omitempty"`
	StorageSystem     string `json:"storage_system"`
	Description       string `json:"description"`
	Location          string `json:"location"`
	EnableDownload    *bool  `json:"enable_download"`
	EnableMkquorumapp *bool  `json:"enable_mkquorumapp"`
	Partnersystem     string `json:"partnersystem"`
	IPQuorumName      string `json:"ipquorum_name"`
	IP6               *bool  `json:"ip6"`
	PartnerIP6        *bool  `json:"partnerip6"`
	NoMetadata        *bool  `json:"nometadata"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	User         *User  `json:"user"`
}

// HealthResponse represents overall system health
type HealthResponse struct {
	Status     string                 `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	Uptime     int64                  `json:"uptime,omitempty"`
	Components map[string]interface{} `json:"components"`
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=admin operator viewer"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role,omitempty"`
	Active   *bool  `json:"active,omitempty"`
}

// DeploymentType represents the type of deployment
type DeploymentType string

const (
	DeploymentTypeSystemd   DeploymentType = "systemd"
	DeploymentTypeContainer DeploymentType = "container"
)

// ContainerInstance represents a containerized IP Quorum instance
type ContainerInstance struct {
	ID            string     `db:"id" json:"id"`
	InstanceID    string     `db:"instance_id" json:"instance_id"`       // Links to Instance.ID
	AgentID       string     `db:"agent_id" json:"agent_id"`             // Which agent manages this container
	ContainerID   string     `db:"container_id" json:"container_id"`     // Docker/Podman container ID
	ContainerName string     `db:"container_name" json:"container_name"` // Container name
	ImageName     string     `db:"image_name" json:"image_name"`         // Container image (e.g., ghcr.io/user/ipquorum:latest)
	ImageTag      string     `db:"image_tag" json:"image_tag"`           // Image tag (e.g., latest, v1.0.0)
	Runtime       string     `db:"runtime" json:"runtime"`               // docker or podman
	Status        string     `db:"status" json:"status"`                 // created, running, paused, stopped, error
	State         string     `db:"state" json:"state"`                   // Container state details
	RestartPolicy string     `db:"restart_policy" json:"restart_policy"` // no, always, on-failure, unless-stopped
	RestartCount  int        `db:"restart_count" json:"restart_count"`   // Number of restarts
	ExitCode      *int       `db:"exit_code" json:"exit_code,omitempty"` // Exit code if stopped
	Error         string     `db:"error" json:"error,omitempty"`         // Error message if any
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	StartedAt     *time.Time `db:"started_at" json:"started_at,omitempty"`
	FinishedAt    *time.Time `db:"finished_at" json:"finished_at,omitempty"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// ContainerConfig represents container configuration
type ContainerConfig struct {
	Ports       []PortMapping        `json:"ports,omitempty"`
	Volumes     []VolumeMount        `json:"volumes,omitempty"`
	Environment map[string]string    `json:"environment,omitempty"`
	Labels      map[string]string    `json:"labels,omitempty"`
	Resources   *ResourceConstraints `json:"resources,omitempty"`
	WorkingDir  string               `json:"working_dir,omitempty"`
}

// PortMapping represents a port mapping
type PortMapping struct {
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"` // tcp, udp
}

// VolumeMount represents a volume mount
type VolumeMount struct {
	Source      string `json:"source"`      // Host path or volume name
	Destination string `json:"destination"` // Container path
	ReadOnly    bool   `json:"read_only"`
}

// ResourceConstraints represents resource limits
type ResourceConstraints struct {
	CPULimit    float64 `json:"cpu_limit,omitempty"`    // CPU cores (e.g., 1.5)
	MemoryLimit int64   `json:"memory_limit,omitempty"` // Memory in bytes
}

// CreateContainerInstanceRequest represents a request to create a container instance
type CreateContainerInstanceRequest struct {
	InstanceID    string           `json:"instance_id" binding:"required"`
	AgentID       string           `json:"agent_id" binding:"required"`
	ImageName     string           `json:"image_name" binding:"required"`
	ImageTag      string           `json:"image_tag"`
	RestartPolicy string           `json:"restart_policy"`
	Config        *ContainerConfig `json:"config,omitempty"`
}

// UpdateContainerInstanceRequest represents a request to update a container instance
type UpdateContainerInstanceRequest struct {
	RestartPolicy string           `json:"restart_policy,omitempty"`
	Config        *ContainerConfig `json:"config,omitempty"`
}

// ContainerStats represents container statistics
type ContainerStats struct {
	ContainerID    string    `json:"container_id"`
	CPUPercent     float64   `json:"cpu_percent"`
	MemoryUsage    int64     `json:"memory_usage"`
	MemoryLimit    int64     `json:"memory_limit"`
	MemoryPercent  float64   `json:"memory_percent"`
	NetworkRxBytes int64     `json:"network_rx_bytes"`
	NetworkTxBytes int64     `json:"network_tx_bytes"`
	BlockRead      int64     `json:"block_read"`
	BlockWrite     int64     `json:"block_write"`
	PIDs           int       `json:"pids"`
	Timestamp      time.Time `json:"timestamp"`
}

// ContainerLogs represents container logs
type ContainerLogs struct {
	ContainerID string    `json:"container_id"`
	Logs        []string  `json:"logs"`
	Timestamp   time.Time `json:"timestamp"`
}

// IPQuorumConfig represents IPQuorum-specific container configuration
type IPQuorumConfig struct {
	APIEndpoint   string `json:"api_endpoint" binding:"required"`
	Username      string `json:"username" binding:"required"`
	Password      string `json:"password" binding:"required"` // Will be stored securely
	PartnerSystem string `json:"partner_system"`
	IP6           bool   `json:"ip6"`
	PartnerIP6    bool   `json:"partnerip6"`
	NoMetadata    bool   `json:"nometadata"`
	Insecure      bool   `json:"insecure"`
	DownloadJAR   bool   `json:"download_jar"`
	CreateQuorum  bool   `json:"create_quorum"`
	Port          int    `json:"port"` // IPQuorum port (default 1260)
	JavaOpts      string `json:"java_opts,omitempty"`
	ForceDownload bool   `json:"force_download,omitempty"`
}

// CreateIPQuorumContainerRequest represents a request to create an IPQuorum container
type CreateIPQuorumContainerRequest struct {
	Name           string          `json:"name" binding:"required"`
	AgentID        string          `json:"agent_id" binding:"required"`
	IPQuorumConfig *IPQuorumConfig `json:"ipquorum_config" binding:"required"`
	RestartPolicy  string          `json:"restart_policy,omitempty"`
	ImageName      string          `json:"image_name,omitempty"` // Optional, defaults to localhost/ipquorum-runner:latest
	ImageTag       string          `json:"image_tag,omitempty"`  // Optional, defaults to latest
}

// IPQuorumContainerResponse represents an IPQuorum container with full details
type IPQuorumContainerResponse struct {
	ContainerInstance *ContainerInstance `json:"container"`
	Instance          *Instance          `json:"instance,omitempty"`
	IPQuorumConfig    *IPQuorumConfig    `json:"ipquorum_config"`
}
