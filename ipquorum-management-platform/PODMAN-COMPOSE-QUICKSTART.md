# Podman Compose Quick Start Guide

## Overview
This guide shows how to quickly start and test the IPQuorum Container Management Platform using Podman Compose.

---

## Prerequisites

### 1. Install Podman
```bash
# macOS
brew install podman
podman machine init
podman machine start

# Linux (Fedora/RHEL)
sudo dnf install podman

# Linux (Ubuntu/Debian)
sudo apt-get install podman
```

### 2. Install podman-compose
```bash
pip3 install podman-compose
```

### 3. Verify Installation
```bash
podman --version
podman-compose --version
podman ps  # Should work without errors
```

---

## Quick Start (3 Commands)

```bash
cd ipquorum-management-platform

# 1. Start all services
./test-podman-compose.sh
# Choose option 1

# 2. Wait 30 seconds for services to start

# 3. Open browser
open http://localhost:5173
```

---

## What Gets Started

The Podman Compose setup starts 4 services:

1. **PostgreSQL** (Port 5432)
   - Database for storing configuration
   - Auto-applies migrations on startup

2. **Container Runtime Agent** (Ports 8081, 9090)
   - Manages Podman containers
   - Provides REST API for container operations
   - Exposes Prometheus metrics

3. **Manager Server** (Port 8080)
   - Central management API
   - Handles authentication
   - Coordinates with agents

4. **Frontend** (Port 5173)
   - React web interface
   - Container management UI
   - Real-time updates

---

## Service Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Podman Compose Stack                    │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  Browser ──▶ Frontend:5173 ──▶ Manager:8080 ──▶ Agent:8081 ──▶ Podman
│                                      │                              │
│                                      ▼                              │
│                                 Postgres:5432                       │
│                                                                     │
│  All services in ipquorum-network bridge                           │
│                                                                     │
└─────────────────────────────────────────────────────────┘
```

---

## Testing Workflow

### Step 1: Start Services

```bash
cd ipquorum-management-platform
./test-podman-compose.sh
# Choose option 1: Start services
```

**Expected Output:**
```
✓ Podman installed
✓ podman-compose installed
Building images...
Starting services...
✅ Services started!

Access points:
  Frontend:  http://localhost:5173
  Manager:   http://localhost:8080
  Agent:     http://localhost:8081
  Metrics:   http://localhost:9090/metrics
```

### Step 2: Verify Services

```bash
# Check service status
podman-compose -f podman-compose.yml ps

# Should show 4 services running:
# - ipquorum-postgres
# - ipquorum-agent
# - ipquorum-manager
# - ipquorum-frontend
```

### Step 3: Run Automated Tests

```bash
./test-podman-compose.sh
# Choose option 2: Run tests

# Tests:
# ✓ Manager health check
# ✓ Agent health check
# ✓ Database connection
# ✓ Authentication
# ✓ Container API
# ✓ Frontend accessibility
# ✓ Podman runtime
```

### Step 4: Access Web UI

1. Open browser: http://localhost:5173
2. Login: `admin` / `admin123`
3. Navigate to "Containers"
4. Click "Create Container"
5. Fill form and create a test container

### Step 5: Verify Container in Podman

```bash
# List all containers
podman ps -a

# Should see your created container
# Plus the 4 compose services
```

---

## Manual Testing

### Test 1: Create Container via UI

1. **Login**: http://localhost:5173
2. **Navigate**: Containers page
3. **Create**: Click "Create Container"
4. **Fill Form**:
   - Instance: (select available)
   - Image: nginx
   - Tag: alpine
   - Add port: 8090 → 80
5. **Submit**: Create container
6. **Verify**: `podman ps -a | grep nginx`

### Test 2: Container Lifecycle

1. **Start**: Click "Start" button
2. **Verify**: `podman ps | grep nginx`
3. **Test**: Open http://localhost:8090
4. **Stop**: Click "Stop" button
5. **Verify**: Container stopped in Podman

### Test 3: API Testing

```bash
# Get auth token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')

# List containers
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/containers | jq '.'

# Create container
curl -X POST http://localhost:8080/api/v1/containers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "instance_id": "your-instance-id",
    "agent_id": "your-agent-id",
    "image_name": "redis",
    "image_tag": "alpine"
  }' | jq '.'
```

---

## Viewing Logs

### All Services
```bash
podman-compose -f podman-compose.yml logs -f
```

### Specific Service
```bash
# Manager logs
podman-compose -f podman-compose.yml logs -f manager

# Agent logs
podman-compose -f podman-compose.yml logs -f agent

# Frontend logs
podman-compose -f podman-compose.yml logs -f frontend

# Database logs
podman-compose -f podman-compose.yml logs -f postgres
```

### Using Test Script
```bash
./test-podman-compose.sh
# Choose option 5: View logs
```

---

## Troubleshooting

### Issue: Services won't start

```bash
# Check Podman is running
podman ps

# Check for port conflicts
lsof -i :5173  # Frontend
lsof -i :8080  # Manager
lsof -i :8081  # Agent
lsof -i :5432  # Postgres

# Clean up and restart
podman-compose -f podman-compose.yml down -v
podman-compose -f podman-compose.yml up -d --build
```

### Issue: Can't connect to Podman socket

```bash
# macOS: Check Podman machine
podman machine list
podman machine start

# Linux: Check socket
ls -la /run/user/$(id -u)/podman/podman.sock

# Enable socket
systemctl --user enable --now podman.socket
```

### Issue: Database connection failed

```bash
# Check database is running
podman exec ipquorum-postgres pg_isready -U postgres

# Check migrations
podman exec ipquorum-postgres psql -U postgres -d ipquorum_platform -c "\dt"

# Restart database
podman-compose -f podman-compose.yml restart postgres
```

### Issue: Frontend can't reach backend

```bash
# Check network
podman network inspect ipquorum-network

# Check manager is accessible
curl http://localhost:8080/api/v1/health

# Check frontend environment
podman exec ipquorum-frontend env | grep VITE_API_URL
```

---

## Cleanup

### Stop Services (Keep Data)
```bash
podman-compose -f podman-compose.yml down
```

### Stop Services (Remove Data)
```bash
podman-compose -f podman-compose.yml down -v
```

### Remove All Images
```bash
podman rmi ipquorum-agent:latest
podman rmi ipquorum-manager:latest
podman rmi ipquorum-frontend:latest
```

### Complete Cleanup
```bash
./test-podman-compose.sh
# Choose option 4: Stop and cleanup
```

---

## Production Deployment

For production use, modify `podman-compose.yml`:

1. **Change Secrets**:
   ```yaml
   environment:
     JWT_SECRET: "use-strong-random-secret-here"
     POSTGRES_PASSWORD: "use-strong-password-here"
   ```

2. **Use Production Frontend**:
   - Build production bundle
   - Use nginx to serve static files
   - Remove dev server

3. **Add TLS**:
   - Use reverse proxy (nginx/traefik)
   - Configure SSL certificates

4. **Persistent Volumes**:
   - Use named volumes for data
   - Configure backup strategy

5. **Resource Limits**:
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '2'
         memory: 2G
   ```

---

## Performance Tips

1. **Use Podman Machine with enough resources** (macOS):
   ```bash
   podman machine stop
   podman machine rm
   podman machine init --cpus 4 --memory 8192 --disk-size 50
   podman machine start
   ```

2. **Enable caching for builds**:
   ```bash
   podman-compose build --no-cache  # First time
   podman-compose build             # Use cache
   ```

3. **Monitor resource usage**:
   ```bash
   podman stats
   ```

---

## Next Steps

After successful testing:

1. **Explore Features**:
   - Create multiple containers
   - Test filters and search
   - Try different container images

2. **Integration Testing**:
   - Test with real workloads
   - Monitor performance
   - Check logs for errors

3. **Customize**:
   - Modify compose file for your needs
   - Add more services
   - Configure networking

4. **Production Deployment**:
   - Follow production checklist
   - Set up monitoring
   - Configure backups

---

## Quick Reference

### Common Commands

```bash
# Start
podman-compose up -d

# Stop
podman-compose down

# Rebuild
podman-compose up -d --build

# Logs
podman-compose logs -f

# Status
podman-compose ps

# Restart service
podman-compose restart manager

# Execute command in container
podman exec -it ipquorum-manager sh
```

### Service URLs

- Frontend: http://localhost:5173
- Manager API: http://localhost:8080
- Agent API: http://localhost:8081
- Metrics: http://localhost:9090/metrics
- Database: localhost:5432

### Default Credentials

- Username: `admin`
- Password: `admin123`
- Database: `postgres` / `postgres123`

---

## Support

If you encounter issues:

1. Check logs: `podman-compose logs -f`
2. Verify services: `podman-compose ps`
3. Check Podman: `podman ps -a`
4. Review documentation: `PODMAN-TESTING-GUIDE.md`
5. Run tests: `./test-podman-compose.sh` (option 2)

---

**Quick Start Version**: 1.0.0
**Last Updated**: 2026-05-11
**Tested With**: Podman 4.0+, podman-compose 1.0+