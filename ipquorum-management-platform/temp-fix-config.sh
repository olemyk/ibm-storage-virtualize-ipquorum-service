#!/usr/bin/env bash
set -euo pipefail

echo "=== Temporary Config Fix for Current Container Images ==="
echo "This creates a minimal config file until new images are built"
echo ""

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "Error: This script must be run as root (use sudo)" 
   exit 1
fi

DEPLOY_DIR="/opt/ipquorum-platform"

# Create config directory
echo "[1/4] Creating config directory..."
mkdir -p "${DEPLOY_DIR}/config"

# Read JWT_SECRET from .env file
JWT_SECRET=$(grep '^JWT_SECRET=' "${DEPLOY_DIR}/.env" | cut -d'=' -f2 | tr -d '"' || echo "change-this-secret-in-production")

# Create minimal config file
echo "[2/4] Creating config.yaml..."
cat > "${DEPLOY_DIR}/config/config.yaml" <<EOF
server:
  port: 8080
  host: "0.0.0.0"
  log_level: "info"
  health_check_interval: 30

database:
  type: "sqlite"
  path: "/data/ipquorum.db"

auth:
  jwt_secret: "${JWT_SECRET}"
  token_expiry: 3600
  refresh_expiry: 604800

agent:
  port: 8444
  scripts_dir: "/app/scripts"
  health_interval: 30
  metrics_interval: 60
EOF

echo "[3/4] Setting permissions..."
chmod 644 "${DEPLOY_DIR}/config/config.yaml"

# Update docker-compose to add config volume mount
echo "[4/4] Updating docker-compose.prod.yml..."
cd "${DEPLOY_DIR}"

# Check if config volume is already mounted
if grep -q "/etc/ipquorum-platform/config.yaml" docker-compose.prod.yml; then
    echo "Config volume already exists in docker-compose.prod.yml"
else
    # Add config volume mount to ipquorum-server service
    # This is a simple sed command to add the volume after the volumes: line
    sed -i '/ipquorum-server:/,/volumes:/{
        /volumes:/a\      - ./config/config.yaml:/etc/ipquorum-platform/config.yaml:ro
    }' docker-compose.prod.yml
    echo "Added config volume mount to docker-compose.prod.yml"
fi

echo ""
echo "✅ Config file created successfully!"
echo ""
echo "Config file location: ${DEPLOY_DIR}/config/config.yaml"
echo ""
echo "Next steps:"
echo "1. Restart the containers:"
echo "   cd ${DEPLOY_DIR}"
echo "   podman-compose -f docker-compose.prod.yml down"
echo "   podman-compose -f docker-compose.prod.yml up -d"
echo ""
echo "2. Check logs:"
echo "   podman logs ipquorum-server"
echo ""
echo "Note: This is a temporary fix. Once new container images are built"
echo "      (with commit sha-97e5a18), they won't need this config file."

# Made with Bob
