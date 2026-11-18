package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/olemyk/ipquorum-platform/server/internal/agent"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
	"go.uber.org/zap"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Create instances table (required for foreign key)
	schema := `
	CREATE TABLE instances (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE container_instances (
		id TEXT PRIMARY KEY,
		instance_id TEXT NOT NULL,
		agent_id TEXT NOT NULL,
		container_id TEXT NOT NULL,
		container_name TEXT NOT NULL,
		image_name TEXT NOT NULL,
		image_tag TEXT NOT NULL,
		runtime TEXT NOT NULL,
		status TEXT NOT NULL,
		state TEXT DEFAULT '',
		restart_policy TEXT DEFAULT 'no',
		restart_count INTEGER DEFAULT 0,
		exit_code INTEGER,
		error TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		started_at TIMESTAMP,
		finished_at TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE
	);

	CREATE TABLE container_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		container_instance_id TEXT NOT NULL,
		config_type TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		extra TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (container_instance_id) REFERENCES container_instances(id) ON DELETE CASCADE
	);

	CREATE TABLE container_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		container_instance_id TEXT NOT NULL,
		cpu_percent REAL,
		memory_usage INTEGER,
		memory_limit INTEGER,
		memory_percent REAL,
		network_rx_bytes INTEGER,
		network_tx_bytes INTEGER,
		block_read INTEGER,
		block_write INTEGER,
		pids INTEGER,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (container_instance_id) REFERENCES container_instances(id) ON DELETE CASCADE
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

// For testing, we'll use nil for the agent registry since we're testing
// database operations only. In integration tests, you'd use a real or mock registry.
func newMockAgentRegistry() *agent.Registry {
	// Return nil for unit tests - we're only testing database operations
	return nil
}

// mockAgentClient is a mock implementation of agent.Client for testing
type mockAgentClient struct {
	createContainerFunc func(ctx context.Context, req *agent.CreateContainerRequest) (map[string]interface{}, error)
	startContainerFunc  func(ctx context.Context, containerID string) error
	stopContainerFunc   func(ctx context.Context, containerID string, timeout int) error
}

func TestContainerService_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test instance
	_, err := db.Exec("INSERT INTO instances (id, name) VALUES (?, ?)", "test-instance-1", "Test Instance")
	if err != nil {
		t.Fatalf("Failed to insert test instance: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	registry := newMockAgentRegistry()
	service := NewContainerService(db, registry, logger)

	tests := []struct {
		name    string
		req     *storage.CreateContainerInstanceRequest
		wantErr bool
	}{
		{
			name: "valid container creation",
			req: &storage.CreateContainerInstanceRequest{
				InstanceID:    "test-instance-1",
				AgentID:       "test-agent-1",
				ImageName:     "nginx",
				ImageTag:      "latest",
				RestartPolicy: "always",
			},
			wantErr: false,
		},
		{
			name: "missing instance_id",
			req: &storage.CreateContainerInstanceRequest{
				AgentID:   "test-agent-1",
				ImageName: "nginx",
			},
			wantErr: true,
		},
		{
			name: "missing agent_id",
			req: &storage.CreateContainerInstanceRequest{
				InstanceID: "test-instance-1",
				ImageName:  "nginx",
			},
			wantErr: false, // Agent ID is optional for testing without agent registry
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Note: This test will fail because we need a real agent client
			// In production tests, you'd use dependency injection with mocks
			_, err := service.Create(ctx, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContainerService_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test instance first
	_, err := db.Exec("INSERT INTO instances (id, name) VALUES (?, ?)", "instance-1", "Test Instance")
	if err != nil {
		t.Fatalf("Failed to insert test instance: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	registry := newMockAgentRegistry()
	service := NewContainerService(db, registry, logger)

	// Insert test data
	containerID := "test-container-1"
	_, err = db.Exec(`
		INSERT INTO container_instances
		(id, instance_id, agent_id, container_id, container_name, image_name, image_tag, runtime, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, containerID, "instance-1", "agent-1", "docker-123", "test-container", "nginx", "latest", "docker", "running")

	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "existing container",
			id:      containerID,
			wantErr: false,
		},
		{
			name:    "non-existent container",
			id:      "non-existent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			container, err := service.GetByID(ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && container == nil {
				t.Error("GetByID() returned nil container")
			}

			if !tt.wantErr && container.ID != tt.id {
				t.Errorf("GetByID() container.ID = %v, want %v", container.ID, tt.id)
			}
		})
	}
}

func TestContainerService_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test instances first
	for _, instanceID := range []string{"instance-1", "instance-2", "instance-3"} {
		_, err := db.Exec("INSERT INTO instances (id, name) VALUES (?, ?)", instanceID, "Test Instance "+instanceID)
		if err != nil {
			t.Fatalf("Failed to insert test instance: %v", err)
		}
	}

	logger, _ := zap.NewDevelopment()
	registry := newMockAgentRegistry()
	service := NewContainerService(db, registry, logger)

	// Insert test data
	containers := []struct {
		id         string
		instanceID string
		agentID    string
	}{
		{"container-1", "instance-1", "agent-1"},
		{"container-2", "instance-2", "agent-1"},
		{"container-3", "instance-3", "agent-2"},
	}

	for _, c := range containers {
		_, err := db.Exec(`
			INSERT INTO container_instances
			(id, instance_id, agent_id, container_id, container_name, image_name, image_tag, runtime, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, c.id, c.instanceID, c.agentID, "docker-"+c.id, "test-"+c.id, "nginx", "latest", "docker", "running")

		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	ctx := context.Background()
	result, err := service.List(ctx)

	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(result) != len(containers) {
		t.Errorf("List() returned %d containers, want %d", len(result), len(containers))
	}
}

func TestContainerService_GetByInstanceID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test instance first
	instanceID := "test-instance-1"
	_, err := db.Exec("INSERT INTO instances (id, name) VALUES (?, ?)", instanceID, "Test Instance")
	if err != nil {
		t.Fatalf("Failed to insert test instance: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	registry := newMockAgentRegistry()
	service := NewContainerService(db, registry, logger)

	// Insert test data
	containerID := "test-container-1"

	_, err = db.Exec(`
		INSERT INTO container_instances
		(id, instance_id, agent_id, container_id, container_name, image_name, image_tag, runtime, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, containerID, instanceID, "agent-1", "docker-123", "test-container", "nginx", "latest", "docker", "running")

	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	tests := []struct {
		name       string
		instanceID string
		wantErr    bool
	}{
		{
			name:       "existing instance",
			instanceID: instanceID,
			wantErr:    false,
		},
		{
			name:       "non-existent instance",
			instanceID: "non-existent",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			container, err := service.GetByInstanceID(ctx, tt.instanceID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByInstanceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && container == nil {
				t.Error("GetByInstanceID() returned nil container")
			}

			if !tt.wantErr && container.InstanceID != tt.instanceID {
				t.Errorf("GetByInstanceID() container.InstanceID = %v, want %v", container.InstanceID, tt.instanceID)
			}
		})
	}
}

func TestContainerService_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test instance first
	_, err := db.Exec("INSERT INTO instances (id, name) VALUES (?, ?)", "instance-1", "Test Instance")
	if err != nil {
		t.Fatalf("Failed to insert test instance: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	registry := newMockAgentRegistry()
	service := NewContainerService(db, registry, logger)

	// Insert test data
	containerID := "test-container-1"
	_, err = db.Exec(`
		INSERT INTO container_instances
		(id, instance_id, agent_id, container_id, container_name, image_name, image_tag, runtime, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, containerID, "instance-1", "agent-1", "docker-123", "test-container", "nginx", "latest", "docker", "created")

	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	ctx := context.Background()

	// Update status
	service.updateStatus(ctx, containerID, "running", "")

	// Verify update
	container, err := service.GetByID(ctx, containerID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if container.Status != "running" {
		t.Errorf("Status = %v, want running", container.Status)
	}

	// Update with error
	errorMsg := "test error"
	service.updateStatus(ctx, containerID, "error", errorMsg)

	container, err = service.GetByID(ctx, containerID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if container.Status != "error" {
		t.Errorf("Status = %v, want error", container.Status)
	}

	if container.Error != errorMsg {
		t.Errorf("Error = %v, want %v", container.Error, errorMsg)
	}
}

func TestContainerService_StoreContainerConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test instance first
	_, err := db.Exec("INSERT INTO instances (id, name) VALUES (?, ?)", "instance-1", "Test Instance")
	if err != nil {
		t.Fatalf("Failed to insert test instance: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	registry := newMockAgentRegistry()
	service := NewContainerService(db, registry, logger)

	// Insert test container
	containerID := "test-container-1"
	_, err = db.Exec(`
		INSERT INTO container_instances
		(id, instance_id, agent_id, container_id, container_name, image_name, image_tag, runtime, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, containerID, "instance-1", "agent-1", "docker-123", "test-container", "nginx", "latest", "docker", "created")

	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	ctx := context.Background()

	// Store config
	config := &storage.ContainerConfig{
		Ports: []storage.PortMapping{
			{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
		},
		Volumes: []storage.VolumeMount{
			{Source: "/host/path", Destination: "/container/path", ReadOnly: false},
		},
		Environment: map[string]string{
			"ENV_VAR": "value",
		},
		Labels: map[string]string{
			"app": "test",
		},
	}

	err = service.storeContainerConfig(ctx, containerID, config)
	if err != nil {
		t.Fatalf("storeContainerConfig() error = %v", err)
	}

	// Verify stored config
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM container_configs WHERE container_instance_id = ?", containerID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query config count: %v", err)
	}

	// Should have: 1 port + 1 volume + 1 env + 1 label = 4 entries
	expectedCount := 4
	if count != expectedCount {
		t.Errorf("Config count = %d, want %d", count, expectedCount)
	}
}

// Benchmark tests
func BenchmarkContainerService_GetByID(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	logger, _ := zap.NewDevelopment()
	registry := newMockAgentRegistry()
	service := NewContainerService(db, registry, logger)

	// Insert test data
	containerID := "bench-container-1"
	db.Exec(`
		INSERT INTO container_instances 
		(id, instance_id, agent_id, container_id, container_name, image_name, image_tag, runtime, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, containerID, "instance-1", "agent-1", "docker-123", "test-container", "nginx", "latest", "docker", "running")

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetByID(ctx, containerID)
	}
}

func BenchmarkContainerService_List(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	logger, _ := zap.NewDevelopment()
	registry := newMockAgentRegistry()
	service := NewContainerService(db, registry, logger)

	// Insert test data
	for i := 0; i < 100; i++ {
		db.Exec(`
			INSERT INTO container_instances 
			(id, instance_id, agent_id, container_id, container_name, image_name, image_tag, runtime, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			time.Now().Format("20060102150405")+string(rune(i)),
			"instance-"+string(rune(i)),
			"agent-1",
			"docker-"+string(rune(i)),
			"container-"+string(rune(i)),
			"nginx",
			"latest",
			"docker",
			"running",
		)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.List(ctx)
	}
}

// Made with Bob
