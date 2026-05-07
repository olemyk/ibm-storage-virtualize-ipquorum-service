package executor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewExecutor(t *testing.T) {
	logger := zap.NewNop()
	scriptsDir := "/tmp/test-scripts"
	timeout := 5 * time.Minute

	exec := NewExecutor(logger, scriptsDir, timeout)

	if exec == nil {
		t.Fatal("NewExecutor returned nil")
	}

	if exec.scriptsDir != scriptsDir {
		t.Errorf("Expected scriptsDir %s, got %s", scriptsDir, exec.scriptsDir)
	}

	if exec.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, exec.timeout)
	}
}

func TestExecute_Success(t *testing.T) {
	// Create temporary script directory
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")

	// Create a simple test script
	scriptContent := `#!/bin/bash
echo "Hello from test script"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	logger := zap.NewNop()
	exec := NewExecutor(logger, tmpDir, 5*time.Second)

	ctx := context.Background()
	result := exec.Execute(ctx, "test-script.sh")

	if result.Error != nil {
		t.Errorf("Execute failed: %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if result.Stdout == "" {
		t.Error("Expected stdout output, got empty string")
	}

	if !strings.Contains(result.Stdout, "Hello from test script") {
		t.Errorf("Expected 'Hello from test script' in stdout, got %q", result.Stdout)
	}
}

func TestExecute_WithArgs(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-args.sh")

	scriptContent := `#!/bin/bash
echo "Args: $@"
echo "Arg1: $1"
echo "Arg2: $2"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	logger := zap.NewNop()
	exec := NewExecutor(logger, tmpDir, 5*time.Second)

	ctx := context.Background()
	result := exec.Execute(ctx, "test-args.sh", "arg1", "arg2")

	if result.Error != nil {
		t.Errorf("Execute failed: %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "Args: arg1 arg2") {
		t.Errorf("Expected 'Args: arg1 arg2' in stdout, got %q", result.Stdout)
	}
}

func TestExecute_ScriptNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	logger := zap.NewNop()
	exec := NewExecutor(logger, tmpDir, 5*time.Second)

	ctx := context.Background()
	result := exec.Execute(ctx, "nonexistent.sh")

	if result.Error == nil {
		t.Error("Expected error for nonexistent script, got nil")
	}

	if result.ExitCode != -1 {
		t.Errorf("Expected exit code -1, got %d", result.ExitCode)
	}
}

func TestExecute_ScriptNotExecutable(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "not-executable.sh")

	scriptContent := `#!/bin/bash
echo "This should not run"
exit 0
`
	// Create script without execute permission
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	logger := zap.NewNop()
	exec := NewExecutor(logger, tmpDir, 5*time.Second)

	ctx := context.Background()
	result := exec.Execute(ctx, "not-executable.sh")

	if result.Error == nil {
		t.Error("Expected error for non-executable script, got nil")
	}

	if result.ExitCode != -1 {
		t.Errorf("Expected exit code -1, got %d", result.ExitCode)
	}
}

func TestExecute_ScriptError(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "error-script.sh")

	scriptContent := `#!/bin/bash
echo "Error message" >&2
exit 1
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	logger := zap.NewNop()
	exec := NewExecutor(logger, tmpDir, 5*time.Second)

	ctx := context.Background()
	result := exec.Execute(ctx, "error-script.sh")

	if result.Error == nil {
		t.Error("Expected error for failing script, got nil")
	}

	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}

	// Stderr should contain error message
	if !strings.Contains(result.Stderr, "Error message") {
		t.Errorf("Expected 'Error message' in stderr, got %q", result.Stderr)
	}
}

func TestExecute_Timeout(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "slow-script.sh")

	scriptContent := `#!/bin/bash
sleep 10
echo "Done"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	logger := zap.NewNop()
	exec := NewExecutor(logger, tmpDir, 1*time.Second)

	ctx := context.Background()
	result := exec.Execute(ctx, "slow-script.sh")

	if result.Error == nil {
		t.Error("Expected timeout error, got nil")
	}

	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for timeout")
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "long-script.sh")

	scriptContent := `#!/bin/bash
sleep 5
echo "Done"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	logger := zap.NewNop()
	exec := NewExecutor(logger, tmpDir, 10*time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after 500ms
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	result := exec.Execute(ctx, "long-script.sh")

	if result.Error == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestParseInstanceStatus(t *testing.T) {
	tests := []struct {
		name           string
		stdout         string
		expectedStatus string
		expectedHealth bool
	}{
		{
			name:           "running",
			stdout:         "Service is active (running)",
			expectedStatus: "running",
			expectedHealth: true,
		},
		{
			name:           "stopped",
			stdout:         "Service is inactive (dead)",
			expectedStatus: "stopped",
			expectedHealth: false,
		},
		{
			name:           "failed",
			stdout:         "Service failed to start",
			expectedStatus: "failed",
			expectedHealth: false,
		},
		{
			name:           "unknown",
			stdout:         "Some other output",
			expectedStatus: "unknown",
			expectedHealth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, healthy := ParseInstanceStatus(tt.stdout)
			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}
			if healthy != tt.expectedHealth {
				t.Errorf("Expected healthy %v, got %v", tt.expectedHealth, healthy)
			}
		})
	}
}

func TestParseNetworkCheck(t *testing.T) {
	tests := []struct {
		name              string
		stdout            string
		expectedReachable bool
		messageContains   string
	}{
		{
			name:              "reachable",
			stdout:            "Network check: OK - endpoint is reachable",
			expectedReachable: true,
			messageContains:   "OK",
		},
		{
			name:              "failed",
			stdout:            "Network check: FAILED - connection timeout",
			expectedReachable: false,
			messageContains:   "FAILED",
		},
		{
			name:              "unknown",
			stdout:            "Some other output",
			expectedReachable: false,
			messageContains:   "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reachable, message := ParseNetworkCheck(tt.stdout)
			if reachable != tt.expectedReachable {
				t.Errorf("Expected reachable %v, got %v", tt.expectedReachable, reachable)
			}
			if !strings.Contains(message, tt.messageContains) {
				t.Errorf("Expected message to contain %q, got %q", tt.messageContains, message)
			}
		})
	}
}

func TestStartInstance(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "ipquorum-instance-manager.sh")

	scriptContent := `#!/bin/bash
echo "Action: $1"
echo "Instance: $2"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	logger := zap.NewNop()
	exec := NewExecutor(logger, tmpDir, 5*time.Second)

	ctx := context.Background()
	result := exec.StartInstance(ctx, "test-instance")

	if result.Error != nil {
		t.Errorf("StartInstance failed: %v", result.Error)
	}

	if !strings.Contains(result.Stdout, "Action: start") {
		t.Errorf("Expected 'Action: start' in stdout, got %q", result.Stdout)
	}

	if !strings.Contains(result.Stdout, "Instance: test-instance") {
		t.Errorf("Expected 'Instance: test-instance' in stdout, got %q", result.Stdout)
	}
}

func TestStopInstance(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "ipquorum-instance-manager.sh")

	scriptContent := `#!/bin/bash
echo "Action: $1"
echo "Instance: $2"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	logger := zap.NewNop()
	exec := NewExecutor(logger, tmpDir, 5*time.Second)

	ctx := context.Background()
	result := exec.StopInstance(ctx, "test-instance")

	if result.Error != nil {
		t.Errorf("StopInstance failed: %v", result.Error)
	}

	if !strings.Contains(result.Stdout, "Action: stop") {
		t.Errorf("Expected 'Action: stop' in stdout, got %q", result.Stdout)
	}
}

func TestRestartInstance(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "ipquorum-instance-manager.sh")

	scriptContent := `#!/bin/bash
echo "Action: $1"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	logger := zap.NewNop()
	exec := NewExecutor(logger, tmpDir, 5*time.Second)

	ctx := context.Background()
	result := exec.RestartInstance(ctx, "test-instance")

	if result.Error != nil {
		t.Errorf("RestartInstance failed: %v", result.Error)
	}

	if !strings.Contains(result.Stdout, "Action: restart") {
		t.Errorf("Expected 'Action: restart' in stdout, got %q", result.Stdout)
	}
}

func TestDeleteInstance(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "ipquorum-instance-manager.sh")

	scriptContent := `#!/bin/bash
echo "Action: $1"
echo "Instance: $2"
if [ "$3" = "--force" ]; then
    echo "Force: true"
fi
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	logger := zap.NewNop()
	exec := NewExecutor(logger, tmpDir, 5*time.Second)

	ctx := context.Background()

	// Test without force
	result := exec.DeleteInstance(ctx, "test-instance", false)
	if result.Error != nil {
		t.Errorf("DeleteInstance failed: %v", result.Error)
	}
	if !strings.Contains(result.Stdout, "Action: delete") {
		t.Errorf("Expected 'Action: delete' in stdout, got %q", result.Stdout)
	}

	// Test with force
	result = exec.DeleteInstance(ctx, "test-instance", true)
	if result.Error != nil {
		t.Errorf("DeleteInstance with force failed: %v", result.Error)
	}
	if !strings.Contains(result.Stdout, "Force: true") {
		t.Errorf("Expected 'Force: true' in stdout, got %q", result.Stdout)
	}
}
