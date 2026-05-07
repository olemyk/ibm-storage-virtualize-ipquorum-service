package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/olemyk/ipquorum-platform/server/internal/agent"
	"go.uber.org/zap"
)

// Executor handles execution of bash scripts and remote agent operations
type Executor struct {
	log           *zap.Logger
	scriptsDir    string
	timeout       time.Duration
	agentRegistry *agent.Registry
}

// NewExecutor creates a new script executor
func NewExecutor(log *zap.Logger, scriptsDir string, timeout time.Duration, agentRegistry *agent.Registry) *Executor {
	return &Executor{
		log:           log,
		scriptsDir:    scriptsDir,
		timeout:       timeout,
		agentRegistry: agentRegistry,
	}
}

// ExecuteResult contains the result of script execution
type ExecuteResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Error    error
}

// Execute runs a bash script with arguments
func (e *Executor) Execute(ctx context.Context, scriptName string, args ...string) *ExecuteResult {
	start := time.Now()
	result := &ExecuteResult{}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Build script path
	scriptPath := filepath.Join(e.scriptsDir, scriptName)

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		result.Error = fmt.Errorf("script not found: %s", scriptPath)
		result.ExitCode = -1
		return result
	}

	// Check if script is executable
	info, err := os.Stat(scriptPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to stat script: %w", err)
		result.ExitCode = -1
		return result
	}
	if info.Mode()&0111 == 0 {
		result.Error = fmt.Errorf("script is not executable: %s", scriptPath)
		result.ExitCode = -1
		return result
	}

	// Prepare command
	cmd := exec.CommandContext(execCtx, "/bin/bash", append([]string{scriptPath}, args...)...)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set environment variables
	cmd.Env = os.Environ()

	e.log.Debug("Executing script",
		zap.String("script", scriptName),
		zap.Strings("args", args),
		zap.String("path", scriptPath),
	)

	// Execute command
	err = cmd.Run()
	result.Duration = time.Since(start)
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Error = err
		e.log.Error("Script execution failed",
			zap.String("script", scriptName),
			zap.Int("exit_code", result.ExitCode),
			zap.Duration("duration", result.Duration),
			zap.Error(err),
		)
	} else {
		result.ExitCode = 0
		e.log.Debug("Script execution succeeded",
			zap.String("script", scriptName),
			zap.Duration("duration", result.Duration),
		)
	}

	return result
}

// ExecuteInstanceManager executes the ipquorum-instance-manager.sh script
func (e *Executor) ExecuteInstanceManager(ctx context.Context, action, instanceName string, extraArgs ...string) *ExecuteResult {
	args := []string{action, instanceName}
	args = append(args, extraArgs...)
	return e.Execute(ctx, "ipquorum-instance-manager.sh", args...)
}

// StartInstance starts an IP Quorum instance
func (e *Executor) StartInstance(ctx context.Context, instanceName string) *ExecuteResult {
	return e.ExecuteInstanceManager(ctx, "start", instanceName)
}

// StopInstance stops an IP Quorum instance
func (e *Executor) StopInstance(ctx context.Context, instanceName string) *ExecuteResult {
	return e.ExecuteInstanceManager(ctx, "stop", instanceName)
}

// RestartInstance restarts an IP Quorum instance
func (e *Executor) RestartInstance(ctx context.Context, instanceName string) *ExecuteResult {
	return e.ExecuteInstanceManager(ctx, "restart", instanceName)
}

// StatusInstance gets the status of an IP Quorum instance
func (e *Executor) StatusInstance(ctx context.Context, instanceName string) *ExecuteResult {
	return e.ExecuteInstanceManager(ctx, "status", instanceName)
}

// CreateInstanceParams holds all parameters for creating an instance
type CreateInstanceParams struct {
	InstanceName      string
	APIEndpoint       string
	Username          string
	Password          string
	EnableDownload    bool
	EnableMkquorumapp bool
	Partnersystem     string
	IPQuorumName      string
	IP6               bool
	PartnerIP6        bool
	NoMetadata        bool
}

// CreateInstance creates a new IP Quorum instance
func (e *Executor) CreateInstance(ctx context.Context, instanceName, apiEndpoint, username, password string, enableDownload, enableMkquorumapp bool, partnersystem string) *ExecuteResult {
	// Use new method with params struct
	params := CreateInstanceParams{
		InstanceName:      instanceName,
		APIEndpoint:       apiEndpoint,
		Username:          username,
		Password:          password,
		EnableDownload:    enableDownload,
		EnableMkquorumapp: enableMkquorumapp,
		Partnersystem:     partnersystem,
		IP6:               false,
		PartnerIP6:        false,
		NoMetadata:        false,
		IPQuorumName:      "",
	}
	return e.CreateInstanceWithParams(ctx, params)
}

// CreateInstanceWithParams creates a new IP Quorum instance with full parameters
func (e *Executor) CreateInstanceWithParams(ctx context.Context, params CreateInstanceParams) *ExecuteResult {
	args := []string{
		"create",
		params.InstanceName,
		"--api-endpoint", params.APIEndpoint,
		"--user", params.Username,
		"--pass", params.Password,
	}

	if params.EnableDownload {
		args = append(args, "--download")
	} else {
		args = append(args, "--no-download")
	}

	if params.EnableMkquorumapp {
		args = append(args, "--mkquorumapp")
		if params.Partnersystem != "" {
			args = append(args, "--partnersystem", params.Partnersystem)
		}
		// Add new mkquorumapp parameters
		if params.IP6 {
			args = append(args, "--ip6=true")
		} else {
			args = append(args, "--ip6=false")
		}
		if params.PartnerIP6 {
			args = append(args, "--partnerip6=true")
		} else {
			args = append(args, "--partnerip6=false")
		}
		if params.NoMetadata {
			args = append(args, "--nometadata=true")
		} else {
			args = append(args, "--nometadata=false")
		}
	} else {
		args = append(args, "--no-mkquorumapp")
	}

	// Add ipquorum_name if provided
	if params.IPQuorumName != "" {
		args = append(args, "--ipquorum-name", params.IPQuorumName)
	}

	return e.Execute(ctx, "ipquorum-instance-manager.sh", args...)
}

// DeleteInstance deletes an IP Quorum instance
func (e *Executor) DeleteInstance(ctx context.Context, instanceName string, force bool) *ExecuteResult {
	args := []string{"delete", instanceName}
	if force {
		args = append(args, "--force")
	}
	return e.Execute(ctx, "ipquorum-instance-manager.sh", args...)
}

// ListInstances lists all IP Quorum instances
func (e *Executor) ListInstances(ctx context.Context) *ExecuteResult {
	return e.ExecuteInstanceManager(ctx, "list", "")
}

// CheckNetwork checks network connectivity to the API endpoint
func (e *Executor) CheckNetwork(ctx context.Context, instanceName string) *ExecuteResult {
	return e.ExecuteInstanceManager(ctx, "check-network", instanceName)
}

// ValidateConfig validates the configuration of an instance
func (e *Executor) ValidateConfig(ctx context.Context, instanceName string) *ExecuteResult {
	return e.ExecuteInstanceManager(ctx, "validate", instanceName)
}

// GetLogs gets logs from a local instance
func (e *Executor) GetLogs(ctx context.Context, instanceName string, lines int) *ExecuteResult {
	args := []string{"logs", instanceName, fmt.Sprintf("--lines=%d", lines)}
	return e.Execute(ctx, "ipquorum-instance-manager.sh", args...)
}

// ParseInstanceStatus parses the status output from the script
func ParseInstanceStatus(stdout string) (status string, healthy bool) {
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "active (running)") {
			return "running", true
		}
		if strings.Contains(line, "inactive") || strings.Contains(line, "dead") {
			return "stopped", false
		}
		if strings.Contains(line, "failed") {
			return "failed", false
		}
	}
	return "unknown", false
}

// ParseNetworkCheck parses the network check output
func ParseNetworkCheck(stdout string) (reachable bool, message string) {
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Network check: OK") || strings.Contains(line, "reachable") {
			return true, line
		}
		if strings.Contains(line, "Network check: FAILED") || strings.Contains(line, "unreachable") {
			return false, line
		}
	}
	return false, "Unknown network status"
}

// Remote execution methods for agent-based operations

// StartInstanceRemote starts an instance on a remote agent
func (e *Executor) StartInstanceRemote(ctx context.Context, serverID, instanceName string) *ExecuteResult {
	result := &ExecuteResult{}
	start := time.Now()

	// Get agent client
	client, err := e.agentRegistry.Get(serverID)
	if err != nil {
		result.Error = fmt.Errorf("failed to get agent: %w", err)
		result.ExitCode = -1
		result.Stderr = err.Error()
		return result
	}

	// Call agent API
	err = client.StartInstance(ctx, instanceName)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		result.ExitCode = 1
		result.Stderr = err.Error()
		e.log.Error("Remote start instance failed",
			zap.String("server_id", serverID),
			zap.String("instance", instanceName),
			zap.Error(err),
		)
	} else {
		result.ExitCode = 0
		result.Stdout = fmt.Sprintf("Instance %s started on server %s", instanceName, serverID)
		e.log.Info("Remote start instance succeeded",
			zap.String("server_id", serverID),
			zap.String("instance", instanceName),
		)
	}

	return result
}

// StopInstanceRemote stops an instance on a remote agent
func (e *Executor) StopInstanceRemote(ctx context.Context, serverID, instanceName string) *ExecuteResult {
	result := &ExecuteResult{}
	start := time.Now()

	client, err := e.agentRegistry.Get(serverID)
	if err != nil {
		result.Error = fmt.Errorf("failed to get agent: %w", err)
		result.ExitCode = -1
		result.Stderr = err.Error()
		return result
	}

	err = client.StopInstance(ctx, instanceName)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		result.ExitCode = 1
		result.Stderr = err.Error()
		e.log.Error("Remote stop instance failed",
			zap.String("server_id", serverID),
			zap.String("instance", instanceName),
			zap.Error(err),
		)
	} else {
		result.ExitCode = 0
		result.Stdout = fmt.Sprintf("Instance %s stopped on server %s", instanceName, serverID)
		e.log.Info("Remote stop instance succeeded",
			zap.String("server_id", serverID),
			zap.String("instance", instanceName),
		)
	}

	return result
}

// RestartInstanceRemote restarts an instance on a remote agent
func (e *Executor) RestartInstanceRemote(ctx context.Context, serverID, instanceName string) *ExecuteResult {
	result := &ExecuteResult{}
	start := time.Now()

	client, err := e.agentRegistry.Get(serverID)
	if err != nil {
		result.Error = fmt.Errorf("failed to get agent: %w", err)
		result.ExitCode = -1
		result.Stderr = err.Error()
		return result
	}

	err = client.RestartInstance(ctx, instanceName)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		result.ExitCode = 1
		result.Stderr = err.Error()
		e.log.Error("Remote restart instance failed",
			zap.String("server_id", serverID),
			zap.String("instance", instanceName),
			zap.Error(err),
		)
	} else {
		result.ExitCode = 0
		result.Stdout = fmt.Sprintf("Instance %s restarted on server %s", instanceName, serverID)
		e.log.Info("Remote restart instance succeeded",
			zap.String("server_id", serverID),
			zap.String("instance", instanceName),
		)
	}

	return result
}

// StatusInstanceRemote gets status of an instance on a remote agent
func (e *Executor) StatusInstanceRemote(ctx context.Context, serverID, instanceName string) *ExecuteResult {
	result := &ExecuteResult{}
	start := time.Now()

	client, err := e.agentRegistry.Get(serverID)
	if err != nil {
		result.Error = fmt.Errorf("failed to get agent: %w", err)
		result.ExitCode = -1
		result.Stderr = err.Error()
		return result
	}

	status, err := client.GetStatus(ctx, instanceName)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		result.ExitCode = 1
		result.Stderr = err.Error()
	} else {
		result.ExitCode = 0
		result.Stdout = fmt.Sprintf("Status: %s, Health: %s, Message: %s", status.Status, status.Health, status.Message)
	}

	return result
}

// GetLogsRemote gets logs from an instance on a remote agent
func (e *Executor) GetLogsRemote(ctx context.Context, serverID, instanceName string, lines int) *ExecuteResult {
	result := &ExecuteResult{}
	start := time.Now()

	client, err := e.agentRegistry.Get(serverID)
	if err != nil {
		result.Error = fmt.Errorf("failed to get agent: %w", err)
		result.ExitCode = -1
		result.Stderr = err.Error()
		return result
	}

	logs, err := client.GetLogs(ctx, instanceName, lines)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		result.ExitCode = 1
		result.Stderr = err.Error()
	} else {
		result.ExitCode = 0
		result.Stdout = logs
	}

	return result
}

// CreateInstanceRemote creates an instance on a remote agent
func (e *Executor) CreateInstanceRemote(ctx context.Context, serverID string, params CreateInstanceParams) *ExecuteResult {
	result := &ExecuteResult{}
	start := time.Now()

	client, err := e.agentRegistry.Get(serverID)
	if err != nil {
		result.Error = fmt.Errorf("failed to get agent: %w", err)
		result.ExitCode = -1
		result.Stderr = err.Error()
		return result
	}

	// Convert params to agent request
	req := &agent.CreateInstanceRequest{
		Name:              params.InstanceName,
		APIEndpoint:       params.APIEndpoint,
		Username:          params.Username,
		Password:          params.Password,
		EnableDownload:    params.EnableDownload,
		EnableMkquorumapp: params.EnableMkquorumapp,
		Partnersystem:     params.Partnersystem,
		IPQuorumName:      params.IPQuorumName,
		IP6:               params.IP6,
		PartnerIP6:        params.PartnerIP6,
		NoMetadata:        params.NoMetadata,
	}

	err = client.CreateInstance(ctx, req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		result.ExitCode = 1
		result.Stderr = err.Error()
		e.log.Error("Remote create instance failed",
			zap.String("server_id", serverID),
			zap.String("instance", params.InstanceName),
			zap.Error(err),
		)
	} else {
		result.ExitCode = 0
		result.Stdout = fmt.Sprintf("Instance %s created on server %s", params.InstanceName, serverID)
		e.log.Info("Remote create instance succeeded",
			zap.String("server_id", serverID),
			zap.String("instance", params.InstanceName),
		)
	}

	return result
}

// DeleteInstanceRemote deletes an instance on a remote agent
func (e *Executor) DeleteInstanceRemote(ctx context.Context, serverID, instanceName string) *ExecuteResult {
	result := &ExecuteResult{}
	start := time.Now()

	client, err := e.agentRegistry.Get(serverID)
	if err != nil {
		result.Error = fmt.Errorf("failed to get agent: %w", err)
		result.ExitCode = -1
		result.Stderr = err.Error()
		return result
	}

	err = client.DeleteInstance(ctx, instanceName)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		result.ExitCode = 1
		result.Stderr = err.Error()
		e.log.Error("Remote delete instance failed",
			zap.String("server_id", serverID),
			zap.String("instance", instanceName),
			zap.Error(err),
		)
	} else {
		result.ExitCode = 0
		result.Stdout = fmt.Sprintf("Instance %s deleted on server %s", instanceName, serverID)
		e.log.Info("Remote delete instance succeeded",
			zap.String("server_id", serverID),
			zap.String("instance", instanceName),
		)
	}

	return result
}

// IsRemoteExecution checks if an operation should be executed remotely
func (e *Executor) IsRemoteExecution(serverID string) bool {
	return serverID != "" && serverID != "local" && e.agentRegistry != nil
}
