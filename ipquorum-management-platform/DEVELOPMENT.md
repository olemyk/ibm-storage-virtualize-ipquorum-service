# IP Quorum Management Platform - Development Guide

## Prerequisites

### Required Software
- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Git** - Version control
- **Make** (optional) - Build automation
- **jq** - JSON processing for testing
- **curl** - API testing

### System Requirements
- Linux/macOS (recommended) or Windows with WSL2
- 2GB RAM minimum
- 500MB disk space

## Project Structure

```
ipquorum-management-platform/
├── server/                      # Management server
│   ├── cmd/server/             # Main entry point
│   ├── internal/               # Private application code
│   │   ├── api/               # REST API handlers
│   │   ├── auth/              # Authentication & authorization
│   │   ├── executor/          # Bash script executor
│   │   ├── health/            # Health check system
│   │   ├── service/           # Business logic
│   │   └── storage/           # Database layer
│   ├── pkg/                    # Public libraries
│   │   ├── config/            # Configuration management
│   │   └── logger/            # Logging utilities
│   └── go.mod                  # Go dependencies
├── agent/                       # Local agent (future)
├── web/                         # Web dashboard (future)
├── ARCHITECTURE.md             # Architecture documentation
├── README.md                   # User documentation
└── DEVELOPMENT.md              # This file
```

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service.git
cd ibm-storage-virtualize-ipquorum-service/ipquorum-management-platform
```

### 2. Install Dependencies

```bash
cd server
go mod download
go mod verify
```

### 3. Build the Server

```bash
# Development build
go build -o bin/ipquorum-server ./cmd/server

# Production build with version info
go build -ldflags="-s -w -X main.version=dev -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/ipquorum-server ./cmd/server
```

### 4. Initialize Configuration

```bash
# Create default configuration
./bin/ipquorum-server init

# This creates:
# - /etc/ipquorum-platform/config.yaml (or ./config.yaml if no permissions)
# - /var/lib/ipquorum-platform/ipquorum.db (SQLite database)
```

### 5. Create Admin User

The init command creates a default admin user:
- **Username**: `admin`
- **Password**: `changeme` (randomly generated, shown during init)

**IMPORTANT**: Change the default password immediately!

## Development Workflow

### Running the Server

#### Development Mode (HTTP, debug logging)

```bash
# Run with debug logging
./bin/ipquorum-server run --debug

# Or with custom config
./bin/ipquorum-server run --config ./dev-config.yaml --debug
```

#### Production Mode (HTTPS, info logging)

```bash
# Generate TLS certificates first
mkdir -p /etc/ipquorum-platform/tls
openssl req -x509 -newkey rsa:4096 -keyout /etc/ipquorum-platform/tls/server.key \
  -out /etc/ipquorum-platform/tls/server.crt -days 365 -nodes \
  -subj "/CN=localhost"

# Run server
./bin/ipquorum-server run --config /etc/ipquorum-platform/config.yaml
```

### Configuration

Edit `config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8443
  tls_cert: "/etc/ipquorum-platform/tls/server.crt"
  tls_key: "/etc/ipquorum-platform/tls/server.key"
  log_level: "info"  # debug, info, warn, error
  health_check_interval: 30  # seconds

database:
  type: "sqlite"  # sqlite or postgres
  path: "/var/lib/ipquorum-platform/ipquorum.db"

auth:
  jwt_secret: "your-secret-key-here"  # Auto-generated during init
  token_expiry: 3600      # 1 hour
  refresh_expiry: 604800  # 7 days

agent:
  port: 8444
  scripts_dir: "/opt/ipquorum/scripts"  # Path to bash scripts
  health_interval: 30
  metrics_interval: 60
```

### Testing the API

#### 1. Get Authentication Token

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"changeme"}' | jq -r '.token')

echo "Token: $TOKEN"
```

#### 2. Create an Instance

```bash
curl -X POST http://localhost:8443/api/v1/instances \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-quorum-01",
    "api_endpoint": "10.33.7.80",
    "username": "superuser",
    "password": "your-password",
    "storage_system": "SVC Cluster 01",
    "description": "Test IP Quorum instance",
    "location": "Datacenter A",
    "enable_download": true,
    "enable_mkquorumapp": true,
    "partnersystem": "svc_cluster02"
  }' | jq
```

#### 3. List Instances

```bash
curl -s http://localhost:8443/api/v1/instances \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### 4. Start Instance

```bash
INSTANCE_ID="<instance-id-from-create>"

curl -X POST http://localhost:8443/api/v1/instances/$INSTANCE_ID/start \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### 5. Check Instance Status

```bash
curl -s http://localhost:8443/api/v1/instances/$INSTANCE_ID/status \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### 6. Get Health Information

```bash
# Get all instances health
curl -s http://localhost:8443/api/v1/health/instances \
  -H "Authorization: Bearer $TOKEN" | jq

# Get specific instance health with history
curl -s "http://localhost:8443/api/v1/health/instances/$INSTANCE_ID?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq

# Trigger immediate health check
curl -s "http://localhost:8443/api/v1/health/instances/$INSTANCE_ID?check=now" \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### 7. Stop Instance

```bash
curl -X POST http://localhost:8443/api/v1/instances/$INSTANCE_ID/stop \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### 8. Delete Instance

```bash
# Normal delete (stops first if running)
curl -X DELETE http://localhost:8443/api/v1/instances/$INSTANCE_ID \
  -H "Authorization: Bearer $TOKEN" | jq

# Force delete (even if running)
curl -X DELETE "http://localhost:8443/api/v1/instances/$INSTANCE_ID?force=true" \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Local Testing Setup

### Option 1: Mock Bash Scripts (No IBM Storage)

For development without access to IBM Storage Virtualize:

1. Create mock scripts directory:

```bash
mkdir -p /opt/ipquorum/scripts
```

2. Create mock `ipquorum-instance-manager.sh`:

```bash
cat > /opt/ipquorum/scripts/ipquorum-instance-manager.sh << 'EOF'
#!/bin/bash
# Mock script for testing

ACTION="$1"
INSTANCE="$2"

case "$ACTION" in
  create)
    echo "Mock: Creating instance $INSTANCE"
    exit 0
    ;;
  start)
    echo "Mock: Starting instance $INSTANCE"
    exit 0
    ;;
  stop)
    echo "Mock: Stopping instance $INSTANCE"
    exit 0
    ;;
  restart)
    echo "Mock: Restarting instance $INSTANCE"
    exit 0
    ;;
  status)
    echo "● ipquorum@$INSTANCE.service - IP Quorum Service"
    echo "   Active: active (running) since $(date)"
    exit 0
    ;;
  delete)
    echo "Mock: Deleting instance $INSTANCE"
    exit 0
    ;;
  check-network)
    echo "Network check: OK - API endpoint reachable"
    exit 0
    ;;
  validate)
    echo "Configuration validation: OK"
    exit 0
    ;;
  *)
    echo "Unknown action: $ACTION"
    exit 1
    ;;
esac
EOF

chmod +x /opt/ipquorum/scripts/ipquorum-instance-manager.sh
```

### Option 2: Real Bash Scripts (With IBM Storage)

1. Copy the actual bash scripts:

```bash
sudo mkdir -p /opt/ipquorum/scripts
sudo cp ../../ipquorum-systemd/multi-instance/ipquorum-instance-manager.sh \
  /opt/ipquorum/scripts/
sudo chmod +x /opt/ipquorum/scripts/ipquorum-instance-manager.sh
```

2. Ensure systemd template is installed:

```bash
sudo cp ../../ipquorum-systemd/multi-instance/ipquorum@.service \
  /etc/systemd/system/
sudo systemctl daemon-reload
```

## Development Tips

### Hot Reload

Use `air` for automatic recompilation:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
cd server
air
```

### Debugging

```bash
# Run with delve debugger
dlv debug ./cmd/server -- run --debug

# Or attach to running process
dlv attach $(pgrep ipquorum-server)
```

### Database Inspection

```bash
# SQLite
sqlite3 /var/lib/ipquorum-platform/ipquorum.db

# List tables
.tables

# Query instances
SELECT * FROM instances;

# Query health checks
SELECT * FROM health_checks ORDER BY checked_at DESC LIMIT 10;
```

### Logs

```bash
# Server logs (if using systemd)
journalctl -u ipquorum-server -f

# Or check application logs
tail -f /var/log/ipquorum-platform/server.log
```

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Run integration tests (requires running server)
go test -tags=integration ./...
```

### Load Testing

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Load test instance listing
hey -n 1000 -c 10 -H "Authorization: Bearer $TOKEN" \
  http://localhost:8443/api/v1/instances
```

## Troubleshooting

### Server Won't Start

1. Check port availability:
```bash
sudo lsof -i :8443
```

2. Check permissions:
```bash
ls -la /var/lib/ipquorum-platform/
ls -la /etc/ipquorum-platform/
```

3. Check logs:
```bash
./bin/ipquorum-server run --debug
```

### Database Issues

```bash
# Reset database
rm /var/lib/ipquorum-platform/ipquorum.db
./bin/ipquorum-server init
```

### Script Execution Fails

1. Check script exists and is executable:
```bash
ls -la /opt/ipquorum/scripts/ipquorum-instance-manager.sh
```

2. Test script manually:
```bash
/opt/ipquorum/scripts/ipquorum-instance-manager.sh status test-instance
```

3. Check script output in logs (debug mode)

## Contributing

1. Create a feature branch
2. Make changes
3. Add tests
4. Run `go fmt ./...`
5. Run `go vet ./...`
6. Run tests
7. Submit pull request

## Next Steps

- [ ] Implement Prometheus metrics
- [ ] Add unit tests
- [ ] Create web dashboard
- [ ] Implement local agent
- [ ] Add backup/restore functionality
- [ ] Implement RBAC policies
- [ ] Add audit log viewer
- [ ] Create deployment scripts

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [SQLite](https://www.sqlite.org/docs.html)
- [JWT](https://jwt.io/)
- [Prometheus](https://prometheus.io/docs/)