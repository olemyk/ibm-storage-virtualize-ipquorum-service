package pool

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/olemyk/ipquorum-management-platform/container-runtime-agent/pkg/runtime"
	"go.uber.org/zap"
)

var (
	// ErrPoolClosed is returned when trying to get a connection from a closed pool
	ErrPoolClosed = errors.New("connection pool is closed")
	// ErrPoolExhausted is returned when the pool has no available connections
	ErrPoolExhausted = errors.New("connection pool exhausted")
	// ErrInvalidConnection is returned when a connection is invalid
	ErrInvalidConnection = errors.New("invalid connection")
)

// Connection represents a pooled runtime connection
type Connection struct {
	runtime   runtime.Runtime
	pool      *Pool
	createdAt time.Time
	lastUsed  time.Time
	useCount  int64
}

// Runtime returns the underlying runtime
func (c *Connection) Runtime() runtime.Runtime {
	return c.runtime
}

// Close returns the connection to the pool
func (c *Connection) Close() error {
	return c.pool.Put(c)
}

// Pool manages a pool of runtime connections
type Pool struct {
	mu sync.RWMutex

	// Configuration
	factory     func() (runtime.Runtime, error)
	minSize     int
	maxSize     int
	maxIdleTime time.Duration
	maxLifetime time.Duration
	healthCheck time.Duration

	// State
	connections []*Connection
	idle        []*Connection
	active      map[*Connection]bool
	closed      bool

	// Metrics
	totalCreated int64
	totalReused  int64
	totalClosed  int64

	// Channels
	connChan  chan *Connection
	closeChan chan struct{}

	logger *zap.Logger
}

// Config holds pool configuration
type Config struct {
	MinSize     int           // Minimum number of connections
	MaxSize     int           // Maximum number of connections
	MaxIdleTime time.Duration // Maximum time a connection can be idle
	MaxLifetime time.Duration // Maximum lifetime of a connection
	HealthCheck time.Duration // Health check interval
}

// DefaultConfig returns default pool configuration
func DefaultConfig() *Config {
	return &Config{
		MinSize:     2,
		MaxSize:     10,
		MaxIdleTime: 5 * time.Minute,
		MaxLifetime: 30 * time.Minute,
		HealthCheck: 1 * time.Minute,
	}
}

// NewPool creates a new connection pool
func NewPool(factory func() (runtime.Runtime, error), config *Config, logger *zap.Logger) (*Pool, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.MinSize < 0 {
		config.MinSize = 0
	}
	if config.MaxSize < config.MinSize {
		config.MaxSize = config.MinSize
	}
	if config.MaxSize == 0 {
		config.MaxSize = 10
	}

	p := &Pool{
		factory:     factory,
		minSize:     config.MinSize,
		maxSize:     config.MaxSize,
		maxIdleTime: config.MaxIdleTime,
		maxLifetime: config.MaxLifetime,
		healthCheck: config.HealthCheck,
		connections: make([]*Connection, 0, config.MaxSize),
		idle:        make([]*Connection, 0, config.MaxSize),
		active:      make(map[*Connection]bool),
		connChan:    make(chan *Connection, config.MaxSize),
		closeChan:   make(chan struct{}),
		logger:      logger,
	}

	// Create minimum connections
	for i := 0; i < config.MinSize; i++ {
		conn, err := p.createConnection()
		if err != nil {
			p.logger.Warn("Failed to create initial connection", zap.Error(err))
			continue
		}
		p.connections = append(p.connections, conn)
		p.idle = append(p.idle, conn)
	}

	// Start maintenance goroutine
	go p.maintain()

	return p, nil
}

// Get retrieves a connection from the pool
func (p *Pool) Get(ctx context.Context) (*Connection, error) {
	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return nil, ErrPoolClosed
	}

	// Try to get an idle connection
	if len(p.idle) > 0 {
		conn := p.idle[len(p.idle)-1]
		p.idle = p.idle[:len(p.idle)-1]
		p.active[conn] = true
		conn.lastUsed = time.Now()
		conn.useCount++
		p.totalReused++
		p.mu.Unlock()

		// Validate connection
		if err := p.validateConnection(ctx, conn); err != nil {
			p.logger.Warn("Connection validation failed", zap.Error(err))
			p.removeConnection(conn)
			return p.Get(ctx) // Retry
		}

		return conn, nil
	}

	// Create new connection if under max size
	if len(p.connections) < p.maxSize {
		conn, err := p.createConnection()
		if err != nil {
			p.mu.Unlock()
			return nil, err
		}
		p.connections = append(p.connections, conn)
		p.active[conn] = true
		conn.lastUsed = time.Now()
		conn.useCount++
		p.mu.Unlock()
		return conn, nil
	}

	p.mu.Unlock()

	// Wait for available connection
	select {
	case conn := <-p.connChan:
		p.mu.Lock()
		p.active[conn] = true
		conn.lastUsed = time.Now()
		conn.useCount++
		p.totalReused++
		p.mu.Unlock()
		return conn, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.closeChan:
		return nil, ErrPoolClosed
	}
}

// Put returns a connection to the pool
func (p *Pool) Put(conn *Connection) error {
	if conn == nil {
		return ErrInvalidConnection
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return p.closeConnection(conn)
	}

	delete(p.active, conn)

	// Check if connection should be closed
	if p.shouldCloseConnection(conn) {
		return p.closeConnection(conn)
	}

	// Return to idle pool
	p.idle = append(p.idle, conn)

	// Try to send to waiting goroutine
	select {
	case p.connChan <- conn:
		p.idle = p.idle[:len(p.idle)-1]
	default:
	}

	return nil
}

// Close closes the pool and all connections
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	close(p.closeChan)

	// Close all connections
	for _, conn := range p.connections {
		p.closeConnection(conn)
	}

	p.connections = nil
	p.idle = nil
	p.active = nil

	return nil
}

// Stats returns pool statistics
func (p *Pool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PoolStats{
		TotalConnections:  len(p.connections),
		IdleConnections:   len(p.idle),
		ActiveConnections: len(p.active),
		TotalCreated:      p.totalCreated,
		TotalReused:       p.totalReused,
		TotalClosed:       p.totalClosed,
	}
}

// PoolStats holds pool statistics
type PoolStats struct {
	TotalConnections  int
	IdleConnections   int
	ActiveConnections int
	TotalCreated      int64
	TotalReused       int64
	TotalClosed       int64
}

// Private methods

func (p *Pool) createConnection() (*Connection, error) {
	rt, err := p.factory()
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		runtime:   rt,
		pool:      p,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		useCount:  0,
	}

	p.totalCreated++
	p.logger.Debug("Created new connection", zap.Int64("total_created", p.totalCreated))

	return conn, nil
}

func (p *Pool) closeConnection(conn *Connection) error {
	if conn == nil {
		return nil
	}

	// Remove from connections slice
	for i, c := range p.connections {
		if c == conn {
			p.connections = append(p.connections[:i], p.connections[i+1:]...)
			break
		}
	}

	// Remove from idle slice
	for i, c := range p.idle {
		if c == conn {
			p.idle = append(p.idle[:i], p.idle[i+1:]...)
			break
		}
	}

	delete(p.active, conn)
	p.totalClosed++

	p.logger.Debug("Closed connection",
		zap.Int64("use_count", conn.useCount),
		zap.Duration("lifetime", time.Since(conn.createdAt)))

	return nil
}

func (p *Pool) removeConnection(conn *Connection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closeConnection(conn)
}

func (p *Pool) shouldCloseConnection(conn *Connection) bool {
	now := time.Now()

	// Check max lifetime
	if p.maxLifetime > 0 && now.Sub(conn.createdAt) > p.maxLifetime {
		return true
	}

	// Check max idle time
	if p.maxIdleTime > 0 && now.Sub(conn.lastUsed) > p.maxIdleTime {
		return true
	}

	// Check if over min size
	if len(p.connections) > p.minSize && len(p.idle) > p.minSize {
		return true
	}

	return false
}

func (p *Pool) validateConnection(ctx context.Context, conn *Connection) error {
	if conn == nil || conn.runtime == nil {
		return ErrInvalidConnection
	}

	// Ping the runtime
	return conn.runtime.Ping(ctx)
}

func (p *Pool) maintain() {
	ticker := time.NewTicker(p.healthCheck)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.performMaintenance()
		case <-p.closeChan:
			return
		}
	}
}

func (p *Pool) performMaintenance() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check idle connections
	var toRemove []*Connection
	for _, conn := range p.idle {
		if p.shouldCloseConnection(conn) {
			toRemove = append(toRemove, conn)
			continue
		}

		// Health check
		if err := p.validateConnection(ctx, conn); err != nil {
			p.logger.Warn("Connection health check failed", zap.Error(err))
			toRemove = append(toRemove, conn)
		}
	}

	// Remove unhealthy connections
	for _, conn := range toRemove {
		p.closeConnection(conn)
	}

	// Ensure minimum connections
	for len(p.connections) < p.minSize {
		conn, err := p.createConnection()
		if err != nil {
			p.logger.Warn("Failed to create connection during maintenance", zap.Error(err))
			break
		}
		p.connections = append(p.connections, conn)
		p.idle = append(p.idle, conn)
	}

	p.logger.Debug("Pool maintenance completed",
		zap.Int("total", len(p.connections)),
		zap.Int("idle", len(p.idle)),
		zap.Int("active", len(p.active)))
}

// Made with Bob
