# IP Quorum Management Platform - Docker Deployment Guide

> Docker-specific deployment instructions for the IP Quorum Management Platform

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Deployment Methods](#deployment-methods)
- [Configuration](#configuration)
- [Management](#management)
- [Networking](#networking)
- [Storage](#storage)
- [Security](#security)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

## Overview

This guide covers Docker-specific deployment scenarios for the IP Quorum Management Platform. For general deployment information, see [DEPLOYMENT-GUIDE.md](./DEPLOYMENT-GUIDE.md).

### Docker Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Host                          │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │         ipquorum-network (bridge)                │  │
│  │                                                  │  │
│  │  ┌──────────────┐  ┌──────────────┐            │  │
│  │  │ ipquorum-web │  │ ipquorum-    │            │  │
│  │  │ (nginx)      │  │ server       │            │  │
│  │  │ Port: 3000   │  │ Port: 8443   │            │  │
│  │  └──────────────┘  └──────────────┘            │  │
│  │         │                  │                    │  │
│  │         └──────────────────┘                    │  │
│  └──────────────────────────────────────────────────┘  │
│                                                         │
│  Volumes:                                               │
│  • ipquorum-data    → /data                            │
│  • ipquorum-logs    → /var/log/ipquorum               │
│  • ipquorum-config  → /config                         │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Prerequisites

### System Requirements

- **OS**: Linux (RHEL 8+, Ubuntu 20.04+, Debian 11+)
- **CPU**: 2+ cores (4+ recommended)
- **RAM**: 4GB minimum (8GB recommended)
- **Disk**: 20GB minimum (50GB recommended)
- **Docker**: 20.10+ (24.0+ recommended)
- **Docker Compose**: 2.0+ (plugin or standalone)

### Install Docker

#### RHEL/CentOS/Rocky Linux

```bash
# Remove old versions
sudo dnf remove docker docker-client docker-client-latest \
    docker-common docker-latest docker-latest-logrotate \
    docker-logrotate docker-engine

# Add Docker repository
sudo dnf install -y dnf-plugins-core
sudo dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

# Install Docker
sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Start and enable Docker
sudo systemctl enable --now docker

# Verify installation
docker --version
docker compose version
```

#### Ubuntu/Debian

```bash
# Remove old versions
sudo apt-get remove docker docker-engine docker.io containerd runc

# Update package index
sudo apt-get update

# Install dependencies
sudo apt-get install -y ca-certificates curl gnupg lsb-release

# Add Docker's official GPG key
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | \
    sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Set up repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Start and enable Docker
sudo systemctl enable --now docker

# Verify installation
docker --version
docker compose version
```

### Post-Installation Steps

```bash
# Add user to docker group (optional, allows non-root access)
sudo usermod -aG docker $USER

# Apply group changes (or logout/login)
newgrp docker

# Verify Docker works without sudo
docker run hello-world

# Configure Docker daemon (optional)
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "storage-driver": "overlay2",
  "live-restore": true
}
EOF

# Restart Docker
sudo systemctl restart docker
```

## Installation

### Quick Start with Automated Script

```bash
# Download and run deployment script
curl -fsSL https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-ipquorum-service/main/ipquorum-management-platform/deploy.sh -o deploy.sh
chmod +x deploy.sh
sudo ./deploy.sh --docker
```

### Manual Installation

#### 1. Create Directory Structure

```bash
# Create deployment directory
sudo mkdir -p /opt/ipquorum-platform
cd /opt/ipquorum-platform

# Create subdirectories
sudo mkdir -p data logs config scripts backups

# Set ownership
sudo chown -R $USER:$USER /opt/ipquorum-platform
```

#### 2. Download Configuration

```bash
# Download docker-compose file
curl -fsSL -o docker-compose.prod.yml \
    https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-ipquorum-service/main/ipquorum-management-platform/docker-compose.prod.yml

# Download environment template
curl -fsSL -o .env \
    https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-ipquorum-service/main/ipquorum-management-platform/.env.example
```

#### 3. Configure Environment

```bash
# Edit .env file
nano .env

# Generate secure secrets
JWT_SECRET=$(openssl rand -base64 32)
GRAFANA_PASSWORD=$(openssl rand -base64 16)

# Update .env with generated secrets
sed -i "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
sed -i "s|GRAFANA_ADMIN_PASSWORD=.*|GRAFANA_ADMIN_PASSWORD=$GRAFANA_PASSWORD|" .env
```

#### 4. Pull Images

```bash
# Pull latest images
docker pull ghcr.io/olemyk/ipquorum-manager:v3.0.0
docker pull ghcr.io/olemyk/ipquorum-web:v3.0.0

# Verify images
docker images | grep ipquorum
```

#### 5. Start Services

```bash
# Start all services
docker compose -f docker-compose.prod.yml up -d

# Verify services are running
docker compose -f docker-compose.prod.yml ps

# Check logs
docker compose -f docker-compose.prod.yml logs -f
```

## Deployment Methods

### Method 1: Docker Compose (Recommended)

```bash
# Start services
docker compose -f docker-compose.prod.yml up -d

# Start with monitoring
docker compose -f docker-compose.prod.yml --profile monitoring up -d

# Stop services
docker compose -f docker-compose.prod.yml down

# Stop and remove volumes (WARNING: destroys data)
docker compose -f docker-compose.prod.yml down -v
```

### Method 2: Docker Run (Manual)

```bash
# Create network
docker network create ipquorum-network

# Create volumes
docker volume create ipquorum-data
docker volume create ipquorum-logs

# Run manager
docker run -d \
  --name ipquorum-server \
  --network ipquorum-network \
  -p 8443:8080 \
  -v ipquorum-data:/data \
  -v ipquorum-logs:/var/log/ipquorum \
  -e JWT_SECRET="your-secret-key" \
  -e DB_PATH="/data/ipquorum.db" \
  -e LOG_LEVEL="info" \
  --restart unless-stopped \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Run web dashboard
docker run -d \
  --name ipquorum-web \
  --network ipquorum-network \
  -p 3000:80 \
  --restart unless-stopped \
  ghcr.io/olemyk/ipquorum-web:v3.0.0
```

### Method 3: Docker Swarm

```bash
# Initialize swarm
docker swarm init

# Create secrets
echo "your-jwt-secret" | docker secret create jwt_secret -
echo "your-db-password" | docker secret create db_password -

# Deploy stack
docker stack deploy -c docker-compose.prod.yml ipquorum

# List services
docker stack services ipquorum

# View logs
docker service logs ipquorum_ipquorum-server

# Remove stack
docker stack rm ipquorum
```

## Configuration

### Environment Variables

Create or edit `.env` file:

```bash
# Security
JWT_SECRET=your-super-secret-jwt-key-change-this
JWT_EXPIRATION=24h

# Database
DB_PATH=/data/ipquorum.db

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Ports
WEB_PORT=3000
API_PORT=8443
PROMETHEUS_PORT=9090
GRAFANA_PORT=3001

# Performance
MAX_CONCURRENT_OPERATIONS=10
OPERATION_TIMEOUT=300s

# Monitoring
PROMETHEUS_RETENTION=15d
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=your-secure-password
```

### Docker Compose Override

Create `docker-compose.override.yml` for custom settings:

```yaml
version: '3.8'

services:
  ipquorum-server:
    environment:
      - LOG_LEVEL=debug
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G

  ipquorum-web:
    ports:
      - "80:80"  # Expose on port 80
```

## Management

### Service Management

```bash
# Start services
docker compose -f docker-compose.prod.yml up -d

# Stop services
docker compose -f docker-compose.prod.yml stop

# Restart services
docker compose -f docker-compose.prod.yml restart

# Restart specific service
docker compose -f docker-compose.prod.yml restart ipquorum-server

# View status
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs -f

# View logs for specific service
docker compose -f docker-compose.prod.yml logs -f ipquorum-server

# Execute command in container
docker compose -f docker-compose.prod.yml exec ipquorum-server sh

# Scale services (if configured)
docker compose -f docker-compose.prod.yml up -d --scale ipquorum-server=3
```

### Container Management

```bash
# List containers
docker ps
docker ps -a  # Include stopped containers

# Inspect container
docker inspect ipquorum-server

# View container logs
docker logs ipquorum-server
docker logs -f ipquorum-server  # Follow logs

# Execute command in container
docker exec -it ipquorum-server sh
docker exec ipquorum-server ls -la /data

# Copy files to/from container
docker cp config.yaml ipquorum-server:/config/
docker cp ipquorum-server:/data/ipquorum.db ./backup/

# View container stats
docker stats
docker stats ipquorum-server ipquorum-web

# Stop container
docker stop ipquorum-server

# Start container
docker start ipquorum-server

# Remove container
docker rm ipquorum-server
docker rm -f ipquorum-server  # Force remove
```

### Image Management

```bash
# List images
docker images

# Pull specific version
docker pull ghcr.io/olemyk/ipquorum-manager:v3.1.0

# Tag image
docker tag ghcr.io/olemyk/ipquorum-manager:v3.0.0 ipquorum-manager:latest

# Remove image
docker rmi ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Remove unused images
docker image prune
docker image prune -a  # Remove all unused images

# Build custom image
docker build -t ipquorum-manager:custom -f Containerfile .
```

## Networking

### Network Management

```bash
# List networks
docker network ls

# Inspect network
docker network inspect ipquorum-network

# Create custom network
docker network create \
  --driver bridge \
  --subnet 172.28.0.0/16 \
  --gateway 172.28.0.1 \
  ipquorum-custom-network

# Connect container to network
docker network connect ipquorum-network container-name

# Disconnect container from network
docker network disconnect ipquorum-network container-name

# Remove network
docker network rm ipquorum-network
```

### Port Mapping

```bash
# View port mappings
docker port ipquorum-server

# Custom port mapping
docker run -d \
  -p 8080:8080 \
  -p 8443:8443 \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Bind to specific interface
docker run -d \
  -p 127.0.0.1:8443:8080 \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

### DNS Configuration

```bash
# Custom DNS servers
docker run -d \
  --dns 8.8.8.8 \
  --dns 8.8.4.4 \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Add to /etc/hosts
docker run -d \
  --add-host=storage1:10.0.0.10 \
  --add-host=storage2:10.0.0.11 \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

## Storage

### Volume Management

```bash
# List volumes
docker volume ls

# Inspect volume
docker volume inspect ipquorum-data

# Create volume
docker volume create ipquorum-data

# Remove volume
docker volume rm ipquorum-data

# Remove unused volumes
docker volume prune

# Backup volume
docker run --rm \
  -v ipquorum-data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/ipquorum-data-backup.tar.gz /data

# Restore volume
docker run --rm \
  -v ipquorum-data:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/ipquorum-data-backup.tar.gz -C /
```

### Bind Mounts

```bash
# Use bind mounts instead of volumes
docker run -d \
  -v /opt/ipquorum-platform/data:/data \
  -v /opt/ipquorum-platform/logs:/var/log/ipquorum \
  -v /opt/ipquorum-platform/config:/config:ro \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

## Security

### TLS/SSL Configuration

```bash
# Generate self-signed certificate
mkdir -p config/tls
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout config/tls/key.pem \
  -out config/tls/cert.pem \
  -subj "/CN=ipquorum.local"

# Mount certificates
docker run -d \
  -v $(pwd)/config/tls:/config/tls:ro \
  -e TLS_ENABLED=true \
  -e TLS_CERT_FILE=/config/tls/cert.pem \
  -e TLS_KEY_FILE=/config/tls/key.pem \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

### Secrets Management

```bash
# Using Docker secrets (Swarm mode)
echo "my-secret-key" | docker secret create jwt_secret -

# Using environment file
cat > .env.secret << 'EOF'
JWT_SECRET=my-secret-key
DB_PASSWORD=my-db-password
EOF
chmod 600 .env.secret

docker compose --env-file .env.secret -f docker-compose.prod.yml up -d
```

### Security Best Practices

```bash
# Run as non-root user (already configured in images)
docker run -d --user 1000:1000 ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Read-only root filesystem
docker run -d --read-only \
  -v ipquorum-data:/data \
  -v ipquorum-logs:/var/log/ipquorum \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Drop capabilities
docker run -d \
  --cap-drop=ALL \
  --cap-add=NET_BIND_SERVICE \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Security scanning
docker scan ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

## Monitoring

### Health Checks

```bash
# View health status
docker ps --format "table {{.Names}}\t{{.Status}}"

# Inspect health check
docker inspect --format='{{json .State.Health}}' ipquorum-server | jq

# Custom health check
docker run -d \
  --health-cmd="curl -f http://localhost:8080/health || exit 1" \
  --health-interval=30s \
  --health-timeout=3s \
  --health-retries=3 \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

### Resource Monitoring

```bash
# Real-time stats
docker stats

# Export stats
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}"

# Container resource limits
docker run -d \
  --cpus="2" \
  --memory="4g" \
  --memory-swap="4g" \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

### Log Management

```bash
# View logs
docker logs ipquorum-server
docker logs -f ipquorum-server  # Follow
docker logs --tail 100 ipquorum-server  # Last 100 lines
docker logs --since 1h ipquorum-server  # Last hour

# Configure logging driver
docker run -d \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Export logs
docker logs ipquorum-server > logs/export-$(date +%Y%m%d).log
```

## Troubleshooting

### Common Issues

#### Container Won't Start

```bash
# Check container status
docker ps -a

# View logs
docker logs ipquorum-server

# Inspect container
docker inspect ipquorum-server

# Check for port conflicts
sudo netstat -tulpn | grep -E ':(3000|8443)'

# Remove and recreate
docker rm -f ipquorum-server
docker compose -f docker-compose.prod.yml up -d
```

#### Network Issues

```bash
# Test network connectivity
docker exec ipquorum-server ping -c 3 ipquorum-web
docker exec ipquorum-web curl http://ipquorum-server:8080/health

# Inspect network
docker network inspect ipquorum-network

# Recreate network
docker compose -f docker-compose.prod.yml down
docker network rm ipquorum-network
docker compose -f docker-compose.prod.yml up -d
```

#### Volume Issues

```bash
# Check volume mounts
docker inspect ipquorum-server | grep -A 10 Mounts

# Check permissions
docker exec ipquorum-server ls -la /data

# Fix permissions
docker exec -u root ipquorum-server chown -R 1000:1000 /data
```

### Debug Mode

```bash
# Run with debug logging
docker run -d \
  -e LOG_LEVEL=debug \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Interactive shell
docker run -it --rm \
  -v ipquorum-data:/data \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0 sh

# Override entrypoint
docker run -it --rm \
  --entrypoint sh \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

## Best Practices

### Production Deployment

1. **Use Docker Compose** for multi-container orchestration
2. **Pin image versions** (avoid `:latest` tag)
3. **Use named volumes** for persistent data
4. **Configure resource limits** to prevent resource exhaustion
5. **Enable health checks** for automatic recovery
6. **Use secrets** for sensitive data
7. **Configure logging** with rotation
8. **Regular backups** of volumes and configuration
9. **Monitor resource usage** and set alerts
10. **Keep Docker updated** for security patches

### Performance Tuning

```yaml
# docker-compose.prod.yml
services:
  ipquorum-server:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G
    ulimits:
      nofile:
        soft: 65536
        hard: 65536
```

### Backup Strategy

```bash
#!/bin/bash
# backup-docker.sh

BACKUP_DIR="/backup/ipquorum"
DATE=$(date +%Y%m%d-%H%M%S)

# Backup volumes
docker run --rm \
  -v ipquorum-data:/data:ro \
  -v $BACKUP_DIR:/backup \
  alpine tar czf /backup/data-$DATE.tar.gz /data

# Backup configuration
tar czf $BACKUP_DIR/config-$DATE.tar.gz .env docker-compose.prod.yml

# Cleanup old backups (keep 30 days)
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete
```

## Next Steps

- [General Deployment Guide](./DEPLOYMENT-GUIDE.md)
- [Podman Deployment](./PODMAN-DEPLOYMENT.md)
- [API Documentation](./API-SPECIFICATION.md)
- [Monitoring Guide](./MONITORING.md)

---

**Last Updated**: 2026-05-07  
**Version**: 3.0.0