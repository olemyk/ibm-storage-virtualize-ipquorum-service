package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"
)

// DockerRuntime implements the Runtime interface for Docker
type DockerRuntime struct {
	client *client.Client
	logger *zap.Logger
}

// NewDockerRuntime creates a new Docker runtime client
func NewDockerRuntime(socketPath string, logger *zap.Logger) (*DockerRuntime, error) {
	var opts []client.Opt

	if socketPath != "" {
		opts = append(opts, client.WithHost("unix://"+socketPath))
	}
	opts = append(opts, client.WithAPIVersionNegotiation())

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerRuntime{
		client: cli,
		logger: logger,
	}, nil
}

// Name returns the runtime name
func (d *DockerRuntime) Name() string {
	return "Docker"
}

// Type returns the runtime type
func (d *DockerRuntime) Type() RuntimeType {
	return RuntimeTypeDocker
}

// Create creates a new container
func (d *DockerRuntime) Create(ctx context.Context, config *ContainerConfig) (*Container, error) {
	d.logger.Info("Creating Docker container", zap.String("name", config.Name), zap.String("image", config.Image))

	// Convert port mappings
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for _, pm := range config.Ports {
		port := nat.Port(fmt.Sprintf("%d/%s", pm.ContainerPort, pm.Protocol))
		if pm.Protocol == "" {
			port = nat.Port(fmt.Sprintf("%d/tcp", pm.ContainerPort))
		}
		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   pm.HostIP,
				HostPort: strconv.Itoa(pm.HostPort),
			},
		}
	}

	// Convert volume mounts
	binds := []string{}
	for _, vm := range config.Volumes {
		bind := fmt.Sprintf("%s:%s", vm.Source, vm.Destination)
		if vm.ReadOnly {
			bind += ":ro"
		}
		binds = append(binds, bind)
	}

	// Convert resources
	resources := container.Resources{}
	if config.Resources != nil {
		if config.Resources.MemoryLimit != "" {
			memBytes, err := parseMemory(config.Resources.MemoryLimit)
			if err == nil {
				resources.Memory = memBytes
			}
		}
		if config.Resources.CPULimit != "" {
			cpuQuota, err := parseCPU(config.Resources.CPULimit)
			if err == nil {
				resources.CPUQuota = cpuQuota
				resources.CPUPeriod = 100000 // Standard period
			}
		}
		if config.Resources.CPUShares > 0 {
			resources.CPUShares = config.Resources.CPUShares
		}
	}

	// Convert environment map to array format for Docker API
	var envArray []string
	if len(config.Env) > 0 {
		envArray = make([]string, 0, len(config.Env))
		for key, value := range config.Env {
			envArray = append(envArray, key+"="+value)
		}
	}

	// Create container config
	containerConfig := &container.Config{
		Image:        config.Image,
		Cmd:          config.Cmd,
		Entrypoint:   config.Entrypoint,
		Env:          envArray,
		Labels:       config.Labels,
		ExposedPorts: exposedPorts,
		WorkingDir:   config.WorkingDir,
		User:         config.User,
		Hostname:     config.Hostname,
	}

	// Create host config
	hostConfig := &container.HostConfig{
		Binds:        binds,
		PortBindings: portBindings,
		NetworkMode:  container.NetworkMode(config.NetworkMode),
		Resources:    resources,
	}

	// Set restart policy
	if config.RestartPolicy != "" {
		hostConfig.RestartPolicy = container.RestartPolicy{
			Name: container.RestartPolicyMode(config.RestartPolicy),
		}
		if config.RestartPolicy == string(RestartPolicyOnFailure) {
			hostConfig.RestartPolicy.MaximumRetryCount = 3
		}
	}

	// Create the container
	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, config.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	d.logger.Info("Docker container created", zap.String("id", resp.ID), zap.String("name", config.Name))

	return &Container{
		ID:      resp.ID,
		Name:    config.Name,
		Image:   config.Image,
		Status:  "created",
		State:   string(StateCreated),
		Created: time.Now(),
		Runtime: RuntimeTypeDocker,
	}, nil
}

// Start starts a container
func (d *DockerRuntime) Start(ctx context.Context, id string) error {
	d.logger.Info("Starting Docker container", zap.String("id", id))

	if err := d.client.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	d.logger.Info("Docker container started", zap.String("id", id))
	return nil
}

// Stop stops a container
func (d *DockerRuntime) Stop(ctx context.Context, id string, timeout time.Duration) error {
	d.logger.Info("Stopping Docker container", zap.String("id", id), zap.Duration("timeout", timeout))

	timeoutSec := int(timeout.Seconds())
	if err := d.client.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeoutSec}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	d.logger.Info("Docker container stopped", zap.String("id", id))
	return nil
}

// Restart restarts a container
func (d *DockerRuntime) Restart(ctx context.Context, id string, timeout time.Duration) error {
	d.logger.Info("Restarting Docker container", zap.String("id", id))

	timeoutSec := int(timeout.Seconds())
	if err := d.client.ContainerRestart(ctx, id, container.StopOptions{Timeout: &timeoutSec}); err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}

	d.logger.Info("Docker container restarted", zap.String("id", id))
	return nil
}

// Remove removes a container
func (d *DockerRuntime) Remove(ctx context.Context, id string, force bool) error {
	d.logger.Info("Removing Docker container", zap.String("id", id), zap.Bool("force", force))

	if err := d.client.ContainerRemove(ctx, id, container.RemoveOptions{Force: force}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	d.logger.Info("Docker container removed", zap.String("id", id))
	return nil
}

// Inspect inspects a container
func (d *DockerRuntime) Inspect(ctx context.Context, id string) (*Container, error) {
	info, err := d.client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	return d.convertContainer(&info), nil
}

// List lists containers
func (d *DockerRuntime) List(ctx context.Context, filterMap map[string]string) ([]*Container, error) {
	// Convert filter map to Docker filters
	args := filters.NewArgs()
	for k, v := range filterMap {
		args.Add(k, v)
	}

	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: args,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]*Container, 0, len(containers))
	for _, c := range containers {
		result = append(result, &Container{
			ID:      c.ID,
			Name:    strings.TrimPrefix(c.Names[0], "/"),
			Image:   c.Image,
			Status:  c.Status,
			State:   c.State,
			Created: time.Unix(c.Created, 0),
			Labels:  c.Labels,
			Runtime: RuntimeTypeDocker,
		})
	}

	return result, nil
}

// ListImages lists all images
func (d *DockerRuntime) ListImages(ctx context.Context) ([]*Image, error) {
	images, err := d.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	result := make([]*Image, 0, len(images))
	for _, img := range images {
		result = append(result, &Image{
			ID:          img.ID,
			RepoTags:    img.RepoTags,
			RepoDigests: img.RepoDigests,
			Size:        img.Size,
			Created:     time.Unix(img.Created, 0),
			Labels:      img.Labels,
		})
	}

	return result, nil
}

// Logs retrieves container logs
func (d *DockerRuntime) Logs(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     opts.Follow,
		Timestamps: opts.Timestamps,
		Details:    opts.Details,
	}

	if opts.Tail > 0 {
		options.Tail = strconv.Itoa(opts.Tail)
	}

	if !opts.Since.IsZero() {
		options.Since = opts.Since.Format(time.RFC3339)
	}

	if !opts.Until.IsZero() {
		options.Until = opts.Until.Format(time.RFC3339)
	}

	logs, err := d.client.ContainerLogs(ctx, id, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	return logs, nil
}

// Stats retrieves container statistics
func (d *DockerRuntime) Stats(ctx context.Context, id string) (*Stats, error) {
	stats, err := d.client.ContainerStats(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()

	var v container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	return d.convertStats(&v), nil
}

// Ping checks if the Docker daemon is accessible
func (d *DockerRuntime) Ping(ctx context.Context) error {
	_, err := d.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("Docker ping failed: %w", err)
	}
	return nil
}

// Version returns Docker version information
func (d *DockerRuntime) Version(ctx context.Context) (*Version, error) {
	v, err := d.client.ServerVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker version: %w", err)
	}

	return &Version{
		Version:       v.Version,
		APIVersion:    v.APIVersion,
		MinAPIVersion: v.MinAPIVersion,
		GoVersion:     v.GoVersion,
		GitCommit:     v.GitCommit,
		BuildTime:     v.BuildTime,
		OS:            v.Os,
		Arch:          v.Arch,
	}, nil
}

// Info returns Docker system information
func (d *DockerRuntime) Info(ctx context.Context) (*RuntimeInfo, error) {
	info, err := d.client.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker info: %w", err)
	}

	version, _ := d.Version(ctx)

	return &RuntimeInfo{
		Name:              "Docker",
		Type:              string(RuntimeTypeDocker),
		Version:           version.Version,
		APIVersion:        version.APIVersion,
		OS:                info.OperatingSystem,
		Arch:              info.Architecture,
		KernelVersion:     info.KernelVersion,
		ContainersRunning: info.ContainersRunning,
		ContainersPaused:  info.ContainersPaused,
		ContainersStopped: info.ContainersStopped,
		Images:            info.Images,
		Driver:            info.Driver,
		MemoryTotal:       info.MemTotal,
		CPUs:              info.NCPU,
	}, nil
}

// Helper functions

func (d *DockerRuntime) convertContainer(info *types.ContainerJSON) *Container {
	c := &Container{
		ID:      info.ID,
		Name:    strings.TrimPrefix(info.Name, "/"),
		Image:   info.Config.Image,
		Status:  info.State.Status,
		State:   info.State.Status,
		Labels:  info.Config.Labels,
		Runtime: RuntimeTypeDocker,
	}

	// Parse Created timestamp
	if t, err := time.Parse(time.RFC3339Nano, info.Created); err == nil {
		c.Created = t
	}

	// Parse Started timestamp
	if info.State.Running && info.State.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, info.State.StartedAt); err == nil {
			c.Started = t
		}
	}

	// Parse Finished timestamp
	if info.State.FinishedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, info.State.FinishedAt); err == nil {
			c.Finished = t
		}
	}

	c.ExitCode = info.State.ExitCode

	// Convert port mappings
	for port, bindings := range info.NetworkSettings.Ports {
		for _, binding := range bindings {
			hostPort, _ := strconv.Atoi(binding.HostPort)
			containerPort, _ := strconv.Atoi(port.Port())
			c.Ports = append(c.Ports, PortMapping{
				HostIP:        binding.HostIP,
				HostPort:      hostPort,
				ContainerPort: containerPort,
				Protocol:      port.Proto(),
			})
		}
	}

	return c
}

func (d *DockerRuntime) convertStats(stats *container.StatsResponse) *Stats {
	// Calculate CPU percentage
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	cpuPercent := 0.0
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	// Calculate memory percentage
	memPercent := 0.0
	if stats.MemoryStats.Limit > 0 {
		memPercent = float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
	}

	// Calculate network I/O
	var networkRx, networkTx uint64
	for _, network := range stats.Networks {
		networkRx += network.RxBytes
		networkTx += network.TxBytes
	}

	// Calculate block I/O
	var blockRead, blockWrite uint64
	for _, bio := range stats.BlkioStats.IoServiceBytesRecursive {
		if bio.Op == "Read" {
			blockRead += bio.Value
		} else if bio.Op == "Write" {
			blockWrite += bio.Value
		}
	}

	return &Stats{
		CPUPercent:    cpuPercent,
		MemoryUsage:   stats.MemoryStats.Usage,
		MemoryLimit:   stats.MemoryStats.Limit,
		MemoryPercent: memPercent,
		NetworkRx:     networkRx,
		NetworkTx:     networkTx,
		BlockRead:     blockRead,
		BlockWrite:    blockWrite,
		PIDs:          stats.PidsStats.Current,
	}
}

// parseMemory converts memory string to bytes (e.g., "512m" -> 536870912)
func parseMemory(mem string) (int64, error) {
	mem = strings.ToLower(strings.TrimSpace(mem))
	if mem == "" {
		return 0, fmt.Errorf("empty memory value")
	}

	multiplier := int64(1)
	if strings.HasSuffix(mem, "k") {
		multiplier = 1024
		mem = strings.TrimSuffix(mem, "k")
	} else if strings.HasSuffix(mem, "m") {
		multiplier = 1024 * 1024
		mem = strings.TrimSuffix(mem, "m")
	} else if strings.HasSuffix(mem, "g") {
		multiplier = 1024 * 1024 * 1024
		mem = strings.TrimSuffix(mem, "g")
	}

	value, err := strconv.ParseInt(mem, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid memory value: %w", err)
	}

	return value * multiplier, nil
}

// parseCPU converts CPU string to quota (e.g., "1.5" -> 150000)
func parseCPU(cpu string) (int64, error) {
	cpu = strings.TrimSpace(cpu)
	if cpu == "" {
		return 0, fmt.Errorf("empty CPU value")
	}

	value, err := strconv.ParseFloat(cpu, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid CPU value: %w", err)
	}

	// Convert to quota (period is 100000 microseconds)
	return int64(value * 100000), nil
}

// Made with Bob
