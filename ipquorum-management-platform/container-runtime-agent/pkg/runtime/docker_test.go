package runtime

import (
	"io"
	"testing"
	"time"

	"github.com/olemyk/ipquorum-management-platform/container-runtime-agent/pkg/runtime/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func getTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func TestNewDockerRuntime(t *testing.T) {
	tests := []struct {
		name        string
		socketPath  string
		expectError bool
	}{
		{
			name:        "create with default socket",
			socketPath:  "",
			expectError: false,
		},
		{
			name:        "create with custom socket",
			socketPath:  "/custom/docker.sock",
			expectError: false,
		},
		{
			name:        "create with non-existent socket",
			socketPath:  "/nonexistent/docker.sock",
			expectError: false, // Client creation succeeds, ping will fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime, err := NewDockerRuntime(tt.socketPath, getTestLogger())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, runtime)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, runtime)
				assert.Equal(t, RuntimeTypeDocker, runtime.Type())
			}
		})
	}
}

func TestDockerRuntime_Ping(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	tests := []struct {
		name        string
		socketPath  string
		expectError bool
	}{
		{
			name:        "ping with default socket",
			socketPath:  "",
			expectError: false,
		},
		{
			name:        "ping with non-existent socket",
			socketPath:  "/nonexistent/docker.sock",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime, err := NewDockerRuntime(tt.socketPath, getTestLogger())
			require.NoError(t, err)

			ctx := testhelpers.TestContext(t, 5*time.Second)
			err = runtime.Ping(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDockerRuntime_Version(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getTestLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 5*time.Second)
	version, err := runtime.Version(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, version)
	assert.NotEmpty(t, version.Version)
	assert.NotEmpty(t, version.APIVersion)
	assert.NotEmpty(t, version.GoVersion)
	assert.NotEmpty(t, version.OS)
	assert.NotEmpty(t, version.Arch)
}

func TestDockerRuntime_Info(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getTestLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 5*time.Second)
	info, err := runtime.Info(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.NotEmpty(t, info.Name)
	assert.NotEmpty(t, info.Version)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Arch)
	assert.GreaterOrEqual(t, info.CPUs, int64(1))
	assert.Greater(t, info.MemoryTotal, int64(0))
}

func TestDockerRuntime_List(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getTestLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 10*time.Second)

	tests := []struct {
		name    string
		filters map[string]string
	}{
		{
			name:    "list all containers",
			filters: nil,
		},
		{
			name: "list with name filter",
			filters: map[string]string{
				"name": "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containers, err := runtime.List(ctx, tt.filters)
			assert.NoError(t, err)
			assert.NotNil(t, containers)
		})
	}
}

func TestDockerRuntime_Inspect(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getTestLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 5*time.Second)

	t.Run("inspect non-existent container", func(t *testing.T) {
		container, err := runtime.Inspect(ctx, "nonexistent-container-12345")
		assert.Error(t, err)
		assert.Nil(t, container)
	})
}

func TestDockerRuntime_Create(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getTestLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 30*time.Second)

	tests := []struct {
		name        string
		config      *ContainerConfig
		expectError bool
	}{
		{
			name: "create simple container",
			config: &ContainerConfig{
				Name:  testhelpers.RandomName("test-container"),
				Image: "alpine:latest",
				Cmd:   []string{"sleep", "10"},
			},
			expectError: false,
		},
		{
			name: "create with environment variables",
			config: &ContainerConfig{
				Name:  testhelpers.RandomName("test-env"),
				Image: "alpine:latest",
				Env: map[string]string{
					"TEST_VAR": "test_value",
				},
				Cmd: []string{"sleep", "10"},
			},
			expectError: false,
		},
		{
			name: "create with invalid image",
			config: &ContainerConfig{
				Name:  testhelpers.RandomName("test-invalid"),
				Image: "nonexistent-image-12345:latest",
				Cmd:   []string{"sleep", "10"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := runtime.Create(ctx, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, container)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, container)
				assert.NotEmpty(t, container.ID)

				// Cleanup
				testhelpers.CleanupContainer(t, func() error {
					return runtime.Remove(ctx, container.ID, true)
				})
			}
		})
	}
}

func TestDockerRuntime_StartStop(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getTestLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 60*time.Second)

	// Create a test container
	config := &ContainerConfig{
		Name:  testhelpers.RandomName("test-lifecycle"),
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "infinity"},
	}

	container, err := runtime.Create(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, container)

	defer func() {
		_ = runtime.Remove(ctx, container.ID, true)
	}()

	t.Run("start container", func(t *testing.T) {
		err := runtime.Start(ctx, container.ID)
		assert.NoError(t, err)

		// Verify container is running
		inspected, err := runtime.Inspect(ctx, container.ID)
		assert.NoError(t, err)
		assert.Equal(t, "running", inspected.State)
	})

	t.Run("stop container", func(t *testing.T) {
		err := runtime.Stop(ctx, container.ID, 10*time.Second)
		assert.NoError(t, err)

		// Verify container is stopped
		inspected, err := runtime.Inspect(ctx, container.ID)
		assert.NoError(t, err)
		assert.NotEqual(t, "running", inspected.State)
	})

	t.Run("restart container", func(t *testing.T) {
		err := runtime.Restart(ctx, container.ID, 10*time.Second)
		assert.NoError(t, err)

		// Verify container is running again
		inspected, err := runtime.Inspect(ctx, container.ID)
		assert.NoError(t, err)
		assert.Equal(t, "running", inspected.State)
	})
}

func TestDockerRuntime_Remove(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getTestLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 30*time.Second)

	tests := []struct {
		name        string
		force       bool
		expectError bool
	}{
		{
			name:        "remove stopped container",
			force:       false,
			expectError: false,
		},
		{
			name:        "force remove running container",
			force:       true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create container
			config := &ContainerConfig{
				Name:  testhelpers.RandomName("test-remove"),
				Image: "alpine:latest",
				Cmd:   []string{"sleep", "10"},
			}

			container, err := runtime.Create(ctx, config)
			require.NoError(t, err)

			if tt.force {
				// Start container for force remove test
				err = runtime.Start(ctx, container.ID)
				require.NoError(t, err)
			}

			// Remove container
			err = runtime.Remove(ctx, container.ID, tt.force)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify container is removed
				_, err = runtime.Inspect(ctx, container.ID)
				assert.Error(t, err)
			}
		})
	}
}

func TestDockerRuntime_Logs(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getTestLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 30*time.Second)

	// Create and start a container that produces logs
	config := &ContainerConfig{
		Name:  testhelpers.RandomName("test-logs"),
		Image: "alpine:latest",
		Cmd:   []string{"sh", "-c", "echo 'test log line' && sleep 5"},
	}

	container, err := runtime.Create(ctx, config)
	require.NoError(t, err)
	defer func() {
		_ = runtime.Remove(ctx, container.ID, true)
	}()

	err = runtime.Start(ctx, container.ID)
	require.NoError(t, err)

	// Wait a bit for logs to be generated
	time.Sleep(2 * time.Second)

	t.Run("get logs", func(t *testing.T) {
		logsReader, err := runtime.Logs(ctx, container.ID, LogOptions{
			Follow:     false,
			Tail:       100,
			Timestamps: false,
		})

		assert.NoError(t, err)
		assert.NotNil(t, logsReader)

		if logsReader != nil {
			defer logsReader.Close()
			logs, _ := io.ReadAll(logsReader)
			assert.NotEmpty(t, logs)
		}
	})

	t.Run("get logs with tail", func(t *testing.T) {
		logsReader, err := runtime.Logs(ctx, container.ID, LogOptions{
			Follow:     false,
			Tail:       10,
			Timestamps: false,
		})

		assert.NoError(t, err)
		assert.NotNil(t, logsReader)

		if logsReader != nil {
			defer logsReader.Close()
		}
	})
}

func TestDockerRuntime_Stats(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getTestLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 30*time.Second)

	// Create and start a container
	config := &ContainerConfig{
		Name:  testhelpers.RandomName("test-stats"),
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "30"},
	}

	container, err := runtime.Create(ctx, config)
	require.NoError(t, err)
	defer func() {
		_ = runtime.Remove(ctx, container.ID, true)
	}()

	err = runtime.Start(ctx, container.ID)
	require.NoError(t, err)

	// Wait a bit for stats to be available
	time.Sleep(2 * time.Second)

	t.Run("get stats", func(t *testing.T) {
		stats, err := runtime.Stats(ctx, container.ID)

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Greater(t, stats.MemoryUsage, uint64(0))
		assert.GreaterOrEqual(t, stats.CPUPercent, 0.0)
	})
}

func TestDockerRuntime_Type(t *testing.T) {
	runtime, err := NewDockerRuntime("", getTestLogger())
	require.NoError(t, err)

	assert.Equal(t, RuntimeTypeDocker, runtime.Type())
	assert.Equal(t, "Docker", runtime.Name())
}

// Made with Bob
