# IPQuorum Management Agent

A lightweight agent for managing IPQuorum instances on remote hosts via REST API.

## Overview

The IPQuorum Agent runs as a systemd service on each host that needs to manage IPQuorum instances. It provides a REST API that the management server can call to perform operations like creating, starting, stopping, and monitoring IPQuorum instances.

## Features

- 🔐 **API Key Authentication** - Secure communication with management server
- 🚀 **Systemd Integration** - Runs as a reliable systemd service
- 📊 **Health Monitoring** - Built-in health check endpoint
- 🔧 **Script Execution** - Executes bash scripts to manage instances
- 📝 **Structured Logging** - JSON logging with configurable levels
- ⚡ **Lightweight** - Minimal resource footprint

## Architecture

```
┌─────────────────────────────────┐
│  Management Server              │
│  (Container or Host)            │
└────────────┬────────────────────┘
             │ HTTPS + API Key
             ↓
┌─────────────────────────────────┐
│  IPQuorum Agent (Port 9090)     │
│  ┌───────────────────────────┐  │
│  │  REST API Server          │  │
│  │  - Authentication         │  │
│  │  - Request Handling       │  │
│  └───────────┬───────────────┘  │
│              ↓                   │
│  ┌───────────────────────────┐  │
│  │  Script Executor          │  │
│  │  - Bash Script Calls      │  │
│  │  - Timeout Management     │  │
│  └───────────┬───────────────┘  │
│              ↓                   │
│  ┌───────────────────────────┐  │
│  │  Systemd Integration      │  │
│  │  - Service Management     │  │
│  │  - Status Monitoring      │  │
│  └───────────────────────────┘  │
└─────────────────────────────────┘
```

## Installation

### Prerequisites

- Go 1.23 or later
- Linux with systemd
- Root access
- IPQuorum instance manager script

### Quick Install

```bash
# Clone the repository
cd ipquorum-management-platform/agent

# Run installation script (as root)
sudo ./install-agent.sh
```

The installation script will:
1. Build the agent binary
2. Install it to `/usr/local/bin/ipquorum-agent`
3. Create configuration directory `/etc/ipquorum-agent`
4. Generate a secure API key
5. Install and enable the systemd service

### Manual Installation

```bash
# Build the agent
make build

# Copy binary
sudo cp build/ipquorum-agent /usr/local/bin/

# Create config directory
sudo mkdir -p /etc/ipquorum-agent

# Copy and edit configuration
sudo cp config.yaml.example /etc/ipquorum-agent/config.yaml
sudo nano /etc/ipquorum-agent/config.yaml

# Install systemd service
sudo cp ipquorum-agent.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable ipquorum-agent
```

## Configuration

Edit `/etc/ipquorum-agent/config.yaml`:

```yaml
agent:
  port: 9090                    # API port
  host: "0.0.0.0"              # Bind address
  api_key: "your-api-key"      # Authentication key
  log_level: "info"            # debug, info, warn, error
  instance_name: "agent-01"    # Agent identifier

server:
  url: "https://mgmt:8080"     # Management server URL
  insecure: false              # Allow insecure TLS

script:
  path: "/usr/local/bin/ipquorum-instance-manager.sh"
  timeout: 300                 # Command timeout (seconds)
```

### Security Notes

- **API Key**: Generate a strong random key: `openssl rand -hex 32`
- **File Permissions**: Config file should be `600` (owner read/write only)
- **Network**: Consider firewall rules to restrict access to port 9090

## Usage

### Start the Agent

```bash
# Start service
sudo systemctl start ipquorum-agent

# Check status
sudo systemctl status ipquorum-agent

# View logs
sudo journalctl -u ipquorum-agent -f
```

### API Endpoints

All endpoints require the `X-API-Key` header with your configured API key.

#### Health Check
```bash
curl http://localhost:9090/health
```

#### Agent Info
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:9090/info
```

#### List Instances
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:9090/instances
```

#### Create Instance
```bash
curl -X POST \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ipquorum-01",
    "api_endpoint": "10.0.0.1",
    "username": "admin",
    "password": "password"
  }' \
  http://localhost:9090/instances
```

#### Get Instance Status
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:9090/instances/ipquorum-01/status
```

#### Start/Stop/Restart Instance
```bash
# Start
curl -X POST -H "X-API-Key: your-api-key" \
  http://localhost:9090/instances/ipquorum-01/start

# Stop
curl -X POST -H "X-API-Key: your-api-key" \
  http://localhost:9090/instances/ipquorum-01/stop

# Restart
curl -X POST -H "X-API-Key: your-api-key" \
  http://localhost:9090/instances/ipquorum-01/restart
```

#### Get Instance Logs
```bash
curl -H "X-API-Key: your-api-key" \
  "http://localhost:9090/instances/ipquorum-01/logs?lines=100"
```

#### Delete Instance
```bash
curl -X DELETE -H "X-API-Key: your-api-key" \
  "http://localhost:9090/instances/ipquorum-01?force=true"
```

## Development

### Build

```bash
# Build for current platform
make build

# Build for Linux
make build-linux

# Run tests
make test

# Show coverage
make coverage
```

### Project Structure

```
agent/
├── cmd/
│   └── agent/          # Main application
│       └── main.go
├── internal/
│   ├── api/           # REST API server
│   │   └── server.go
│   ├── config/        # Configuration management
│   │   └── config.go
│   └── executor/      # Script execution
│       └── executor.go
├── pkg/
│   └── logger/        # Logging utilities
│       └── logger.go
├── config.yaml.example
├── ipquorum-agent.service
├── install-agent.sh
├── Makefile
└── README.md
```

## Troubleshooting

### Agent Won't Start

```bash
# Check service status
sudo systemctl status ipquorum-agent

# View detailed logs
sudo journalctl -u ipquorum-agent -n 50

# Check configuration
sudo cat /etc/ipquorum-agent/config.yaml

# Verify script exists
ls -l /usr/local/bin/ipquorum-instance-manager.sh
```

### API Returns 401 Unauthorized

- Verify API key in configuration matches the one you're using
- Check that `X-API-Key` header is set correctly
- Ensure config file permissions are correct (600)

### Script Execution Fails

- Verify script path in configuration
- Check script has execute permissions
- Test script manually: `/usr/local/bin/ipquorum-instance-manager.sh list`
- Check timeout setting if operations are slow

### Port Already in Use

```bash
# Check what's using port 9090
sudo lsof -i :9090

# Change port in config.yaml and restart
sudo systemctl restart ipquorum-agent
```

## Integration with Management Server

To connect the agent to the management server, you'll need to:

1. Note the agent's API key from `/etc/ipquorum-agent/config.yaml`
2. Configure the management server to communicate with this agent
3. Ensure network connectivity between server and agent (port 9090)
4. Test connectivity: `curl -H "X-API-Key: KEY" http://agent-host:9090/health`

## Security Considerations

- ✅ Run agent as root (required for systemd operations)
- ✅ Use strong API keys (32+ characters, random)
- ✅ Restrict network access to port 9090
- ✅ Use TLS/HTTPS in production (configure reverse proxy)
- ✅ Regularly rotate API keys
- ✅ Monitor agent logs for suspicious activity
- ✅ Keep agent binary updated

## License

Same as the main IPQuorum Management Platform project.

## Support

For issues, questions, or contributions, please refer to the main project repository.