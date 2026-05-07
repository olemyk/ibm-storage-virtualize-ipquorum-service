package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Agent    AgentConfig    `mapstructure:"agent"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port                int    `mapstructure:"port"`
	Host                string `mapstructure:"host"`
	TLSCert             string `mapstructure:"tls_cert"`
	TLSKey              string `mapstructure:"tls_key"`
	LogLevel            string `mapstructure:"log_level"`
	HealthCheckInterval int    `mapstructure:"health_check_interval"` // seconds
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type string `mapstructure:"type"` // sqlite or postgres
	Path string `mapstructure:"path"` // for sqlite
	DSN  string `mapstructure:"dsn"`  // for postgres
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string `mapstructure:"jwt_secret"`
	TokenExpiry   int    `mapstructure:"token_expiry"`   // in seconds
	RefreshExpiry int    `mapstructure:"refresh_expiry"` // in seconds
}

// AgentConfig holds agent communication configuration
type AgentConfig struct {
	Port            int    `mapstructure:"port"`
	TLSCert         string `mapstructure:"tls_cert"`
	TLSKey          string `mapstructure:"tls_key"`
	ScriptsDir      string `mapstructure:"scripts_dir"`      // path to bash scripts
	HealthInterval  int    `mapstructure:"health_interval"`  // in seconds
	MetricsInterval int    `mapstructure:"metrics_interval"` // in seconds
}

// Default returns a default configuration
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port:                8443,
			Host:                "0.0.0.0",
			TLSCert:             "/etc/ipquorum-platform/tls/server.crt",
			TLSKey:              "/etc/ipquorum-platform/tls/server.key",
			LogLevel:            "info",
			HealthCheckInterval: 30, // 30 seconds
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			Path: "/var/lib/ipquorum-platform/ipquorum.db",
		},
		Auth: AuthConfig{
			JWTSecret:     generateSecret(),
			TokenExpiry:   3600,   // 1 hour
			RefreshExpiry: 604800, // 7 days
		},
		Agent: AgentConfig{
			Port:            8444,
			TLSCert:         "/etc/ipquorum-platform/tls/agent.crt",
			TLSKey:          "/etc/ipquorum-platform/tls/agent.key",
			ScriptsDir:      "/opt/ipquorum/scripts",
			HealthInterval:  30,
			MetricsInterval: 60,
		},
	}
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// Set defaults
	cfg := Default()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		// If config file doesn't exist, use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Unmarshal config
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	viper.Set("server", c.Server)
	viper.Set("database", c.Database)
	viper.Set("auth", c.Auth)
	viper.Set("agent", c.Agent)

	if err := viper.WriteConfigAs(path); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Type != "sqlite" && c.Database.Type != "postgres" {
		return fmt.Errorf("invalid database type: %s", c.Database.Type)
	}

	if c.Database.Type == "sqlite" && c.Database.Path == "" {
		return fmt.Errorf("database path is required for sqlite")
	}

	if c.Database.Type == "postgres" && c.Database.DSN == "" {
		return fmt.Errorf("database DSN is required for postgres")
	}

	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if c.Auth.TokenExpiry < 60 {
		return fmt.Errorf("token expiry must be at least 60 seconds")
	}

	return nil
}

// generateSecret generates a random secret for JWT
func generateSecret() string {
	// In production, this should be a secure random string
	// For now, return a placeholder that will be replaced
	return "CHANGE_ME_IN_PRODUCTION"
}
