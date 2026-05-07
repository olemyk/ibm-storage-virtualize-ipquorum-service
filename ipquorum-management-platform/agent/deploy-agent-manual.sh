#!/usr/bin/env bash
set -euo pipefail

# Manual deployment script for IPQuorum Agent
# Run this script on your local machine after the binary has been copied to /tmp/

echo "=== IPQuorum Agent Manual Deployment ==="
echo ""
echo "The agent binary has been copied to the RHEL server at:"
echo "  /tmp/ipquorum-agent"
echo ""
echo "Please run the following commands on the RHEL server (packer@10.33.3.215):"
echo ""
echo "# Stop the agent service"
echo "sudo systemctl stop ipquorum-agent"
echo ""
echo "# Replace the agent binary"
echo "sudo cp /tmp/ipquorum-agent /usr/local/bin/ipquorum-agent"
echo "sudo chmod +x /usr/local/bin/ipquorum-agent"
echo ""
echo "# Start the agent service"
echo "sudo systemctl start ipquorum-agent"
echo ""
echo "# Check the status"
echo "sudo systemctl status ipquorum-agent"
echo ""
echo "# Verify the agent is responding"
echo "curl http://localhost:9090/health"
echo ""
echo "=== After deployment ==="
echo "The health checker will automatically detect the updated agent on the next check (every 30 seconds)."
echo "You should see:"
echo "  - Uptime displayed (e.g., '2h 15m')"
echo "  - Last Health Check timestamp updating"
echo "  - started_at populated in the database"


