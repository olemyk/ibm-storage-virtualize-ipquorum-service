package storage

import (
	"time"
)

// Instance represents an IP Quorum instance
type Instance struct {
	ID                string     `db:"id" json:"id"`
	Name              string     `db:"name" json:"name"`
	ServerID          string     `db:"server_id" json:"server_id"`
	APIEndpoint       string     `db:"api_endpoint" json:"api_endpoint"`
	Username          string     `db:"username" json:"username"`
	Password          string     `db:"-" json:"-"` // Never stored in DB, only used for creation
	StorageSystem     string     `db:"storage_system" json:"storage_system"`
	Description       string     `db:"description" json:"description"`
	Location          string     `db:"location" json:"location"`
	Status            string     `db:"status" json:"status"`                                 // running, stopped, error
	Health            string     `db:"health" json:"health"`                                 // healthy, degraded, unhealthy
	StartedAt         *time.Time `db:"started_at" json:"started_at,omitempty"`               // When instance was started (NULL if not running)
	Uptime            int64      `db:"-" json:"uptime,omitempty"`                            // Calculated uptime in seconds (not stored in DB)
	LastHealthCheck   *time.Time `db:"last_health_check" json:"last_health_check,omitempty"` // Last health check timestamp
	EnableDownload    bool       `db:"enable_download" json:"enable_download"`
	EnableMkquorumapp bool       `db:"enable_mkquorumapp" json:"enable_mkquorumapp"`
	Partnersystem     string     `db:"partnersystem" json:"partnersystem"`
	IPQuorumName      string     `db:"ipquorum_name" json:"ipquorum_name"`
	IP6               bool       `db:"ip6" json:"ip6"`
	PartnerIP6        bool       `db:"partnerip6" json:"partnerip6"`
	NoMetadata        bool       `db:"nometadata" json:"nometadata"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
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
	Name              string `json:"name" binding:"required"`
	ServerID          string `json:"server_id"`
	APIEndpoint       string `json:"api_endpoint" binding:"required"`
	Username          string `json:"username" binding:"required"`
	Password          string `json:"password" binding:"required"`
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
