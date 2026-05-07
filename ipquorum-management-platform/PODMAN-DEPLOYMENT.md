# IP Quorum Management Platform - Podman Deployment Guide

> Podman-specific deployment instructions for the IP Quorum Management Platform

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Deployment Methods](#deployment-methods)
- [Configuration](#configuration)
- [Management](#management)
- [Networking](#networking)
- [Storage](#storage)
- [Systemd Integration](#systemd-integration)
- [Rootless Containers](#rootless-containers)
- [Security](#security)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

## Overview

This guide covers Podman-specific deployment scenarios for the IP Quorum Management Platform. Podman is a daemonless container engine that provides a Docker-compatible CLI and is the default container runtime for RHEL 8+.

### Why Podman?

- **Daemonless**: No background daemon required
- **Rootless**: Run containers as non-root user
- **Systemd Integration**: Native systemd service generation
- **Docker Compatible**: Drop-in replacement for Docker CLI
- **Security**: Enhanced security with user namespaces
- **RHEL Default**: Default container runtime in RHEL 8+

### Podman Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Host System                          │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  User Namespace (Rootless)                              │
│  ┌──────────────────────────────────────────────────┐  │
│  │         ipquorum-pod                             │  │
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
│  Volumes (User-owned):                                  │
│  • ~/.local/share/containers/storage/volumes/          │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Prerequisites

### System Requirements

- **OS**: Linux (RHEL 8+, CentOS Stream 8+, Fedora 34+, Ubuntu 20.10+)
- **CPU**: 2+ cores (4+ recommended)
- **RAM**: 4GB minimum (8GB recommended)
- **Disk**: 20GB minimum (50GB recommended)
- **Podman**: 3.0+ (4.0+ recommended)
- **podman-compose**: 1.0+ (optional but recommended)

### Install Podman

#### RHEL/CentOS/Rocky Linux

```bash
# Podman is included by default in RHEL 8+
sudo dnf install -y podman podman-compose

# Verify installation
podman --version
podman-compose --version
```

#### Fedora

```bash
# Install Podman
sudo dnf install -y podman podman-compose

# Verify installation
podman --version
podman-compose --version
```

#### Ubuntu/Debian

```bash
# Add repository
sudo apt-get update
sudo apt-get install -y software-properties-common
sudo add-apt-repository -y ppa:projectatomic/ppa

# Install Podman
sudo apt-get update
sudo apt-get install -y podman

# Install podman-compose (via pip)
sudo apt-get install -y python3-pip
pip3 install podman-compose

# Verify installation
podman --version
podman-compose --version
```

### Post-Installation Configuration

```bash
# Enable user namespaces (if not already enabled)
echo "user.max_user_namespaces=28633" | sudo tee -a /etc/sysctl.d/userns.conf
sudo sysctl -p /etc/sysctl.d/userns.conf

# Configure subuid/subgid for rootless (usually automatic)
grep $USER /etc/subuid /etc/subgid

# If not present, add entries
echo "$USER:100000:65536" | sudo tee -a /etc/subuid
echo "$USER:100000:65536" | sudo tee -a /etc/subgid

# Configure registries
mkdir -p ~/.config/containers
cat > ~/.config/containers/registries.conf << 'EOF'
[registries.search]
registries = ['docker.io', 'quay.io', 'ghcr.io']

[registries.insecure]
registries = []

[registries.block]
registries = []
EOF

# Test Podman
podman run --rm hello-world
```

## Installation

### Quick Start with Automated Script

```bash
# Download and run deployment script
curl -fsSL https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-ipquorum-service/main/ipquorum-management-platform/deploy.sh -o deploy.sh
chmod +x deploy.sh
./deploy.sh --podman
```

### Manual Installation (Rootless)

#### 1. Create Directory Structure

```bash
# Create deployment directory in user space
mkdir -p ~/ipquorum-platform
cd ~/ipquorum-platform

# Create subdirectories
mkdir -p data logs config scripts backups
```

#### 2. Download Configuration

```bash
# Download docker-compose file (compatible with podman-compose)
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
podman pull ghcr.io/olemyk/ipquorum-manager:v3.0.0
podman pull ghcr.io/olemyk/ipquorum-web:v3.0.0

# Verify images
podman images | grep ipquorum
```

#### 5. Start Services

```bash
# Start all services
podman-compose -f docker-compose.prod.yml up -d

# Verify services are running
podman-compose -f docker-compose.prod.yml ps

# Check logs
podman-compose -f docker-compose.prod.yml logs -f
```

## Deployment Methods

### Method 1: Podman Compose (Recommended)

```bash
# Start services
podman-compose -f docker-compose.prod.yml up -d

# Start with monitoring
podman-compose -f docker-compose.prod.yml --profile monitoring up -d

# Stop services
podman-compose -f docker-compose.prod.yml down

# Stop and remove volumes
podman-compose -f docker-compose.prod.yml down -v
```

### Method 2: Podman Run (Manual)

```bash
# Create network
podman network create ipquorum-network

# Create volumes
podman volume create ipquorum-data
podman volume create ipquorum-logs

# Run manager
podman run -d \
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
podman run -d \
  --name ipquorum-web \
  --network ipquorum-network \
  -p 3000:80 \
  --restart unless-stopped \
  ghcr.io/olemyk/ipquorum-web:v3.0.0
```

### Method 3: Podman Pod

```bash
# Create pod with shared network
podman pod create \
  --name ipquorum-pod \
  -p 3000:80 \
  -p 8443:8080

# Run manager in pod
podman run -d \
  --pod ipquorum-pod \
  --name ipquorum-server \
  -v ipquorum-data:/data \
  -e JWT_SECRET="your-secret-key" \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Run web in pod
podman run -d \
  --pod ipquorum-pod \
  --name ipquorum-web \
  ghcr.io/olemyk/ipquorum-web:v3.0.0

# Manage pod
podman pod start ipquorum-pod
podman pod stop ipquorum-pod
podman pod rm ipquorum-pod
```

### Method 4: Kubernetes YAML (Podman Play)

Create `ipquorum-kube.yaml`:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: ipquorum-pod
spec:
  containers:
  - name: ipquorum-server
    image: ghcr.io/olemyk/ipquorum-manager:v3.0.0
    ports:
    - containerPort: 8080
      hostPort: 8443
    env:
    - name: JWT_SECRET
      value: "your-secret-key"
    volumeMounts:
    - name: data
      mountPath: /data
    - name: logs
      mountPath: /var/log/ipquorum
  
  - name: ipquorum-web
    image: ghcr.io/olemyk/ipquorum-web:v3.0.0
    ports:
    - containerPort: 80
      hostPort: 3000
  
  volumes:
  - name: data
    hostPath:
      path: /home/user/ipquorum-platform/data
  - name: logs
    hostPath:
      path: /home/user/ipquorum-platform/logs
```

Deploy:

```bash
# Deploy from YAML
podman play kube ipquorum-kube.yaml

# Stop and remove
podman play kube --down ipquorum-kube.yaml

# Generate YAML from running pod
podman generate kube ipquorum-pod > ipquorum-kube.yaml
```

## Configuration

### Rootless vs Rootful

#### Rootless (Recommended)

```bash
# Run as regular user (no sudo)
podman run -d \
  --name ipquorum-server \
  -p 8443:8080 \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Containers run in user namespace
# Storage: ~/.local/share/containers/storage/
# Config: ~/.config/containers/
```

#### Rootful (When Needed)

```bash
# Run with sudo for privileged operations
sudo podman run -d \
  --name ipquorum-server \
  -p 443:8080 \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Containers run as root
# Storage: /var/lib/containers/storage/
# Config: /etc/containers/
```

### Port Mapping for Rootless

Rootless containers cannot bind to ports < 1024 by default:

```bash
# Option 1: Use high ports (recommended)
podman run -d -p 8443:8080 ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Option 2: Enable low port binding
echo "net.ipv4.ip_unprivileged_port_start=80" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Option 3: Use port forwarding
sudo firewall-cmd --add-forward-port=port=443:proto=tcp:toport=8443
```

## Management

### Container Management

```bash
# List containers
podman ps
podman ps -a  # Include stopped

# Inspect container
podman inspect ipquorum-server

# View logs
podman logs ipquorum-server
podman logs -f ipquorum-server  # Follow

# Execute command
podman exec -it ipquorum-server sh

# Copy files
podman cp config.yaml ipquorum-server:/config/
podman cp ipquorum-server:/data/ipquorum.db ./backup/

# Stats
podman stats
podman stats ipquorum-server

# Stop/start/restart
podman stop ipquorum-server
podman start ipquorum-server
podman restart ipquorum-server

# Remove
podman rm ipquorum-server
podman rm -f ipquorum-server  # Force
```

### Pod Management

```bash
# List pods
podman pod ps

# Inspect pod
podman pod inspect ipquorum-pod

# Start/stop pod
podman pod start ipquorum-pod
podman pod stop ipquorum-pod
podman pod restart ipquorum-pod

# Remove pod
podman pod rm ipquorum-pod
podman pod rm -f ipquorum-pod  # Force

# List containers in pod
podman ps --pod
```

### Image Management

```bash
# List images
podman images

# Pull image
podman pull ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Tag image
podman tag ghcr.io/olemyk/ipquorum-manager:v3.0.0 ipquorum-manager:latest

# Remove image
podman rmi ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Prune unused images
podman image prune
podman image prune -a

# Build image
podman build -t ipquorum-manager:custom -f Containerfile .
```

## Networking

### Network Management

```bash
# List networks
podman network ls

# Inspect network
podman network inspect ipquorum-network

# Create network
podman network create ipquorum-network

# Connect container to network
podman network connect ipquorum-network container-name

# Disconnect
podman network disconnect ipquorum-network container-name

# Remove network
podman network rm ipquorum-network
```

### DNS and Hostname Resolution

```bash
# Containers in same pod can use localhost
podman pod create --name ipquorum-pod -p 3000:80 -p 8443:8080

# Containers in same network can use container names
podman run -d --network ipquorum-network --name server ghcr.io/olemyk/ipquorum-manager:v3.0.0
podman run -d --network ipquorum-network --name web ghcr.io/olemyk/ipquorum-web:v3.0.0
podman exec web curl http://server:8080/health
```

## Storage

### Volume Management

```bash
# List volumes
podman volume ls

# Inspect volume
podman volume inspect ipquorum-data

# Create volume
podman volume create ipquorum-data

# Remove volume
podman volume rm ipquorum-data

# Prune unused volumes
podman volume prune

# Backup volume
podman run --rm \
  -v ipquorum-data:/data:ro \
  -v $(pwd):/backup \
  alpine tar czf /backup/ipquorum-data-backup.tar.gz /data

# Restore volume
podman run --rm \
  -v ipquorum-data:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/ipquorum-data-backup.tar.gz -C /
```

### Volume Location

```bash
# Rootless volumes
~/.local/share/containers/storage/volumes/

# Rootful volumes
/var/lib/containers/storage/volumes/

# List volume location
podman volume inspect ipquorum-data | grep Mountpoint
```

## Systemd Integration

### Generate Systemd Service

```bash
# Generate service file for container
podman generate systemd --new --name ipquorum-server > ~/.config/systemd/user/ipquorum-server.service

# Generate service file for pod
podman generate systemd --new --name ipquorum-pod > ~/.config/systemd/user/ipquorum-pod.service

# Reload systemd
systemctl --user daemon-reload

# Enable and start service
systemctl --user enable ipquorum-server.service
systemctl --user start ipquorum-server.service

# Check status
systemctl --user status ipquorum-server.service

# Enable linger (keep services running after logout)
loginctl enable-linger $USER
```

### Complete Systemd Setup

```bash
# 1. Create pod
podman pod create --name ipquorum-pod -p 3000:80 -p 8443:8080

# 2. Run containers in pod
podman run -d --pod ipquorum-pod --name ipquorum-server \
  -v ipquorum-data:/data \
  -e JWT_SECRET="your-secret" \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

podman run -d --pod ipquorum-pod --name ipquorum-web \
  ghcr.io/olemyk/ipquorum-web:v3.0.0

# 3. Generate systemd files
mkdir -p ~/.config/systemd/user
cd ~/.config/systemd/user
podman generate systemd --new --files --name ipquorum-pod

# 4. Stop and remove pod (systemd will manage it)
podman pod stop ipquorum-pod
podman pod rm ipquorum-pod

# 5. Enable and start services
systemctl --user daemon-reload
systemctl --user enable pod-ipquorum-pod.service
systemctl --user start pod-ipquorum-pod.service

# 6. Enable linger
loginctl enable-linger $USER

# 7. Check status
systemctl --user status pod-ipquorum-pod.service
```

### Systemd Service Management

```bash
# Start service
systemctl --user start ipquorum-server.service

# Stop service
systemctl --user stop ipquorum-server.service

# Restart service
systemctl --user restart ipquorum-server.service

# Enable on boot
systemctl --user enable ipquorum-server.service

# Disable
systemctl --user disable ipquorum-server.service

# View logs
journalctl --user -u ipquorum-server.service
journalctl --user -u ipquorum-server.service -f  # Follow
```

## Rootless Containers

### Benefits

- **Security**: Containers run without root privileges
- **Isolation**: User namespace isolation
- **No Daemon**: No privileged daemon required
- **Multi-user**: Each user has their own containers

### Setup Rootless Environment

```bash
# Check if rootless is configured
podman info | grep rootless

# Configure subuid/subgid
grep $USER /etc/subuid /etc/subgid

# If not present, add (requires root)
echo "$USER:100000:65536" | sudo tee -a /etc/subuid
echo "$USER:100000:65536" | sudo tee -a /etc/subgid

# Verify configuration
podman system migrate
podman info
```

### Rootless Limitations and Solutions

#### Port Binding < 1024

```bash
# Problem: Cannot bind to ports < 1024
# Solution 1: Use high ports
podman run -d -p 8443:8080 ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Solution 2: Enable unprivileged port binding
echo "net.ipv4.ip_unprivileged_port_start=80" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Solution 3: Use firewall port forwarding
sudo firewall-cmd --add-forward-port=port=443:proto=tcp:toport=8443:toaddr=127.0.0.1
```

#### Volume Permissions

```bash
# Rootless containers use user namespace
# Files created in volumes are owned by subuid range

# Check ownership
ls -ln ~/ipquorum-platform/data

# Fix permissions if needed
podman unshare chown -R 0:0 ~/ipquorum-platform/data
```

## Security

### SELinux Configuration

```bash
# Check SELinux status
getenforce

# Label volumes for SELinux
podman run -d \
  -v ~/ipquorum-platform/data:/data:Z \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# :z - shared label (multiple containers)
# :Z - private label (single container)
```

### Security Best Practices

```bash
# 1. Use rootless containers
podman run -d --name ipquorum-server ghcr.io/olemyk/ipquorum-manager:v3.0.0

# 2. Drop capabilities
podman run -d \
  --cap-drop=ALL \
  --cap-add=NET_BIND_SERVICE \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# 3. Read-only root filesystem
podman run -d \
  --read-only \
  -v ipquorum-data:/data \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# 4. No new privileges
podman run -d \
  --security-opt=no-new-privileges \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# 5. Use secrets
podman secret create jwt_secret jwt_secret.txt
podman run -d \
  --secret jwt_secret \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

## Troubleshooting

### Common Issues

#### Rootless Networking Issues

```bash
# Problem: Cannot reach containers
# Solution: Check slirp4netns
podman info | grep -A 5 networkBackend

# Restart networking
podman system reset --force
```

#### Permission Denied Errors

```bash
# Problem: Permission denied accessing volumes
# Solution: Use podman unshare
podman unshare chown -R 0:0 ~/ipquorum-platform/data

# Or use :Z label
podman run -d -v ~/ipquorum-platform/data:/data:Z ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

#### Port Already in Use

```bash
# Check what's using the port
ss -tulpn | grep :8443

# Use different port
podman run -d -p 8444:8080 ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

### Debug Commands

```bash
# System information
podman info
podman version

# Check events
podman events

# Inspect container
podman inspect ipquorum-server

# Check logs
podman logs ipquorum-server
journalctl --user -u ipquorum-server.service

# Network debugging
podman network inspect ipquorum-network
podman exec ipquorum-server ip addr
podman exec ipquorum-server ping -c 3 ipquorum-web
```

## Best Practices

### Production Deployment

1. **Use Rootless** for enhanced security
2. **Enable Systemd** for automatic startup
3. **Use Pods** for related containers
4. **Configure SELinux** properly with :Z labels
5. **Enable Linger** for persistent services
6. **Regular Backups** of volumes
7. **Monitor Resources** with podman stats
8. **Use Secrets** for sensitive data
9. **Pin Image Versions** (avoid :latest)
10. **Test Updates** before production deployment

### Performance Tuning

```bash
# Increase ulimits
podman run -d \
  --ulimit nofile=65536:65536 \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0

# Set resource limits
podman run -d \
  --cpus=2 \
  --memory=4g \
  --memory-swap=4g \
  ghcr.io/olemyk/ipquorum-manager:v3.0.0
```

### Backup Strategy

```bash
#!/bin/bash
# backup-podman.sh

BACKUP_DIR="$HOME/backups/ipquorum"
DATE=$(date +%Y%m%d-%H%M%S)

# Backup volumes
podman run --rm \
  -v ipquorum-data:/data:ro \
  -v $BACKUP_DIR:/backup \
  alpine tar czf /backup/data-$DATE.tar.gz /data

# Backup configuration
tar czf $BACKUP_DIR/config-$DATE.tar.gz \
  ~/.config/systemd/user/ipquorum*.service \
  ~/ipquorum-platform/.env \
  ~/ipquorum-platform/docker-compose.prod.yml

# Cleanup old backups
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete
```

## Next Steps

- [General Deployment Guide](./DEPLOYMENT-GUIDE.md)
- [Docker Deployment](./DOCKER-DEPLOYMENT.md)
- [API Documentation](./API-SPECIFICATION.md)
- [Monitoring Guide](./MONITORING.md)

---

**Last Updated**: 2026-05-07  
**Version**: 3.0.0