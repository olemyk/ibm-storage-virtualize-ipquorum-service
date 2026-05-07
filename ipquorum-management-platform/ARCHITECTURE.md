# IP Quorum Management Platform - Architecture Design

## Overview

A modern Go-based management platform for IBM Storage Virtualize IP Quorum services with web dashboard, REST API, monitoring, and multi-server orchestration capabilities.

## Current State (v2.0.7)

```
┌─────────────────────────────────────────────────────────┐
│ Current Architecture (Production Ready)                 │
├─────────────────────────────────────────────────────────┤
│ Bash Scripts                                            │
│ ├─ install-ipquorum-service.sh (Installer)             │
│ ├─ ipquorum-instance-manager.sh (Instance Management)  │
│ └─ Systemd Integration (ipquorum@.service)             │
│                                                         │
│ Go Binary                                               │
│ └─ ipquorum-download-go (Download & API Operations)    │
└─────────────────────────────────────────────────────────┘
```

## Target Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    IP Quorum Management Platform                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┐         ┌──────────────────┐            │
│  │  Web Dashboard   │◄────────┤   REST API       │            │
│  │  (React + TS)    │         │   (Go + Gin)     │            │
│  └──────────────────┘         └──────────────────┘            │
│           │                            │                        │
│           │ WebSocket                  │ HTTP/JSON              │
│           ▼                            ▼                        │
│  ┌─────────────────────────────────────────────────┐          │
│  │         Management Server (Go)                   │          │
│  ├─────────────────────────────────────────────────┤          │
│  │ • Instance Management                            │          │
│  │ • Configuration Management                       │          │
│  │ • Health Monitoring                              │          │
│  │ • Metrics Collection                             │          │
│  │ • Multi-Server Orchestration                     │          │
│  │ • Authentication & Authorization                 │          │
│  └─────────────────────────────────────────────────┘          │
│           │                            │                        │
│           │                            │                        │
│           ▼                            ▼                        │
│  ┌──────────────────┐         ┌──────────────────┐            │
│  │  Local Agent     │         │   Database       │            │
│  │  (Go Binary)     │         │   (SQLite/Bolt)  │            │
│  └──────────────────┘         └──────────────────┘            │
│           │                                                     │
│           ▼                                                     │
│  ┌──────────────────────────────────────────────┐             │
│  │  Existing Bash Scripts (Backward Compatible) │             │
│  │  ├─ ipquorum-instance-manager.sh             │             │
│  │  └─ Systemd Services                         │             │
│  └──────────────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. Management Server (Go Backend)

**Purpose**: Central management and orchestration server

**Technology Stack**:
- **Framework**: Gin (HTTP router)
- **Database**: SQLite (single server) / PostgreSQL (multi-server)
- **Cache**: Redis (optional, for session management)
- **Metrics**: Prometheus client
- **Logging**: Zap (structured logging)

**Core Modules**:

```go
ipquorum-management-server/
├── cmd/
│   └── server/
│       └── main.go                 // Entry point
├── internal/
│   ├── api/                        // REST API handlers
│   │   ├── instances.go            // Instance CRUD operations
│   │   ├── monitoring.go           // Health & metrics endpoints
│   │   ├── orchestration.go        // Multi-server operations
│   │   └── auth.go                 // Authentication endpoints
│   ├── service/                    // Business logic
│   │   ├── instance_manager.go     // Instance lifecycle management
│   │   ├── health_checker.go       // Health monitoring
│   │   ├── metrics_collector.go    // Metrics aggregation
│   │   └── orchestrator.go         // Multi-server coordination
│   ├── agent/                      // Local agent communication
│   │   ├── client.go               // Agent client
│   │   └── protocol.go             // Agent protocol
│   ├── storage/                    // Data persistence
│   │   ├── database.go             // Database interface
│   │   ├── models.go               // Data models
│   │   └── migrations.go           // Schema migrations
│   └── auth/                       // Authentication & Authorization
│       ├── jwt.go                  // JWT token management
│       ├── rbac.go                 // Role-based access control
│       └── middleware.go           // Auth middleware
├── pkg/
│   ├── config/                     // Configuration management
│   ├── logger/                     // Logging utilities
│   └── errors/                     // Error handling
└── web/                            // Embedded web dashboard
    └── dist/                       // Built React app
```

### 2. Local Agent (Go Binary)

**Purpose**: Runs on each server to execute local operations

**Responsibilities**:
- Execute systemd commands (start, stop, restart)
- Monitor local instance health
- Collect local metrics
- Report status to management server
- Execute bash scripts for backward compatibility

```go
ipquorum-agent/
├── cmd/
│   └── agent/
│       └── main.go
├── internal/
│   ├── executor/                   // Command execution
│   │   ├── systemd.go              // Systemd operations
│   │   ├── bash.go                 // Bash script execution
│   │   └── health.go               // Health checks
│   ├── collector/                  // Metrics collection
│   │   ├── instance_metrics.go
│   │   └── system_metrics.go
│   └── reporter/                   // Report to server
│       └── client.go
└── pkg/
    └── protocol/                   // Communication protocol
```

### 3. Web Dashboard (React + TypeScript)

**Purpose**: Modern web interface for management

**Technology Stack**:
- **Framework**: React 18 + TypeScript
- **UI Library**: Carbon Design System (IBM)
- **State Management**: Redux Toolkit
- **API Client**: Axios
- **Real-time**: WebSocket
- **Charts**: Recharts / Chart.js
- **Build**: Vite

**Component Structure**:

```
ipquorum-dashboard/
├── src/
│   ├── components/
│   │   ├── Dashboard/              // Main dashboard
│   │   ├── Instances/              // Instance management
│   │   │   ├── InstanceList.tsx
│   │   │   ├── InstanceCreate.tsx
│   │   │   ├── InstanceDetails.tsx
│   │   │   └── InstanceActions.tsx
│   │   ├── Monitoring/             // Health & metrics
│   │   │   ├── HealthOverview.tsx
│   │   │   ├── MetricsCharts.tsx
│   │   │   └── AlertsList.tsx
│   │   ├── Servers/                // Multi-server view
│   │   │   ├── ServerList.tsx
│   │   │   └── ServerDetails.tsx
│   │   └── Settings/               // Configuration
│   ├── services/
│   │   ├── api.ts                  // API client
│   │   ├── websocket.ts            // WebSocket client
│   │   └── auth.ts                 // Authentication
│   ├── store/                      // Redux store
│   │   ├── instances.ts
│   │   ├── monitoring.ts
│   │   └── auth.ts
│   └── types/                      // TypeScript types
└── public/
```

## REST API Design

### Base URL
```
https://<server>:8443/api/v1
```

### Authentication
```
POST   /auth/login              // Login with credentials
POST   /auth/logout             // Logout
POST   /auth/refresh            // Refresh JWT token
GET    /auth/me                 // Get current user info
```

### Instance Management
```
GET    /instances               // List all instances
POST   /instances               // Create new instance
GET    /instances/:id           // Get instance details
PUT    /instances/:id           // Update instance
DELETE /instances/:id           // Delete instance
POST   /instances/:id/start     // Start instance
POST   /instances/:id/stop      // Stop instance
POST   /instances/:id/restart   // Restart instance
GET    /instances/:id/logs      // Get instance logs
GET    /instances/:id/status    // Get instance status
```

### Health & Monitoring
```
GET    /health                  // Overall system health
GET    /health/instances        // All instances health
GET    /health/instances/:id    // Specific instance health
GET    /metrics                 // Prometheus metrics
GET    /metrics/instances/:id   // Instance-specific metrics
```

### Multi-Server Orchestration
```
GET    /servers                 // List managed servers
POST   /servers                 // Register new server
GET    /servers/:id             // Get server details
DELETE /servers/:id             // Unregister server
GET    /servers/:id/instances   // List instances on server
POST   /orchestration/deploy    // Deploy to multiple servers
POST   /orchestration/sync      // Sync configuration
```

### Configuration
```
GET    /config                  // Get global configuration
PUT    /config                  // Update global configuration
GET    /config/templates        // Get instance templates
POST   /config/templates        // Create template
```

## Data Models

### Instance Model
```go
type Instance struct {
    ID                string    `json:"id" db:"id"`
    Name              string    `json:"name" db:"name"`
    ServerID          string    `json:"server_id" db:"server_id"`
    APIEndpoint       string    `json:"api_endpoint" db:"api_endpoint"`
    Username          string    `json:"username" db:"username"`
    StorageSystem     string    `json:"storage_system" db:"storage_system"`
    Description       string    `json:"description" db:"description"`
    Location          string    `json:"location" db:"location"`
    Status            string    `json:"status" db:"status"` // running, stopped, error
    Health            string    `json:"health" db:"health"` // healthy, degraded, unhealthy
    EnableDownload    bool      `json:"enable_download" db:"enable_download"`
    EnableMkquorumapp bool      `json:"enable_mkquorumapp" db:"enable_mkquorumapp"`
    CreatedAt         time.Time `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
```

### Server Model
```go
type Server struct {
    ID          string    `json:"id" db:"id"`
    Hostname    string    `json:"hostname" db:"hostname"`
    IPAddress   string    `json:"ip_address" db:"ip_address"`
    AgentPort   int       `json:"agent_port" db:"agent_port"`
    Status      string    `json:"status" db:"status"` // online, offline, error
    Version     string    `json:"version" db:"version"`
    LastSeen    time.Time `json:"last_seen" db:"last_seen"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
```

### Health Check Model
```go
type HealthCheck struct {
    InstanceID    string    `json:"instance_id"`
    Timestamp     time.Time `json:"timestamp"`
    Status        string    `json:"status"` // healthy, degraded, unhealthy
    Port7443      bool      `json:"port_7443"`
    Port1260      bool      `json:"port_1260"`
    ProcessRunning bool     `json:"process_running"`
    ResponseTime  int64     `json:"response_time_ms"`
    ErrorMessage  string    `json:"error_message,omitempty"`
}
```

### Metrics Model
```go
type Metrics struct {
    InstanceID      string    `json:"instance_id"`
    Timestamp       time.Time `json:"timestamp"`
    CPUUsage        float64   `json:"cpu_usage"`
    MemoryUsage     int64     `json:"memory_usage_bytes"`
    NetworkRxBytes  int64     `json:"network_rx_bytes"`
    NetworkTxBytes  int64     `json:"network_tx_bytes"`
    Uptime          int64     `json:"uptime_seconds"`
}
```

## Database Schema

### SQLite Schema (Single Server)
```sql
-- Instances table
CREATE TABLE instances (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    server_id TEXT DEFAULT 'local',
    api_endpoint TEXT NOT NULL,
    username TEXT NOT NULL,
    storage_system TEXT,
    description TEXT,
    location TEXT,
    status TEXT NOT NULL,
    health TEXT NOT NULL,
    enable_download BOOLEAN DEFAULT 1,
    enable_mkquorumapp BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Health checks table
CREATE TABLE health_checks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    instance_id TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL,
    port_7443 BOOLEAN,
    port_1260 BOOLEAN,
    process_running BOOLEAN,
    response_time_ms INTEGER,
    error_message TEXT,
    FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE
);

-- Metrics table
CREATE TABLE metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    instance_id TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    cpu_usage REAL,
    memory_usage_bytes INTEGER,
    network_rx_bytes INTEGER,
    network_tx_bytes INTEGER,
    uptime_seconds INTEGER,
    FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE
);

-- Servers table (for multi-server)
CREATE TABLE servers (
    id TEXT PRIMARY KEY,
    hostname TEXT UNIQUE NOT NULL,
    ip_address TEXT NOT NULL,
    agent_port INTEGER DEFAULT 8444,
    status TEXT NOT NULL,
    version TEXT,
    last_seen DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Users table (authentication)
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL, -- admin, operator, viewer
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME
);

-- Audit log table
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    details TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

## Security Architecture

### Authentication
- **JWT Tokens**: Stateless authentication
- **Token Expiry**: 1 hour access token, 7 days refresh token
- **Password Hashing**: bcrypt with cost factor 12
- **Session Management**: Redis-backed (optional)

### Authorization (RBAC)
```
Roles:
├── Admin
│   ├── Full system access
│   ├── User management
│   └── Configuration changes
├── Operator
│   ├── Instance management (create, start, stop, delete)
│   ├── View monitoring
│   └── No user/config changes
└── Viewer
    ├── Read-only access
    └── View instances and monitoring
```

### TLS/SSL
- **Management Server**: TLS 1.3, self-signed or Let's Encrypt
- **Agent Communication**: Mutual TLS (mTLS)
- **Web Dashboard**: HTTPS only

### API Security
- Rate limiting (100 req/min per IP)
- CORS configuration
- Input validation
- SQL injection prevention (parameterized queries)
- XSS protection

## Monitoring & Observability

### Metrics (Prometheus)
```
# Instance metrics
ipquorum_instance_status{instance="name", server="hostname"}
ipquorum_instance_health{instance="name", server="hostname"}
ipquorum_instance_uptime_seconds{instance="name"}
ipquorum_instance_cpu_usage{instance="name"}
ipquorum_instance_memory_bytes{instance="name"}

# System metrics
ipquorum_api_requests_total{method="GET", endpoint="/instances", status="200"}
ipquorum_api_request_duration_seconds{method="GET", endpoint="/instances"}
ipquorum_health_checks_total{instance="name", result="success"}
ipquorum_servers_online{server="hostname"}
```

### Logging
```
Structured JSON logging with levels:
- DEBUG: Detailed debugging information
- INFO: General informational messages
- WARN: Warning messages
- ERROR: Error messages
- FATAL: Fatal errors causing shutdown

Log fields:
- timestamp
- level
- message
- instance_id (if applicable)
- user_id (if applicable)
- request_id (for API calls)
- error (if error occurred)
```

### Health Checks
```
Endpoint: GET /health

Response:
{
  "status": "healthy",
  "timestamp": "2026-04-24T21:00:00Z",
  "components": {
    "database": "healthy",
    "agents": "healthy",
    "instances": {
      "total": 5,
      "healthy": 4,
      "degraded": 1,
      "unhealthy": 0
    }
  }
}
```

## Deployment Architecture

### Single Server Deployment
```
┌─────────────────────────────────────┐
│         Linux Server                │
├─────────────────────────────────────┤
│ Management Server (Port 8443)      │
│ ├─ REST API                         │
│ ├─ Web Dashboard                    │
│ └─ Database (SQLite)                │
│                                     │
│ Local Agent (Port 8444)             │
│ └─ Executes local operations        │
│                                     │
│ IP Quorum Instances                 │
│ ├─ instance1 (systemd)              │
│ ├─ instance2 (systemd)              │
│ └─ instance3 (systemd)              │
└─────────────────────────────────────┘
```

### Multi-Server Deployment
```
┌──────────────────────────────────────────────────────────┐
│              Management Server                           │
│              (Central Control)                           │
├──────────────────────────────────────────────────────────┤
│ Management Server (Port 8443)                           │
│ ├─ REST API                                              │
│ ├─ Web Dashboard                                         │
│ ├─ Database (PostgreSQL)                                │
│ └─ Orchestration Engine                                 │
└──────────────────────────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  Server 1    │ │  Server 2    │ │  Server 3    │
├──────────────┤ ├──────────────┤ ├──────────────┤
│ Agent        │ │ Agent        │ │ Agent        │
│ Instances    │ │ Instances    │ │ Instances    │
└──────────────┘ └──────────────┘ └──────────────┘
```

## Installation & Deployment

### Management Server Installation
```bash
# Download binary
wget https://github.com/olemyk/ipquorum-platform/releases/latest/download/ipquorum-server-linux-amd64

# Install
sudo mv ipquorum-server-linux-amd64 /usr/local/bin/ipquorum-server
sudo chmod +x /usr/local/bin/ipquorum-server

# Create systemd service
sudo systemctl enable ipquorum-server
sudo systemctl start ipquorum-server

# Access web dashboard
https://localhost:8443
```

### Agent Installation
```bash
# Download binary
wget https://github.com/olemyk/ipquorum-platform/releases/latest/download/ipquorum-agent-linux-amd64

# Install
sudo mv ipquorum-agent-linux-amd64 /usr/local/bin/ipquorum-agent
sudo chmod +x /usr/local/bin/ipquorum-agent

# Configure
sudo ipquorum-agent configure --server https://management-server:8443

# Start
sudo systemctl enable ipquorum-agent
sudo systemctl start ipquorum-agent
```

## Backward Compatibility

### Integration with Existing Bash Scripts
```go
// Agent executes bash scripts for operations
func (a *Agent) CreateInstance(config InstanceConfig) error {
    // Call existing bash script
    cmd := exec.Command(
        "/usr/local/bin/ipquorum-instance-manager.sh",
        "create",
        config.Name,
        "--non-interactive",
    )
    
    // Set environment variables
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("API_ENDPOINT=%s", config.APIEndpoint),
        fmt.Sprintf("VIRTUALIZE_USERNAME=%s", config.Username),
        // ... other env vars
    )
    
    return cmd.Run()
}
```

### Migration Path
1. **Phase 1**: Install management server alongside existing setup
2. **Phase 2**: Import existing instances into management platform
3. **Phase 3**: Use web dashboard for new instances
4. **Phase 4**: Gradually migrate management to platform
5. **Phase 5**: Bash scripts remain as execution layer

## Development Roadmap

### Phase 1: Foundation (v3.0.0)
- [ ] Management server core
- [ ] REST API implementation
- [ ] SQLite database
- [ ] Basic authentication
- [ ] Local agent
- [ ] Bash script integration

### Phase 2: Web Dashboard (v3.1.0)
- [ ] React dashboard
- [ ] Instance management UI
- [ ] Real-time status updates
- [ ] Basic monitoring charts

### Phase 3: Monitoring (v3.2.0)
- [ ] Health check system
- [ ] Metrics collection
- [ ] Prometheus integration
- [ ] Alerting system

### Phase 4: Multi-Server (v3.3.0)
- [ ] Server registration
- [ ] Multi-server orchestration
- [ ] Distributed health checks
- [ ] PostgreSQL support

### Phase 5: Advanced Features (v3.4.0)
- [ ] Advanced RBAC
- [ ] Audit logging
- [ ] Backup/restore
- [ ] Configuration templates
- [ ] Bulk operations

## Technology Decisions

### Why Go?
- ✅ Single binary deployment
- ✅ Excellent concurrency (goroutines)
- ✅ Strong standard library
- ✅ Cross-platform compilation
- ✅ Great performance
- ✅ Existing codebase (ipquorum-download-go)

### Why React + TypeScript?
- ✅ Modern, component-based UI
- ✅ Type safety with TypeScript
- ✅ Large ecosystem
- ✅ Carbon Design System (IBM)
- ✅ Excellent developer experience

### Why SQLite (initially)?
- ✅ Zero configuration
- ✅ Embedded database
- ✅ Perfect for single server
- ✅ Easy migration to PostgreSQL later

### Why Gin Framework?
- ✅ Fast HTTP router
- ✅ Middleware support
- ✅ Good documentation
- ✅ Active community

## Performance Targets

- **API Response Time**: < 100ms (p95)
- **Dashboard Load Time**: < 2s
- **Health Check Interval**: 30s
- **Metrics Collection**: 60s
- **Concurrent Instances**: 100+ per server
- **API Throughput**: 1000 req/s

## Conclusion

This architecture provides a solid foundation for building a modern, scalable IP Quorum management platform while maintaining backward compatibility with existing Bash-based infrastructure. The phased approach allows for incremental development and deployment without disrupting current operations.