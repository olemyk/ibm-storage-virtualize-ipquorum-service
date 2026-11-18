package runtime

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// FactoryConfig holds configuration for runtime factory
type FactoryConfig struct {
	// PreferredRuntime specifies which runtime to prefer if both are available
	// Valid values: "docker", "podman", "auto" (default)
	PreferredRuntime string

	// DockerSocketPath is the path to Docker socket (default: /var/run/docker.sock)
	DockerSocketPath string

	// PodmanSocketPath is the path to Podman socket (default: /run/podman/podman.sock)
	PodmanSocketPath string

	// Timeout for runtime detection operations
	DetectionTimeout time.Duration
}

// DefaultFactoryConfig returns default factory configuration
func DefaultFactoryConfig() *FactoryConfig {
	return &FactoryConfig{
		PreferredRuntime: "auto",
		DockerSocketPath: "/var/run/docker.sock",
		PodmanSocketPath: "/run/podman/podman.sock",
		DetectionTimeout: 5 * time.Second,
	}
}

// Factory creates and manages container runtime instances
type Factory struct {
	config *FactoryConfig
}

// NewFactory creates a new runtime factory with the given configuration
func NewFactory(config *FactoryConfig) *Factory {
	if config == nil {
		config = DefaultFactoryConfig()
	}
	return &Factory{config: config}
}

// DetectAvailableRuntimes detects which container runtimes are available on the system
func (f *Factory) DetectAvailableRuntimes(ctx context.Context) ([]RuntimeType, error) {
	var available []RuntimeType

	// Check Docker
	if f.isDockerAvailable(ctx) {
		available = append(available, RuntimeTypeDocker)
	}

	// Check Podman
	if f.isPodmanAvailable(ctx) {
		available = append(available, RuntimeTypePodman)
	}

	return available, nil
}

// CreateRuntime creates a runtime instance based on configuration and availability
func (f *Factory) CreateRuntime(ctx context.Context) (Runtime, error) {
	// Validate configuration before probing the host
	switch strings.ToLower(f.config.PreferredRuntime) {
	case "docker", "podman", "auto", "":
	default:
		return nil, fmt.Errorf("invalid preferred runtime: %s (must be 'docker', 'podman', or 'auto')", f.config.PreferredRuntime)
	}

	available, err := f.DetectAvailableRuntimes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect available runtimes: %w", err)
	}

	if len(available) == 0 {
		return nil, fmt.Errorf("no container runtime available (checked Docker and Podman)")
	}

	// Determine which runtime to use
	var runtimeType RuntimeType
	switch strings.ToLower(f.config.PreferredRuntime) {
	case "docker":
		if f.contains(available, RuntimeTypeDocker) {
			runtimeType = RuntimeTypeDocker
		} else {
			return nil, fmt.Errorf("preferred runtime 'docker' is not available")
		}
	case "podman":
		if f.contains(available, RuntimeTypePodman) {
			runtimeType = RuntimeTypePodman
		} else {
			return nil, fmt.Errorf("preferred runtime 'podman' is not available")
		}
	case "auto", "":
		// Auto-select: prefer Docker if available, otherwise Podman
		if f.contains(available, RuntimeTypeDocker) {
			runtimeType = RuntimeTypeDocker
		} else {
			runtimeType = RuntimeTypePodman
		}
	default:
		return nil, fmt.Errorf("invalid preferred runtime: %s (must be 'docker', 'podman', or 'auto')", f.config.PreferredRuntime)
	}

	// Create the runtime instance
	return f.createRuntimeByType(ctx, runtimeType)
}

// CreateRuntimeByType creates a specific runtime instance
func (f *Factory) createRuntimeByType(ctx context.Context, runtimeType RuntimeType) (Runtime, error) {
	// Create a basic logger for runtime instances
	logger, _ := zap.NewProduction()

	switch runtimeType {
	case RuntimeTypeDocker:
		return NewDockerRuntime(f.config.DockerSocketPath, logger)
	case RuntimeTypePodman:
		return NewPodmanRuntime(f.config.PodmanSocketPath, logger)
	default:
		return nil, fmt.Errorf("unsupported runtime type: %s", runtimeType)
	}
}

// isDockerAvailable checks if Docker socket is available and responsive
func (f *Factory) isDockerAvailable(ctx context.Context) bool {
	socketPath := f.config.DockerSocketPath

	// Skip file existence check for sockets mounted in containers
	// Unix sockets don't always appear as stat-able files when mounted
	// Just try to connect directly

	// Try to connect to the socket and make a simple API call
	return f.testSocket(ctx, socketPath, "/version")
}

// isPodmanAvailable checks if Podman socket is available and responsive
func (f *Factory) isPodmanAvailable(ctx context.Context) bool {
	socketPath := f.config.PodmanSocketPath

	// Skip file existence check for sockets mounted in containers
	// Unix sockets don't always appear as stat-able files when mounted
	// Just try to connect directly

	// Try to connect to the socket and make a simple API call
	return f.testSocket(ctx, socketPath, "/v4.0.0/libpod/info")
}

// testSocket tests if a Unix socket is accessible by making an HTTP request
func (f *Factory) testSocket(ctx context.Context, socketPath, endpoint string) bool {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, f.config.DetectionTimeout)
	defer cancel()

	// Create HTTP client with Unix socket transport
	client := &http.Client{
		Timeout: f.config.DetectionTimeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Make a simple GET request
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost"+endpoint, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Check if we got a successful response
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// contains checks if a slice contains a specific runtime type
func (f *Factory) contains(slice []RuntimeType, item RuntimeType) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetRuntimeInfo returns information about a specific runtime
func (f *Factory) GetRuntimeInfo(ctx context.Context, runtimeType RuntimeType) (*RuntimeInfo, error) {
	runtime, err := f.createRuntimeByType(ctx, runtimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}

	return runtime.Info(ctx)
}

// ValidateRuntime validates that a runtime is properly configured and accessible
func (f *Factory) ValidateRuntime(ctx context.Context, runtimeType RuntimeType) error {
	runtime, err := f.createRuntimeByType(ctx, runtimeType)
	if err != nil {
		return fmt.Errorf("failed to create runtime: %w", err)
	}

	// Try to ping the runtime
	if err := runtime.Ping(ctx); err != nil {
		return fmt.Errorf("runtime ping failed: %w", err)
	}

	// Try to get version info
	if _, err := runtime.Version(ctx); err != nil {
		return fmt.Errorf("failed to get runtime version: %w", err)
	}

	return nil
}

// GetPreferredRuntime returns the preferred runtime based on configuration and availability
func (f *Factory) GetPreferredRuntime(ctx context.Context) (RuntimeType, error) {
	available, err := f.DetectAvailableRuntimes(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to detect available runtimes: %w", err)
	}

	if len(available) == 0 {
		return "", fmt.Errorf("no container runtime available")
	}

	switch strings.ToLower(f.config.PreferredRuntime) {
	case "docker":
		if f.contains(available, RuntimeTypeDocker) {
			return RuntimeTypeDocker, nil
		}
		return "", fmt.Errorf("preferred runtime 'docker' is not available")
	case "podman":
		if f.contains(available, RuntimeTypePodman) {
			return RuntimeTypePodman, nil
		}
		return "", fmt.Errorf("preferred runtime 'podman' is not available")
	case "auto", "":
		// Auto-select: prefer Docker if available, otherwise Podman
		if f.contains(available, RuntimeTypeDocker) {
			return RuntimeTypeDocker, nil
		}
		return RuntimeTypePodman, nil
	default:
		return "", fmt.Errorf("invalid preferred runtime: %s", f.config.PreferredRuntime)
	}
}

// Made with Bob
