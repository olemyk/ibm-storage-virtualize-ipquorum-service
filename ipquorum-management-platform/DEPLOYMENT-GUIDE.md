# IP Quorum Management Platform - Deployment Guide

> Comprehensive guide for deploying the IP Quorum Management Platform using pre-built container images

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Deployment Options](#deployment-options)
- [Configuration](#configuration)
- [Security](#security)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)
- [Upgrade Guide](#upgrade-guide)

## Overview

The IP Quorum Management Platform is distributed as pre-built container images via GitHub Container Registry (ghcr.io). This guide covers deployment using Docker or Podman.

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Load Balancer (Optional)             │
│                    nginx/traefik/haproxy                │
└─────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼────────┐  ┌──────▼───────┐  ┌────────▼────────┐
│  Web Dashboard │  │  API Server  │  │   Prometheus    │
│  (Port 3000)   │  │  (Port 8443) │  │   (Port 9090)   │
│  ipquorum-web  │  │  ipquorum-   │  │   (Optional)    │
│                │  │   manager    │  │                 │
└────────────────┘  └──────────────┘  └─────────────────┘
                            │
                    ┌───────┴────────┐
                    │   SQLite DB    │
                    │   (Volume)     │
                    └────────────────┘
```

### Container Images

- **Manager (Backend)**: `ghcr.io/olemyk/ipquorum-manager:v3.0.0`
- **Web (Frontend)**: `ghcr.io/olemyk/ipquorum-web:v3.0.0`

## Prerequisites

### System Requirements

**Minimum:**
- 2 CPU cores
- 4GB RAM
- 20GB disk space
- Linux OS (RHEL 8+, Ubuntu 20.04+, Debian 11+)

**Recommended:**
- 4 CPU cores
- 8GB RAM
- 50GB disk space
- SSD storage

### Software Requirements

Choose one of the following container runtimes:

#### Option 1: Docker (Recommended)

```bash
# RHEL/CentOS/Rocky Linux
sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y docker.io docker-compose-v2

# Start and enable Docker
sudo systemctl enable --now docker

# Add user to docker group (optional)
sudo usermod -aG docker $USER
newgrp docker

# Verify installation
docker --version
docker compose version
```

#### Option 2: Podman

```bash
# RHEL/CentOS/Rocky Linux
sudo dnf install -y podman podman-compose

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y podman podman-compose

# Verify installation
podman --version
podman-compose --version
```

### Network Requirements

- **Inbound Ports:**
  - 3000/tcp - Web Dashboard
  - 8443/tcp - API Server
  - 9090/tcp - Prometheus (optional)
  - 3001/tcp - Grafana (optional)

- **Outbound Connectivity:**
  - ghcr.io (443/tcp) - Pull container images
  - IBM Storage Virtualize systems - IP Quorum communication

### Firewall Configuration

```bash
# RHEL/CentOS/Rocky Linux (firewalld)
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --permanent --add-port=8443/tcp
sudo firewall-cmd --permanent --add-port=9090/tcp
sudo firewall-cmd --reload

# Ubuntu/Debian (ufw)
sudo ufw allow 3000/tcp
sudo ufw allow 8443/tcp
sudo ufw allow 9090/tcp
sudo ufw reload
```

## Quick Start

### 1. Create Deployment Directory

```bash
# Create directory structure
sudo mkdir -p /opt/ipquorum-platform/{data,logs,config,scripts}
cd /opt/ipquorum-platform

# Set permissions
sudo chown -R $USER:$USER /opt/ipquorum-platform
```

### 2. Download Configuration Files

```bash
# Download docker-compose file
wget https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-ipquorum-service/main/ipquorum-management-platform/docker-compose.prod.yml

# Download environment template
wget -O .env https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-ipquorum-service/main/ipquorum-management-platform/.env.example
```

### 3. Configure Environment

Edit the `.env` file:

```bash
nano .env
```

**Minimum required configuration:**

```bash
# Security - CHANGE THESE!
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
GRAFANA_ADMIN_PASSWORD=your-secure-grafana-password

# Database
DB_PATH=/data/ipquorum.db

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Ports (optional, defaults shown)
WEB_PORT=3000
API_PORT=8443
PROMETHEUS_PORT=9090
GRAFANA_PORT=3001
```

### 4. Pull Container Images

```bash
# Using Docker
docker pull ghcr.io/olemyk/ipquorum-manager:v3.0.0
docker pull ghcr.io/olemyk/ipquorum-web:v3.0.0

# Using Podman
podman pull ghcr.io/olemyk/ipquorum-manager:v3.0.0
podman pull ghcr.io/olemyk/ipquorum-web:v3.0.0
```

### 5. Start Services

```bash
# Using Docker
docker compose -f docker-compose.prod.yml up -d

# Using Podman
podman-compose -f docker-compose.prod.yml up -d

# Verify services are running
docker compose -f docker-compose.prod.yml ps
# or
podman-compose -f docker-compose.prod.yml ps
```

### 6. Verify Deployment

```bash
# Check API health
curl http://localhost:8443/health

# Check Web dashboard
curl http://localhost:3000/

# View logs
docker compose -f docker-compose.prod.yml logs -f
# or
podman-compose -f docker-compose.prod.yml logs -f
```

### 7. Access Web Dashboard

Open your browser and navigate to:
- **Web Dashboard**: http://localhost:3000
- **Default Credentials**: 
  - Username: `admin`
  - Password: `changeme` (change immediately!)

### 8. Initial Setup

1. **Change Default Password**
   - Login with default credentials
   - Navigate to Settings → Security
   - Change admin password

2. **Configure First Server**
   - Go to Servers → Add Server
   - Enter server details
   - Test connection

3. **Create First Instance**
   - Go to Instances → Create Instance
   - Fill in IBM Storage Virtualize details
   - Start instance

## Deployment Options

### Option 1: Standalone Deployment (Single Host)

Best for: Development, testing, small deployments

```bash
cd /opt/ipquorum-platform
docker compose -f docker-compose.prod.yml up -d
```

### Option 2: Production Deployment with Monitoring

Includes Prometheus and Grafana for monitoring:

```bash
# Start all services including monitoring
docker compose -f docker-compose.prod.yml --profile monitoring up -d

# Access Grafana
# URL: http://localhost:3001
# Default: admin / (password from .env)
```

### Option 3: High Availability Deployment

For production environments requiring HA:

```bash
# Use external load balancer (nginx, haproxy, traefik)
# Deploy multiple instances behind load balancer
# Use shared storage for database (NFS, Ceph, etc.)

# Example with 3 nodes:
# Node 1: Primary manager + web
# Node 2: Secondary manager + web
# Node 3: Monitoring (Prometheus + Grafana)
```

### Option 4: Kubernetes Deployment

See [KUBERNETES-DEPLOYMENT.md](./KUBERNETES-DEPLOYMENT.md) for Kubernetes-specific instructions.

## Configuration

### Environment Variables

#### Core Settings

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_PATH` | Database file path | `/data/ipquorum.db` | Yes |
| `JWT_SECRET` | JWT signing key | - | Yes |
| `JWT_EXPIRATION` | Token expiration | `24h` | No |
| `LOG_LEVEL` | Logging level | `info` | No |
| `LOG_FORMAT` | Log format (json/text) | `json` | No |

#### Network Settings

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `WEB_PORT` | Web dashboard port | `3000` | No |
| `API_PORT` | API server port | `8443` | No |
| `PROMETHEUS_PORT` | Prometheus port | `9090` | No |
| `GRAFANA_PORT` | Grafana port | `3001` | No |

#### Performance Settings

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `MAX_CONCURRENT_OPERATIONS` | Max parallel operations | `10` | No |
| `OPERATION_TIMEOUT` | Operation timeout | `300s` | No |
| `NGINX_WORKER_PROCESSES` | Nginx workers | `auto` | No |
| `NGINX_WORKER_CONNECTIONS` | Nginx connections | `1024` | No |

### Volume Mounts

```yaml
volumes:
  - ./data:/data                    # Database and persistent data
  - ./logs:/var/log/ipquorum       # Application logs
  - ./scripts:/app/scripts:ro      # Custom scripts (read-only)
  - ./config:/config:ro            # Configuration files (read-only)
```

### Custom Configuration File

Create `config/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8443
  tls:
    enabled: false
    cert_file: "/config/tls/cert.pem"
    key_file: "/config/tls/key.pem"

database:
  path: "/data/ipquorum.db"
  max_connections: 25
  connection_timeout: 30s

security:
  jwt_secret: "${JWT_SECRET}"
  jwt_expiration: "24h"
  password_min_length: 12
  session_timeout: "30m"

logging:
  level: "info"
  format: "json"
  output: "/var/log/ipquorum/app.log"

monitoring:
  enabled: true
  prometheus_port: 9090
  health_check_interval: "30s"
```

## Security

### TLS/SSL Configuration

#### Generate Self-Signed Certificate

```bash
# Create TLS directory
mkdir -p config/tls

# Generate certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout config/tls/key.pem \
  -out config/tls/cert.pem \
  -subj "/CN=ipquorum.local"

# Set permissions
chmod 600 config/tls/key.pem
chmod 644 config/tls/cert.pem
```

#### Use Let's Encrypt Certificate

```bash
# Install certbot
sudo dnf install -y certbot

# Obtain certificate
sudo certbot certonly --standalone -d your-domain.com

# Copy certificates
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem config/tls/cert.pem
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem config/tls/key.pem
```

### Secrets Management

#### Using Docker Secrets

```bash
# Create secrets
echo "your-jwt-secret" | docker secret create jwt_secret -
echo "your-db-password" | docker secret create db_password -

# Update docker-compose.prod.yml to use secrets
```

#### Using Environment Files

```bash
# Create secure .env file
cat > .env.secure << 'EOF'
JWT_SECRET=$(openssl rand -base64 32)
DB_PASSWORD=$(openssl rand -base64 24)
GRAFANA_ADMIN_PASSWORD=$(openssl rand -base64 16)
EOF

# Set restrictive permissions
chmod 600 .env.secure

# Use in docker-compose
docker compose --env-file .env.secure -f docker-compose.prod.yml up -d
```

### Network Security

#### Restrict Access with Firewall

```bash
# Allow only specific IPs
sudo firewall-cmd --permanent --add-rich-rule='rule family="ipv4" source address="10.0.0.0/8" port port="8443" protocol="tcp" accept'
sudo firewall-cmd --reload
```

#### Use Reverse Proxy

Example nginx configuration:

```nginx
upstream ipquorum_backend {
    server localhost:8443;
}

upstream ipquorum_frontend {
    server localhost:3000;
}

server {
    listen 443 ssl http2;
    server_name ipquorum.example.com;

    ssl_certificate /etc/ssl/certs/cert.pem;
    ssl_certificate_key /etc/ssl/private/key.pem;

    # API
    location /api/ {
        proxy_pass http://ipquorum_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Web Dashboard
    location / {
        proxy_pass http://ipquorum_frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Monitoring

### Prometheus Metrics

Access Prometheus at: http://localhost:9090

**Key Metrics:**
- `ipquorum_instance_status` - Instance status (0=stopped, 1=running)
- `ipquorum_instance_health` - Health check status
- `ipquorum_api_requests_total` - Total API requests
- `ipquorum_api_request_duration_seconds` - Request duration

### Grafana Dashboards

Access Grafana at: http://localhost:3001

**Pre-configured Dashboards:**
1. IP Quorum Overview
2. Instance Health
3. API Performance
4. System Resources

### Health Checks

```bash
# API health
curl http://localhost:8443/health

# Detailed health
curl http://localhost:8443/api/v1/health/detailed

# Instance health
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8443/api/v1/instances/instance-name/health
```

### Log Management

```bash
# View all logs
docker compose -f docker-compose.prod.yml logs

# Follow logs
docker compose -f docker-compose.prod.yml logs -f

# View specific service logs
docker compose -f docker-compose.prod.yml logs ipquorum-server

# Export logs
docker compose -f docker-compose.prod.yml logs > logs/export-$(date +%Y%m%d).log
```

## Troubleshooting

### Common Issues

#### 1. Container Won't Start

```bash
# Check container status
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs ipquorum-server

# Check for port conflicts
sudo netstat -tulpn | grep -E ':(3000|8443|9090)'

# Verify image integrity
docker images | grep ipquorum
```

#### 2. Database Connection Issues

```bash
# Check database file permissions
ls -la data/ipquorum.db

# Verify volume mount
docker inspect ipquorum-server | grep -A 10 Mounts

# Reset database (WARNING: destroys data)
docker compose -f docker-compose.prod.yml down -v
rm -f data/ipquorum.db
docker compose -f docker-compose.prod.yml up -d
```

#### 3. Authentication Failures

```bash
# Reset admin password
docker exec -it ipquorum-server /app/ipquorum-server reset-password --user admin

# Check JWT secret
docker exec ipquorum-server env | grep JWT_SECRET

# Verify token expiration
curl -v http://localhost:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"changeme"}'
```

#### 4. Network Connectivity Issues

```bash
# Test container network
docker network inspect ipquorum-network

# Check DNS resolution
docker exec ipquorum-server nslookup ipquorum-web

# Test inter-container communication
docker exec ipquorum-web curl http://ipquorum-server:8443/health
```

#### 5. Performance Issues

```bash
# Check resource usage
docker stats

# Increase resource limits in docker-compose.prod.yml
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
```

### Debug Mode

Enable debug logging:

```bash
# Update .env
LOG_LEVEL=debug

# Restart services
docker compose -f docker-compose.prod.yml restart

# View debug logs
docker compose -f docker-compose.prod.yml logs -f | grep DEBUG
```

### Support Resources

- **Documentation**: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service
- **Issues**: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/issues
- **Discussions**: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/discussions

## Upgrade Guide

### Backup Before Upgrade

```bash
# Stop services
docker compose -f docker-compose.prod.yml down

# Backup database
cp data/ipquorum.db data/ipquorum.db.backup-$(date +%Y%m%d)

# Backup configuration
tar -czf config-backup-$(date +%Y%m%d).tar.gz .env config/
```

### Upgrade Process

```bash
# Pull new images
docker pull ghcr.io/olemyk/ipquorum-manager:v3.1.0
docker pull ghcr.io/olemyk/ipquorum-web:v3.1.0

# Update docker-compose.prod.yml with new versions
sed -i 's|:v3.0.0|:v3.1.0|g' docker-compose.prod.yml

# Start services
docker compose -f docker-compose.prod.yml up -d

# Verify upgrade
docker compose -f docker-compose.prod.yml ps
curl http://localhost:8443/health
```

### Rollback

```bash
# Stop services
docker compose -f docker-compose.prod.yml down

# Restore database
cp data/ipquorum.db.backup-YYYYMMDD data/ipquorum.db

# Revert to previous version
sed -i 's|:v3.1.0|:v3.0.0|g' docker-compose.prod.yml

# Start services
docker compose -f docker-compose.prod.yml up -d
```

## Maintenance

### Regular Tasks

#### Daily
- Monitor logs for errors
- Check instance health
- Verify backup completion

#### Weekly
- Review resource usage
- Update security patches
- Test backup restoration

#### Monthly
- Rotate logs
- Update container images
- Review access logs
- Performance tuning

### Backup Strategy

```bash
#!/bin/bash
# backup.sh - Daily backup script

BACKUP_DIR="/backup/ipquorum"
DATE=$(date +%Y%m%d)

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup database
docker exec ipquorum-server sqlite3 /data/ipquorum.db ".backup /data/backup-$DATE.db"
docker cp ipquorum-server:/data/backup-$DATE.db $BACKUP_DIR/

# Backup configuration
tar -czf $BACKUP_DIR/config-$DATE.tar.gz .env config/

# Cleanup old backups (keep 30 days)
find $BACKUP_DIR -name "*.db" -mtime +30 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "Backup completed: $DATE"
```

### Log Rotation

```bash
# Create logrotate configuration
sudo tee /etc/logrotate.d/ipquorum << 'EOF'
/opt/ipquorum-platform/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 root root
    sharedscripts
    postrotate
        docker compose -f /opt/ipquorum-platform/docker-compose.prod.yml restart
    endscript
}
EOF
```

## Next Steps

- [Docker-Specific Deployment](./DOCKER-DEPLOYMENT.md)
- [Podman-Specific Deployment](./PODMAN-DEPLOYMENT.md)
- [Kubernetes Deployment](./KUBERNETES-DEPLOYMENT.md)
- [API Documentation](./API-SPECIFICATION.md)
- [Architecture Overview](./ARCHITECTURE.md)

---

**Last Updated**: 2026-05-07  
**Version**: 3.0.0