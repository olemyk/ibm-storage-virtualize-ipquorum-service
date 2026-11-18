# API Key Setup Guide

## Overview

This guide explains how to set up and use API keys for agent authentication in the IP Quorum Management Platform.

## Generated API Key

**Your Development API Key:**
```
<your-agent-api-key>
```

⚠️ **Important**: Save this key securely. It's required for registering agents in the web UI.

## What is the API Key Used For?

The API key provides authentication between:
- **Management Server** ← API Key → **Agent**

It ensures that only authorized agents can be registered and managed by the platform.

## Setup Steps

### 1. API Key Already Configured

The setup script has already:
- ✅ Generated a secure API key
- ✅ Updated `.env` file with `AGENT_API_KEY`
- ✅ Created backup of previous `.env`
- ✅ Saved key to `.api-key-20260512.txt`

### 2. Configure the Agent

The agent needs to be configured with the same API key. Choose one method:

#### Option A: Environment Variable (Recommended for Containers)

When starting the agent container:
```bash
podman run -d \
  --name ipquorum-agent \
  -e IPQUORUM_AGENT_API_KEY="<your-agent-api-key>" \
  # ... other options
  localhost/ipquorum_ipquorum-agent:latest
```

#### Option B: Configuration File

Update `container-runtime-agent/config.yaml`:
```yaml
auth:
  api_key: "<your-agent-api-key>"
  jwt_secret: "dev-secret-change-in-production"
  token_expiry: 24h
```

### 3. Restart Services

After configuring the agent, restart all services:

```bash
cd ipquorum-test

# Stop all services
podman stop ipquorum-server ipquorum-web ipquorum-agent
podman rm ipquorum-server ipquorum-web ipquorum-agent

# Restart with compose
podman-compose -f podman-compose.dev.yml up -d

# Or restart agent manually with API key
podman run -d \
  --name ipquorum-agent \
  --network ipquorum_ipquorum-network \
  -p 8081:8080 \
  -p 9090:9090 \
  -v /run/user/1000/podman/podman.sock:/run/podman/podman.sock:rw \
  -v ~/ipquorum-test/container-runtime-agent/config.yaml:/app/config.yaml:ro \
  -e IPQUORUM_AGENT_API_KEY="<your-agent-api-key>" \
  -e IPQUORUM_AGENT_SERVER_PORT=8080 \
  -e IPQUORUM_AGENT_RUNTIME_PREFERRED_RUNTIME=podman \
  -e IPQUORUM_AGENT_RUNTIME_PODMAN_SOCKET_PATH=/run/podman/podman.sock \
  -e IPQUORUM_AGENT_AUTH_ENABLED=true \
  -e IPQUORUM_AGENT_AUTH_JWT_SECRET=test-secret-key-change-in-production \
  -e IPQUORUM_AGENT_LOGGING_LEVEL=debug \
  -e IPQUORUM_AGENT_METRICS_ENABLED=true \
  --user 0:0 \
  localhost/ipquorum_ipquorum-agent:latest
```

## Registering Servers in Web UI

### Step 1: Access Web Dashboard

Navigate to: **https://10.33.3.215:3443**

Login credentials:
- Username: `admin`
- Password: `admin123`

### Step 2: Register Server

1. Click **"Servers"** in navigation
2. Click **"Register Server"** button
3. Fill in the form:

| Field | Value | Required |
|-------|-------|----------|
| **Hostname** | `ipquorum-agent` | ✅ Yes |
| **IP Address** | `ipquorum-agent` | ✅ Yes |
| **Agent Port** | `8080` | ✅ Yes |
| **API Key** | `<your-agent-api-key>` | ✅ **Yes** |
| **TLS Enabled** | `false` | No |
| **TLS Verify** | `false` | No |

4. Click **"Register"**

### Step 3: Verify Registration

After successful registration:
- ✅ Server appears in servers list
- ✅ Status shows "online" (green)
- ✅ You can view server details

## API Registration (Alternative)

You can also register via REST API:

```bash
# Get authentication token
TOKEN=$(curl -k -X POST https://10.33.3.215:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r .token)

# Register server with API key
curl -k -X POST https://10.33.3.215:8443/api/v1/servers/register \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "ipquorum-agent",
    "ip_address": "ipquorum-agent",
    "agent_port": 8080,
    "api_key": "<your-agent-api-key>",
    "tls_enabled": false,
    "tls_verify": false
  }'
```

## Troubleshooting

### Error: "Invalid API Key"

**Cause**: API key mismatch between server and agent

**Solution**:
1. Verify API key in `.env` matches agent configuration
2. Check environment variables are loaded: `podman inspect ipquorum-agent | grep API_KEY`
3. Restart agent after config changes

### Error: "Key: 'APIKey' Error:Field validation for 'APIKey' failed on the 'required' tag"

**Cause**: API key field is empty in registration request

**Solution**: Enter the API key in the web UI form or API request

### Agent Shows as "Offline"

**Possible causes**:
1. Agent not running
2. API key not configured on agent
3. Network connectivity issues

**Solutions**:
```bash
# Check agent is running
podman ps | grep ipquorum-agent

# Check agent logs for API key errors
podman logs ipquorum-agent | grep -i "api"

# Verify agent has API key configured
podman inspect ipquorum-agent | grep IPQUORUM_AGENT_API_KEY

# Test agent connectivity
curl http://ipquorum-agent:8080/health
```

## Security Best Practices

### Development Environment
- ✅ Use generated API key
- ✅ Store key in `.env` file
- ✅ Add `.env` to `.gitignore`
- ✅ Keep `.api-key-*.txt` files secure

### Production Environment
- ⚠️ **Generate new API keys for production**
- ⚠️ **Use strong, unique keys per environment**
- ⚠️ **Enable TLS on agents**
- ⚠️ **Use valid SSL certificates**
- ⚠️ **Rotate API keys regularly (every 90 days)**
- ⚠️ **Store keys in secrets management system**
- ⚠️ **Never commit API keys to version control**
- ⚠️ **Use different keys for different agents**

## Generating New API Keys

To generate a new API key:

```bash
# Run the setup script
cd ipquorum-management-platform
./setup-api-keys.sh

# Or generate manually
openssl rand -base64 32
```

## Key Rotation

When rotating API keys:

1. Generate new key
2. Update `.env` file
3. Update agent configuration
4. Restart all services
5. Re-register agents with new key
6. Securely delete old key

## Files Modified

- `.env` - Contains `AGENT_API_KEY`
- `.api-key-20260512.txt` - Secure note with key (chmod 600)
- `.env.backup.*` - Backup of previous configuration

## Related Documentation

- [Web UI Server Registration Guide](WEB-UI-SERVER-REGISTRATION-GUIDE.md)
- [Agent Connection Guide](AGENT-CONNECTION-COMPLETE.md)
- [API Specification](API-SPECIFICATION.md)

## Quick Reference

**API Key**: `<your-agent-api-key>`

**Use this key when**:
- Registering servers in web UI
- Configuring agent containers
- Making API calls to register servers
- Setting up new agent instances