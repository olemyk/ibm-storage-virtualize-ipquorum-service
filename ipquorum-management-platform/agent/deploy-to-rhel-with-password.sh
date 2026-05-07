#!/usr/bin/env bash
set -euo pipefail

# Deployment script for IPQuorum Agent to RHEL server with password support
# This script:
# 1. Builds the agent binary for Linux
# 2. Copies updated bash script to server
# 3. Deploys agent binary
# 4. Restarts agent service

RHEL_HOST="${RHEL_HOST:-packer@10.33.3.215}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "=== IPQuorum Agent Deployment ==="
echo "Target: ${RHEL_HOST}"
echo ""

# Prompt for sudo password once
read -sp "Enter sudo password for ${RHEL_HOST}: " SUDO_PASS
echo ""
echo ""

# Step 1: Build agent binary for Linux
echo "Step 1: Building agent binary for Linux..."
cd "${SCRIPT_DIR}"
if [[ ! -f "build-for-linux.sh" ]]; then
    echo "Error: build-for-linux.sh not found"
    exit 1
fi

./build-for-linux.sh
if [[ ! -f "build/ipquorum-agent-linux-amd64" ]]; then
    echo "Error: Build failed - binary not found"
    exit 1
fi
echo "✓ Agent binary built successfully"
echo ""

# Step 2: Copy updated bash script
echo "Step 2: Copying updated bash script..."
BASH_SCRIPT="${PROJECT_ROOT}/ipquorum-systemd/multi-instance/ipquorum-instance-manager.sh"
if [[ ! -f "${BASH_SCRIPT}" ]]; then
    echo "Error: Bash script not found at ${BASH_SCRIPT}"
    exit 1
fi

scp "${BASH_SCRIPT}" "${RHEL_HOST}:/tmp/ipquorum-instance-manager.sh"
ssh -t "${RHEL_HOST}" "echo '${SUDO_PASS}' | sudo -S cp /tmp/ipquorum-instance-manager.sh /usr/local/bin/ipquorum-instance-manager.sh && \
                     echo '${SUDO_PASS}' | sudo -S chmod +x /usr/local/bin/ipquorum-instance-manager.sh && \
                     rm /tmp/ipquorum-instance-manager.sh"
echo "✓ Bash script deployed"
echo ""

# Step 3: Deploy agent binary
echo "Step 3: Deploying agent binary..."
scp "build/ipquorum-agent-linux-amd64" "${RHEL_HOST}:/tmp/ipquorum-agent"
ssh -t "${RHEL_HOST}" "echo '${SUDO_PASS}' | sudo -S systemctl stop ipquorum-agent && \
                     echo '${SUDO_PASS}' | sudo -S cp /tmp/ipquorum-agent /usr/local/bin/ipquorum-agent && \
                     echo '${SUDO_PASS}' | sudo -S chmod +x /usr/local/bin/ipquorum-agent && \
                     rm /tmp/ipquorum-agent"
echo "✓ Agent binary deployed"
echo ""

# Step 4: Restart agent service
echo "Step 4: Restarting agent service..."
ssh -t "${RHEL_HOST}" "echo '${SUDO_PASS}' | sudo -S systemctl start ipquorum-agent && \
                     echo '${SUDO_PASS}' | sudo -S systemctl status ipquorum-agent --no-pager -l"
echo "✓ Agent service restarted"
echo ""

# Step 5: Verify deployment
echo "Step 5: Verifying deployment..."
echo ""
echo "Agent health check:"
ssh "${RHEL_HOST}" "curl -s http://localhost:8444/health | jq ."
echo ""

echo "=== Deployment Complete ==="
echo ""
echo "The agent now has /api/v1 prefix for all routes."
echo "Test the logs endpoint:"
echo "  curl -H 'X-API-Key:ea17f6738a9db12f157265dc833154615f05c8c294269b4ced57d20684e6d4d7' \\"
echo "    http://10.33.3.215:8444/api/v1/instances/testgui8/logs?lines=50"
