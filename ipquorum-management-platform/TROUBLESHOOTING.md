# Troubleshooting Guide

## Database Permission Issues

### Symptom
```
Error: failed to initialize database: failed to enable foreign keys: unable to open database file: no such file or directory
```

### Root Cause
The container runs as UID 1000 (appuser), but the host `./data` directory is owned by a different user, preventing the container from writing to the database file.

### Solution

**Option 1: Use Podman `:U` flag (RECOMMENDED - already configured)**

The compose file uses Podman's `:U` volume flag which automatically handles permissions:
```yaml
volumes:
  - ./data:/data:U
```

This is already configured in [`podman-compose.dev.yml`](podman-compose.dev.yml). Just redeploy:
```bash
./start-web-dev.sh --clean
```

The `:U` flag tells Podman to automatically chown the volume to match the container user (UID 1000). No manual intervention needed!

**Option 2: Manual fix (if `:U` flag not working)**
```bash
# Set ownership to UID 1000
sudo chown -R 1000:1000 ./data

# Or for rootless Podman, use your user
chown -R $(id -u):$(id -g) ./data
```

**Option 3: Use podman unshare (for rootless Podman)**
```bash
# Enter user namespace and fix ownership
podman unshare chown -R 1000:1000 ./data
```

### Verification
```bash
# Check directory ownership
ls -ld ./data
# Should show: drwxr-xr-x ... 1000 1000 ... ./data

# Check if database file exists and is writable
ls -l ./data/ipquorum.db
# Should show: -rw-r--r-- ... 1000 1000 ... ./data/ipquorum.db
```

## Port Already in Use

### Symptom
```
rootlessport listen tcp 0.0.0.0:8443: bind: address already in use
```

### Root Cause
Stale `conmon` or `rootlessport` processes from previous failed container starts.

### Solution
```bash
# Kill stale conmon processes
sudo pkill -9 conmon

# Or find and kill specific process
sudo lsof -i :8443
sudo kill -9 <PID>

# Then restart containers
./start-web-dev.sh --clean
```

## Network Subnet Conflicts

### Symptom
```
Command 'podman network create ... --subnet 172.28.0.0/16' returned non-zero exit status 125
```

### Root Cause
The subnet conflicts with existing network routes on the host system.

### Solution
The compose file uses `10.89.0.0/16` which is less commonly used. If this still conflicts:

1. Check existing networks:
```bash
podman network ls
ip route show
```

2. Edit `podman-compose.dev.yml` and change the subnet:
```yaml
networks:
  ipquorum-net:
    driver: bridge
    ipam:
      config:
        - subnet: 10.99.0.0/16  # Change to unused subnet
```

## HEALTHCHECK Warnings

### Symptom
```
WARN[0254] HEALTHCHECK is not supported for OCI image format
```

### Root Cause
Podman defaults to OCI image format which doesn't support HEALTHCHECK directive.

### Solution
Set `BUILDAH_FORMAT=docker` in `.env` file:
```bash
echo "BUILDAH_FORMAT=docker" >> .env
```

Or export before building:
```bash
export BUILDAH_FORMAT=docker
podman-compose build
```

See [`HEALTHCHECK-FIX.md`](HEALTHCHECK-FIX.md) for detailed explanation.

## Container Health Check Failures

### Symptom
Containers show as "unhealthy" in `podman ps` output.

### Root Cause
Health check endpoint returns non-200 status or connection fails.

### Diagnosis
```bash
# Check container logs
podman logs ipquorum-server

# Test health endpoint manually
podman exec ipquorum-server wget --spider --no-check-certificate https://localhost:8080/api/v1/health

# Check if service is listening
podman exec ipquorum-server netstat -tlnp | grep 8080
```

### Solution
1. Verify the service started correctly (check logs)
2. Ensure health check uses correct protocol (HTTPS vs HTTP)
3. Verify port mapping in compose file
4. Check if TLS certificates are valid

## Go Version Mismatch

### Symptom
```
go: go.mod requires go >= 1.25.0 (running go 1.23.12)
```

### Root Cause
Dockerfile uses older Go version than required by `go.mod`.

### Solution
Update Dockerfile to use correct Go version:
```dockerfile
FROM docker.io/library/golang:1.25-alpine AS builder
```

## Permission Denied on Scripts

### Symptom
```
bash: ./start-web-dev.sh: Permission denied
```

### Solution
```bash
chmod +x start-web-dev.sh
chmod +x deploy-to-linux.sh
```

## Container Build Failures

### Symptom
Build fails with various errors during `podman-compose build`.

### Diagnosis
```bash
# Build with verbose output
podman-compose build --no-cache --progress=plain

# Check available disk space
df -h

# Check Podman storage
podman system df
```

### Solution
```bash
# Clean up old images and containers
podman system prune -a -f

# Reset storage if needed
podman system reset
```

## Web UI Not Accessible

### Symptom
Cannot access web UI at https://localhost:3443

### Diagnosis
```bash
# Check if container is running
podman ps | grep ipquorum-web

# Check port mapping
podman port ipquorum-web

# Test from inside container
podman exec ipquorum-web curl -k https://localhost:3000
```

### Solution
1. Verify container is running and healthy
2. Check firewall rules allow port 3443
3. Try accessing via server IP instead of localhost
4. Check nginx logs: `podman logs ipquorum-web`

## Agent Connection Issues

### Symptom
Agent shows as disconnected in web UI or server logs show connection errors.

### Diagnosis
```bash
# Check agent is running
podman ps | grep ipquorum-agent

# Check agent logs
podman logs ipquorum-agent

# Test agent API
curl http://localhost:8081/health
```

### Solution
1. Verify agent is running: `podman ps`
2. Check agent configuration in `container-runtime-agent/config.yaml`
3. Verify network connectivity between containers
4. Check JWT token configuration matches between server and agent

## Getting Help

If issues persist:

1. **Collect logs:**
```bash
podman logs ipquorum-server > server.log
podman logs ipquorum-web > web.log
podman logs ipquorum-agent > agent.log  # if using agent
```

2. **Check system info:**
```bash
podman version
podman info
uname -a
```

3. **Review documentation:**
- [`README.md`](README.md) - Main documentation
- [`DEPLOYMENT-GUIDE.md`](DEPLOYMENT-GUIDE.md) - Deployment instructions
- [`HEALTHCHECK-FIX.md`](HEALTHCHECK-FIX.md) - HEALTHCHECK issues
- [`CONTAINER-DEPLOYMENT.md`](CONTAINER-DEPLOYMENT.md) - Container details

4. **Enable debug mode:**
```bash
# In .env file
LOG_LEVEL=debug

# Restart containers
./start-web-dev.sh --clean