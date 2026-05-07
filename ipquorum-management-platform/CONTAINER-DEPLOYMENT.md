# Container Deployment Guide

This guide explains how to deploy the IP Quorum Management Platform using containers (Podman or Docker).

## Prerequisites

### Podman (Recommended)
```bash
# macOS
brew install podman

# Linux (Debian/Ubuntu)
sudo apt-get install podman

# Linux (Fedora/RHEL)
sudo dnf install podman
```

### Docker (Alternative)
```bash
# macOS
brew install docker

# Linux
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
```

### Additional Tools
```bash
# Install jq for JSON parsing (required for tests)
brew install jq  # macOS
sudo apt-get install jq  # Linux
```

## Quick Start

### Using the Helper Script (Easiest)

```bash
cd ipquorum-management-platform

# Build and run the container
./run-container.sh run

# View logs
./run-container.sh logs

# Run API tests
./run-container.sh test

# Stop the container
./run-container.sh stop

# Clean up everything
./run-container.sh clean
```

### Manual Podman Commands

```bash
cd ipquorum-management-platform

# Build the image
podman build -t ipquorum-management-server:latest -f Containerfile .

# Create a volume for persistent data
podman volume create ipquorum-data

# Run the container
podman run -d \
  --name ipquorum-server \
  -p 8080:8080 \
  -v ipquorum-data:/data \
  -v ./test-scripts:/scripts:ro \
  -v ./test-config/config.yaml:/etc/ipquorum/config.yaml:ro \
  -e CONFIG_PATH=/etc/ipquorum/config.yaml \
  -e LOG_LEVEL=info \
  ipquorum-management-server:latest

# View logs
podman logs -f ipquorum-server

# Stop the container
podman stop ipquorum-server

# Remove the container
podman rm ipquorum-server
```

### Using Docker Compose / Podman Compose

```bash
cd ipquorum-management-platform

# Start services
podman-compose up -d
# or
docker-compose up -d

# View logs
podman-compose logs -f
# or
docker-compose logs -f

# Stop services
podman-compose down
# or
docker-compose down
```

## Container Structure

### Multi-Stage Build

The Containerfile uses a multi-stage build for optimal image size:

1. **Builder Stage**: Compiles the Go application
   - Base: `golang:1.21-alpine`
   - Installs build dependencies
   - Compiles static binary with CGO enabled (for SQLite)

2. **Runtime Stage**: Minimal runtime environment
   - Base: `alpine:latest`
   - Installs runtime dependencies (bash, curl, ca-certificates)
   - Copies compiled binary and scripts
   - Non-root user execution

### Directory Structure in Container

```
/app/
├── server              # Compiled Go binary
└── scripts/           # Bash scripts directory

/data/                 # Persistent data (volume mount)
└── ipquorum.db       # SQLite database

/etc/ipquorum/        # Configuration
└── config.yaml       # Server configuration
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIG_PATH` | Path to config file | `/etc/ipquorum/config.yaml` |
| `LOG_LEVEL` | Logging level | `info` |
| `SERVER_PORT` | API server port | `8080` |
| `JWT_SECRET` | JWT signing secret | (from config) |

### Volume Mounts

| Host Path | Container Path | Purpose |
|-----------|----------------|---------|
| `ipquorum-data` | `/data` | Persistent database storage |
| `./test-scripts` | `/scripts` | Mock bash scripts for testing |
| `./test-config/config.yaml` | `/etc/ipquorum/config.yaml` | Server configuration |

### Port Mappings

| Host Port | Container Port | Service |
|-----------|----------------|---------|
| `8080` | `8080` | REST API |

## Testing

### Run API Tests

```bash
# Using the helper script
./run-container.sh test

# Or manually
./test-api.sh
```

The test script will:
1. Check API health
2. Login and obtain JWT token
3. Create a test instance
4. Start/stop/restart the instance
5. Check instance status and health
6. Delete the instance
7. Verify cleanup

### Expected Test Output

```
==========================================
IP Quorum Management Platform API Tests
==========================================

[INFO] Test 1: Health Check
[PASS] Health check passed

[INFO] Test 2: User Login
[PASS] Login successful, token obtained

[INFO] Test 3: Get Current User
[PASS] User info retrieved successfully

...

==========================================
Test Summary
==========================================
Total Tests: 15
Passed: 15
Failed: 0

✓ All tests passed!
```

## Troubleshooting

### Container Won't Start

```bash
# Check container logs
podman logs ipquorum-server

# Check if port is already in use
lsof -i :8080

# Verify image was built successfully
podman images | grep ipquorum
```

### Database Issues

```bash
# Check database file permissions
podman exec ipquorum-server ls -la /data/

# Reset database (WARNING: deletes all data)
podman volume rm ipquorum-data
podman volume create ipquorum-data
```

### Script Execution Errors

```bash
# Verify scripts are mounted correctly
podman exec ipquorum-server ls -la /scripts/

# Check script permissions
podman exec ipquorum-server ls -la /scripts/ipquorum-instance-manager.sh

# Test script manually
podman exec ipquorum-server /scripts/ipquorum-instance-manager.sh list
```

### Network Issues

```bash
# Check if container is running
podman ps

# Test API from inside container
podman exec ipquorum-server wget -O- http://localhost:8080/health

# Test API from host
curl http://localhost:8080/health
```

### Build Failures

```bash
# Clean build cache
podman system prune -a

# Rebuild with no cache
podman build --no-cache -t ipquorum-management-server:latest -f Containerfile .

# Check Go module issues
cd server && go mod tidy && go mod verify
```

## Production Deployment

### Security Considerations

1. **Change Default Credentials**
   ```yaml
   # In config.yaml, update JWT secret
   auth:
     jwt_secret: "your-secure-random-secret-here"
   ```

2. **Use Secrets Management**
   ```bash
   # Use Podman secrets
   echo "your-jwt-secret" | podman secret create jwt_secret -
   
   # Mount secret in container
   podman run -d \
     --secret jwt_secret \
     -e JWT_SECRET_FILE=/run/secrets/jwt_secret \
     ...
   ```

3. **Enable TLS**
   - Use a reverse proxy (nginx, traefik) for TLS termination
   - Or configure TLS directly in the application

4. **Restrict Network Access**
   ```bash
   # Only expose to localhost
   podman run -d -p 127.0.0.1:8080:8080 ...
   ```

### Resource Limits

```bash
# Set memory and CPU limits
podman run -d \
  --memory=512m \
  --cpus=1.0 \
  --name ipquorum-server \
  ...
```

### Health Checks

```bash
# Container with health check
podman run -d \
  --health-cmd="wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  --health-start-period=10s \
  --name ipquorum-server \
  ...
```

### Backup and Restore

```bash
# Backup database
podman exec ipquorum-server cat /data/ipquorum.db > backup-$(date +%Y%m%d).db

# Restore database
podman cp backup-20260426.db ipquorum-server:/data/ipquorum.db
podman restart ipquorum-server
```

## Integration with Real IBM Storage

To use with real IBM Storage Virtualize systems:

1. **Replace Mock Scripts**
   ```bash
   # Mount real scripts instead of test-scripts
   -v /path/to/real/scripts:/scripts:ro
   ```

2. **Update Configuration**
   ```yaml
   # In config.yaml
   agent:
     scripts_dir: "/scripts"
     timeout: 300s
   ```

3. **Network Access**
   - Ensure container can reach IBM Storage management IPs
   - Configure firewall rules if needed
   - Use host network mode if required: `--network=host`

## Next Steps

- Review [ARCHITECTURE.md](./ARCHITECTURE.md) for system design
- Check [DEVELOPMENT.md](./DEVELOPMENT.md) for development setup
- See [README.md](./README.md) for API documentation
- Explore Phase 2 features (web dashboard, advanced monitoring)