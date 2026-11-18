# Agent Setup Guide

## Overview
This guide helps you set up the IPQuorum agent on server 10.33.3.215 to enable container creation from the web UI.

## Current Status
✅ Manager (Web UI + API) is running and working
❌ Agent needs configuration and registration

## Prerequisites
- Root/sudo access to the server
- Agent binary already installed at `/usr/local/bin/ipquorum-agent`
- Agent systemd service exists but needs configuration

---

## Step 1: Create Agent Configuration

Run these commands with root access:

```bash
# Create config directory
sudo mkdir -p /etc/ipquorum-agent

# Create config file
sudo tee /etc/ipquorum-agent/config.yaml > /dev/null << 'EOF'
server:
  host: "0.0.0.0"
  port: 8081
  read_timeout: 30s
  write_timeout: 30s

auth:
  jwt_secret: "dev-secret-change-in-production"
  token_expiry: 24h

metrics:
  enabled: true
  port: 9090

logging:
  level: "info"
  format: "json"

container:
  runtime: "podman"
  socket: "/run/podman/podman.sock"
  
cache:
  enabled: true
  ttl: 5m
  max_size: 1000

rate_limit:
  enabled: true
  requests_per_second: 100
  burst: 200
EOF

# Set proper permissions
sudo chmod 644 /etc/ipquorum-agent/config.yaml
```

---

## Step 2: Restart Agent Service

```bash
# Restart the agent
sudo systemctl restart ipquorum-agent

# Check status
sudo systemctl status ipquorum-agent

# Verify it's listening
ss -tlnp | grep 8081
```

---

## Step 3: Test Agent Health

```bash
# Test health endpoint
curl http://localhost:8081/health

# Expected output:
# {"status":"healthy","timestamp":"..."}
```

---

## Step 4: Register Agent with Manager

### Option A: Via Web UI (Recommended)

1. Open browser: https://10.33.3.215:3443
2. Login: admin / admin123
3. Navigate to "Servers" page
4. Click "Add Server" or "Register Server"
5. Fill in the form:
   - **Hostname**: `local-server`
   - **IP Address**: `10.33.3.215`
   - **Port**: `8081`
   - **API Key**: `dev-secret-change-in-production`
   - **TLS Enabled**: No (unchecked)
6. Click "Register" or "Submit"

### Option B: Via API

```bash
# Get authentication token
TOKEN=$(curl -k -s -X POST https://localhost:3443/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')

# Register the server
curl -k -X POST https://localhost:3443/api/v1/servers \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "hostname": "local-server",
    "ip_address": "10.33.3.215",
    "agent_port": 8081,
    "api_key": "dev-secret-change-in-production",
    "tls_enabled": false,
    "tls_verify": false
  }'

# Verify registration
curl -k -s https://localhost:3443/api/v1/servers \
  -H "Authorization: Bearer $TOKEN" | jq .
```

---

## Step 5: Verify Registration

### Check in Web UI
1. Go to "Servers" page
2. You should see "local-server" listed
3. Status should show "online" or "healthy"

### Check via API
```bash
TOKEN=$(curl -k -s -X POST https://localhost:3443/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')

curl -k -s https://localhost:3443/api/v1/servers \
  -H "Authorization: Bearer $TOKEN" | jq .
```

---

## Step 6: Create Your First Container

Once the agent is registered:

1. Go to "Containers" page in web UI
2. Click "Create Container"
3. Select "local-server" from the Instance dropdown
4. Fill in container details:
   - Name: `test-container`
   - Storage System: Your IBM Storage Virtualize system
   - API Endpoint: Your storage system IP
   - Username/Password: Storage system credentials
5. Click "Create"

---

## Troubleshooting

### Agent Not Starting
```bash
# Check logs
sudo journalctl -u ipquorum-agent -n 50 --no-pager

# Check if config file exists
sudo cat /etc/ipquorum-agent/config.yaml

# Check if port is in use
ss -tlnp | grep 8081
```

### Agent Not Listening
```bash
# Verify agent process
ps aux | grep ipquorum-agent

# Check if podman socket exists
ls -la /run/podman/podman.sock

# Test podman access
podman ps
```

### Registration Fails
```bash
# Test agent health from manager
curl http://10.33.3.215:8081/health

# Check firewall
sudo firewall-cmd --list-ports
sudo firewall-cmd --add-port=8081/tcp --permanent
sudo firewall-cmd --reload
```

### No Instances in Dropdown
- Verify agent is registered (Step 5)
- Check agent status in "Servers" page
- Refresh the web page
- Check browser console for errors

---

## Architecture

```
┌─────────────────────────────────────────┐
│  Manager (10.33.3.215:3443)            │
│  - Web UI                               │
│  - API Server (port 8443)               │
└──────────────┬──────────────────────────┘
               │
               │ Manages
               │
┌──────────────▼──────────────────────────┐
│  Agent (10.33.3.215:8081)              │
│  - Receives commands from manager       │
│  - Manages podman containers            │
│  - Reports status back to manager       │
└─────────────────────────────────────────┘
```

---

## Security Notes

### Production Deployment
For production, change these values:

1. **JWT Secret**: Change `dev-secret-change-in-production` to a strong random string
2. **Enable TLS**: Set `tls_enabled: true` and provide certificates
3. **Firewall**: Restrict port 8081 to manager IP only
4. **API Key**: Use a strong, unique API key per agent

### Generate Secure Values
```bash
# Generate random JWT secret
openssl rand -base64 32

# Generate random API key
openssl rand -hex 32
```

---

## Next Steps

After successful agent registration:
1. Create test containers
2. Monitor container status
3. Test container lifecycle (start/stop/restart)
4. Review logs and metrics
5. Set up additional agents on other servers

---

## Support

For issues or questions:
- Check `BUGFIX-SUMMARY.md` for recent fixes
- Review `DEPLOYMENT-GUIDE.md` for deployment details
- Check `MANAGER-AGENT-CONNECTION-GUIDE.md` for connection troubleshooting