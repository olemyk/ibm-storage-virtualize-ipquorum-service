package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config holds the complete configuration for the Container Runtime Agent
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Runtime  RuntimeConfig  `mapstructure:"runtime"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Security SecurityConfig `mapstructure:"security"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	TLS          TLSConfig     `mapstructure:"tls"`
}

// TLSConfig holds TLS/HTTPS configuration
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// RuntimeConfig holds container runtime configuration
type RuntimeConfig struct {
	PreferredRuntime  string            `mapstructure:"preferred_runtime"`
	DockerSocketPath  string            `mapstructure:"docker_socket_path"`
	PodmanSocketPath  string            `mapstructure:"podman_socket_path"`
	DetectionTimeout  time.Duration     `mapstructure:"detection_timeout"`
	DefaultNamespace  string            `mapstructure:"default_namespace"`
	MaxContainers     int               `mapstructure:"max_containers"`
	ContainerDefaults ContainerDefaults `mapstructure:"container_defaults"`
}

// ContainerDefaults holds default values for container creation
type ContainerDefaults struct {
	RestartPolicy string            `mapstructure:"restart_policy"`
	NetworkMode   string            `mapstructure:"network_mode"`
	LogDriver     string            `mapstructure:"log_driver"`
	Labels        map[string]string `mapstructure:"labels"`
	Resources     ResourceDefaults  `mapstructure:"resources"`
}

// ResourceDefaults holds default resource limits
type ResourceDefaults struct {
	MemoryLimit int64   `mapstructure:"memory_limit"`
	CPUShares   int64   `mapstructure:"cpu_shares"`
	CPUQuota    int64   `mapstructure:"cpu_quota"`
	CPUPeriod   int64   `mapstructure:"cpu_period"`
	CPUs        float64 `mapstructure:"cpus"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled       bool            `mapstructure:"enabled"`
	JWTSecret     string          `mapstructure:"jwt_secret"`
	JWTExpiration time.Duration   `mapstructure:"jwt_expiration"`
	APIKeys       []string        `mapstructure:"api_keys"`
	BasicAuth     BasicAuthConfig `mapstructure:"basic_auth"`
}

// BasicAuthConfig holds basic authentication configuration
type BasicAuthConfig struct {
	Enabled bool              `mapstructure:"enabled"`
	Users   map[string]string `mapstructure:"users"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
	Port    int    `mapstructure:"port"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	AllowedOrigins []string        `mapstructure:"allowed_origins"`
	RateLimiting   RateLimitConfig `mapstructure:"rate_limiting"`
	IPWhitelist    []string        `mapstructure:"ip_whitelist"`
	IPBlacklist    []string        `mapstructure:"ip_blacklist"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	RequestsPerMinute int           `mapstructure:"requests_per_minute"`
	BurstSize         int           `mapstructure:"burst_size"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
			TLS: TLSConfig{
				Enabled:  false,
				CertFile: "",
				KeyFile:  "",
			},
		},
		Runtime: RuntimeConfig{
			PreferredRuntime: "auto",
			DockerSocketPath: "/var/run/docker.sock",
			PodmanSocketPath: "/run/podman/podman.sock",
			DetectionTimeout: 5 * time.Second,
			DefaultNamespace: "ipquorum",
			MaxContainers:    100,
			ContainerDefaults: ContainerDefaults{
				RestartPolicy: "unless-stopped",
				NetworkMode:   "bridge",
				LogDriver:     "json-file",
				Labels: map[string]string{
					"managed-by": "ipquorum-agent",
				},
				Resources: ResourceDefaults{
					MemoryLimit: 512 * 1024 * 1024, // 512MB
					CPUShares:   1024,
					CPUQuota:    0,
					CPUPeriod:   100000,
					CPUs:        1.0,
				},
			},
		},
		Auth: AuthConfig{
			Enabled:       true,
			JWTSecret:     "",
			JWTExpiration: 24 * time.Hour,
			APIKeys:       []string{},
			BasicAuth: BasicAuthConfig{
				Enabled: false,
				Users:   map[string]string{},
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			File:       "",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
			Port:    9090,
		},
		Security: SecurityConfig{
			AllowedOrigins: []string{"*"},
			RateLimiting: RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 60,
				BurstSize:         10,
				CleanupInterval:   5 * time.Minute,
			},
			IPWhitelist: []string{},
			IPBlacklist: []string{},
		},
	}
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Set up Viper
	v := viper.New()
	v.SetConfigType("yaml")

	// If config file is provided, read it
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Environment variables override config file
	v.SetEnvPrefix("IPQUORUM_AGENT")
	v.AutomaticEnv()

	// Unmarshal into config struct
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadFromEnv loads configuration from environment variables only
func LoadFromEnv() (*Config, error) {
	return Load("")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be 1-65535)", c.Server.Port)
	}

	if c.Server.TLS.Enabled {
		if c.Server.TLS.CertFile == "" {
			return fmt.Errorf("TLS enabled but cert_file is empty")
		}
		if c.Server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS enabled but key_file is empty")
		}
		if _, err := os.Stat(c.Server.TLS.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("TLS cert file does not exist: %s", c.Server.TLS.CertFile)
		}
		if _, err := os.Stat(c.Server.TLS.KeyFile); os.IsNotExist(err) {
			return fmt.Errorf("TLS key file does not exist: %s", c.Server.TLS.KeyFile)
		}
	}

	// Validate runtime config
	validRuntimes := map[string]bool{"auto": true, "docker": true, "podman": true}
	if !validRuntimes[c.Runtime.PreferredRuntime] {
		return fmt.Errorf("invalid preferred_runtime: %s (must be 'auto', 'docker', or 'podman')", c.Runtime.PreferredRuntime)
	}

	if c.Runtime.MaxContainers < 1 {
		return fmt.Errorf("max_containers must be at least 1")
	}

	// Validate auth config
	if c.Auth.Enabled && c.Auth.JWTSecret == "" && len(c.Auth.APIKeys) == 0 && !c.Auth.BasicAuth.Enabled {
		return fmt.Errorf("authentication enabled but no auth method configured (jwt_secret, api_keys, or basic_auth required)")
	}

	// Validate logging config
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be 'debug', 'info', 'warn', or 'error')", c.Logging.Level)
	}

	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[c.Logging.Format] {
		return fmt.Errorf("invalid log format: %s (must be 'json' or 'text')", c.Logging.Format)
	}

	// Validate metrics config
	if c.Metrics.Enabled && (c.Metrics.Port < 1 || c.Metrics.Port > 65535) {
		return fmt.Errorf("invalid metrics port: %d (must be 1-65535)", c.Metrics.Port)
	}

	return nil
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetMetricsAddress returns the full metrics server address
func (c *Config) GetMetricsAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Metrics.Port)
}

// IsProductionMode returns true if running in production mode
func (c *Config) IsProductionMode() bool {
	return c.Logging.Level != "debug"
}

// Made with Bob
