# IP Quorum Management Platform

> Modern Go-based management platform for IBM Storage Virtualize IP Quorum services with web dashboard, REST API, monitoring, and multi-server orchestration.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react)](https://react.dev/)

## 🎯 Overview

The IP Quorum Management Platform is a next-generation management solution that extends the existing Bash-based IP Quorum service with:

- 🌐 **Web Dashboard** - Modern React-based UI for easy management
- 🔌 **REST API** - Programmatic access for automation and integration
- 📊 **Monitoring** - Real-time health checks and metrics collection
- 🖥️ **Multi-Server** - Centralized management of multiple servers
- 🔐 **Security** - JWT authentication with role-based access control
- 🔄 **Backward Compatible** - Works alongside existing Bash scripts

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    IP Quorum Management Platform                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┐         ┌──────────────────┐            │
│  │  Web Dashboard   │◄────────┤   REST API       │            │
│  │  (React + TS)    │         │   (Go + Gin)     │            │
│  └──────────────────┘         └──────────────────┘            │
│           │                            │                        │
│           │ WebSocket                  │ HTTP/JSON              │
│           ▼                            ▼                        │
│  ┌─────────────────────────────────────────────────┐          │
│  │         Management Server (Go)                   │          │
│  │  • Instance Management                           │          │
│  │  • Health Monitoring                             │          │
│  │  • Multi-Server Orchestration                    │          │
│  └─────────────────────────────────────────────────┘          │
│           │                            │                        │
│           ▼                            ▼                        │
│  ┌──────────────────┐         ┌──────────────────┐            │
│  │  Local Agent     │         │   Database       │            │
│  │  (Go Binary)     │         │   (SQLite/PG)    │            │
│  └──────────────────┘         └──────────────────┘            │
│           │                                                     │
│           ▼                                                     │
│  ┌──────────────────────────────────────────────┐             │
│  │  Existing Bash Scripts (Backward Compatible) │             │
│  └──────────────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

## ✨ Features

### Current (v2.0.7 - Bash-based)
- ✅ Multi-instance systemd service management
- ✅ Interactive instance creation
- ✅ Network connectivity checks
- ✅ Automatic JAR download
- ✅ Password management
- ✅ Health validation

### Planned (v3.0.0+ - Platform)
- 🚀 Web-based management dashboard
- 🚀 REST API for automation
- 🚀 Real-time monitoring and alerts
- 🚀 Multi-server orchestration
- 🚀 Centralized configuration management
- 🚀 Advanced health checks and diagnostics
- 🚀 Prometheus metrics export
- 🚀 Role-based access control
- 🚀 Audit logging
- 🚀 Backup and restore

## 📋 Prerequisites

### Management Server
- Linux (RHEL 8+, Ubuntu 20.04+)
- Go 1.21+ (for building from source)
- 2GB RAM minimum
- 10GB disk space

### Managed Servers
- Linux with systemd
- Bash 4.0+
- Java 11+ (for IP Quorum JAR)
- Network access to IBM Storage Virtualize

## 🚀 Quick Start

### Option 1: Binary Installation (Recommended)

```bash
# Download management server
wget https://github.com/olemyk/ipquorum-platform/releases/latest/download/ipquorum-server-linux-amd64
sudo mv ipquorum-server-linux-amd64 /usr/local/bin/ipquorum-server
sudo chmod +x /usr/local/bin/ipquorum-server

# Create configuration
sudo mkdir -p /etc/ipquorum-platform
sudo ipquorum-server init

# Start server
sudo systemctl enable ipquorum-server
sudo systemctl start ipquorum-server

# Access web dashboard
https://localhost:8443
# Default credentials: admin / changeme
```

### Option 2: Docker

```bash
# Pull image
docker pull ghcr.io/olemyk/ipquorum-platform:latest

# Run container
docker run -d \
  --name ipquorum-platform \
  -p 8443:8443 \
  -v /var/lib/ipquorum-platform:/data \
  ghcr.io/olemyk/ipquorum-platform:latest

# Access dashboard
https://localhost:8443
```

### Option 3: Build from Source

```bash
# Clone repository
git clone https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service.git
cd ibm-storage-virtualize-ipquorum-service/ipquorum-management-platform

# Build management server
cd server
go build -o ipquorum-server cmd/server/main.go

# Build web dashboard
cd ../dashboard
npm install
npm run build

# Build agent
cd ../agent
go build -o ipquorum-agent cmd/agent/main.go

# Run server
./server/ipquorum-server
```

## 📖 Documentation

- [Architecture Design](./ARCHITECTURE.md) - Detailed system architecture
- [API Documentation](./API.md) - REST API reference
- [Development Guide](./DEVELOPMENT.md) - Setup development environment
- [Deployment Guide](./DEPLOYMENT.md) - Production deployment
- [Migration Guide](./MIGRATION.md) - Migrate from Bash scripts

## 🔌 REST API

### Authentication
```bash
# Login
curl -X POST https://localhost:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"changeme"}'

# Response
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600
}
```

### Instance Management
```bash
# List instances
curl -H "Authorization: Bearer $TOKEN" \
  https://localhost:8443/api/v1/instances

# Create instance
curl -X POST https://localhost:8443/api/v1/instances \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prod-cluster01",
    "api_endpoint": "10.33.7.80",
    "username": "superuser",
    "password": "secret",
    "enable_download": true,
    "enable_mkquorumapp": true
  }'

# Start instance
curl -X POST https://localhost:8443/api/v1/instances/prod-cluster01/start \
  -H "Authorization: Bearer $TOKEN"

# Get instance status
curl -H "Authorization: Bearer $TOKEN" \
  https://localhost:8443/api/v1/instances/prod-cluster01/status
```

### Health & Monitoring
```bash
# System health
curl https://localhost:8443/api/v1/health

# Instance metrics
curl -H "Authorization: Bearer $TOKEN" \
  https://localhost:8443/api/v1/metrics/instances/prod-cluster01

# Prometheus metrics
curl https://localhost:8443/metrics
```

## 🖥️ Web Dashboard

### Features
- **Dashboard Overview** - System health at a glance
- **Instance Management** - Create, start, stop, delete instances
- **Real-time Monitoring** - Live status updates via WebSocket
- **Health Checks** - Port connectivity and process status
- **Metrics Visualization** - CPU, memory, network charts
- **Multi-Server View** - Manage instances across servers
- **Audit Logs** - Track all user actions
- **Settings** - Configure system preferences

### Screenshots

#### Dashboard Overview
```
┌─────────────────────────────────────────────────────────┐
│ IP Quorum Management Platform                           │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  System Health: ● Healthy                              │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐            │
│  │ Instances│  │  Servers │  │  Alerts  │            │
│  │    12    │  │     3    │  │     2    │            │
│  └──────────┘  └──────────┘  └──────────┘            │
│                                                         │
│  Recent Activity:                                       │
│  • instance-prod01 started                             │
│  • instance-dev02 health check passed                  │
│  • New server registered: server03                     │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## 🔐 Security

### Authentication
- JWT-based authentication
- Secure password hashing (bcrypt)
- Token expiration and refresh
- Session management

### Authorization (RBAC)
- **Admin** - Full system access
- **Operator** - Instance management
- **Viewer** - Read-only access

### Network Security
- TLS 1.3 encryption
- Mutual TLS for agent communication
- Rate limiting
- CORS protection

### Best Practices
```bash
# Change default password immediately
curl -X PUT https://localhost:8443/api/v1/auth/password \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"old_password":"changeme","new_password":"StrongP@ssw0rd!"}'

# Use TLS certificates
ipquorum-server --tls-cert /path/to/cert.pem --tls-key /path/to/key.pem

# Enable audit logging
ipquorum-server --audit-log /var/log/ipquorum/audit.log
```

## 📊 Monitoring

### Prometheus Metrics
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'ipquorum-platform'
    static_configs:
      - targets: ['localhost:8443']
    metrics_path: '/metrics'
```

### Available Metrics
```
# Instance metrics
ipquorum_instance_status{instance="name", server="hostname"}
ipquorum_instance_health{instance="name", server="hostname"}
ipquorum_instance_uptime_seconds{instance="name"}

# System metrics
ipquorum_api_requests_total{method="GET", endpoint="/instances"}
ipquorum_api_request_duration_seconds{method="GET"}
ipquorum_health_checks_total{instance="name", result="success"}
```

### Grafana Dashboard
```bash
# Import dashboard
curl -X POST http://grafana:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @grafana-dashboard.json
```

## 🔄 Backward Compatibility

The platform maintains full backward compatibility with existing Bash scripts:

```bash
# Existing Bash scripts continue to work
sudo ipquorum-instance-manager.sh create prod-cluster01

# Platform agent executes Bash scripts internally
# No migration required for existing instances
```

### Migration Path
1. Install management platform alongside existing setup
2. Import existing instances into platform
3. Use web dashboard for new instances
4. Gradually migrate management to platform
5. Bash scripts remain as execution layer

## 🛠️ Development

### Prerequisites
- Go 1.21+
- Node.js 18+
- npm or yarn
- Docker (optional)

### Setup Development Environment
```bash
# Clone repository
git clone https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service.git
cd ibm-storage-virtualize-ipquorum-service/ipquorum-management-platform

# Install dependencies
make deps

# Run tests
make test

# Start development server
make dev

# Build for production
make build
```

### Project Structure
```
ipquorum-management-platform/
├── server/                 # Go backend
│   ├── cmd/
│   ├── internal/
│   └── pkg/
├── agent/                  # Go agent
│   ├── cmd/
│   └── internal/
├── dashboard/              # React frontend
│   ├── src/
│   └── public/
├── docs/                   # Documentation
├── scripts/                # Build scripts
└── deployments/            # Deployment configs
```

## 🗺️ Roadmap

### Phase 1: Foundation (v3.0.0) - Q2 2026
- [x] Architecture design
- [ ] Management server core
- [ ] REST API implementation
- [ ] Local agent
- [ ] Basic authentication
- [ ] Bash script integration

### Phase 2: Web Dashboard (v3.1.0) - Q3 2026
- [ ] React dashboard
- [ ] Instance management UI
- [ ] Real-time status updates
- [ ] Basic monitoring

### Phase 3: Monitoring (v3.2.0) - Q4 2026
- [ ] Health check system
- [ ] Metrics collection
- [ ] Prometheus integration
- [ ] Alerting

### Phase 4: Multi-Server (v3.3.0) - Q1 2027
- [ ] Server registration
- [ ] Multi-server orchestration
- [ ] Distributed monitoring

### Phase 5: Advanced Features (v3.4.0) - Q2 2027
- [ ] Advanced RBAC
- [ ] Audit logging
- [ ] Backup/restore
- [ ] Configuration templates

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](../LICENSE) file for details.

## 🙏 Acknowledgments

- IBM Storage Virtualize team for the IP Quorum service
- Carbon Design System for UI components
- Go and React communities

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/issues)
- **Discussions**: [GitHub Discussions](https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/discussions)
- **Documentation**: [Wiki](https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/wiki)

## 🔗 Related Projects

- [ipquorum-download-go](../ipquorum-download-go/) - Go-based download tool
- [ipquorum-systemd](../ipquorum-systemd/) - Systemd service (current v2.0.7)

---

**Note**: This is a new platform currently in planning phase. The existing Bash-based solution (v2.0.7) is production-ready and fully functional. This platform will extend and enhance the current capabilities while maintaining backward compatibility.