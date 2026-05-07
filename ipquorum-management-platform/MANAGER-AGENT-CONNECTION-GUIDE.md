# IPQuorum Manager-Agent Connection Guide

## Overview

This guide explains how to connect the IPQuorum Management Platform (Manager) to IPQuorum Agents running on remote hosts. Understanding the current architecture and deployment models is crucial for successful integration.

## Current Architecture (v2.0.8)

### What We Have Now

```
┌─────────────────────────────────────────────────────────┐
│  IPQuorum Management Platform (Manager)                 │
│  Running on: localhost (or dedicated server)            │
├─────────────────────────────────────────────────────────┤
│  ┌──────────────────┐         ┌──────────────────┐    │
│  │  Web Dashboard   │◄────────┤   REST API       │    │
│  │  Port: 3000      │         │   Port: 8443     │    │
│  └──────────────────┘         └──────────────────┘    │
│           │                            │                │
│           │                            │                │
│           ▼                            ▼                │
│  ┌─────────────────────────────────────────────────┐  │
│  │         Management Server (Go)                   │  │
│  │  • Instance Management (Database)                │  │
│  │  • Configuration Storage                         │  │
│  │  • User Authentication                           │  │
│  │  • Health Monitoring                             │  │
│  └─────────────────────────────────────────────────┘  │
│           │                                             │
│           │ Currently: LOCAL execution only             │
│           ▼                                             │
│  ┌─────────────────────────────────────────────────┐  │
│  │  Local Executor (Bash Scripts)                   │  │
│  │  • Executes ipquorum-instance-manager.sh         │  │
│  │  • Manages systemd services locally              │  │
│  └─────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### What We Need for Remote Management

```
┌─────────────────────────────────────────────────────────┐
│  IPQuorum Management Platform (Manager)                 │
│  Running on: Central Server                             │
├─────────────────────────────────────────────────────────┤
│  ┌──────────────────┐         ┌──────────────────┐    │
│  │  Web Dashboard   │◄────────┤   REST API       │    │
│  │  Port: 3000      │         │   Port: 8443     │    │
│  └──────────────────┘         └──────────────────┘    │
│           │                            │                │
│           │                            │                │
│           ▼                            ▼                │
│  ┌─────────────────────────────────────────────────┐  │
│  │         Management Server (Go)                   │  │
│  │  • Instance Management (Database)                │  │
│  │  • Agent Communication Layer (HTTP Client)       │  │
│  │  • Multi-Server Orchestration                    │  │
│  └─────────────────────────────────────────────────┘  │
└──────────────────┬──────────────────────────────────────┘
                   │
                   │ HTTPS/REST API
                   │ Port: 9090
                   │
         ┌─────────┼─────────┬─────────┐
         │                   │         │
         ▼                   ▼         ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  Agent Host 1   │  │  Agent Host 2   │  │  Agent Host 3   │
│  10.0.0.100     │  │  10.0.0.101     │  │  10.0.0.102     │
├─────────────────┤  ├─────────────────┤  ├─────────────────┤
│ IPQuorum Agent  │  │ IPQuorum Agent  │  │ IPQuorum Agent  │
│ Port: 9090      │  │ Port: 9090      │  │ Port: 9090      │
│ • REST API      │  │ • REST API      │  │ • REST API      │
│ • Executor      │  │ • Executor      │  │ • Executor      │
│ • Monitor       │  │ • Monitor       │  │ • Monitor       │
├─────────────────┤  ├─────────────────┤  ├─────────────────┤
│ IPQuorum        │  │ IPQuorum        │  │ IPQuorum        │
│ Instances       │  │ Instances       │  │ Instances       │
│ • instance1     │  │ • instance3     │  │ • instance5     │
│ • instance2     │  │ • instance4     │  │ • instance6     │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

## Two Deployment Models

### Model 1: Local Deployment (Current - v2.0.8)

**Use Case**: Single server with Manager and Agent on the same host

**Architecture**:
```
┌─────────────────────────────────────────┐
│  Single Linux Server                    │
├─────────────────────────────────────────┤
│  Manager (Containers)                   │
│  ├─ ipquorum-server (Port 8443)        │
│  └─ ipquorum-web (Port 3000)           │
│                                         │
│  Agent (Systemd Service)                │
│  └─ ipquorum-agent (Port 9090)         │
│                                         │
│  IPQuorum Instances (Systemd)           │
│  ├─ ipquorum@instance1.service         │
│  └─ ipquorum@instance2.service         │
└─────────────────────────────────────────┘
```

**Connection**: Manager connects to `http://localhost:9090` or `http://127.0.0.1:9090`

**Pros**:
- ✅ Simple setup
- ✅ No network configuration needed
- ✅ Fast communication (localhost)
- ✅ Good for testing and small deployments

**Cons**:
- ❌ Single point of failure
- ❌ Limited scalability
- ❌ All instances on one server

### Model 2: Remote Deployment (Target Architecture)

**Use Case**: Multiple servers with centralized management

**Architecture**:
```
┌─────────────────────────────────────────┐
│  Management Server (10.0.0.50)          │
├─────────────────────────────────────────┤
│  Manager (Containers)                   │
│  ├─ ipquorum-server (Port 8443)        │
│  └─ ipquorum-web (Port 3000)           │
└──────────────┬──────────────────────────┘
               │
               │ Network: 10.0.0.0/24
               │
    ┌──────────┼──────────┐
    │                     │
    ▼                     ▼
┌─────────────────┐  ┌─────────────────┐
│ Agent Host 1    │  │ Agent Host 2    │
│ 10.0.0.100      │  │ 10.0.0.101      │
├─────────────────┤  ├─────────────────┤
│ Agent + IPQ     │  │ Agent + IPQ     │
└─────────────────┘  └─────────────────┘
```

**Connection**: Manager connects to `https://10.0.0.100:9090`, `https://10.0.0.101:9090`, etc.

**Pros**:
- ✅ Scalable to many servers
- ✅ Centralized management
- ✅ Fault isolation
- ✅ Load distribution

**Cons**:
- ❌ More complex setup
- ❌ Network configuration required
- ❌ TLS/security considerations
- ❌ Firewall rules needed

## Current Implementation Status

### ✅ What's Working (v2.0.8)

1. **Manager Server**:
   - ✅ REST API for instance management
   - ✅ Database for storing instance configurations
   - ✅ Web dashboard for UI
   - ✅ Authentication (JWT)
   - ✅ All new agent fields supported

2. **Agent**:
   - ✅ REST API endpoints (start/stop/restart/status/logs)
   - ✅ Systemd integration
   - ✅ Configuration management
   - ✅ Health monitoring
   - ✅ Successfully deployed on RHEL

3. **Integration**:
   - ✅ Manager can store instance configurations
   - ✅ Manager has executor for local operations
   - ✅ Agent has all required endpoints

### ❌ What's Missing (Gap Analysis)

1. **Manager → Agent Communication**:
   - ❌ HTTP client in Manager to call Agent API
   - ❌ Agent registration in Manager database
   - ❌ Agent health monitoring from Manager
   - ❌ Multi-agent orchestration logic

2. **Configuration**:
   - ❌ Agent connection settings in Manager
   - ❌ API key management
   - ❌ TLS certificate handling

3. **UI**:
   - ❌ Server/Agent management page
   - ❌ Agent status indicators
   - ❌ Multi-server instance view

## How to Connect: Step-by-Step Guide

### Scenario 1: Local Deployment (Same Server)

This is the **simplest** way to get started and test the integration.

#### Prerequisites
- Manager running (containers on localhost:8443 and localhost:3000)
- Agent deployed on same server (see AGENT-DEPLOYMENT-GUIDE.md)

#### Step 1: Verify Agent is Running

```bash
# Check agent status
sudo systemctl status ipquorum-agent

# Test agent API
curl -k https://localhost:9090/health

# Should return: {"status":"healthy"}
```

#### Step 2: Get Agent API Key

```bash
# View agent configuration
sudo cat /etc/ipquorum-agent/config.yaml | grep api_key

# Example output:
# api_key: a1b2c3d4e5f6...
```

#### Step 3: Configure Manager to Use Agent

**Option A: Environment Variables** (Recommended for containers)

```bash
cd ipquorum-management-platform

# Edit .env file
vi .env

# Add these lines:
AGENT_ENABLED=true
AGENT_HOST=localhost
AGENT_PORT=9090
AGENT_API_KEY=your-api-key-from-step-2
AGENT_TLS_ENABLED=true
AGENT_TLS_VERIFY=false  # For self-signed certs
```

**Option B: Configuration File**

```bash
# Edit server config
vi ipquorum-management-platform/server/config.yaml

# Add agent section:
agent:
  enabled: true
  host: localhost
  port: 9090
  api_key: your-api-key-from-step-2
  tls:
    enabled: true
    verify: false
```

#### Step 4: Restart Manager

```bash
cd ipquorum-management-platform

# Restart containers
podman-compose -f docker-compose.prod.yml restart ipquorum-server

# Check logs
podman logs ipquorum-server --tail 50
```

#### Step 5: Test Connection

```bash
# Get Manager auth token
TOKEN=$(curl -s -X POST http://localhost:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')

# Test agent connection through Manager
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8443/api/v1/agents/status | jq '.'
```

#### Step 6: Create Instance via Manager

```bash
# Create instance - Manager will call Agent API
curl -s -X POST http://localhost:8443/api/v1/instances \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-instance",
    "api_endpoint": "10.0.0.100",
    "username": "superuser",
    "password": "password",
    "partnersystem": "remote-cluster",
    "enable_download": true,
    "enable_mkquorumapp": true
  }' | jq '.'
```

### Scenario 2: Remote Deployment (Multiple Servers)

This is for **production** deployments with multiple IPQuorum hosts.

#### Prerequisites
- Manager running on central server (e.g., 10.0.0.50)
- Agent deployed on remote hosts (e.g., 10.0.0.100, 10.0.0.101)
- Network connectivity between Manager and Agents
- Firewall rules allowing port 9090

#### Step 1: Deploy Agent on Remote Hosts

On each IPQuorum host (10.0.0.100, 10.0.0.101, etc.):

```bash
# Download and run agent installer
curl -L https://your-repo/agent-install.sh | sudo bash

# Or manually follow AGENT-DEPLOYMENT-GUIDE.md
```

#### Step 2: Configure Firewall on Agent Hosts

```bash
# Allow Manager IP to access agent port
sudo firewall-cmd --permanent --add-rich-rule='
  rule family="ipv4"
  source address="10.0.0.50/32"
  port protocol="tcp" port="9090" accept'

sudo firewall-cmd --reload
```

#### Step 3: Test Agent Connectivity from Manager

From Manager server (10.0.0.50):

```bash
# Test network connectivity
curl -k https://10.0.0.100:9090/health
curl -k https://10.0.0.101:9090/health

# Test with API key
curl -k -H "X-API-Key: agent-api-key" \
  https://10.0.0.100:9090/api/v1/instances
```

#### Step 4: Register Agents in Manager

**Currently**: This requires manual database entry or API implementation.

**Future**: Will be done via Web UI or API.

**Manual Method** (temporary):

```bash
# Connect to Manager database
podman exec -it ipquorum-server sqlite3 /data/ipquorum.db

# Insert agent/server record
INSERT INTO servers (id, hostname, ip_address, agent_port, status, version, created_at)
VALUES (
  'agent-host-1',
  'rhel94.example.com',
  '10.33.3.215',
  9090,
  'online',
  '1.0.5',
  datetime('now')
);

# Repeat for each agent
```

#### Step 5: Configure Manager for Multi-Agent

```bash
# Edit Manager .env
vi ipquorum-management-platform/.env

# Add:
AGENT_ENABLED=true
MULTI_AGENT_MODE=true
AGENT_DISCOVERY=manual  # or 'auto' when implemented
```

#### Step 6: Create Instances on Specific Agents

```bash
# Create instance on specific agent
curl -s -X POST http://10.0.0.50:8443/api/v1/instances \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "instance-on-host1",
    "server_id": "agent-host-1",
    "api_endpoint": "10.0.0.200",
    "username": "superuser",
    "password": "password",
    "partnersystem": "remote-cluster"
  }' | jq '.'
```

## Implementation Roadmap

To fully enable Manager-Agent communication, these components need to be implemented:

### Phase 1: Basic Agent Communication (Next Steps)

1. **Manager HTTP Client**:
   ```go
   // server/internal/agent/client.go
   type AgentClient struct {
       baseURL    string
       apiKey     string
       httpClient *http.Client
   }
   
   func (c *AgentClient) StartInstance(name string) error
   func (c *AgentClient) StopInstance(name string) error
   func (c *AgentClient) GetStatus(name string) (*Status, error)
   ```

2. **Agent Registry**:
   ```go
   // server/internal/service/agent_registry.go
   type AgentRegistry struct {
       agents map[string]*AgentClient
   }
   
   func (r *AgentRegistry) GetAgent(serverID string) (*AgentClient, error)
   func (r *AgentRegistry) RegisterAgent(serverID, host string, port int, apiKey string) error
   ```

3. **Update Executor**:
   ```go
   // server/internal/executor/executor.go
   func (e *Executor) StartInstance(ctx context.Context, instance *Instance) error {
       if instance.ServerID == "local" {
           // Use local bash scripts
           return e.startLocal(ctx, instance)
       } else {
           // Use agent API
           agent := e.agentRegistry.GetAgent(instance.ServerID)
           return agent.StartInstance(instance.Name)
       }
   }
   ```

### Phase 2: Agent Management UI

1. **Server Management Page**:
   - List all registered agents
   - Show agent status (online/offline)
   - Add/remove agents
   - Test agent connectivity

2. **Instance Assignment**:
   - Select target server when creating instance
   - Show which server each instance is on
   - Move instances between servers

### Phase 3: Advanced Features

1. **Agent Health Monitoring**:
   - Periodic health checks
   - Alert on agent failures
   - Auto-reconnect logic

2. **Load Balancing**:
   - Distribute instances across agents
   - Consider server capacity
   - Automatic failover

3. **Bulk Operations**:
   - Start/stop multiple instances
   - Deploy to multiple servers
   - Configuration sync

## Troubleshooting

### Connection Issues

**Problem**: Manager can't connect to Agent

```bash
# Check 1: Network connectivity
ping 10.0.0.100
telnet 10.0.0.100 9090

# Check 2: Agent is running
ssh user@10.0.0.100
sudo systemctl status ipquorum-agent

# Check 3: Firewall
sudo firewall-cmd --list-all

# Check 4: TLS certificate
openssl s_client -connect 10.0.0.100:9090
```

**Problem**: Authentication fails

```bash
# Verify API key matches
# On Agent:
sudo cat /etc/ipquorum-agent/config.yaml | grep api_key

# On Manager:
cat ipquorum-management-platform/.env | grep AGENT_API_KEY

# They must match!
```

**Problem**: TLS verification fails

```bash
# Temporary fix: Disable TLS verification
# In Manager .env:
AGENT_TLS_VERIFY=false

# Permanent fix: Use proper certificates
# See AGENT-DEPLOYMENT-GUIDE.md for certificate setup
```

### Performance Issues

**Problem**: Slow response from Agent

```bash
# Check Agent logs
sudo journalctl -u ipquorum-agent -n 100

# Check system resources
top
df -h
```

**Problem**: Timeout errors

```bash
# Increase timeout in Manager config
# In .env:
AGENT_TIMEOUT=60  # seconds
```

## Security Considerations

### 1. API Key Management

```bash
# Generate strong API keys
openssl rand -hex 32

# Rotate keys regularly
# Store securely (environment variables, secrets manager)
```

### 2. TLS/SSL

```bash
# Always use TLS in production
AGENT_TLS_ENABLED=true
AGENT_TLS_VERIFY=true

# Use CA-signed certificates
# Or Let's Encrypt for public-facing servers
```

### 3. Network Security

```bash
# Restrict access by IP
sudo firewall-cmd --permanent --add-rich-rule='
  rule family="ipv4"
  source address="10.0.0.50/32"
  port protocol="tcp" port="9090" accept'

# Use VPN for remote access
# Consider mTLS for mutual authentication
```

### 4. Audit Logging

```bash
# Enable audit logging on Agent
# In config.yaml:
logging:
  audit: true
  audit_file: /var/log/ipquorum-agent/audit.log

# Monitor access
sudo tail -f /var/log/ipquorum-agent/audit.log
```

## Quick Reference

### Manager Endpoints (for Agent Integration)

```bash
# List agents
GET /api/v1/agents

# Get agent status
GET /api/v1/agents/{id}/status

# Register agent
POST /api/v1/agents
{
  "hostname": "host1.example.com",
  "ip_address": "10.0.0.100",
  "agent_port": 9090,
  "api_key": "..."
}

# Test agent connection
POST /api/v1/agents/{id}/test
```

### Agent Endpoints (Called by Manager)

```bash
# Health check
GET /health

# List instances
GET /api/v1/instances

# Start instance
POST /api/v1/instances/{name}/start

# Stop instance
POST /api/v1/instances/{name}/stop

# Get status
GET /api/v1/instances/{name}/status

# Get logs
GET /api/v1/instances/{name}/logs?lines=100
```

## Next Steps

1. **For Local Testing**:
   - Follow "Scenario 1: Local Deployment"
   - Test basic Manager-Agent communication
   - Create instances via Manager

2. **For Production**:
   - Deploy agents on all IPQuorum hosts
   - Implement Manager HTTP client (Phase 1)
   - Add agent management UI (Phase 2)
   - Enable advanced features (Phase 3)

3. **For Development**:
   - Review `server/internal/executor/executor.go`
   - Implement `server/internal/agent/client.go`
   - Update API handlers to support multi-agent
   - Add agent management endpoints

## Conclusion

The IPQuorum Management Platform is designed to support both local and remote agent deployments. While the current v2.0.8 release has all the foundational components (Manager, Agent, API), the Manager-Agent communication layer needs to be implemented for full remote management capabilities.

For immediate use, deploy in **Local Mode** where Manager and Agent run on the same server. For production multi-server deployments, follow the implementation roadmap to add the necessary communication layer.

#