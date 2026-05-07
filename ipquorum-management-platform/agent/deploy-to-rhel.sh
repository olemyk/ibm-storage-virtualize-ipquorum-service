#!/usr/bin/env bash
set -euo pipefail

# Deployment script for IPQuorum Agent to RHEL server
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
ssh "${RHEL_HOST}" "sudo cp /tmp/ipquorum-instance-manager.sh /usr/local/bin/ipquorum-instance-manager.sh && \
                     sudo chmod +x /usr/local/bin/ipquorum-instance-manager.sh && \
                     rm /tmp/ipquorum-instance-manager.sh"
echo "✓ Bash script deployed"
echo ""

# Step 3: Deploy agent binary
echo "Step 3: Deploying agent binary..."
scp "build/ipquorum-agent-linux-amd64" "${RHEL_HOST}:/tmp/ipquorum-agent"
ssh "${RHEL_HOST}" "sudo systemctl stop ipquorum-agent && \
                     sudo cp /tmp/ipquorum-agent /usr/local/bin/ipquorum-agent && \
                     sudo chmod +x /usr/local/bin/ipquorum-agent && \
                     rm /tmp/ipquorum-agent"
echo "✓ Agent binary deployed"
echo ""

# Step 4: Restart agent service
echo "Step 4: Restarting agent service..."
ssh "${RHEL_HOST}" "sudo systemctl start ipquorum-agent && \
                     sudo systemctl status ipquorum-agent --no-pager -l"
echo "✓ Agent service restarted"
echo ""

# Step 5: Verify deployment
echo "Step 5: Verifying deployment..."
echo ""
echo "Checking for updateInstanceConfig function:"
ssh "${RHEL_HOST}" "strings /usr/local/bin/ipquorum-agent | grep -q updateInstanceConfig && echo '✓ Function found' || echo '✗ Function NOT found'"
echo ""
echo "Agent version:"
ssh "${RHEL_HOST}" "curl -s http://localhost:9090/health | jq -r '.version // \"unknown\"'"
echo ""

echo "=== Deployment Complete ==="
echo ""
echo "Next steps:"
echo "1. Test instance creation:"
echo "   curl -X POST -H 'X-API-Key:YOUR_KEY' -H 'Content-Type: application/json' \\"
echo "     -d '{\"name\":\"test_new\",\"api_endpoint\":\"10.33.7.80\",\"username\":\"superuser\",\"password\":\"password\"}' \\"
echo "     http://localhost:9090/instances"
echo ""
echo "2. Verify config was updated:"
echo "   sudo cat /etc/ipquorum/instances/test_new.conf | grep API_ENDPOINT"
echo ""
echo "3. Test deletion:"
echo "   curl -X DELETE -H 'X-API-Key:YOUR_KEY' http://localhost:9090/instances/test_new"


