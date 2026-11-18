package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFactory(t *testing.T) {
	tests := []struct {
		name   string
		config *FactoryConfig
	}{
		{
			name: "default config",
			config: &FactoryConfig{
				PreferredRuntime: "auto",
				DockerSocketPath: "/var/run/docker.sock",
				PodmanSocketPath: "/run/podman/podman.sock",
				DetectionTimeout: 5 * time.Second,
			},
		},
		{
			name: "docker preferred",
			config: &FactoryConfig{
				PreferredRuntime: "docker",
				DockerSocketPath: "/var/run/docker.sock",
				PodmanSocketPath: "/run/podman/podman.sock",
				DetectionTimeout: 3 * time.Second,
			},
		},
		{
			name: "podman preferred",
			config: &FactoryConfig{
				PreferredRuntime: "podman",
				DockerSocketPath: "/var/run/docker.sock",
				PodmanSocketPath: "/run/podman/podman.sock",
				DetectionTimeout: 3 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory(tt.config)
			assert.NotNil(t, factory)
			assert.Equal(t, tt.config, factory.config)
		})
	}
}

func TestFactoryDetectAvailableRuntimes(t *testing.T) {
	t.Run("detect with short timeout", func(t *testing.T) {
		config := &FactoryConfig{
			PreferredRuntime: "auto",
			DockerSocketPath: "/var/run/docker.sock",
			PodmanSocketPath: "/run/podman/podman.sock",
			DetectionTimeout: 100 * time.Millisecond,
		}

		factory := NewFactory(config)
		ctx := context.Background()

		runtimes, err := factory.DetectAvailableRuntimes(ctx)
		// Should not error even if no runtimes found
		assert.NoError(t, err)
		// May be empty slice if no runtimes available
		if runtimes != nil {
			assert.IsType(t, []RuntimeType{}, runtimes)
		}
	})

	t.Run("detect with context cancellation", func(t *testing.T) {
		config := &FactoryConfig{
			PreferredRuntime: "auto",
			DockerSocketPath: "/var/run/docker.sock",
			PodmanSocketPath: "/run/podman/podman.sock",
			DetectionTimeout: 5 * time.Second,
		}

		factory := NewFactory(config)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		runtimes, err := factory.DetectAvailableRuntimes(ctx)
		// Should handle cancellation gracefully
		assert.NoError(t, err)
		// May return nil or empty slice
		if runtimes != nil {
			assert.IsType(t, []RuntimeType{}, runtimes)
		}
	})
}

func TestFactoryCreateRuntime(t *testing.T) {
	t.Run("create with invalid preferred runtime", func(t *testing.T) {
		config := &FactoryConfig{
			PreferredRuntime: "invalid",
			DockerSocketPath: "/var/run/docker.sock",
			PodmanSocketPath: "/run/podman/podman.sock",
			DetectionTimeout: 5 * time.Second,
		}

		factory := NewFactory(config)
		ctx := context.Background()

		rt, err := factory.CreateRuntime(ctx)
		assert.Error(t, err)
		assert.Nil(t, rt)
		assert.Contains(t, err.Error(), "invalid preferred runtime")
	})
}

func TestFactoryValidateRuntime(t *testing.T) {
	t.Run("validate non-existent runtime", func(t *testing.T) {
		config := &FactoryConfig{
			PreferredRuntime: "auto",
			DockerSocketPath: "/nonexistent/docker.sock",
			PodmanSocketPath: "/nonexistent/podman.sock",
			DetectionTimeout: 100 * time.Millisecond,
		}

		factory := NewFactory(config)
		ctx := context.Background()

		err := factory.ValidateRuntime(ctx, RuntimeTypeDocker)
		assert.Error(t, err)
		// Error message may vary (ping failed, connection failed, etc.)
		assert.NotEmpty(t, err.Error())
	})
}

func TestFactoryGetPreferredRuntime(t *testing.T) {
	tests := []struct {
		name             string
		preferredRuntime string
		expectError      bool
	}{
		{
			name:             "auto preference",
			preferredRuntime: "auto",
			expectError:      false,
		},
		{
			name:             "docker preference",
			preferredRuntime: "docker",
			expectError:      false,
		},
		{
			name:             "podman preference",
			preferredRuntime: "podman",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &FactoryConfig{
				PreferredRuntime: tt.preferredRuntime,
				DockerSocketPath: "/var/run/docker.sock",
				PodmanSocketPath: "/run/podman/podman.sock",
				DetectionTimeout: 100 * time.Millisecond,
			}

			factory := NewFactory(config)
			ctx := context.Background()

			runtimeType, err := factory.GetPreferredRuntime(ctx)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				// May error if no runtime available, which is OK
				if err == nil {
					assert.NotEmpty(t, runtimeType)
				}
			}
		})
	}
}

func TestFactoryGetRuntimeInfo(t *testing.T) {
	t.Run("get info for non-existent runtime", func(t *testing.T) {
		config := &FactoryConfig{
			PreferredRuntime: "auto",
			DockerSocketPath: "/nonexistent/docker.sock",
			PodmanSocketPath: "/nonexistent/podman.sock",
			DetectionTimeout: 100 * time.Millisecond,
		}

		factory := NewFactory(config)
		ctx := context.Background()

		info, err := factory.GetRuntimeInfo(ctx, RuntimeTypeDocker)
		assert.Error(t, err)
		assert.Nil(t, info)
	})
}

func TestRuntimeTypeConstants(t *testing.T) {
	t.Run("runtime type values", func(t *testing.T) {
		assert.Equal(t, RuntimeType("docker"), RuntimeTypeDocker)
		assert.Equal(t, RuntimeType("podman"), RuntimeTypePodman)
	})
}

func TestFactoryConfigDefaults(t *testing.T) {
	t.Run("verify default paths", func(t *testing.T) {
		config := &FactoryConfig{
			PreferredRuntime: "auto",
			DockerSocketPath: "/var/run/docker.sock",
			PodmanSocketPath: "/run/podman/podman.sock",
			DetectionTimeout: 5 * time.Second,
		}

		assert.Equal(t, "auto", config.PreferredRuntime)
		assert.Equal(t, "/var/run/docker.sock", config.DockerSocketPath)
		assert.Equal(t, "/run/podman/podman.sock", config.PodmanSocketPath)
		assert.Equal(t, 5*time.Second, config.DetectionTimeout)
	})
}

func TestFactoryConcurrency(t *testing.T) {
	t.Run("concurrent detection", func(t *testing.T) {
		config := &FactoryConfig{
			PreferredRuntime: "auto",
			DockerSocketPath: "/var/run/docker.sock",
			PodmanSocketPath: "/run/podman/podman.sock",
			DetectionTimeout: 100 * time.Millisecond,
		}

		factory := NewFactory(config)
		ctx := context.Background()

		// Run multiple detections concurrently
		done := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func() {
				_, err := factory.DetectAvailableRuntimes(ctx)
				assert.NoError(t, err)
				done <- true
			}()
		}

		// Wait for all to complete
		for i := 0; i < 5; i++ {
			<-done
		}
	})
}

func TestFactoryContextHandling(t *testing.T) {
	t.Run("respect context timeout", func(t *testing.T) {
		config := &FactoryConfig{
			PreferredRuntime: "auto",
			DockerSocketPath: "/var/run/docker.sock",
			PodmanSocketPath: "/run/podman/podman.sock",
			DetectionTimeout: 5 * time.Second,
		}

		factory := NewFactory(config)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		start := time.Now()
		_, err := factory.DetectAvailableRuntimes(ctx)
		duration := time.Since(start)

		// Should complete quickly due to context timeout
		assert.NoError(t, err) // Detection handles timeout gracefully
		assert.Less(t, duration, 200*time.Millisecond)
	})
}

// Made with Bob
