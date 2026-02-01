package config

import (
	"fmt"
	"time"

	"github.com/olemyk/ipquorum-go/internal/errors"
)

// Config holds all configuration for IPQuorum operations
type Config struct {
	// API connection settings
	APIEndpoint string
	Username    string
	Password    string
	APIPort     int
	VerifySSL   bool

	// Operation flags
	MkQuorumApp bool
	Download    bool

	// Output settings
	OutputFile string
	Debug      bool

	// Retry settings
	MaxRetries int
	BaseDelay  time.Duration

	// mkquorumapp parameters
	IP6           bool
	NoMetadata    bool
	PartnerSystem string
	PartnerIP6    bool
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{
		APIPort:     7443,
		VerifySSL:   false,
		MkQuorumApp: true,
		Download:    true,
		OutputFile:  "ip_quorum.jar",
		Debug:       false,
		MaxRetries:  8,
		BaseDelay:   3 * time.Second,
		IP6:         false,
		NoMetadata:  false,
		PartnerIP6:  false,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Password == "" {
		return errors.NewValidationError("password is required")
	}

	if c.APIEndpoint == "" {
		return errors.NewValidationError("--api-endpoint <host> is required")
	}

	if c.Username == "" {
		return errors.NewValidationError("--user <username> is required")
	}

	if c.MkQuorumApp && c.PartnerSystem == "" {
		return errors.NewValidationError(
			"--partnersystem <name> is MANDATORY when --mkquorumapp is enabled",
		)
	}

	if c.MaxRetries < 1 {
		return errors.NewValidationError("max retries must be at least 1")
	}

	if c.BaseDelay < 0 {
		return errors.NewValidationError("base delay cannot be negative")
	}

	return nil
}

// String returns a string representation of the config (with masked password)
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{APIEndpoint: %s, Username: %s, Password: ********, "+
			"MkQuorumApp: %t, Download: %t, PartnerSystem: %s}",
		c.APIEndpoint, c.Username, c.MkQuorumApp, c.Download, c.PartnerSystem,
	)
}

// Made with help from Bob
