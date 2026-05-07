# IPQuorum Agent Deployment Guide

This guide explains how to deploy the IPQuorum Agent on existing IPQuorum installations to enable remote management via the IPQuorum Management Platform.

## Overview

The IPQuorum Agent is a lightweight service that runs on hosts with existing IPQuorum installations. It provides a REST API for the Management Platform to:

- Start/stop IPQuorum instances
- Monitor instance status
- Retrieve logs and metrics
- Manage configurations

## Architecture

```
┌─────────────────────────────────────┐
│  IPQuorum Management Platform       │
│  (Central Control)                  │
│  - Web Dashboard                    │
│  - API Server                       │
└──────────────┬──────────────────────┘
               │ HTTPS/TLS
               │ Port 9090
               ▼
┌─────────────────────────────────────┐
│  IPQuorum Agent                     │
│  (On IPQuorum Host)                 │
│  - REST API                         │
│  - Instance Manager                 │
│  - Health Monitor                   │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Existing IPQuorum Installation     │
│  - systemd services                 │
│  - ip_quorum.jar instances          │
│  - Configuration files              │
└─────────────────────────────────────┘
```

## Prerequisites

### On IPQuorum Host

- **Existing IPQuorum Installation**: Working IPQuorum setup with systemd services
- **OS**: Linux (RHEL/CentOS 8+, Ubuntu 20.04+)
- **Python**: 3.8+ or Go 1.21+ (depending on agent implementation)
- **Network**: Port 9090 accessible from Management Platform
- **Permissions**: Root or sudo access for systemd management

### On Management Platform

- Management Platform deployed and running
- Network connectivity to IPQuorum host
- TLS certificates (for secure communication)

## Agent Installation Methods

### Method 1: Systemd Service (Recommended)

This method integrates the agent as a systemd service alongside existing IPQuorum instances.

#### Step 1: Download Agent

```bash
# On IPQuorum host
cd /opt
sudo mkdir -p ipquorum-agent
cd ipquorum-agent

# Download agent binary (example - adjust URL)
sudo curl -L https://github.com/your-org/ipquorum-agent/releases/latest/download/ipquorum-agent-linux-amd64 -o ipquorum-agent
sudo chmod +x ipquorum-agent
```

#### Step 2: Create Configuration

```bash
# Create agent configuration
sudo tee /etc/ipquorum-agent/config.yaml <<EOF
# IPQuorum Agent Configuration

# Server settings
server:
  host: 0.0.0.0
  port: 9090
  
# Security
security:
  tls_enabled: true
  tls_cert: /etc/ipquorum-agent/certs/server.crt
  tls_key: /etc/ipquorum-agent/certs/server.key
  api_key: $(openssl rand -hex 32)
  
# IPQuorum settings
ipquorum:
  systemd_prefix: "ibm-virtualize-ipquorum"
  config_dir: /etc/ipquorum
  data_dir: /var/lib/ipquorum
  log_dir: /var/log/ipquorum
  
# Monitoring
monitoring:
  enabled: true
  interval: 30s
  
# Logging
logging:
  level: info
  format: json
  file: /var/log/ipquorum-agent/agent.log
EOF
```

**Important**: Save the generated `api_key` - you'll need it for the Management Platform configuration!

#### Step 3: Generate TLS Certificates

```bash
# Create certificate directory
sudo mkdir -p /etc/ipquorum-agent/certs

# Generate self-signed certificate (for testing)
sudo openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout /etc/ipquorum-agent/certs/server.key \
  -out /etc/ipquorum-agent/certs/server.crt \
  -days 365 \
  -subj "/CN=$(hostname -f)"

# Set permissions
sudo chmod 600 /etc/ipquorum-agent/certs/server.key
sudo chmod 644 /etc/ipquorum-agent/certs/server.crt
```

**Production**: Use proper CA-signed certificates or Let's Encrypt.

#### Step 4: Create Systemd Service

```bash
sudo tee /etc/systemd/system/ipquorum-agent.service <<EOF
[Unit]
Description=IPQuorum Management Agent
Documentation=https://github.com/your-org/ipquorum-agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/opt/ipquorum-agent/ipquorum-agent --config /etc/ipquorum-agent/config.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=ipquorum-agent

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/ipquorum /var/log/ipquorum-agent

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF
```

#### Step 5: Start Agent

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable and start agent
sudo systemctl enable ipquorum-agent
sudo systemctl start ipquorum-agent

# Check status
sudo systemctl status ipquorum-agent

# View logs
sudo journalctl -u ipquorum-agent -f
```

#### Step 6: Verify Agent

```bash
# Test health endpoint
curl -k https://localhost:9090/health

# Test with API key
curl -k -H "X-API-Key: YOUR_API_KEY" https://localhost:9090/api/v1/instances
```

### Method 2: Container Deployment

Deploy agent as a container with access to host systemd.

```bash
# Create agent container
podman run -d \
  --name ipquorum-agent \
  --network host \
  -v /run/systemd:/run/systemd:ro \
  -v /etc/ipquorum:/etc/ipquorum:ro \
  -v /var/lib/ipquorum:/var/lib/ipquorum \
  -v /var/log/ipquorum:/var/log/ipquorum:ro \
  -e AGENT_API_KEY=$(openssl rand -hex 32) \
  -e AGENT_TLS_ENABLED=true \
  --restart unless-stopped \
  ipquorum-agent:latest
```

### Method 3: Standalone Script

For simple deployments or testing.

```bash
# Create agent script
sudo tee /usr/local/bin/ipquorum-agent.sh <<'EOF'
#!/bin/bash
# Simple IPQuorum Agent

API_KEY="${AGENT_API_KEY:-changeme}"
PORT="${AGENT_PORT:-9090}"

# Start simple HTTP server
python3 -m http.server $PORT &
echo "Agent started on port $PORT"
EOF

sudo chmod +x /usr/local/bin/ipquorum-agent.sh
```

## Firewall Configuration

```bash
# Allow agent port
sudo firewall-cmd --permanent --add-port=9090/tcp
sudo firewall-cmd --reload

# Or with iptables
sudo iptables -A INPUT -p tcp --dport 9090 -j ACCEPT
sudo iptables-save > /etc/iptables/rules.v4
```

## Configure Management Platform

After agent is deployed, configure the Management Platform to connect to it.

### Step 1: Update Management Platform .env

On the Management Platform host:

```bash
cd ipquorum-management-platform

# Edit .env file
vi .env
```

Add/update these settings:

```bash
# Enable agent integration
IPQUORUM_AGENT_ENABLED=true

# Agent connection details
IPQUORUM_AGENT_HOST=10.0.0.100  # IPQuorum host IP
IPQUORUM_AGENT_PORT=9090
IPQUORUM_AGENT_API_KEY=your-api-key-from-agent-config

# TLS settings
IPQUORUM_AGENT_TLS_ENABLED=true
IPQUORUM_AGENT_TLS_VERIFY=true  # Set to false for self-signed certs
```

### Step 2: Restart Management Platform

```bash
# Restart services
./deploy-prod.sh restart

# Or with podman-compose
podman-compose -f docker-compose.prod.yml restart ipquorum-server
```

### Step 3: Verify Connection

```bash
# Check logs
./deploy-prod.sh logs ipquorum-server

# Test API endpoint
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/agent/status
```

## Agent API Endpoints

The agent exposes these endpoints:

### Health Check
```bash
GET /health
```

### List Instances
```bash
GET /api/v1/instances
Headers: X-API-Key: your-api-key
```

### Get Instance Status
```bash
GET /api/v1/instances/{name}/status
Headers: X-API-Key: your-api-key
```

### Start Instance
```bash
POST /api/v1/instances/{name}/start
Headers: X-API-Key: your-api-key
```

### Stop Instance
```bash
POST /api/v1/instances/{name}/stop
Headers: X-API-Key: your-api-key
```

### Get Logs
```bash
GET /api/v1/instances/{name}/logs?lines=100
Headers: X-API-Key: your-api-key
```

## Security Best Practices

### 1. Use Strong API Keys

```bash
# Generate secure API key
openssl rand -hex 32
```

### 2. Enable TLS

Always use TLS in production:
- Use CA-signed certificates
- Enable certificate verification
- Use TLS 1.2 or higher

### 3. Restrict Network Access

```bash
# Allow only Management Platform IP
sudo firewall-cmd --permanent --add-rich-rule='
  rule family="ipv4"
  source address="10.0.0.50/32"
  port protocol="tcp" port="9090" accept'
```

### 4. Use Least Privilege

Run agent with minimal required permissions:
- Read-only access to configs
- Write access only to necessary directories
- Use systemd security features

### 5. Monitor Agent Activity

```bash
# Enable audit logging
sudo auditctl -w /opt/ipquorum-agent/ipquorum-agent -p x -k ipquorum_agent

# Monitor logs
sudo journalctl -u ipquorum-agent -f
```

## Troubleshooting

### Agent Won't Start

```bash
# Check logs
sudo journalctl -u ipquorum-agent -n 50

# Check configuration
sudo /opt/ipquorum-agent/ipquorum-agent --config /etc/ipquorum-agent/config.yaml --validate

# Check permissions
ls -la /opt/ipquorum-agent/
ls -la /etc/ipquorum-agent/
```

### Connection Refused

```bash
# Check if agent is listening
sudo netstat -tlnp | grep 9090

# Test locally
curl -k https://localhost:9090/health

# Check firewall
sudo firewall-cmd --list-all
```

### TLS Certificate Issues

```bash
# Verify certificate
openssl x509 -in /etc/ipquorum-agent/certs/server.crt -text -noout

# Test TLS connection
openssl s_client -connect localhost:9090

# Disable TLS verification temporarily (testing only)
# In Management Platform .env:
IPQUORUM_AGENT_TLS_VERIFY=false
```

### Permission Denied

```bash
# Check systemd access
sudo systemctl --user status

# Verify agent can access systemd
sudo -u root systemctl list-units | grep ipquorum

# Check file permissions
ls -la /var/lib/ipquorum
ls -la /etc/ipquorum
```

## Monitoring Agent Health

### Prometheus Metrics

Agent exposes Prometheus metrics:

```bash
curl -k https://localhost:9090/metrics
```

### Health Checks

```bash
# Simple health check
curl -k https://localhost:9090/health

# Detailed status
curl -k -H "X-API-Key: YOUR_KEY" https://localhost:9090/api/v1/status
```

### Log Monitoring

```bash
# Follow agent logs
sudo journalctl -u ipquorum-agent -f

# Search for errors
sudo journalctl -u ipquorum-agent | grep -i error

# Export logs
sudo journalctl -u ipquorum-agent --since "1 hour ago" > agent-logs.txt
```

## Upgrading Agent

```bash
# Stop agent
sudo systemctl stop ipquorum-agent

# Backup configuration
sudo cp /etc/ipquorum-agent/config.yaml /etc/ipquorum-agent/config.yaml.bak

# Download new version
sudo curl -L https://github.com/your-org/ipquorum-agent/releases/latest/download/ipquorum-agent-linux-amd64 -o /opt/ipquorum-agent/ipquorum-agent.new

# Replace binary
sudo mv /opt/ipquorum-agent/ipquorum-agent.new /opt/ipquorum-agent/ipquorum-agent
sudo chmod +x /opt/ipquorum-agent/ipquorum-agent

# Start agent
sudo systemctl start ipquorum-agent

# Verify
sudo systemctl status ipquorum-agent
```

## Uninstalling Agent

```bash
# Stop and disable service
sudo systemctl stop ipquorum-agent
sudo systemctl disable ipquorum-agent

# Remove files
sudo rm -rf /opt/ipquorum-agent
sudo rm -rf /etc/ipquorum-agent
sudo rm /etc/systemd/system/ipquorum-agent.service

# Reload systemd
sudo systemctl daemon-reload

# Remove firewall rule
sudo firewall-cmd --permanent --remove-port=9090/tcp
sudo firewall-cmd --reload
```

## Example: Complete Setup

Here's a complete example for a typical deployment:

```bash
#!/bin/bash
# Complete IPQuorum Agent Setup

set -e

echo "Installing IPQuorum Agent..."

# 1. Create directories
sudo mkdir -p /opt/ipquorum-agent
sudo mkdir -p /etc/ipquorum-agent/certs
sudo mkdir -p /var/log/ipquorum-agent

# 2. Download agent (replace with actual URL)
cd /opt/ipquorum-agent
sudo curl -L https://example.com/ipquorum-agent -o ipquorum-agent
sudo chmod +x ipquorum-agent

# 3. Generate API key
API_KEY=$(openssl rand -hex 32)
echo "Generated API Key: $API_KEY"
echo "SAVE THIS KEY - You'll need it for the Management Platform!"

# 4. Generate TLS certificate
sudo openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout /etc/ipquorum-agent/certs/server.key \
  -out /etc/ipquorum-agent/certs/server.crt \
  -days 365 \
  -subj "/CN=$(hostname -f)"

# 5. Create configuration
sudo tee /etc/ipquorum-agent/config.yaml <<EOF
server:
  host: 0.0.0.0
  port: 9090
security:
  tls_enabled: true
  tls_cert: /etc/ipquorum-agent/certs/server.crt
  tls_key: /etc/ipquorum-agent/certs/server.key
  api_key: $API_KEY
ipquorum:
  systemd_prefix: "ibm-virtualize-ipquorum"
  config_dir: /etc/ipquorum
  data_dir: /var/lib/ipquorum
logging:
  level: info
  file: /var/log/ipquorum-agent/agent.log
EOF

# 6. Create systemd service
sudo tee /etc/systemd/system/ipquorum-agent.service <<'EOF'
[Unit]
Description=IPQuorum Management Agent
After=network-online.target

[Service]
Type=simple
User=root
ExecStart=/opt/ipquorum-agent/ipquorum-agent --config /etc/ipquorum-agent/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# 7. Configure firewall
sudo firewall-cmd --permanent --add-port=9090/tcp
sudo firewall-cmd --reload

# 8. Start service
sudo systemctl daemon-reload
sudo systemctl enable ipquorum-agent
sudo systemctl start ipquorum-agent

# 9. Verify
sleep 2
sudo systemctl status ipquorum-agent

echo ""
echo "Agent installation complete!"
echo "API Key: $API_KEY"
echo "Agent URL: https://$(hostname -f):9090"
echo ""
echo "Next steps:"
echo "1. Save the API key securely"
echo "2. Configure Management Platform with this host's IP and API key"
echo "3. Set IPQUORUM_AGENT_TLS_VERIFY=false if using self-signed certificate"
```

#