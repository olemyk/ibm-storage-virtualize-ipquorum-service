package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the agent configuration
type Config struct {
	Agent  AgentConfig  `yaml:"agent"`
	Server ServerConfig `yaml:"server"`
	Script ScriptConfig `yaml:"script"`
}

// AgentConfig contains agent-specific settings
type AgentConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	APIKey       string `yaml:"api_key"`
	LogLevel     string `yaml:"log_level"`
	InstanceName string `yaml:"instance_name"`
}

// ServerConfig contains management server connection settings
type ServerConfig struct {
	URL      string `yaml:"url"`
	Insecure bool   `yaml:"insecure"`
}

// ScriptConfig contains script execution settings
type ScriptConfig struct {
	Path    string `yaml:"path"`
	Timeout int    `yaml:"timeout"` // seconds
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.Agent.Port == 0 {
		cfg.Agent.Port = 9090
	}
	if cfg.Agent.Host == "" {
		cfg.Agent.Host = "0.0.0.0"
	}
	if cfg.Agent.LogLevel == "" {
		cfg.Agent.LogLevel = "info"
	}
	if cfg.Script.Timeout == 0 {
		cfg.Script.Timeout = 300 // 5 minutes default
	}

	return &cfg, nil
}
