#!/bin/bash
# Deploy config fix to RHEL server

set -euo pipefail

REMOTE_HOST="${1:-packer@10.33.3.215}"
REMOTE_DIR="/home/packer/agent"

echo "========================================="
echo "Deploying Config Fix to RHEL Server"
echo "========================================="
echo ""

# Step 1: Sync the new binary
echo "Step 1: Syncing new agent binary..."
rsync -avz --progress ipquorum-agent-linux "${REMOTE_HOST}:${REMOTE_DIR}/"

echo ""
echo "Step 2: Installing and restarting agent on remote server..."
ssh "${REMOTE_HOST}" << 'ENDSSH'
cd /home/packer/agent

# Stop the agent
echo "Stopping agent..."
sudo systemctl stop ipquorum-agent

# Backup old binary
echo "Backing up old binary..."
sudo cp /usr/local/bin/ipquorum-agent /usr/local/bin/ipquorum-agent.backup.$(date +%Y%m%d_%H%M%S) || true

# Install new binary
echo "Installing new binary..."
sudo install -m 755 ipquorum-agent-linux /usr/local/bin/ipquorum-agent

# Verify version
echo "Verifying installation..."
/usr/local/bin/ipquorum-agent --version || echo "Version check not available"

# Start the agent
echo "Starting agent..."
sudo systemctl start ipquorum-agent

# Check status
echo "Checking status..."
sudo systemctl status ipquorum-agent --no-pager

echo ""
echo "✓ Agent deployed and restarted successfully!"
ENDSSH

echo ""
echo "========================================="
echo "Deployment Complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "  1. Test instance creation:"
echo "     curl -X POST http://localhost:9090/instances \\"
echo "       -H \"X-API-Key: YOUR_KEY\" \\"
echo "       -H \"Content-Type: application/json\" \\"
echo "       -d '{\"name\": \"test_config_fix\", \"api_endpoint\": \"10.33.7.80\", \"username\": \"superuser\", \"password\": \"passw0rd\", \"partnersystem\": \"svc_cluster02\"}'"
echo ""
echo "  2. Verify config file:"
echo "     sudo cat /etc/ipquorum/instances/test_config_fix.conf | grep -E \"API_ENDPOINT|VIRTUALIZE_USERNAME\""
echo ""
echo "  3. Check logs:"
echo "     sudo journalctl -u ipquorum-agent -f"


