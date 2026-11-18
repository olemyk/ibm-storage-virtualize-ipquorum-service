package runtime

import (
	"context"
	"io"
	"time"
)

// Runtime defines the interface for container runtimes (Docker/Podman)
type Runtime interface {
	// Lifecycle operations
	Create(ctx context.Context, config *ContainerConfig) (*Container, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string, timeout time.Duration) error
	Restart(ctx context.Context, id string, timeout time.Duration) error
	Remove(ctx context.Context, id string, force bool) error

	// Inspection operations
	Inspect(ctx context.Context, id string) (*Container, error)
	List(ctx context.Context, filters map[string]string) ([]*Container, error)

	// Image operations
	ListImages(ctx context.Context) ([]*Image, error)

	// Monitoring operations
	Logs(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error)
	Stats(ctx context.Context, id string) (*Stats, error)

	// Health operations
	Ping(ctx context.Context) error
	Version(ctx context.Context) (*Version, error)
	Info(ctx context.Context) (*RuntimeInfo, error)

	// Runtime identification
	Name() string
	Type() RuntimeType
}

// RuntimeType identifies the container runtime
type RuntimeType string

const (
	RuntimeTypeDocker RuntimeType = "docker"
	RuntimeTypePodman RuntimeType = "podman"
)

// ContainerConfig defines the configuration for creating a container
type ContainerConfig struct {
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	Cmd           []string          `json:"cmd,omitempty"`
	Entrypoint    []string          `json:"entrypoint,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Volumes       []VolumeMount     `json:"volumes,omitempty"`
	Ports         []PortMapping     `json:"ports,omitempty"`
	NetworkMode   string            `json:"network_mode,omitempty"`
	RestartPolicy string            `json:"restart_policy,omitempty"`
	Resources     *Resources        `json:"resources,omitempty"`
	WorkingDir    string            `json:"working_dir,omitempty"`
	User          string            `json:"user,omitempty"`
	Hostname      string            `json:"hostname,omitempty"`
}

// Container represents a container instance
type Container struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Image    string            `json:"image"`
	Status   string            `json:"status"`
	State    string            `json:"state"`
	Created  time.Time         `json:"created"`
	Started  time.Time         `json:"started,omitempty"`
	Finished time.Time         `json:"finished,omitempty"`
	ExitCode int               `json:"exit_code,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
	Ports    []PortMapping     `json:"ports,omitempty"`
	Runtime  RuntimeType       `json:"runtime"`
}

// Image represents a container image
type Image struct {
	ID          string            `json:"id"`
	RepoTags    []string          `json:"repo_tags"`
	RepoDigests []string          `json:"repo_digests,omitempty"`
	Size        int64             `json:"size"`
	Created     time.Time         `json:"created"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// Stats represents container resource usage statistics
type Stats struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   uint64  `json:"memory_usage"`
	MemoryLimit   uint64  `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRx     uint64  `json:"network_rx"`
	NetworkTx     uint64  `json:"network_tx"`
	BlockRead     uint64  `json:"block_read"`
	BlockWrite    uint64  `json:"block_write"`
	PIDs          uint64  `json:"pids"`
}

// Resources defines resource constraints for a container
type Resources struct {
	CPULimit    string `json:"cpu_limit,omitempty"`    // "1.0", "2.0"
	MemoryLimit string `json:"memory_limit,omitempty"` // "512m", "1g"
	CPUShares   int64  `json:"cpu_shares,omitempty"`
	MemorySwap  string `json:"memory_swap,omitempty"`
}

// VolumeMount defines a volume mount
type VolumeMount struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	ReadOnly    bool   `json:"read_only,omitempty"`
	Type        string `json:"type,omitempty"` // "bind", "volume", "tmpfs"
}

// PortMapping defines a port mapping
type PortMapping struct {
	HostIP        string `json:"host_ip,omitempty"`
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol,omitempty"` // "tcp", "udp"
}

// LogOptions defines options for retrieving container logs
type LogOptions struct {
	Follow     bool      `json:"follow"`
	Tail       int       `json:"tail"`
	Since      time.Time `json:"since,omitempty"`
	Until      time.Time `json:"until,omitempty"`
	Timestamps bool      `json:"timestamps"`
	Details    bool      `json:"details"`
}

// Version represents runtime version information
type Version struct {
	Version       string `json:"version"`
	APIVersion    string `json:"api_version"`
	MinAPIVersion string `json:"min_api_version,omitempty"`
	GoVersion     string `json:"go_version"`
	GitCommit     string `json:"git_commit,omitempty"`
	BuildTime     string `json:"build_time,omitempty"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
}

// RuntimeInfo represents runtime system information
type RuntimeInfo struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	Version           string   `json:"version"`
	APIVersion        string   `json:"api_version"`
	OS                string   `json:"os"`
	Arch              string   `json:"arch"`
	KernelVersion     string   `json:"kernel_version,omitempty"`
	ContainersRunning int      `json:"containers_running"`
	ContainersPaused  int      `json:"containers_paused"`
	ContainersStopped int      `json:"containers_stopped"`
	Images            int      `json:"images"`
	Driver            string   `json:"driver,omitempty"`
	DriverStatus      []string `json:"driver_status,omitempty"`
	MemoryTotal       int64    `json:"memory_total,omitempty"`
	CPUs              int      `json:"cpus,omitempty"`
}

// ContainerState represents the state of a container
type ContainerState string

const (
	StateCreated    ContainerState = "created"
	StateRunning    ContainerState = "running"
	StatePaused     ContainerState = "paused"
	StateRestarting ContainerState = "restarting"
	StateExited     ContainerState = "exited"
	StateDead       ContainerState = "dead"
)

// RestartPolicy defines container restart behavior
type RestartPolicy string

const (
	RestartPolicyNo            RestartPolicy = "no"
	RestartPolicyAlways        RestartPolicy = "always"
	RestartPolicyOnFailure     RestartPolicy = "on-failure"
	RestartPolicyUnlessStopped RestartPolicy = "unless-stopped"
)

// Made with Bob
