package mocks

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/olemyk/ipquorum-management-platform/container-runtime-agent/pkg/runtime"
)

// MockRuntime is a mock implementation of the Runtime interface for testing
type MockRuntime struct {
	mu          sync.RWMutex
	name        string
	runtimeType runtime.RuntimeType
	containers  map[string]*runtime.Container
	images      []string

	// Mock behavior controls
	ShouldFailPing    bool
	ShouldFailCreate  bool
	ShouldFailStart   bool
	ShouldFailStop    bool
	ShouldFailRemove  bool
	ShouldFailList    bool
	ShouldFailInspect bool
	ShouldFailLogs    bool
	ShouldFailStats   bool

	// Call tracking
	PingCalls    int
	CreateCalls  int
	StartCalls   int
	StopCalls    int
	RemoveCalls  int
	ListCalls    int
	InspectCalls int
	LogsCalls    int
	StatsCalls   int
}

// NewMockRuntime creates a new mock runtime
func NewMockRuntime(name string, runtimeType runtime.RuntimeType) *MockRuntime {
	return &MockRuntime{
		name:        name,
		runtimeType: runtimeType,
		containers:  make(map[string]*runtime.Container),
		images:      []string{"alpine:latest", "nginx:latest", "redis:latest"},
	}
}

// Name returns the runtime name
func (m *MockRuntime) Name() string {
	return m.name
}

// Type returns the runtime type
func (m *MockRuntime) Type() runtime.RuntimeType {
	return m.runtimeType
}

// Ping checks if the runtime is accessible
func (m *MockRuntime) Ping(ctx context.Context) error {
	m.mu.Lock()
	m.PingCalls++
	m.mu.Unlock()

	if m.ShouldFailPing {
		return fmt.Errorf("mock ping failed")
	}
	return nil
}

// Version returns runtime version information
func (m *MockRuntime) Version(ctx context.Context) (*runtime.Version, error) {
	return &runtime.Version{
		Version:    "1.0.0-mock",
		APIVersion: "1.41",
		OS:         "linux",
		Arch:       "amd64",
		GoVersion:  "go1.21",
	}, nil
}

// Info returns runtime information
func (m *MockRuntime) Info(ctx context.Context) (*runtime.RuntimeInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	running := 0
	stopped := 0
	for _, c := range m.containers {
		if c.State == "running" {
			running++
		} else {
			stopped++
		}
	}

	return &runtime.RuntimeInfo{
		Name:              m.name,
		Type:              string(m.runtimeType),
		Version:           "1.0.0-mock",
		APIVersion:        "1.41",
		OS:                "linux",
		Arch:              "amd64",
		KernelVersion:     "5.10.0",
		ContainersRunning: running,
		ContainersPaused:  0,
		ContainersStopped: stopped,
		Images:            len(m.images),
		Driver:            "overlay2",
		MemoryTotal:       8589934592, // 8GB
		CPUs:              4,
	}, nil
}

// Create creates a new container
func (m *MockRuntime) Create(ctx context.Context, config *runtime.ContainerConfig) (*runtime.Container, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CreateCalls++

	if m.ShouldFailCreate {
		return nil, fmt.Errorf("mock create failed")
	}

	container := &runtime.Container{
		ID:      fmt.Sprintf("mock-%s-%d", config.Name, time.Now().Unix()),
		Name:    config.Name,
		Image:   config.Image,
		State:   "created",
		Status:  "Created",
		Created: time.Now(),
		Labels:  config.Labels,
	}

	m.containers[container.ID] = container
	return container, nil
}

// Start starts a container
func (m *MockRuntime) Start(ctx context.Context, containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StartCalls++

	if m.ShouldFailStart {
		return fmt.Errorf("mock start failed")
	}

	container, exists := m.containers[containerID]
	if !exists {
		return fmt.Errorf("container not found: %s", containerID)
	}

	container.State = "running"
	container.Status = "Up"
	return nil
}

// Stop stops a container
func (m *MockRuntime) Stop(ctx context.Context, containerID string, timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StopCalls++

	if m.ShouldFailStop {
		return fmt.Errorf("mock stop failed")
	}

	container, exists := m.containers[containerID]
	if !exists {
		return fmt.Errorf("container not found: %s", containerID)
	}

	container.State = "exited"
	container.Status = "Exited (0)"
	return nil
}

// Restart restarts a container
func (m *MockRuntime) Restart(ctx context.Context, containerID string, timeout time.Duration) error {
	if err := m.Stop(ctx, containerID, timeout); err != nil {
		return err
	}
	return m.Start(ctx, containerID)
}

// Remove removes a container
func (m *MockRuntime) Remove(ctx context.Context, containerID string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RemoveCalls++

	if m.ShouldFailRemove {
		return fmt.Errorf("mock remove failed")
	}

	if _, exists := m.containers[containerID]; !exists {
		return fmt.Errorf("container not found: %s", containerID)
	}

	delete(m.containers, containerID)
	return nil
}

// List lists containers
func (m *MockRuntime) List(ctx context.Context, filters map[string]string) ([]*runtime.Container, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.ListCalls++

	if m.ShouldFailList {
		return nil, fmt.Errorf("mock list failed")
	}

	var containers []*runtime.Container
	for _, c := range m.containers {
		// Apply filters
		if len(filters) > 0 {
			match := true
			for k, v := range filters {
				switch k {
				case "name":
					if !strings.Contains(c.Name, v) {
						match = false
					}
				case "status":
					if c.State != v {
						match = false
					}
				}
			}
			if !match {
				continue
			}
		}
		containers = append(containers, c)
	}

	return containers, nil
}

// Inspect inspects a container
func (m *MockRuntime) Inspect(ctx context.Context, containerID string) (*runtime.Container, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.InspectCalls++

	if m.ShouldFailInspect {
		return nil, fmt.Errorf("mock inspect failed")
	}

	container, exists := m.containers[containerID]
	if !exists {
		return nil, fmt.Errorf("container not found: %s", containerID)
	}

	return container, nil
}

// Logs retrieves container logs
func (m *MockRuntime) Logs(ctx context.Context, containerID string, opts runtime.LogOptions) (io.ReadCloser, error) {
	m.mu.Lock()
	m.LogsCalls++
	m.mu.Unlock()

	if m.ShouldFailLogs {
		return nil, fmt.Errorf("mock logs failed")
	}

	m.mu.RLock()
	container, exists := m.containers[containerID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("container not found: %s", containerID)
	}

	logs := fmt.Sprintf("Mock logs for container %s (%s)\n", container.Name, container.ID)
	return io.NopCloser(strings.NewReader(logs)), nil
}

// Stats retrieves container statistics
func (m *MockRuntime) Stats(ctx context.Context, containerID string) (*runtime.Stats, error) {
	m.mu.Lock()
	m.StatsCalls++
	m.mu.Unlock()

	if m.ShouldFailStats {
		return nil, fmt.Errorf("mock stats failed")
	}

	m.mu.RLock()
	_, exists := m.containers[containerID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("container not found: %s", containerID)
	}

	return &runtime.Stats{
		CPUPercent:    25.5,
		MemoryUsage:   536870912,  // 512MB
		MemoryLimit:   2147483648, // 2GB
		MemoryPercent: 25.0,
		NetworkRx:     1024000,
		NetworkTx:     512000,
		BlockRead:     2048000,
		BlockWrite:    1024000,
		PIDs:          10,
	}, nil
}

// ListImages lists the mock images
func (m *MockRuntime) ListImages(ctx context.Context) ([]*runtime.Image, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	images := make([]*runtime.Image, 0, len(m.images))
	for _, name := range m.images {
		images = append(images, &runtime.Image{ID: name, RepoTags: []string{name}})
	}
	return images, nil
}

// AddContainer adds a container to the mock runtime (for testing)
func (m *MockRuntime) AddContainer(container *runtime.Container) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.containers[container.ID] = container
}

// GetContainer gets a container from the mock runtime (for testing)
func (m *MockRuntime) GetContainer(containerID string) (*runtime.Container, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	container, exists := m.containers[containerID]
	return container, exists
}

// Reset resets the mock runtime state
func (m *MockRuntime) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.containers = make(map[string]*runtime.Container)
	m.ShouldFailPing = false
	m.ShouldFailCreate = false
	m.ShouldFailStart = false
	m.ShouldFailStop = false
	m.ShouldFailRemove = false
	m.ShouldFailList = false
	m.ShouldFailInspect = false
	m.ShouldFailLogs = false
	m.ShouldFailStats = false

	m.PingCalls = 0
	m.CreateCalls = 0
	m.StartCalls = 0
	m.StopCalls = 0
	m.RemoveCalls = 0
	m.ListCalls = 0
	m.InspectCalls = 0
	m.LogsCalls = 0
	m.StatsCalls = 0
}

// Made with Bob
