package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CreateInstanceRequest contains all parameters for creating an instance
type CreateInstanceRequest struct {
	Name               string
	APIEndpoint        string
	Username           string
	Password           string
	StorageSystem      string
	StorageDescription string
	StorageLocation    string
	IPQuorumName       string
	DownloadEnabled    bool
	MkQuorumAppEnabled bool
	PartnerSystem      string
	IP6                bool
	PartnerIP6         bool
	NoMetadata         bool
}

// Executor handles script execution
type Executor struct {
	scriptPath string
	timeout    time.Duration
	log        *zap.Logger
}

// NewExecutor creates a new executor
func NewExecutor(scriptPath string, timeout int, log *zap.Logger) *Executor {
	return &Executor{
		scriptPath: scriptPath,
		timeout:    time.Duration(timeout) * time.Second,
		log:        log,
	}
}

// ExecuteCommand runs a command with the script
func (e *Executor) ExecuteCommand(ctx context.Context, action string, args ...string) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Build command: script action args...
	cmdArgs := append([]string{action}, args...)

	e.log.Info("Executing command",
		zap.String("script", e.scriptPath),
		zap.String("action", action),
		zap.Strings("args", args))

	cmd := exec.CommandContext(cmdCtx, e.scriptPath, cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := stdout.String()
	errOutput := stderr.String()

	if err != nil {
		e.log.Error("Command execution failed",
			zap.String("action", action),
			zap.Error(err),
			zap.String("stdout", output),
			zap.String("stderr", errOutput))

		return "", fmt.Errorf("command failed: %w\nstderr: %s", err, errOutput)
	}

	e.log.Debug("Command executed successfully",
		zap.String("action", action),
		zap.String("output", output))

	return strings.TrimSpace(output), nil
}

// CreateInstance creates a new IPQuorum instance
func (e *Executor) CreateInstance(ctx context.Context, req CreateInstanceRequest) error {
	// The script expects environment variables for non-interactive mode
	// We need to create a config file or use a different approach
	// For now, call with --non-interactive and then manually configure
	_, err := e.ExecuteCommand(ctx, "create", req.Name, "--non-interactive")
	if err != nil {
		return err
	}

	// After creating the instance, we need to update its configuration
	// This is a workaround until we have a better solution
	// The config file is at /etc/ipquorum/instances/<name>.conf
	return e.updateInstanceConfig(ctx, req)
}

// updateInstanceConfig updates the instance configuration file
func (e *Executor) updateInstanceConfig(ctx context.Context, req CreateInstanceRequest) error {
	configPath := fmt.Sprintf("/etc/ipquorum/instances/%s.conf", req.Name)

	// Use sed to update the configuration file (with sudo)
	// Note: Config file uses format without quotes for most fields
	commands := [][]string{
		// Required fields
		{"sudo", "sed", "-i", fmt.Sprintf("s|^API_ENDPOINT=<API_ENDPOINT>|API_ENDPOINT=%s|g", req.APIEndpoint), configPath},
		{"sudo", "sed", "-i", fmt.Sprintf("s|^VIRTUALIZE_USERNAME=<USERNAME>|VIRTUALIZE_USERNAME=%s|g", req.Username), configPath},

		// Documentation fields (with quotes in config)
		{"sudo", "sed", "-i", fmt.Sprintf("s|^IBM_STORAGE_SYSTEM=\"<HOSTNAME_OR_IP>\"|IBM_STORAGE_SYSTEM=\"%s\"|g", req.StorageSystem), configPath},
		{"sudo", "sed", "-i", fmt.Sprintf("s|^IBM_STORAGE_DESCRIPTION=\"\"|IBM_STORAGE_DESCRIPTION=\"%s\"|g", req.StorageDescription), configPath},
		{"sudo", "sed", "-i", fmt.Sprintf("s|^IBM_STORAGE_LOCATION=\"\"|IBM_STORAGE_LOCATION=\"%s\"|g", req.StorageLocation), configPath},

		// IP Quorum name
		{"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_NAME=.*|IPQUORUM_NAME=%s|g", req.IPQuorumName), configPath},

		// Download configuration
		{"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_DOWNLOAD_ENABLED=.*|IPQUORUM_DOWNLOAD_ENABLED=%t|g", req.DownloadEnabled), configPath},

		// mkquorumapp configuration
		{"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_MKQUORUMAPP_ENABLED=.*|IPQUORUM_MKQUORUMAPP_ENABLED=%t|g", req.MkQuorumAppEnabled), configPath},
		{"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_PARTNERSYSTEM=.*|IPQUORUM_PARTNERSYSTEM=%s|g", req.PartnerSystem), configPath},
		{"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_IP6=.*|IPQUORUM_IP6=%t|g", req.IP6), configPath},
		{"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_PARTNERIP6=.*|IPQUORUM_PARTNERIP6=%t|g", req.PartnerIP6), configPath},
		{"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_NOMETADATA=.*|IPQUORUM_NOMETADATA=%t|g", req.NoMetadata), configPath},
	}

	for _, cmdArgs := range commands {
		cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			e.log.Error("Failed to update config",
				zap.String("name", req.Name),
				zap.String("command", strings.Join(cmdArgs, " ")),
				zap.String("output", string(output)),
				zap.Error(err))
			return fmt.Errorf("failed to update config: %w", err)
		}
	}

	// Store password in secure location
	passwordDir := "/var/lib/ipquorum/.passwords"
	passwordFile := fmt.Sprintf("%s/%s.pass", passwordDir, req.Name)

	// Create password directory if it doesn't exist (with sudo)
	cmd := exec.CommandContext(ctx, "sudo", "mkdir", "-p", passwordDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		e.log.Error("Failed to create password directory", zap.String("output", string(output)), zap.Error(err))
		return fmt.Errorf("failed to create password directory: %w", err)
	}

	// Write password to file (with sudo)
	cmd = exec.CommandContext(ctx, "sudo", "bash", "-c",
		fmt.Sprintf("echo '%s' > %s && chmod 400 %s && chown ipquorum:ipquorum %s",
			req.Password, passwordFile, passwordFile, passwordFile))
	if output, err := cmd.CombinedOutput(); err != nil {
		e.log.Error("Failed to write password", zap.String("output", string(output)), zap.Error(err))
		return fmt.Errorf("failed to write password: %w", err)
	}

	// Update config to use password file (with sudo)
	// The template has: VIRTUALIZE_PASSWORD_FILE=/var/lib/ipquorum/.passwords/${INSTANCE_NAME}.password
	// We need to replace the variable with the actual path
	cmd = exec.CommandContext(ctx, "sudo", "sed", "-i",
		fmt.Sprintf("s|^VIRTUALIZE_PASSWORD_FILE=.*|VIRTUALIZE_PASSWORD_FILE=%s|g", passwordFile), configPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		e.log.Error("Failed to update password file path", zap.String("output", string(output)), zap.Error(err))
		return fmt.Errorf("failed to update password file path: %w", err)
	}

	e.log.Info("Successfully updated instance config",
		zap.String("name", req.Name),
		zap.String("api_endpoint", req.APIEndpoint),
		zap.String("username", req.Username),
		zap.String("storage_system", req.StorageSystem),
		zap.String("mkquorumapp_enabled", fmt.Sprintf("%t", req.MkQuorumAppEnabled)),
		zap.String("password_file", passwordFile))

	return nil
}

// DeleteInstance deletes an IPQuorum instance
func (e *Executor) DeleteInstance(ctx context.Context, name string, force bool) error {
	// Always use --force for non-interactive deletion via API
	args := []string{name, "--force"}
	if force {
		// If force is true, also delete data and logs
		args = append(args, "--delete-data")
	}
	_, err := e.ExecuteCommand(ctx, "delete", args...)
	return err
}

// StartInstance starts an IPQuorum instance
func (e *Executor) StartInstance(ctx context.Context, name string) error {
	_, err := e.ExecuteCommand(ctx, "start", name)
	return err
}

// StopInstance stops an IPQuorum instance
func (e *Executor) StopInstance(ctx context.Context, name string) error {
	_, err := e.ExecuteCommand(ctx, "stop", name)
	return err
}

// RestartInstance restarts an IPQuorum instance
func (e *Executor) RestartInstance(ctx context.Context, name string) error {
	_, err := e.ExecuteCommand(ctx, "restart", name)
	return err
}

// GetStatus gets the status of an IPQuorum instance
func (e *Executor) GetStatus(ctx context.Context, name string) (string, error) {
	return e.ExecuteCommand(ctx, "status", name)
}

// ListInstances lists all IPQuorum instances
func (e *Executor) ListInstances(ctx context.Context) (string, error) {
	return e.ExecuteCommand(ctx, "list")
}

// GetLogs gets logs for an IPQuorum instance
func (e *Executor) GetLogs(ctx context.Context, name string, lines int) (string, error) {
	return e.ExecuteCommand(ctx, "logs", name, fmt.Sprintf("%d", lines))
}
