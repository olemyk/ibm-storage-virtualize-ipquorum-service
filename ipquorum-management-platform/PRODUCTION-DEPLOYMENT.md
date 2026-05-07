# IPQuorum Management Platform - Production Deployment Guide

This guide covers production deployment of the IPQuorum Management Platform using Docker or Podman.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Configuration](#configuration)
4. [IPQuorum Agent Integration](#ipquorum-agent-integration)
5. [Deployment Options](#deployment-options)
6. [Security Hardening](#security-hardening)
7. [Monitoring Setup](#monitoring-setup)
8. [Backup and Recovery](#backup-and-recovery)
9. [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **CPU**: 2+ cores recommended
- **RAM**: 4GB minimum, 8GB recommended
- **Disk**: 20GB minimum for application and logs
- **OS**: Linux (RHEL/CentOS 8+, Ubuntu 20.04+), macOS, Windows with WSL2

### Software Requirements

Choose one container runtime:

**Option 1: Docker**
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

**Option 2: Podman (Recommended for RHEL/CentOS)**
```bash
# RHEL/CentOS
sudo dnf install -y podman podman-compose

# Ubuntu
sudo apt-get install -y podman podman-compose
```

## Quick Start

### 1. Clone and Navigate

```bash
cd ipquorum-management-platform
```

### 2. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration
vi .env
```

**Minimum required changes:**
```bash
# Generate secure JWT secret
JWT_SECRET=$(openssl rand -base64 32)

# Set secure Grafana password
GRAFANA_ADMIN_PASSWORD=your-secure-password
```

### 3. Deploy

**Interactive Mode:**
```bash
./deploy-prod.sh
```

**Command Line Mode:**
```bash
# Basic deployment (API + Web)
./deploy-prod.sh start

# With monitoring stack
./deploy-prod.sh start monitoring
```

### 4. Access Services

- **Web Dashboard**: http://localhost:3000
- **API Server**: http://localhost:8080
- **Prometheus**: http://localhost:9090 (if monitoring enabled)
- **Grafana**: http://localhost:3001 (if monitoring enabled)

**Default Credentials:**
- Username: `admin`
- Password: `admin123`

⚠️ **Change the default password immediately after first login!**

## Configuration

### Environment Variables

Edit `.env` file to customize deployment:

#### Core Settings

```bash
# Security
JWT_SECRET=your-secure-random-string
JWT_EXPIRATION=24h

# Logging
LOG_LEVEL=info          # debug, info, warn, error
LOG_FORMAT=json         # json, text

# Performance
MAX_CONCURRENT_OPERATIONS=10
OPERATION_TIMEOUT=300s
```

#### Port Configuration

```bash
# Service ports
WEB_PORT=3000
PROMETHEUS_PORT=9090
GRAFANA_PORT=3001
ALERTMANAGER_PORT=9093
```

#### Nginx Tuning

```bash
NGINX_WORKER_PROCESSES=auto
NGINX_WORKER_CONNECTIONS=1024
```

## IPQuorum Agent Integration

The platform can integrate with an IPQuorum Agent for advanced instance management.

### Agent Configuration

Edit `.env` file:

```bash
# Enable agent integration
IPQUORUM_AGENT_ENABLED=true

# Agent connection details
IPQUORUM_AGENT_HOST=10.0.0.100
IPQUORUM_AGENT_PORT=9090
IPQUORUM_AGENT_API_KEY=your-agent-api-key

# TLS settings
IPQUORUM_AGENT_TLS_ENABLED=true
IPQUORUM_AGENT_TLS_VERIFY=true
```

### Agent Setup

1. **Deploy IPQuorum Agent** on target host:
   ```bash
   # On the agent host
   systemctl start ipquorum-agent
   systemctl enable ipquorum-agent
   ```

2. **Generate API Key** on agent:
   ```bash
   ipquorum-agent generate-key
   ```

3. **Configure Management Platform** with agent details in `.env`

4. **Restart Services**:
   ```bash
   ./deploy-prod.sh restart
   ```

### Verify Agent Connection

```bash
# Check logs
./deploy-prod.sh logs ipquorum-server

# Test API endpoint
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/v1/agent/status
```

## Deployment Options

### Standalone Deployment (No Agent)

For testing or standalone operation without agent:

```bash
# In .env
IPQUORUM_AGENT_ENABLED=false
```

Deploy:
```bash
./deploy-prod.sh start
```

### With Monitoring Stack

Deploy with Prometheus, Grafana, and AlertManager:

```bash
./deploy-prod.sh start monitoring
```

Access monitoring:
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001
- **AlertManager**: http://localhost:9093

### High Availability Setup

For production HA deployment:

1. **Use external database** (PostgreSQL instead of SQLite)
2. **Deploy multiple API instances** behind load balancer
3. **Use shared storage** for persistent data
4. **Configure backup** and disaster recovery

Example with external PostgreSQL:

```yaml
# docker-compose.prod.yml (custom)
services:
  ipquorum-server:
    environment:
      - DB_TYPE=postgres
      - DB_HOST=postgres-server
      - DB_PORT=5432
      - DB_NAME=ipquorum
      - DB_USER=ipquorum
      - DB_PASSWORD=${DB_PASSWORD}
```

## Security Hardening

### 1. Generate Secure Secrets

```bash
# JWT Secret (32 bytes)
openssl rand -base64 32

# Grafana Password (16 bytes)
openssl rand -base64 16

# Agent API Key (32 bytes)
openssl rand -hex 32
```

### 2. Enable TLS/HTTPS

**Option A: Reverse Proxy (Recommended)**

Use nginx or Traefik as reverse proxy with Let's Encrypt:

```nginx
# /etc/nginx/sites-available/ipquorum
server {
    listen 443 ssl http2;
    server_name ipquorum.example.com;
    
    ssl_certificate /etc/letsencrypt/live/ipquorum.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/ipquorum.example.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Option B: Direct TLS in Container**

Mount certificates:
```yaml
volumes:
  - /etc/letsencrypt:/etc/letsencrypt:ro
```

### 3. Firewall Configuration

```bash
# Allow only necessary ports
sudo firewall-cmd --permanent --add-port=3000/tcp  # Web
sudo firewall-cmd --permanent --add-port=8080/tcp  # API (if exposed)
sudo firewall-cmd --reload
```

### 4. File Permissions

```bash
# Secure .env file
chmod 600 .env

# Secure data directory
chmod 700 data/
```

### 5. Network Isolation

Use Docker/Podman networks to isolate services:

```yaml
networks:
  frontend:
    driver: bridge
  backend:
    driver: bridge
    internal: true  # No external access
```

## Monitoring Setup

### Grafana Configuration

1. **Access Grafana**: http://localhost:3001
2. **Login** with admin credentials from `.env`
3. **Import Dashboard**:
   - Go to Dashboards → Import
   - Upload `grafana/dashboards/ipquorum-overview.json`
   - Select Prometheus datasource

### Alert Configuration

Alerts are pre-configured in `prometheus/alerts/ipquorum-alerts.yml`:

- **Critical**: Instance down, high error rate, database issues
- **Warning**: High latency, resource usage, failed operations
- **Info**: System events, configuration changes

### Email Notifications

Configure AlertManager for email alerts:

```yaml
# prometheus/alertmanager.yml
global:
  smtp_smarthost: 'smtp.example.com:587'
  smtp_from: 'alerts@example.com'
  smtp_auth_username: 'alerts@example.com'
  smtp_auth_password: 'password'

route:
  receiver: 'email-notifications'

receivers:
  - name: 'email-notifications'
    email_configs:
      - to: 'ops-team@example.com'
```

## Backup and Recovery

### Database Backup

```bash
# Backup SQLite database
docker exec ipquorum-server sqlite3 /data/ipquorum.db ".backup '/data/backup-$(date +%Y%m%d).db'"

# Copy to host
docker cp ipquorum-server:/data/backup-$(date +%Y%m%d).db ./backups/
```

### Automated Backup Script

```bash
#!/bin/bash
# backup-ipquorum.sh

BACKUP_DIR="/backups/ipquorum"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p "$BACKUP_DIR"

# Backup database
docker exec ipquorum-server sqlite3 /data/ipquorum.db ".backup '/data/backup.db'"
docker cp ipquorum-server:/data/backup.db "$BACKUP_DIR/ipquorum-$DATE.db"

# Backup configuration
cp .env "$BACKUP_DIR/env-$DATE.bak"

# Cleanup old backups (keep 30 days)
find "$BACKUP_DIR" -name "*.db" -mtime +30 -delete

echo "Backup completed: $BACKUP_DIR/ipquorum-$DATE.db"
```

Schedule with cron:
```bash
# Daily backup at 2 AM
0 2 * * * /path/to/backup-ipquorum.sh
```

### Restore from Backup

```bash
# Stop services
./deploy-prod.sh stop

# Restore database
cp backups/ipquorum-20240127.db data/ipquorum.db

# Start services
./deploy-prod.sh start
```

## Troubleshooting

### Service Won't Start

**Check logs:**
```bash
./deploy-prod.sh logs ipquorum-server
./deploy-prod.sh logs ipquorum-web
```

**Common issues:**
- Port already in use: Change ports in `.env`
- Permission denied: Check file permissions on `data/` directory
- Database locked: Stop all services and restart

### API Not Responding

```bash
# Check health endpoint
curl http://localhost:8080/health

# Check container status
./deploy-prod.sh status

# Restart API server
docker restart ipquorum-server
```

### Web Dashboard Not Loading

```bash
# Check nginx logs
docker logs ipquorum-web

# Verify API connectivity
docker exec ipquorum-web curl http://ipquorum-server:8080/health

# Check network
docker network inspect ipquorum-network
```

### Database Issues

```bash
# Check database integrity
docker exec ipquorum-server sqlite3 /data/ipquorum.db "PRAGMA integrity_check;"

# Vacuum database
docker exec ipquorum-server sqlite3 /data/ipquorum.db "VACUUM;"
```

### Agent Connection Failed

```bash
# Test agent connectivity
curl -k https://${IPQUORUM_AGENT_HOST}:${IPQUORUM_AGENT_PORT}/health

# Check TLS certificate
openssl s_client -connect ${IPQUORUM_AGENT_HOST}:${IPQUORUM_AGENT_PORT}

# Verify API key
echo $IPQUORUM_AGENT_API_KEY
```

### Performance Issues

**Check resource usage:**
```bash
docker stats ipquorum-server ipquorum-web
```

**Optimize:**
- Increase `MAX_CONCURRENT_OPERATIONS` in `.env`
- Add more CPU/RAM to containers
- Enable database connection pooling
- Use external PostgreSQL for better performance

### View All Logs

```bash
# All services
./deploy-prod.sh logs

# Specific service
./deploy-prod.sh logs ipquorum-server

# Follow logs in real-time
docker logs -f ipquorum-server
```

## Maintenance

### Update Platform

```bash
# Pull latest changes
git pull origin main

# Rebuild images
./deploy-prod.sh stop
./deploy-prod.sh start
```

### Clean Up

```bash
# Remove containers and volumes
./deploy-prod.sh clean

# Remove unused images
docker system prune -a
```

### Health Checks

```bash
# Run health check
./deploy-prod.sh health

# Check all endpoints
curl http://localhost:8080/health
curl http://localhost:8080/metrics
curl http://localhost:3000
```

## Production Checklist

Before going to production:

- [ ] Change default admin password
- [ ] Generate secure JWT secret
- [ ] Configure TLS/HTTPS
- [ ] Set up firewall rules
- [ ] Configure backup automation
- [ ] Set up monitoring alerts
- [ ] Configure email notifications
- [ ] Test disaster recovery
- [ ] Document custom configuration
- [ ] Set up log rotation
- [ ] Configure agent integration (if needed)
- [ ] Test all critical workflows
- [ ] Review security settings

## Support

For issues and questions:
- Check logs: `./deploy-prod.sh logs`
- Review documentation in `docs/` directory
- Check GitHub issues
- Contact support team

#