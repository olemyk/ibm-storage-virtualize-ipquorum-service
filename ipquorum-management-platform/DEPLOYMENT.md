# IPQuorum Management Platform - Deployment Guide

## Table of Contents
- [Quick Start](#quick-start)
- [Docker Compose Deployment](#docker-compose-deployment)
- [Service Architecture](#service-architecture)
- [Configuration](#configuration)
- [Monitoring Stack](#monitoring-stack)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites
- Docker or Podman installed
- Docker Compose or Podman Compose
- 2GB+ RAM available
- Ports 3000, 8080 available (9090, 3001 for monitoring)

### Basic Deployment (Backend + Frontend)

```bash
# Clone the repository
cd ipquorum-management-platform

# Start the services
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f
```

Access the services:
- **Web Dashboard**: http://localhost:3000
- **API Server**: http://localhost:8080
- **API Documentation**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics

### With Monitoring Stack

```bash
# Start all services including Prometheus and Grafana
docker-compose --profile monitoring up -d

# Access monitoring tools
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3001 (admin/admin)
```

## Docker Compose Deployment

### Service Overview

The `docker-compose.yml` defines the following services:

1. **ipquorum-server** (Backend API)
   - Port: 8080
   - Database: SQLite (/data/ipquorum.db)
   - Scripts: /app/scripts

2. **ipquorum-web** (Frontend Dashboard)
   - Port: 3000 (mapped to container port 80)
   - Nginx serving React app
   - Proxies API requests to backend

3. **prometheus** (Metrics Collection) - Optional
   - Port: 9090
   - Profile: monitoring
   - Scrapes metrics from backend

4. **grafana** (Metrics Visualization) - Optional
   - Port: 3001
   - Profile: monitoring
   - Default credentials: admin/admin

### Service Commands

```bash
# Start all services
docker-compose up -d

# Start only core services (no monitoring)
docker-compose up -d ipquorum-server ipquorum-web

# Start with monitoring
docker-compose --profile monitoring up -d

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Restart a specific service
docker-compose restart ipquorum-server

# View logs
docker-compose logs -f ipquorum-server
docker-compose logs -f ipquorum-web

# Scale services (not applicable for this setup)
# docker-compose up -d --scale ipquorum-server=2

# Rebuild containers
docker-compose build
docker-compose up -d --build
```

## Service Architecture

### Network Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Network                        │
│                  (ipquorum-network)                      │
│                                                          │
│  ┌──────────────┐      ┌──────────────┐                │
│  │              │      │              │                │
│  │  Frontend    │─────▶│   Backend    │                │
│  │  (Nginx)     │      │   (Go API)   │                │
│  │  Port: 3000  │      │   Port: 8080 │                │
│  │              │      │              │                │
│  └──────────────┘      └──────┬───────┘                │
│         │                     │                         │
│         │                     │                         │
│         │              ┌──────▼───────┐                │
│         │              │              │                │
│         │              │  Prometheus  │                │
│         │              │  Port: 9090  │                │
│         │              │              │                │
│         │              └──────┬───────┘                │
│         │                     │                         │
│         │              ┌──────▼───────┐                │
│         │              │              │                │
│         └─────────────▶│   Grafana    │                │
│                        │  Port: 3001  │                │
│                        │              │                │
│                        └──────────────┘                │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### Request Flow

1. **User → Frontend (Port 3000)**
   - User accesses web dashboard
   - Nginx serves React SPA

2. **Frontend → Backend (via Nginx proxy)**
   - API requests to `/api/*` are proxied to `ipquorum-server:8080`
   - Health checks to `/health` are proxied
   - Metrics to `/metrics` are proxied

3. **Backend → Database**
   - SQLite database at `/data/ipquorum.db`
   - Persistent volume mount

4. **Prometheus → Backend**
   - Scrapes metrics from `/metrics` endpoint
   - 10-second interval

5. **Grafana → Prometheus**
   - Queries metrics data
   - Visualizes dashboards

## Configuration

### Environment Variables

#### Backend (ipquorum-server)

```yaml
environment:
  - DB_PATH=/data/ipquorum.db          # Database file path
  - JWT_SECRET=your-secret-key         # JWT signing key (CHANGE IN PRODUCTION!)
  - LOG_LEVEL=info                     # Log level: debug, info, warn, error
  - SCRIPTS_DIR=/app/scripts           # Directory for bash scripts
  - PORT=8080                          # API server port (optional)
```

#### Frontend (ipquorum-web)

No environment variables needed. Configuration is baked into the build.

#### Prometheus

Configuration in `prometheus.yml`:
```yaml
scrape_interval: 15s      # How often to scrape metrics
scrape_timeout: 5s        # Timeout for scrape requests
```

#### Grafana

```yaml
environment:
  - GF_SECURITY_ADMIN_USER=admin       # Admin username
  - GF_SECURITY_ADMIN_PASSWORD=admin   # Admin password (CHANGE IN PRODUCTION!)
  - GF_USERS_ALLOW_SIGN_UP=false       # Disable user registration
```

### Volume Mounts

```yaml
volumes:
  # Backend data persistence
  - ./data:/data                       # Database and logs
  - ./scripts:/app/scripts:ro          # Bash scripts (read-only)
  
  # Prometheus data
  - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
  - prometheus-data:/prometheus        # Named volume for metrics data
  
  # Grafana data
  - grafana-data:/var/lib/grafana      # Named volume for dashboards/config
```

### Port Mappings

| Service | Container Port | Host Port | Description |
|---------|---------------|-----------|-------------|
| ipquorum-server | 8080 | 8080 | Backend API |
| ipquorum-web | 80 | 3000 | Frontend Dashboard |
| prometheus | 9090 | 9090 | Prometheus UI |
| grafana | 3000 | 3001 | Grafana UI |

## Monitoring Stack

### Enabling Monitoring

The monitoring stack (Prometheus + Grafana) is optional and uses Docker Compose profiles:

```bash
# Start with monitoring
docker-compose --profile monitoring up -d

# Stop monitoring services
docker-compose --profile monitoring down
```

### Prometheus Setup

1. **Access Prometheus**: http://localhost:9090
2. **Query Metrics**: Use PromQL to query metrics
3. **Example Queries**:
   ```promql
   # HTTP request rate
   rate(http_requests_total[5m])
   
   # Active instances
   ipquorum_instances_total{status="running"}
   
   # API latency
   histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
   ```

### Grafana Setup

1. **Access Grafana**: http://localhost:3001
2. **Login**: admin/admin (change on first login)
3. **Add Prometheus Data Source**:
   - URL: `http://prometheus:9090`
   - Access: Server (default)
   - Save & Test

4. **Import Dashboards**:
   - Create custom dashboards
   - Import community dashboards
   - Use provided dashboard templates (coming soon)

### Available Metrics

See [METRICS.md](./METRICS.md) for complete list of available metrics.

Key metrics:
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `ipquorum_instances_total` - Instance count by status
- `ipquorum_health_checks_total` - Health check results
- `ipquorum_script_executions_total` - Script execution count
- `ipquorum_auth_operations_total` - Authentication operations

## Production Deployment

### Security Checklist

- [ ] Change JWT_SECRET to a strong random value
- [ ] Change Grafana admin password
- [ ] Use HTTPS/TLS for all services
- [ ] Restrict network access (firewall rules)
- [ ] Use secrets management (Docker secrets, Vault)
- [ ] Enable authentication on Prometheus/Grafana
- [ ] Regular security updates
- [ ] Backup database regularly

### Production Configuration

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  ipquorum-server:
    environment:
      - JWT_SECRET=${JWT_SECRET}  # From environment or .env file
      - LOG_LEVEL=warn
    restart: always
    
  ipquorum-web:
    restart: always
    
  prometheus:
    restart: always
    
  grafana:
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    restart: always
```

### Using Environment Files

Create `.env` file:
```bash
JWT_SECRET=your-very-long-random-secret-key-here
GRAFANA_PASSWORD=your-secure-grafana-password
DB_PATH=/data/ipquorum.db
LOG_LEVEL=info
```

Start with environment file:
```bash
docker-compose --env-file .env up -d
```

### Backup and Restore

#### Backup Database

```bash
# Backup SQLite database
docker-compose exec ipquorum-server sqlite3 /data/ipquorum.db ".backup /data/backup.db"

# Copy backup to host
docker cp ipquorum-server:/data/backup.db ./backups/ipquorum-$(date +%Y%m%d).db
```

#### Restore Database

```bash
# Stop services
docker-compose down

# Restore database
cp ./backups/ipquorum-20260426.db ./data/ipquorum.db

# Start services
docker-compose up -d
```

### Health Checks

All services have health checks configured:

```bash
# Check service health
docker-compose ps

# View health check logs
docker inspect ipquorum-server | jq '.[0].State.Health'
```

## Troubleshooting

### Common Issues

#### 1. Port Already in Use

```bash
# Error: port 8080 already allocated
# Solution: Change port mapping in docker-compose.yml
ports:
  - "8081:8080"  # Use different host port
```

#### 2. Database Permission Issues

```bash
# Error: unable to open database file
# Solution: Fix permissions
sudo chown -R 1000:1000 ./data
chmod 755 ./data
```

#### 3. Frontend Can't Connect to Backend

```bash
# Check if backend is running
docker-compose logs ipquorum-server

# Check network connectivity
docker-compose exec ipquorum-web wget -O- http://ipquorum-server:8080/health
```

#### 4. Prometheus Not Scraping Metrics

```bash
# Check Prometheus targets
# Visit: http://localhost:9090/targets

# Check backend metrics endpoint
curl http://localhost:8080/metrics
```

### Debugging Commands

```bash
# View all logs
docker-compose logs

# Follow logs for specific service
docker-compose logs -f ipquorum-server

# Execute command in container
docker-compose exec ipquorum-server sh

# Check container resource usage
docker stats

# Inspect container
docker inspect ipquorum-server

# Check network
docker network inspect ipquorum-management-platform_ipquorum-network
```

### Performance Tuning

#### Backend

```yaml
# Increase resources
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
    reservations:
      cpus: '1'
      memory: 1G
```

#### Database

```bash
# Optimize SQLite
docker-compose exec ipquorum-server sqlite3 /data/ipquorum.db "VACUUM;"
docker-compose exec ipquorum-server sqlite3 /data/ipquorum.db "ANALYZE;"
```

### Logs and Monitoring

```bash
# View logs with timestamps
docker-compose logs -f --timestamps

# Export logs
docker-compose logs > logs-$(date +%Y%m%d).txt

# Monitor resource usage
docker stats --no-stream

# Check disk usage
docker system df
```

## Next Steps

1. **Configure Authentication**: Create admin user via API
2. **Add Instances**: Register IP Quorum instances
3. **Setup Monitoring**: Configure Grafana dashboards
4. **Enable Alerts**: Configure alerting rules
5. **Backup Strategy**: Implement regular backups
6. **Security Hardening**: Follow security checklist

## Support

- **Documentation**: See [README.md](./README.md)
- **API Reference**: See [API-SPECIFICATION.md](./API-SPECIFICATION.md)
- **Metrics**: See [METRICS.md](./METRICS.md)
- **Issues**: Report issues on GitHub

## License

See [LICENSE](../LICENSE) file for details.