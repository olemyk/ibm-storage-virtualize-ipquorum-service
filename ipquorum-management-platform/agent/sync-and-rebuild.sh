#!/usr/bin/env bash
set -euo pipefail

# Sync code to RHEL server and rebuild agent
# This ensures the server has the latest code with updateInstanceConfig

RHEL_HOST="${RHEL_HOST:-packer@10.33.3.215}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Syncing Agent Code to RHEL Server ==="
echo "Target: ${RHEL_HOST}"
echo ""

# Step 1: Sync agent code to server
echo "Step 1: Syncing agent source code..."
ssh "${RHEL_HOST}" "mkdir -p ~/ibm-storage-virtualize-ipquorum-service/ipquorum-management-platform"

rsync -av --delete \
  "${SCRIPT_DIR}/" \
  "${RHEL_HOST}:~/ibm-storage-virtualize-ipquorum-service/ipquorum-management-platform/agent/"

echo "✓ Code synced"
echo ""

# Step 2: Rebuild on server
echo "Step 2: Rebuilding agent on server..."
ssh "${RHEL_HOST}" "cd ~/ibm-storage-virtualize-ipquorum-service/ipquorum-management-platform/agent && \
                     go build -o ipquorum-agent \
                       -ldflags \"-X main.version=1.0.2 -X main.commit=\$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')\" \
                       ./cmd/agent"

echo "✓ Agent rebuilt"
echo ""

# Step 3: Stop service and deploy
echo "Step 3: Deploying new binary..."
ssh "${RHEL_HOST}" "sudo systemctl stop ipquorum-agent && \
                     sudo cp ~/ibm-storage-virtualize-ipquorum-service/ipquorum-management-platform/agent/ipquorum-agent /usr/local/bin/ && \
                     sudo chmod +x /usr/local/bin/ipquorum-agent"

echo "✓ Binary deployed"
echo ""

# Step 4: Start service
echo "Step 4: Starting agent service..."
ssh "${RHEL_HOST}" "sudo systemctl start ipquorum-agent && sleep 2 && sudo systemctl status ipquorum-agent --no-pager -l"

echo "✓ Service started"
echo ""

# Step 5: Verify
echo "Step 5: Verifying deployment..."
echo ""
echo "Checking for updateInstanceConfig function:"
ssh "${RHEL_HOST}" "strings /usr/local/bin/ipquorum-agent | grep -q updateInstanceConfig && echo '✓ Function found in binary' || echo '✗ Function NOT found'"
echo ""
echo "Agent version:"
ssh "${RHEL_HOST}" "curl -s http://localhost:9090/health | jq -r '.version // \"unknown\"'"
echo ""

echo "=== Deployment Complete ==="
echo ""
echo "Test with:"
echo "  curl -X POST -H 'X-API-Key:YOUR_KEY' -H 'Content-Type: application/json' \\"
echo "    -d '{\"name\":\"test_fixed\",\"api_endpoint\":\"10.33.7.80\",\"username\":\"superuser\",\"password\":\"password\"}' \\"
echo "    http://localhost:9090/instances"
echo ""
echo "  sudo cat /etc/ipquorum/instances/test_fixed.conf | grep API_ENDPOINT"


