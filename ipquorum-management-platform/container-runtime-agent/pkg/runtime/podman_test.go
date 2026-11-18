package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Test helper to create a test Podman runtime with mock server
func setupPodmanTest(t *testing.T, handler http.HandlerFunc) (*PodmanRuntime, *httptest.Server) {
	server := httptest.NewServer(handler)
	logger, _ := zap.NewDevelopment()

	runtime, err := NewPodmanRuntimeHTTP(server.URL, logger)
	require.NoError(t, err)

	return runtime, server
}

func TestNewPodmanRuntime(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("Unix socket", func(t *testing.T) {
		runtime, err := NewPodmanRuntime("/run/podman/podman.sock", logger)
		assert.NoError(t, err)
		assert.NotNil(t, runtime)
		assert.Equal(t, "Podman", runtime.Name())
		assert.Equal(t, RuntimeTypePodman, runtime.Type())
	})

	t.Run("HTTP", func(t *testing.T) {
		runtime, err := NewPodmanRuntimeHTTP("http://localhost:8080", logger)
		assert.NoError(t, err)
		assert.NotNil(t, runtime)
	})
}

func TestPodmanRuntime_Name(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	runtime, _ := NewPodmanRuntime("/run/podman/podman.sock", logger)

	assert.Equal(t, "Podman", runtime.Name())
}

func TestPodmanRuntime_Type(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	runtime, _ := NewPodmanRuntime("/run/podman/podman.sock", logger)

	assert.Equal(t, RuntimeTypePodman, runtime.Type())
}

func TestPodmanRuntime_Ping(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/_ping")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		defer server.Close()

		err := runtime.Ping(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Failure", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		})
		defer server.Close()

		err := runtime.Ping(context.Background())
		assert.Error(t, err)
	})
}

func TestPodmanRuntime_Version(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/version")

			response := map[string]interface{}{
				"Version":       "4.5.0",
				"ApiVersion":    "4.5.0",
				"MinAPIVersion": "4.0.0",
				"GoVersion":     "go1.20",
				"GitCommit":     "abc123",
				"Built":         "2024-01-01",
				"OsArch":        "linux/amd64",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})
		defer server.Close()

		version, err := runtime.Version(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, version)
		assert.Equal(t, "4.5.0", version.Version)
		assert.Equal(t, "4.5.0", version.APIVersion)
		assert.Equal(t, "4.0.0", version.MinAPIVersion)
		assert.Equal(t, "go1.20", version.GoVersion)
		assert.Equal(t, "abc123", version.GitCommit)
		assert.Equal(t, "2024-01-01", version.BuildTime)
		assert.Equal(t, "linux", version.OS)
		assert.Equal(t, "amd64", version.Arch)
	})

	t.Run("Error", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		defer server.Close()

		version, err := runtime.Version(context.Background())
		assert.Error(t, err)
		assert.Nil(t, version)
	})
}

func TestPodmanRuntime_Info(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v4.0.0/libpod/version" {
				response := map[string]interface{}{
					"Version":    "4.5.0",
					"ApiVersion": "4.5.0",
					"OsArch":     "linux/amd64",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}

			assert.Contains(t, r.URL.Path, "/info")

			response := map[string]interface{}{
				"host": map[string]interface{}{
					"kernel":   "5.15.0",
					"os":       "linux",
					"arch":     "amd64",
					"cpus":     float64(8),
					"memTotal": float64(16000000000),
				},
				"store": map[string]interface{}{
					"graphDriverName": "overlay2",
					"containerStore": map[string]interface{}{
						"number": float64(5),
					},
					"imageStore": map[string]interface{}{
						"number": float64(10),
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})
		defer server.Close()

		info, err := runtime.Info(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "Podman", info.Name)
		assert.Equal(t, "podman", info.Type)
		assert.Equal(t, "5.15.0", info.KernelVersion)
		assert.Equal(t, 8, info.CPUs)
		assert.Equal(t, int64(16000000000), info.MemoryTotal)
		assert.Equal(t, "overlay2", info.Driver)
		assert.Equal(t, 10, info.Images)
	})
}

func TestPodmanRuntime_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/containers/create")

			var payload map[string]interface{}
			json.NewDecoder(r.Body).Decode(&payload)

			assert.Equal(t, "test-container", payload["name"])
			assert.Equal(t, "alpine:latest", payload["image"])

			response := map[string]interface{}{
				"Id": "container123",
			}

			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})
		defer server.Close()

		config := &ContainerConfig{
			Name:  "test-container",
			Image: "alpine:latest",
			Cmd:   []string{"sleep", "300"},
			Labels: map[string]string{
				"test": "true",
			},
		}

		container, err := runtime.Create(context.Background(), config)
		assert.NoError(t, err)
		assert.NotNil(t, container)
		assert.Equal(t, "container123", container.ID)
		assert.Equal(t, "test-container", container.Name)
		assert.Equal(t, "alpine:latest", container.Image)
		assert.Equal(t, "created", container.Status)
	})

	t.Run("With ports and volumes", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			var payload map[string]interface{}
			json.NewDecoder(r.Body).Decode(&payload)

			assert.NotNil(t, payload["portmappings"])
			assert.NotNil(t, payload["mounts"])

			response := map[string]interface{}{
				"Id": "container456",
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		})
		defer server.Close()

		config := &ContainerConfig{
			Name:  "web-server",
			Image: "nginx:latest",
			Ports: []PortMapping{
				{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
			},
			Volumes: []VolumeMount{
				{Source: "/host/data", Destination: "/data", ReadOnly: false},
			},
		}

		container, err := runtime.Create(context.Background(), config)
		assert.NoError(t, err)
		assert.Equal(t, "container456", container.ID)
	})

	t.Run("Error", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid configuration"))
		})
		defer server.Close()

		config := &ContainerConfig{
			Name:  "bad-container",
			Image: "invalid:image",
		}

		container, err := runtime.Create(context.Background(), config)
		assert.Error(t, err)
		assert.Nil(t, container)
	})
}

func TestPodmanRuntime_Start(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/containers/container123/start")
			w.WriteHeader(http.StatusNoContent)
		})
		defer server.Close()

		err := runtime.Start(context.Background(), "container123")
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Container not found"))
		})
		defer server.Close()

		err := runtime.Start(context.Background(), "nonexistent")
		assert.Error(t, err)
	})
}

func TestPodmanRuntime_Stop(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/containers/container123/stop")
			assert.Contains(t, r.URL.RawQuery, "timeout=10")
			w.WriteHeader(http.StatusNoContent)
		})
		defer server.Close()

		err := runtime.Stop(context.Background(), "container123", 10*time.Second)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		defer server.Close()

		err := runtime.Stop(context.Background(), "container123", 10*time.Second)
		assert.Error(t, err)
	})
}

func TestPodmanRuntime_Restart(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/containers/container123/restart")
			w.WriteHeader(http.StatusNoContent)
		})
		defer server.Close()

		err := runtime.Restart(context.Background(), "container123", 10*time.Second)
		assert.NoError(t, err)
	})
}

func TestPodmanRuntime_Remove(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DELETE", r.Method)
			assert.Contains(t, r.URL.Path, "/containers/container123")
			assert.Contains(t, r.URL.RawQuery, "force=true")
			w.WriteHeader(http.StatusNoContent)
		})
		defer server.Close()

		err := runtime.Remove(context.Background(), "container123", true)
		assert.NoError(t, err)
	})

	t.Run("Without force", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.RawQuery, "force=false")
			w.WriteHeader(http.StatusNoContent)
		})
		defer server.Close()

		err := runtime.Remove(context.Background(), "container123", false)
		assert.NoError(t, err)
	})
}

func TestPodmanRuntime_Inspect(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/containers/container123/json")

			response := map[string]interface{}{
				"Id":      "container123",
				"Name":    "/test-container",
				"Image":   "alpine:latest",
				"Status":  "running",
				"State":   "running",
				"Created": time.Now().Format(time.RFC3339),
				"Labels": map[string]interface{}{
					"app": "test",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})
		defer server.Close()

		container, err := runtime.Inspect(context.Background(), "container123")
		assert.NoError(t, err)
		assert.NotNil(t, container)
		assert.Equal(t, "container123", container.ID)
		assert.Equal(t, "test-container", container.Name)
		assert.Equal(t, "alpine:latest", container.Image)
		assert.Equal(t, "running", container.Status)
	})
}

func TestPodmanRuntime_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/containers/json")
			assert.Contains(t, r.URL.RawQuery, "all=true")

			response := []map[string]interface{}{
				{
					"Id":     "container1",
					"Name":   "/container-1",
					"Image":  "alpine:latest",
					"Status": "running",
					"State":  "running",
				},
				{
					"Id":     "container2",
					"Name":   "/container-2",
					"Image":  "nginx:latest",
					"Status": "exited",
					"State":  "exited",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})
		defer server.Close()

		containers, err := runtime.List(context.Background(), nil)
		assert.NoError(t, err)
		assert.Len(t, containers, 2)
		assert.Equal(t, "container1", containers[0].ID)
		assert.Equal(t, "container-1", containers[0].Name)
		assert.Equal(t, "container2", containers[1].ID)
	})

	t.Run("With filters", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.RawQuery, "filters")

			response := []map[string]interface{}{}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})
		defer server.Close()

		filters := map[string]string{
			"status": "running",
		}

		containers, err := runtime.List(context.Background(), filters)
		assert.NoError(t, err)
		assert.Len(t, containers, 0)
	})
}

func TestPodmanRuntime_Logs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/containers/container123/logs")
			assert.Contains(t, r.URL.RawQuery, "stdout=true")
			assert.Contains(t, r.URL.RawQuery, "stderr=true")

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Container log output\n"))
		})
		defer server.Close()

		opts := LogOptions{
			Follow:     false,
			Timestamps: true,
			Tail:       100,
		}

		logs, err := runtime.Logs(context.Background(), "container123", opts)
		assert.NoError(t, err)
		assert.NotNil(t, logs)
		logs.Close()
	})
}

func TestPodmanRuntime_Stats(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		runtime, server := setupPodmanTest(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/containers/container123/stats")
			assert.Contains(t, r.URL.RawQuery, "stream=false")

			response := map[string]interface{}{
				"cpu_stats": map[string]interface{}{
					"cpu_usage": map[string]interface{}{
						"total_usage": float64(1000000000),
					},
				},
				"memory_stats": map[string]interface{}{
					"usage": float64(104857600),
					"limit": float64(1073741824),
				},
				"networks": map[string]interface{}{
					"eth0": map[string]interface{}{
						"rx_bytes": float64(1024),
						"tx_bytes": float64(2048),
					},
				},
				"pids_stats": map[string]interface{}{
					"current": float64(5),
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})
		defer server.Close()

		stats, err := runtime.Stats(context.Background(), "container123")
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, uint64(104857600), stats.MemoryUsage)
		assert.Equal(t, uint64(1073741824), stats.MemoryLimit)
		assert.Equal(t, uint64(1024), stats.NetworkRx)
		assert.Equal(t, uint64(2048), stats.NetworkTx)
		assert.Equal(t, uint64(5), stats.PIDs)
	})
}

func TestPodmanRuntime_ConvertContainer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	runtime, _ := NewPodmanRuntime("/run/podman/podman.sock", logger)

	data := map[string]interface{}{
		"Id":      "abc123",
		"Name":    "/my-container",
		"Image":   "alpine:latest",
		"Status":  "running",
		"State":   "running",
		"Created": time.Now().Format(time.RFC3339),
		"Labels": map[string]interface{}{
			"app":     "test",
			"version": "1.0",
		},
		"Ports": []interface{}{
			map[string]interface{}{
				"host_ip":        "0.0.0.0",
				"host_port":      float64(8080),
				"container_port": float64(80),
				"protocol":       "tcp",
			},
		},
	}

	container := runtime.convertContainer(data)
	assert.Equal(t, "abc123", container.ID)
	assert.Equal(t, "my-container", container.Name)
	assert.Equal(t, "alpine:latest", container.Image)
	assert.Equal(t, "running", container.Status)
	assert.Equal(t, "running", container.State)
	assert.Equal(t, RuntimeTypePodman, container.Runtime)
	assert.Len(t, container.Labels, 2)
	assert.Equal(t, "test", container.Labels["app"])
	assert.Len(t, container.Ports, 1)
	assert.Equal(t, 8080, container.Ports[0].HostPort)
}

func TestPodmanRuntime_ConvertStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	runtime, _ := NewPodmanRuntime("/run/podman/podman.sock", logger)

	data := map[string]interface{}{
		"cpu_stats": map[string]interface{}{
			"cpu_usage": map[string]interface{}{
				"total_usage": float64(2000000000),
			},
		},
		"memory_stats": map[string]interface{}{
			"usage": float64(209715200),
			"limit": float64(2147483648),
		},
		"networks": map[string]interface{}{
			"eth0": map[string]interface{}{
				"rx_bytes": float64(4096),
				"tx_bytes": float64(8192),
			},
		},
		"blkio_stats": map[string]interface{}{
			"io_service_bytes_recursive": []interface{}{
				map[string]interface{}{
					"op":    "Read",
					"value": float64(1024),
				},
				map[string]interface{}{
					"op":    "Write",
					"value": float64(2048),
				},
			},
		},
		"pids_stats": map[string]interface{}{
			"current": float64(10),
		},
	}

	stats := runtime.convertStats(data)
	assert.Equal(t, uint64(209715200), stats.MemoryUsage)
	assert.Equal(t, uint64(2147483648), stats.MemoryLimit)
	assert.InDelta(t, 9.76, stats.MemoryPercent, 0.1)
	assert.Equal(t, uint64(4096), stats.NetworkRx)
	assert.Equal(t, uint64(8192), stats.NetworkTx)
	assert.Equal(t, uint64(1024), stats.BlockRead)
	assert.Equal(t, uint64(2048), stats.BlockWrite)
	assert.Equal(t, uint64(10), stats.PIDs)
}

// Made with Bob
