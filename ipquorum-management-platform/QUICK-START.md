# IPQuorum Management Platform - Quick Start Guide

Get up and running in 5 minutes! 🚀

## Prerequisites

- Docker or Podman installed
- 4GB RAM minimum
- Ports 3000, 8080 available

## Installation

### 1. Setup Environment

```bash
cd ipquorum-management-platform

# Copy environment template
cp .env.example .env

# Generate secure secrets (Linux/macOS)
sed -i "s|JWT_SECRET=.*|JWT_SECRET=$(openssl rand -base64 32)|g" .env
sed -i "s|GRAFANA_ADMIN_PASSWORD=.*|GRAFANA_ADMIN_PASSWORD=$(openssl rand -base64 16)|g" .env
```

### 2. Deploy

**Interactive Mode:**
```bash
./deploy-prod.sh
# Select option 1 for basic deployment
# Select option 2 for deployment with monitoring
```

**Command Line:**
```bash
# Basic deployment
./deploy-prod.sh start

# With monitoring (Prometheus + Grafana)
./deploy-prod.sh start monitoring
```

### 3. Access

Open your browser:
- **Web Dashboard**: http://localhost:3000
- **API Server**: http://localhost:8080

**Login:**
- Username: `admin`
- Password: `admin123`

⚠️ **Change password after first login!**

## IPQuorum Agent Configuration

To connect to an IPQuorum Agent, edit `.env`:

```bash
# Enable agent
IPQUORUM_AGENT_ENABLED=true

# Agent details
IPQUORUM_AGENT_HOST=10.0.0.100
IPQUORUM_AGENT_PORT=9090
IPQUORUM_AGENT_API_KEY=your-api-key

# TLS settings
IPQUORUM_AGENT_TLS_ENABLED=true
IPQUORUM_AGENT_TLS_VERIFY=true
```

Then restart:
```bash
./deploy-prod.sh restart
```

## Common Commands

```bash
# Start services
./deploy-prod.sh start

# Stop services
./deploy-prod.sh stop

# Restart services
./deploy-prod.sh restart

# View status
./deploy-prod.sh status

# View logs
./deploy-prod.sh logs

# View specific service logs
./deploy-prod.sh logs ipquorum-server

# Health check
./deploy-prod.sh health

# Clean up (removes containers and volumes)
./deploy-prod.sh clean
```

## Monitoring (Optional)

If deployed with monitoring profile:

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001
  - Username: `admin`
  - Password: (from `.env` file)

## Troubleshooting

### Services won't start
```bash
# Check logs
./deploy-prod.sh logs

# Check if ports are in use
netstat -tuln | grep -E '3000|8080'

# Change ports in .env if needed
WEB_PORT=3001
```

### Can't access web dashboard
```bash
# Check service status
./deploy-prod.sh status

# Restart services
./deploy-prod.sh restart

# Check health
curl http://localhost:8080/health
```

### Database issues
```bash
# Stop services
./deploy-prod.sh stop

# Remove data directory
rm -rf data/

# Start fresh
./deploy-prod.sh start
```

## Configuration Files

- `.env` - Environment variables
- `docker-compose.prod.yml` - Production compose file
- `data/` - Database and persistent data
- `logs/` - Application logs

## Next Steps

1. **Change default password** in web dashboard
2. **Configure IPQuorum Agent** (if needed)
3. **Set up monitoring** (optional)
4. **Configure backups** (see PRODUCTION-DEPLOYMENT.md)
5. **Review security settings** (see PRODUCTION-DEPLOYMENT.md)

## Documentation

- **Full Deployment Guide**: [PRODUCTION-DEPLOYMENT.md](PRODUCTION-DEPLOYMENT.md)
- **Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)
- **API Documentation**: [API-SPECIFICATION.md](API-SPECIFICATION.md)
- **Monitoring**: [MONITORING.md](MONITORING.md)
- **Testing**: [TESTING-GUIDE.md](TESTING-GUIDE.md)
- **WebSocket**: [WEBSOCKET.md](WEBSOCKET.md)

## Support

Need help?
1. Check logs: `./deploy-prod.sh logs`
2. Review [PRODUCTION-DEPLOYMENT.md](PRODUCTION-DEPLOYMENT.md)
3. Check [Troubleshooting section](#troubleshooting)

#