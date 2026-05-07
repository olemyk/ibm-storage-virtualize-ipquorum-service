package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Registry manages agent clients
type Registry struct {
	agents map[string]*Client
	mu     sync.RWMutex
	log    *zap.Logger
}

// NewRegistry creates a new agent registry
func NewRegistry(log *zap.Logger) *Registry {
	return &Registry{
		agents: make(map[string]*Client),
		log:    log,
	}
}

// Register registers a new agent
func (r *Registry) Register(serverID string, cfg Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	client := NewClient(cfg)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Health(ctx); err != nil {
		return fmt.Errorf("failed to connect to agent: %w", err)
	}

	r.agents[serverID] = client
	r.log.Info("Agent registered",
		zap.String("server_id", serverID),
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
	)

	return nil
}

// Unregister removes an agent
func (r *Registry) Unregister(serverID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.agents, serverID)
	r.log.Info("Agent unregistered", zap.String("server_id", serverID))
}

// Get retrieves an agent client
func (r *Registry) Get(serverID string) (*Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, exists := r.agents[serverID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", serverID)
	}

	return client, nil
}

// List returns all registered agent server IDs
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	serverIDs := make([]string, 0, len(r.agents))
	for id := range r.agents {
		serverIDs = append(serverIDs, id)
	}

	return serverIDs
}

// Count returns the number of registered agents
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.agents)
}

// HealthCheck checks health of all agents
func (r *Registry) HealthCheck(ctx context.Context) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)
	for serverID, client := range r.agents {
		results[serverID] = client.Health(ctx)
	}

	return results
}

// HealthCheckOne checks health of a specific agent
func (r *Registry) HealthCheckOne(ctx context.Context, serverID string) error {
	client, err := r.Get(serverID)
	if err != nil {
		return err
	}

	return client.Health(ctx)
}
