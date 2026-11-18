//go:build integration
// +build integration

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

// Integration tests require Docker or Podman to be running
// Run with: go test -tags=integration -v ./pkg/runtime

func getIntegrationLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func TestIntegration_ContainerLifecycle(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getIntegrationLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 120*time.Second)

	t.Run("complete lifecycle", func(t *testing.T) {
		// 1. Create container
		config := &ContainerConfig{
			Name:  testhelpers.RandomName("integration-lifecycle"),
			Image: "alpine:latest",
			Cmd:   []string{"sh", "-c", "echo 'Starting...' && sleep 30 && echo 'Done'"},
			Env: []string{
				"TEST_ENV=integration",
			},
			Labels: map[string]string{
				"test": "integration",
				"type": "lifecycle",
			},
		}

		container, err := runtime.Create(ctx, config)
		require.NoError(t, err)
		require.NotNil(t, container)
		assert.NotEmpty(t, container.ID)
		assert.Equal(t, config.Name, container.Name)

		defer func() {
			_ = runtime.Remove(ctx, container.ID, true)
		}()

		// 2. Verify created state
		inspected, err := runtime.Inspect(ctx, container.ID)
		require.NoError(t, err)
		assert.Equal(t, "created", inspected.State)

		// 3. Start container
		err = runtime.Start(ctx, container.ID)
		require.NoError(t, err)

		// 4. Verify running state
		time.Sleep(1 * time.Second)
		inspected, err = runtime.Inspect(ctx, container.ID)
		require.NoError(t, err)
		assert.Equal(t, "running", inspected.State)

		// 5. Get stats while running
		stats, err := runtime.Stats(ctx, container.ID)
		require.NoError(t, err)
		assert.Greater(t, stats.MemoryUsage, uint64(0))

		// 6. Get logs
		logsReader, err := runtime.Logs(ctx, container.ID, LogOptions{
			Follow:     false,
			Tail:       100,
			Timestamps: true,
		})
		require.NoError(t, err)
		defer logsReader.Close()

		logs, err := io.ReadAll(logsReader)
		require.NoError(t, err)
		assert.Contains(t, string(logs), "Starting")

		// 7. Stop container
		err = runtime.Stop(ctx, container.ID, 10*time.Second)
		require.NoError(t, err)

		// 8. Verify stopped state
		inspected, err = runtime.Inspect(ctx, container.ID)
		require.NoError(t, err)
		assert.NotEqual(t, "running", inspected.State)

		// 9. Restart container
		err = runtime.Restart(ctx, container.ID, 10*time.Second)
		require.NoError(t, err)

		// 10. Verify running again
		time.Sleep(1 * time.Second)
		inspected, err = runtime.Inspect(ctx, container.ID)
		require.NoError(t, err)
		assert.Equal(t, "running", inspected.State)

		// 11. Force remove while running
		err = runtime.Remove(ctx, container.ID, true)
		require.NoError(t, err)

		// 12. Verify removed
		_, err = runtime.Inspect(ctx, container.ID)
		assert.Error(t, err)
	})
}

func TestIntegration_ContainerLogs(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getIntegrationLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 60*time.Second)

	t.Run("log streaming", func(t *testing.T) {
		// Create container that generates logs
		config := &ContainerConfig{
			Name:  testhelpers.RandomName("integration-logs"),
			Image: "alpine:latest",
			Cmd: []string{"sh", "-c", `
				for i in 1 2 3 4 5; do
					echo "Log line $i"
					sleep 1
				done
			`},
		}

		container, err := runtime.Create(ctx, config)
		require.NoError(t, err)
		defer func() {
			_ = runtime.Remove(ctx, container.ID, true)
		}()

		err = runtime.Start(ctx, container.ID)
		require.NoError(t, err)

		// Wait for logs to be generated
		time.Sleep(3 * time.Second)

		// Get logs with tail
		logsReader, err := runtime.Logs(ctx, container.ID, LogOptions{
			Follow:     false,
			Tail:       3,
			Timestamps: false,
		})
		require.NoError(t, err)
		defer logsReader.Close()

		logs, err := io.ReadAll(logsReader)
		require.NoError(t, err)
		assert.NotEmpty(t, logs)

		// Get logs with timestamps
		logsReader2, err := runtime.Logs(ctx, container.ID, LogOptions{
			Follow:     false,
			Tail:       100,
			Timestamps: true,
		})
		require.NoError(t, err)
		defer logsReader2.Close()

		logsWithTime, err := io.ReadAll(logsReader2)
		require.NoError(t, err)
		assert.NotEmpty(t, logsWithTime)
	})
}

func TestIntegration_ContainerStats(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getIntegrationLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 60*time.Second)

	t.Run("stats monitoring", func(t *testing.T) {
		// Create container with some workload
		config := &ContainerConfig{
			Name:  testhelpers.RandomName("integration-stats"),
			Image: "alpine:latest",
			Cmd:   []string{"sh", "-c", "while true; do echo test; sleep 1; done"},
		}

		container, err := runtime.Create(ctx, config)
		require.NoError(t, err)
		defer func() {
			_ = runtime.Remove(ctx, container.ID, true)
		}()

		err = runtime.Start(ctx, container.ID)
		require.NoError(t, err)

		// Wait for container to stabilize
		time.Sleep(2 * time.Second)

		// Get stats multiple times
		for i := 0; i < 3; i++ {
			stats, err := runtime.Stats(ctx, container.ID)
			require.NoError(t, err)

			assert.Greater(t, stats.MemoryUsage, uint64(0))
			assert.GreaterOrEqual(t, stats.CPUPercent, 0.0)
			assert.Greater(t, stats.MemoryLimit, uint64(0))

			t.Logf("Stats iteration %d: CPU=%.2f%%, Memory=%d bytes",
				i+1, stats.CPUPercent, stats.MemoryUsage)

			time.Sleep(1 * time.Second)
		}
	})
}

func TestIntegration_RuntimeSwitching(t *testing.T) {
	ctx := testhelpers.TestContext(t, 60*time.Second)

	t.Run("detect and switch runtimes", func(t *testing.T) {
		factory := NewFactory(&FactoryConfig{
			DockerSocket: "",
			PodmanSocket: "",
			Timeout:      5 * time.Second,
		})

		// Detect available runtimes
		runtimes, err := factory.DetectAvailableRuntimes(ctx)
		require.NoError(t, err)

		if len(runtimes) == 0 {
			t.Skip("No container runtimes available")
		}

		t.Logf("Available runtimes: %v", runtimes)

		// Try each available runtime
		for _, rt := range runtimes {
			t.Run(string(rt), func(t *testing.T) {
				runtime, err := factory.CreateRuntime(ctx, rt)
				require.NoError(t, err)

				// Verify runtime works
				err = runtime.Ping(ctx)
				assert.NoError(t, err)

				version, err := runtime.Version(ctx)
				assert.NoError(t, err)
				assert.NotEmpty(t, version.Version)

				t.Logf("Runtime %s version: %s", rt, version.Version)
			})
		}
	})
}

func TestIntegration_ErrorScenarios(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getIntegrationLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 60*time.Second)

	t.Run("invalid image", func(t *testing.T) {
		config := &ContainerConfig{
			Name:  testhelpers.RandomName("integration-error"),
			Image: "nonexistent-image-12345:latest",
			Cmd:   []string{"echo", "test"},
		}

		_, err := runtime.Create(ctx, config)
		assert.Error(t, err)
	})

	t.Run("stop non-existent container", func(t *testing.T) {
		err := runtime.Stop(ctx, "nonexistent-container-12345", 10*time.Second)
		assert.Error(t, err)
	})

	t.Run("remove non-existent container", func(t *testing.T) {
		err := runtime.Remove(ctx, "nonexistent-container-12345", false)
		assert.Error(t, err)
	})

	t.Run("inspect non-existent container", func(t *testing.T) {
		_, err := runtime.Inspect(ctx, "nonexistent-container-12345")
		assert.Error(t, err)
	})

	t.Run("logs from non-existent container", func(t *testing.T) {
		_, err := runtime.Logs(ctx, "nonexistent-container-12345", LogOptions{})
		assert.Error(t, err)
	})

	t.Run("stats from non-existent container", func(t *testing.T) {
		_, err := runtime.Stats(ctx, "nonexistent-container-12345")
		assert.Error(t, err)
	})
}

func TestIntegration_ConcurrentOperations(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getIntegrationLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 120*time.Second)

	t.Run("concurrent container creation", func(t *testing.T) {
		numContainers := 5
		containers := make([]*Container, numContainers)
		errors := make([]error, numContainers)

		// Create containers concurrently
		done := make(chan bool, numContainers)
		for i := 0; i < numContainers; i++ {
			go func(index int) {
				config := &ContainerConfig{
					Name:  testhelpers.RandomName("integration-concurrent"),
					Image: "alpine:latest",
					Cmd:   []string{"sleep", "30"},
				}

				containers[index], errors[index] = runtime.Create(ctx, config)
				done <- true
			}(i)
		}

		// Wait for all to complete
		for i := 0; i < numContainers; i++ {
			<-done
		}

		// Verify all succeeded
		for i := 0; i < numContainers; i++ {
			assert.NoError(t, errors[i])
			assert.NotNil(t, containers[i])
		}

		// Cleanup
		for _, container := range containers {
			if container != nil {
				_ = runtime.Remove(ctx, container.ID, true)
			}
		}
	})
}

func TestIntegration_ResourceConstraints(t *testing.T) {
	testhelpers.SkipIfNoDocker(t)

	runtime, err := NewDockerRuntime("", getIntegrationLogger())
	require.NoError(t, err)

	ctx := testhelpers.TestContext(t, 60*time.Second)

	t.Run("memory limit", func(t *testing.T) {
		config := &ContainerConfig{
			Name:  testhelpers.RandomName("integration-memory"),
			Image: "alpine:latest",
			Cmd:   []string{"sleep", "30"},
			Resources: &Resources{
				MemoryLimit: "128m",
			},
		}

		container, err := runtime.Create(ctx, config)
		require.NoError(t, err)
		defer func() {
			_ = runtime.Remove(ctx, container.ID, true)
		}()

		err = runtime.Start(ctx, container.ID)
		require.NoError(t, err)

		time.Sleep(2 * time.Second)

		stats, err := runtime.Stats(ctx, container.ID)
		require.NoError(t, err)

		// Memory limit should be set (approximately 128MB)
		assert.Greater(t, stats.MemoryLimit, uint64(100*1024*1024))
		assert.Less(t, stats.MemoryLimit, uint64(150*1024*1024))
	})

	t.Run("cpu limit", func(t *testing.T) {
		config := &ContainerConfig{
			Name:  testhelpers.RandomName("integration-cpu"),
			Image: "alpine:latest",
			Cmd:   []string{"sleep", "30"},
			Resources: &Resources{
				CPULimit: "0.5",
			},
		}

		container, err := runtime.Create(ctx, config)
		require.NoError(t, err)
		defer func() {
			_ = runtime.Remove(ctx, container.ID, true)
		}()

		err = runtime.Start(ctx, container.ID)
		require.NoError(t, err)

		// Verify container is running with CPU constraints
		inspected, err := runtime.Inspect(ctx, container.ID)
		require.NoError(t, err)
		assert.Equal(t, "running", inspected.State)
	})
}

// Made with Bob
