package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// PodmanRuntime implements the Runtime interface for Podman
type PodmanRuntime struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewPodmanRuntime creates a new Podman runtime client
func NewPodmanRuntime(socketPath string, logger *zap.Logger) (*PodmanRuntime, error) {
	// Create HTTP client with Unix socket transport
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	return &PodmanRuntime{
		baseURL: "http://localhost/v4.0.0/libpod",
		client:  client,
		logger:  logger,
	}, nil
}

// NewPodmanRuntimeHTTP creates a new Podman runtime client using HTTP
func NewPodmanRuntimeHTTP(apiURL string, logger *zap.Logger) (*PodmanRuntime, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &PodmanRuntime{
		baseURL: apiURL + "/v4.0.0/libpod",
		client:  client,
		logger:  logger,
	}, nil
}

// Name returns the runtime name
func (p *PodmanRuntime) Name() string {
	return "Podman"
}

// Type returns the runtime type
func (p *PodmanRuntime) Type() RuntimeType {
	return RuntimeTypePodman
}

// Create creates a new container
func (p *PodmanRuntime) Create(ctx context.Context, config *ContainerConfig) (*Container, error) {
	p.logger.Info("Creating Podman container", zap.String("name", config.Name), zap.String("image", config.Image))

	// Build Podman-specific container config
	payload := map[string]interface{}{
		"name":  config.Name,
		"image": config.Image,
	}

	if len(config.Cmd) > 0 {
		payload["command"] = config.Cmd
	}

	if len(config.Entrypoint) > 0 {
		payload["entrypoint"] = config.Entrypoint
	}

	if len(config.Env) > 0 {
		// Podman libpod API expects env as map[string]string
		payload["env"] = config.Env
	}

	if len(config.Labels) > 0 {
		payload["labels"] = config.Labels
	}

	if config.WorkingDir != "" {
		payload["work_dir"] = config.WorkingDir
	}

	if config.User != "" {
		payload["user"] = config.User
	}

	if config.Hostname != "" {
		payload["hostname"] = config.Hostname
	}

	// Port mappings
	if len(config.Ports) > 0 {
		portMappings := make([]map[string]interface{}, 0, len(config.Ports))
		for _, pm := range config.Ports {
			protocol := pm.Protocol
			if protocol == "" {
				protocol = "tcp"
			}
			portMappings = append(portMappings, map[string]interface{}{
				"host_ip":        pm.HostIP,
				"host_port":      pm.HostPort,
				"container_port": pm.ContainerPort,
				"protocol":       protocol,
			})
		}
		payload["portmappings"] = portMappings
	}

	// Volume mounts
	if len(config.Volumes) > 0 {
		mounts := make([]map[string]interface{}, 0, len(config.Volumes))
		for _, vm := range config.Volumes {
			mount := map[string]interface{}{
				"source":      vm.Source,
				"destination": vm.Destination,
				"type":        "bind",
			}
			if vm.ReadOnly {
				mount["options"] = []string{"ro"}
			}
			mounts = append(mounts, mount)
		}
		payload["mounts"] = mounts
	}

	// Network mode
	if config.NetworkMode != "" {
		payload["netns"] = map[string]interface{}{
			"nsmode": config.NetworkMode,
		}
	}

	// Restart policy
	if config.RestartPolicy != "" {
		payload["restart_policy"] = config.RestartPolicy
	}

	// Resources
	if config.Resources != nil {
		resourceConfig := make(map[string]interface{})

		if config.Resources.MemoryLimit != "" {
			memBytes, err := parseMemory(config.Resources.MemoryLimit)
			if err == nil {
				resourceConfig["memory"] = memBytes
			}
		}

		if config.Resources.CPULimit != "" {
			cpuQuota, err := parseCPU(config.Resources.CPULimit)
			if err == nil {
				resourceConfig["cpu_quota"] = cpuQuota
				resourceConfig["cpu_period"] = 100000
			}
		}

		if config.Resources.CPUShares > 0 {
			resourceConfig["cpu_shares"] = config.Resources.CPUShares
		}

		if len(resourceConfig) > 0 {
			payload["resource_limits"] = resourceConfig
		}
	}

	// Make API request
	resp, err := p.post(ctx, "/containers/create", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create container: %s - %s", resp.Status, string(body))
	}

	var result struct {
		ID string `json:"Id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	p.logger.Info("Podman container created", zap.String("id", result.ID), zap.String("name", config.Name))

	return &Container{
		ID:      result.ID,
		Name:    config.Name,
		Image:   config.Image,
		Status:  "created",
		State:   string(StateCreated),
		Created: time.Now(),
		Runtime: RuntimeTypePodman,
	}, nil
}

// Start starts a container
func (p *PodmanRuntime) Start(ctx context.Context, id string) error {
	p.logger.Info("Starting Podman container", zap.String("id", id))

	resp, err := p.post(ctx, fmt.Sprintf("/containers/%s/start", id), nil)
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to start container: %s - %s", resp.Status, string(body))
	}

	p.logger.Info("Podman container started", zap.String("id", id))
	return nil
}

// Stop stops a container
func (p *PodmanRuntime) Stop(ctx context.Context, id string, timeout time.Duration) error {
	p.logger.Info("Stopping Podman container", zap.String("id", id), zap.Duration("timeout", timeout))

	timeoutSec := int(timeout.Seconds())
	endpoint := fmt.Sprintf("/containers/%s/stop?timeout=%d", id, timeoutSec)

	resp, err := p.post(ctx, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to stop container: %s - %s", resp.Status, string(body))
	}

	p.logger.Info("Podman container stopped", zap.String("id", id))
	return nil
}

// Restart restarts a container
func (p *PodmanRuntime) Restart(ctx context.Context, id string, timeout time.Duration) error {
	p.logger.Info("Restarting Podman container", zap.String("id", id))

	timeoutSec := int(timeout.Seconds())
	endpoint := fmt.Sprintf("/containers/%s/restart?timeout=%d", id, timeoutSec)

	resp, err := p.post(ctx, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to restart container: %s - %s", resp.Status, string(body))
	}

	p.logger.Info("Podman container restarted", zap.String("id", id))
	return nil
}

// Remove removes a container
func (p *PodmanRuntime) Remove(ctx context.Context, id string, force bool) error {
	p.logger.Info("Removing Podman container", zap.String("id", id), zap.Bool("force", force))

	endpoint := fmt.Sprintf("/containers/%s?force=%t", id, force)

	resp, err := p.delete(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to remove container: %s - %s", resp.Status, string(body))
	}

	p.logger.Info("Podman container removed", zap.String("id", id))
	return nil
}

// Inspect inspects a container
func (p *PodmanRuntime) Inspect(ctx context.Context, id string) (*Container, error) {
	resp, err := p.get(ctx, fmt.Sprintf("/containers/%s/json", id))
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to inspect container: %s - %s", resp.Status, string(body))
	}

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.convertContainer(info), nil
}

// List lists containers
func (p *PodmanRuntime) List(ctx context.Context, filterMap map[string]string) ([]*Container, error) {
	// Build query parameters
	query := url.Values{}
	query.Set("all", "true")

	// Add filters - Podman API expects map[string][]string format
	if len(filterMap) > 0 {
		// Convert map[string]string to map[string][]string
		filters := make(map[string][]string)
		for k, v := range filterMap {
			filters[k] = []string{v}
		}
		filtersJSON, _ := json.Marshal(filters)
		query.Set("filters", string(filtersJSON))
	}

	endpoint := "/containers/json?" + query.Encode()

	resp, err := p.get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list containers: %s - %s", resp.Status, string(body))
	}

	var containers []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := make([]*Container, 0, len(containers))
	for _, c := range containers {
		result = append(result, p.convertContainer(c))
	}

	return result, nil
}

// ListImages lists all images
func (p *PodmanRuntime) ListImages(ctx context.Context) ([]*Image, error) {
	resp, err := p.get(ctx, "/images/json")
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list images: %s - %s", resp.Status, string(body))
	}

	var images []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&images); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := make([]*Image, 0, len(images))
	for _, img := range images {
		image := &Image{}

		if id, ok := img["Id"].(string); ok {
			image.ID = id
		}

		if repoTags, ok := img["RepoTags"].([]interface{}); ok {
			image.RepoTags = make([]string, 0, len(repoTags))
			for _, tag := range repoTags {
				if tagStr, ok := tag.(string); ok {
					image.RepoTags = append(image.RepoTags, tagStr)
				}
			}
		}

		if repoDigests, ok := img["RepoDigests"].([]interface{}); ok {
			image.RepoDigests = make([]string, 0, len(repoDigests))
			for _, digest := range repoDigests {
				if digestStr, ok := digest.(string); ok {
					image.RepoDigests = append(image.RepoDigests, digestStr)
				}
			}
		}

		if size, ok := img["Size"].(float64); ok {
			image.Size = int64(size)
		}

		if created, ok := img["Created"].(float64); ok {
			image.Created = time.Unix(int64(created), 0)
		}

		if labels, ok := img["Labels"].(map[string]interface{}); ok {
			image.Labels = make(map[string]string)
			for k, v := range labels {
				if vStr, ok := v.(string); ok {
					image.Labels[k] = vStr
				}
			}
		}

		result = append(result, image)
	}

	return result, nil
}

// Logs retrieves container logs
func (p *PodmanRuntime) Logs(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error) {
	query := url.Values{}
	query.Set("stdout", "true")
	query.Set("stderr", "true")

	if opts.Follow {
		query.Set("follow", "true")
	}

	if opts.Timestamps {
		query.Set("timestamps", "true")
	}

	if opts.Tail > 0 {
		query.Set("tail", strconv.Itoa(opts.Tail))
	}

	if !opts.Since.IsZero() {
		query.Set("since", strconv.FormatInt(opts.Since.Unix(), 10))
	}

	if !opts.Until.IsZero() {
		query.Set("until", strconv.FormatInt(opts.Until.Unix(), 10))
	}

	endpoint := fmt.Sprintf("/containers/%s/logs?%s", id, query.Encode())

	resp, err := p.get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("failed to get container logs: %s - %s", resp.Status, string(body))
	}

	return resp.Body, nil
}

// Stats retrieves container statistics
func (p *PodmanRuntime) Stats(ctx context.Context, id string) (*Stats, error) {
	endpoint := fmt.Sprintf("/containers/%s/stats?stream=false", id)

	resp, err := p.get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get container stats: %s - %s", resp.Status, string(body))
	}

	var statsData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&statsData); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	return p.convertStats(statsData), nil
}

// Ping checks if the Podman daemon is accessible
func (p *PodmanRuntime) Ping(ctx context.Context) error {
	resp, err := p.get(ctx, "/_ping")
	if err != nil {
		return fmt.Errorf("Podman ping failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Podman ping failed: %s", resp.Status)
	}

	return nil
}

// Version returns Podman version information
func (p *PodmanRuntime) Version(ctx context.Context) (*Version, error) {
	resp, err := p.get(ctx, "/version")
	if err != nil {
		return nil, fmt.Errorf("failed to get Podman version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get Podman version: %s - %s", resp.Status, string(body))
	}

	var versionData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&versionData); err != nil {
		return nil, fmt.Errorf("failed to decode version: %w", err)
	}

	version := &Version{
		OS:   "linux",
		Arch: "amd64",
	}

	if v, ok := versionData["Version"].(string); ok {
		version.Version = v
	}

	if v, ok := versionData["ApiVersion"].(string); ok {
		version.APIVersion = v
	}

	if v, ok := versionData["MinAPIVersion"].(string); ok {
		version.MinAPIVersion = v
	}

	if v, ok := versionData["GoVersion"].(string); ok {
		version.GoVersion = v
	}

	if v, ok := versionData["GitCommit"].(string); ok {
		version.GitCommit = v
	}

	if v, ok := versionData["Built"].(string); ok {
		version.BuildTime = v
	}

	if v, ok := versionData["OsArch"].(string); ok {
		parts := strings.Split(v, "/")
		if len(parts) == 2 {
			version.OS = parts[0]
			version.Arch = parts[1]
		}
	}

	return version, nil
}

// Info returns Podman system information
func (p *PodmanRuntime) Info(ctx context.Context) (*RuntimeInfo, error) {
	resp, err := p.get(ctx, "/info")
	if err != nil {
		return nil, fmt.Errorf("failed to get Podman info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get Podman info: %s - %s", resp.Status, string(body))
	}

	var infoData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&infoData); err != nil {
		return nil, fmt.Errorf("failed to decode info: %w", err)
	}

	version, _ := p.Version(ctx)

	info := &RuntimeInfo{
		Name: "Podman",
		Type: string(RuntimeTypePodman),
	}

	if version != nil {
		info.Version = version.Version
		info.APIVersion = version.APIVersion
		info.OS = version.OS
		info.Arch = version.Arch
	}

	// Extract host info
	if host, ok := infoData["host"].(map[string]interface{}); ok {
		if v, ok := host["kernel"].(string); ok {
			info.KernelVersion = v
		}
		if v, ok := host["os"].(string); ok {
			info.OS = v
		}
		if v, ok := host["arch"].(string); ok {
			info.Arch = v
		}
		if v, ok := host["cpus"].(float64); ok {
			info.CPUs = int(v)
		}
		if v, ok := host["memTotal"].(float64); ok {
			info.MemoryTotal = int64(v)
		}
	}

	// Extract store info
	if store, ok := infoData["store"].(map[string]interface{}); ok {
		if v, ok := store["graphDriverName"].(string); ok {
			info.Driver = v
		}
		if v, ok := store["containerStore"].(map[string]interface{}); ok {
			if num, ok := v["number"].(float64); ok {
				// Podman doesn't separate running/stopped in store info
				// We'll need to list containers to get accurate counts
				info.ContainersRunning = 0
				info.ContainersStopped = int(num)
			}
		}
		if v, ok := store["imageStore"].(map[string]interface{}); ok {
			if num, ok := v["number"].(float64); ok {
				info.Images = int(num)
			}
		}
	}

	return info, nil
}

// Helper methods

func (p *PodmanRuntime) get(ctx context.Context, endpoint string) (*http.Response, error) {
	url := p.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	return p.client.Do(req)
}

func (p *PodmanRuntime) post(ctx context.Context, endpoint string, payload interface{}) (*http.Response, error) {
	url := p.baseURL + endpoint

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return p.client.Do(req)
}

func (p *PodmanRuntime) delete(ctx context.Context, endpoint string) (*http.Response, error) {
	url := p.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	return p.client.Do(req)
}

func (p *PodmanRuntime) convertContainer(data map[string]interface{}) *Container {
	c := &Container{
		Runtime: RuntimeTypePodman,
	}

	if v, ok := data["Id"].(string); ok {
		c.ID = v
	}

	if v, ok := data["Name"].(string); ok {
		c.Name = strings.TrimPrefix(v, "/")
	}

	if v, ok := data["Image"].(string); ok {
		c.Image = v
	}

	// Extract status from nested State object (Podman inspect format)
	if stateObj, ok := data["State"].(map[string]interface{}); ok {
		// Extract Status from State.Status
		if v, ok := stateObj["Status"].(string); ok {
			c.State = v
			c.Status = v // Set both for compatibility
		}

		// Extract StartedAt from State.StartedAt
		if v, ok := stateObj["StartedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				c.Started = t
			}
		}

		// Extract FinishedAt from State.FinishedAt
		if v, ok := stateObj["FinishedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				c.Finished = t
			}
		}

		// Extract ExitCode from State.ExitCode
		if v, ok := stateObj["ExitCode"].(float64); ok {
			c.ExitCode = int(v)
		}
	} else {
		// Fallback to flat structure for backward compatibility
		if v, ok := data["Status"].(string); ok {
			c.Status = v
		}
		if v, ok := data["State"].(string); ok {
			c.State = v
		}
	}

	if v, ok := data["Created"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			c.Created = t
		}
	}

	// StartedAt, FinishedAt, and ExitCode are now extracted from State object above

	if v, ok := data["Labels"].(map[string]interface{}); ok {
		c.Labels = make(map[string]string)
		for k, val := range v {
			if str, ok := val.(string); ok {
				c.Labels[k] = str
			}
		}
	}

	// Extract port mappings
	if ports, ok := data["Ports"].([]interface{}); ok {
		for _, port := range ports {
			if pm, ok := port.(map[string]interface{}); ok {
				portMapping := PortMapping{}
				if v, ok := pm["host_ip"].(string); ok {
					portMapping.HostIP = v
				}
				if v, ok := pm["host_port"].(float64); ok {
					portMapping.HostPort = int(v)
				}
				if v, ok := pm["container_port"].(float64); ok {
					portMapping.ContainerPort = int(v)
				}
				if v, ok := pm["protocol"].(string); ok {
					portMapping.Protocol = v
				}
				c.Ports = append(c.Ports, portMapping)
			}
		}
	}

	return c
}

func (p *PodmanRuntime) convertStats(data map[string]interface{}) *Stats {
	stats := &Stats{}

	// Extract CPU stats
	if cpu, ok := data["cpu_stats"].(map[string]interface{}); ok {
		if usage, ok := cpu["cpu_usage"].(map[string]interface{}); ok {
			if total, ok := usage["total_usage"].(float64); ok {
				// Calculate CPU percentage (simplified)
				stats.CPUPercent = total / 1000000000.0 // Convert to seconds
			}
		}
	}

	// Extract memory stats
	if mem, ok := data["memory_stats"].(map[string]interface{}); ok {
		if usage, ok := mem["usage"].(float64); ok {
			stats.MemoryUsage = uint64(usage)
		}
		if limit, ok := mem["limit"].(float64); ok {
			stats.MemoryLimit = uint64(limit)
			if stats.MemoryLimit > 0 {
				stats.MemoryPercent = float64(stats.MemoryUsage) / float64(stats.MemoryLimit) * 100.0
			}
		}
	}

	// Extract network stats
	if networks, ok := data["networks"].(map[string]interface{}); ok {
		for _, network := range networks {
			if net, ok := network.(map[string]interface{}); ok {
				if rx, ok := net["rx_bytes"].(float64); ok {
					stats.NetworkRx += uint64(rx)
				}
				if tx, ok := net["tx_bytes"].(float64); ok {
					stats.NetworkTx += uint64(tx)
				}
			}
		}
	}

	// Extract block I/O stats
	if blkio, ok := data["blkio_stats"].(map[string]interface{}); ok {
		if ioService, ok := blkio["io_service_bytes_recursive"].([]interface{}); ok {
			for _, io := range ioService {
				if ioMap, ok := io.(map[string]interface{}); ok {
					op, _ := ioMap["op"].(string)
					value, _ := ioMap["value"].(float64)
					if op == "Read" {
						stats.BlockRead += uint64(value)
					} else if op == "Write" {
						stats.BlockWrite += uint64(value)
					}
				}
			}
		}
	}

	// Extract PIDs
	if pids, ok := data["pids_stats"].(map[string]interface{}); ok {
		if current, ok := pids["current"].(float64); ok {
			stats.PIDs = uint64(current)
		}
	}

	return stats
}

// Made with Bob
