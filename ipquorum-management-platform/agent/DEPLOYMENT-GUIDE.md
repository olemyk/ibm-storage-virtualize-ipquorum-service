# Agent Deployment Guide - Update to v1.0.5

## Overview
This guide explains how to update an existing IPQuorum Agent installation to version 1.0.5 with enhanced API support.

---

## Prerequisites

- Existing agent installation (v1.0.0 - v1.0.4)
- SSH access to the server
- Sudo privileges
- Built binary: `build/ipquorum-agent`

---

## Quick Update (Recommended)

### Step 1: Build the New Binary (On Your Local Machine)

```bash
cd ipquorum-management-platform/agent
make build
```

**Output**: Binary created at `build/ipquorum-agent`

### Step 2: Copy Binary to Server

```bash
# Replace with your server details
SERVER_USER="your-username"
SERVER_HOST="your-server-ip"

scp build/ipquorum-agent ${SERVER_USER}@${SERVER_HOST}:/tmp/
```

### Step 3: Update Agent on Server

```bash
# SSH to server
ssh ${SERVER_USER}@${SERVER_HOST}

# Switch to root
sudo su -

# Stop the agent
systemctl stop ipquorum-agent

# Backup current binary (optional but recommended)
cp /usr/local/bin/ipquorum-agent /usr/local/bin/ipquorum-agent.backup

# Install new binary
cp /tmp/ipquorum-agent /usr/local/bin/
chmod +x /usr/local/bin/ipquorum-agent
chown root:root /usr/local/bin/ipquorum-agent

# Verify version (optional)
/usr/local/bin/ipquorum-agent --version

# Start the agent
systemctl start ipquorum-agent

# Check status
systemctl status ipquorum-agent

# View logs
journalctl -u ipquorum-agent -f
```

---

## Automated Deployment Script

Create this script on your local machine:

```bash
#!/bin/bash
# deploy-agent-update.sh

set -e

SERVER_USER="${1:-root}"
SERVER_HOST="${2}"

if [[ -z "$SERVER_HOST" ]]; then
    echo "Usage: $0 [user] <server-host>"
    echo "Example: $0 root 10.33.7.80"
    exit 1
fi

echo "=== Deploying Agent Update to $SERVER_HOST ==="

# Build binary
echo "Building agent..."
cd ipquorum-management-platform/agent
make build

# Copy to server
echo "Copying binary to server..."
scp build/ipquorum-agent ${SERVER_USER}@${SERVER_HOST}:/tmp/

# Deploy on server
echo "Deploying on server..."
ssh ${SERVER_USER}@${SERVER_HOST} << 'EOF'
set -e

echo "Stopping agent..."
systemctl stop ipquorum-agent

echo "Backing up current binary..."
cp /usr/local/bin/ipquorum-agent /usr/local/bin/ipquorum-agent.backup

echo "Installing new binary..."
cp /tmp/ipquorum-agent /usr/local/bin/
chmod +x /usr/local/bin/ipquorum-agent
chown root:root /usr/local/bin/ipquorum-agent

echo "Starting agent..."
systemctl start ipquorum-agent

echo "Checking status..."
systemctl status ipquorum-agent --no-pager

echo "Deployment complete!"
EOF

echo "=== Deployment Complete ==="
echo "Check logs with: ssh ${SERVER_USER}@${SERVER_HOST} 'journalctl -u ipquorum-agent -f'"
```

**Usage**:
```bash
chmod +x deploy-agent-update.sh
./deploy-agent-update.sh root 10.33.7.80
```

---

## Verification Steps

### 1. Check Agent Status
```bash
systemctl status ipquorum-agent
```

**Expected**: `active (running)`

### 2. Check Agent Logs
```bash
journalctl -u ipquorum-agent -n 50
```

**Look for**: No errors, successful startup messages

### 3. Test Health Endpoint
```bash
curl http://localhost:8081/health
```

**Expected**:
```json
{
  "status": "healthy",
  "timestamp": "2026-04-30T12:00:00Z",
  "agent": "ipquorum-agent-01"
}
```

### 4. Test Enhanced API (Minimal Request)
```bash
API_KEY="your-api-key"

curl -X POST http://localhost:8081/instances \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-update",
    "api_endpoint": "10.33.7.80",
    "username": "monitor",
    "password": "test123"
  }'
```

**Expected**: `{"message":"Instance created successfully"}`

### 5. Verify Config File
```bash
sudo cat /etc/ipquorum/instances/test-update.conf | grep -E "IBM_STORAGE_SYSTEM|IPQUORUM_MKQUORUMAPP_ENABLED"
```

**Expected**: New fields present with default values

### 6. Clean Up Test Instance
```bash
curl -X DELETE http://localhost:8081/instances/test-update \
  -H "X-API-Key: $API_KEY"
```

---

## Apply Sudo Fix (If Not Already Applied)

If you're updating from v1.0.4 or earlier and haven't applied the sudo fix:

```bash
# On server, as root
cd /tmp

# Copy setup-sudo.sh from your local machine
# (from ipquorum-management-platform/agent/setup-sudo.sh)

# Run setup
chmod +x setup-sudo.sh
./setup-sudo.sh

# Verify
sudo -l -U ipquorum
```

**Expected**: List of allowed sudo commands

---

## Rollback Procedure (If Needed)

If something goes wrong, rollback to the previous version:

```bash
# On server, as root
systemctl stop ipquorum-agent

# Restore backup
cp /usr/local/bin/ipquorum-agent.backup /usr/local/bin/ipquorum-agent

# Start agent
systemctl start ipquorum-agent

# Verify
systemctl status ipquorum-agent
```

---

## Configuration Changes

### No Configuration Changes Required

The enhanced agent is **100% backward compatible**. Your existing configuration files remain unchanged:

- `/etc/ipquorum/agent.yaml` - No changes needed
- `/etc/systemd/system/ipquorum-agent.service` - No changes needed
- API key - Remains the same

### New Features Available Immediately

After update, you can immediately use:
- All new API fields (optional)
- Enhanced validation
- Smart defaults

---

## Testing the Enhanced API

Use the provided test script:

```bash
# Copy test script to server
scp ipquorum-management-platform/agent/test-enhanced-api.sh root@server:/tmp/

# On server
chmod +x /tmp/test-enhanced-api.sh
API_KEY="your-api-key" /tmp/test-enhanced-api.sh
```

---

## Troubleshooting

### Agent Won't Start

**Check logs**:
```bash
journalctl -u ipquorum-agent -n 100 --no-pager
```

**Common issues**:
- Port 8081 already in use: `netstat -tlnp | grep 8081`
- Config file errors: `cat /etc/ipquorum/agent.yaml`
- Binary permissions: `ls -l /usr/local/bin/ipquorum-agent`

### API Returns Errors

**Check agent logs**:
```bash
journalctl -u ipquorum-agent -f
```

**Test with curl verbose**:
```bash
curl -v -X POST http://localhost:8081/instances \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name":"test","api_endpoint":"10.33.7.80","username":"user","password":"pass"}'
```

### Sudo Permissions Issues

**Verify sudo config**:
```bash
sudo -l -U ipquorum
```

**If missing, apply fix**:
```bash
cd /tmp
# Copy setup-sudo.sh from local machine
chmod +x setup-sudo.sh
./setup-sudo.sh
```

---

## Post-Deployment Checklist

- [ ] Agent service is running
- [ ] Health endpoint responds
- [ ] Can create test instance with minimal request
- [ ] Can create test instance with full request (mkquorumapp)
- [ ] Config files are updated correctly
- [ ] Sudo permissions are working
- [ ] Old test instances cleaned up
- [ ] Logs show no errors

---

## Support

### View Agent Logs
```bash
# Real-time logs
journalctl -u ipquorum-agent -f

# Last 100 lines
journalctl -u ipquorum-agent -n 100

# Since specific time
journalctl -u ipquorum-agent --since "10 minutes ago"
```

### Check Agent Configuration
```bash
cat /etc/ipquorum/agent.yaml
```

### List Instances
```bash
curl http://localhost:8081/instances -H "X-API-Key: $API_KEY" | jq .
```

### Get Instance Status
```bash
curl http://localhost:8081/instances/instance-name/status -H "X-API-Key: $API_KEY"
```

---

## Version History

- **v1.0.5** (2026-04-30): Enhanced API with full field support
- **v1.0.4** (2026-04-29): Config pattern fixes
- **v1.0.3** (2026-04-28): Sudo permission fixes
- **v1.0.0** (2026-04-27): Initial release

---

## Next Steps After Deployment

1. **Test Enhanced Features**: Try creating instances with new fields
2. **Update Documentation**: Update your internal docs with new API capabilities
3. **Update Automation**: Enhance your automation scripts to use new fields
4. **Monitor**: Watch logs for any issues
5. **Feedback**: Report any issues or suggestions

---

## Quick Reference Commands

```bash
# Build
make build

# Deploy
scp build/ipquorum-agent root@server:/tmp/
ssh root@server 'systemctl stop ipquorum-agent && cp /tmp/ipquorum-agent /usr/local/bin/ && chmod +x /usr/local/bin/ipquorum-agent && systemctl start ipquorum-agent'

# Verify
ssh root@server 'systemctl status ipquorum-agent && journalctl -u ipquorum-agent -n 20'

# Test
curl http://server:8081/health
```

---

**Deployment Status**: Ready for Production ✅