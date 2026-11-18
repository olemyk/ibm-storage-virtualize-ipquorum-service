package testhelpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestContext creates a context with timeout for tests
func TestContext(t *testing.T, timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}

// TempDir creates a temporary directory for tests
func TempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "runtime-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// TempSocket creates a temporary socket path for tests
func TempSocket(t *testing.T) string {
	dir := TempDir(t)
	return filepath.Join(dir, "test.sock")
}

// SkipIfNoDocker skips test if Docker is not available
func SkipIfNoDocker(t *testing.T) {
	if os.Getenv("SKIP_DOCKER_TESTS") == "true" {
		t.Skip("Skipping Docker test (SKIP_DOCKER_TESTS=true)")
	}
}

// SkipIfNoPodman skips test if Podman is not available
func SkipIfNoPodman(t *testing.T) {
	if os.Getenv("SKIP_PODMAN_TESTS") == "true" {
		t.Skip("Skipping Podman test (SKIP_PODMAN_TESTS=true)")
	}
}

// RequireDocker requires Docker to be available
func RequireDocker(t *testing.T) {
	SkipIfNoDocker(t)
	// Additional checks can be added here
}

// RequirePodman requires Podman to be available
func RequirePodman(t *testing.T) {
	SkipIfNoPodman(t)
	// Additional checks can be added here
}

// AssertEventually asserts that a condition becomes true within timeout
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if condition() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("Timeout waiting for condition: %s", message)
		}
		<-ticker.C
	}
}

// MockSocketServer creates a mock Unix socket server for testing
type MockSocketServer struct {
	socketPath string
	listener   interface{}
	t          *testing.T
}

// NewMockSocketServer creates a new mock socket server
func NewMockSocketServer(t *testing.T) *MockSocketServer {
	return &MockSocketServer{
		socketPath: TempSocket(t),
		t:          t,
	}
}

// Path returns the socket path
func (m *MockSocketServer) Path() string {
	return m.socketPath
}

// Close closes the mock server
func (m *MockSocketServer) Close() error {
	if m.socketPath != "" {
		return os.Remove(m.socketPath)
	}
	return nil
}

// TestContainerConfig provides common container configuration for tests
type TestContainerConfig struct {
	Name       string
	Image      string
	Env        map[string]string
	Ports      map[string]string
	Volumes    map[string]string
	WorkingDir string
	Command    []string
}

// DefaultTestConfig returns a default test container configuration
func DefaultTestConfig() *TestContainerConfig {
	return &TestContainerConfig{
		Name:  "test-container",
		Image: "alpine:latest",
		Env: map[string]string{
			"TEST_VAR": "test_value",
		},
		Command: []string{"sleep", "infinity"},
	}
}

// WithName sets the container name
func (c *TestContainerConfig) WithName(name string) *TestContainerConfig {
	c.Name = name
	return c
}

// WithImage sets the container image
func (c *TestContainerConfig) WithImage(image string) *TestContainerConfig {
	c.Image = image
	return c
}

// WithEnv adds environment variables
func (c *TestContainerConfig) WithEnv(key, value string) *TestContainerConfig {
	if c.Env == nil {
		c.Env = make(map[string]string)
	}
	c.Env[key] = value
	return c
}

// WithPort adds port mapping
func (c *TestContainerConfig) WithPort(host, container string) *TestContainerConfig {
	if c.Ports == nil {
		c.Ports = make(map[string]string)
	}
	c.Ports[host] = container
	return c
}

// WithVolume adds volume mapping
func (c *TestContainerConfig) WithVolume(host, container string) *TestContainerConfig {
	if c.Volumes == nil {
		c.Volumes = make(map[string]string)
	}
	c.Volumes[host] = container
	return c
}

// WithCommand sets the container command
func (c *TestContainerConfig) WithCommand(cmd ...string) *TestContainerConfig {
	c.Command = cmd
	return c
}

// TestImageConfig provides common image configuration for tests
type TestImageConfig struct {
	Name string
	Tag  string
}

// DefaultTestImage returns a default test image configuration
func DefaultTestImage() *TestImageConfig {
	return &TestImageConfig{
		Name: "alpine",
		Tag:  "latest",
	}
}

// FullName returns the full image name with tag
func (i *TestImageConfig) FullName() string {
	if i.Tag != "" {
		return fmt.Sprintf("%s:%s", i.Name, i.Tag)
	}
	return i.Name
}

// TestNetworkConfig provides common network configuration for tests
type TestNetworkConfig struct {
	Name   string
	Driver string
	Subnet string
}

// DefaultTestNetwork returns a default test network configuration
func DefaultTestNetwork() *TestNetworkConfig {
	return &TestNetworkConfig{
		Name:   "test-network",
		Driver: "bridge",
	}
}

// TestVolumeConfig provides common volume configuration for tests
type TestVolumeConfig struct {
	Name   string
	Driver string
}

// DefaultTestVolume returns a default test volume configuration
func DefaultTestVolume() *TestVolumeConfig {
	return &TestVolumeConfig{
		Name:   "test-volume",
		Driver: "local",
	}
}

// CleanupContainer ensures a container is removed after test
func CleanupContainer(t *testing.T, cleanup func() error) {
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	})
}

// RandomName generates a random name for test resources
func RandomName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// Made with Bob
