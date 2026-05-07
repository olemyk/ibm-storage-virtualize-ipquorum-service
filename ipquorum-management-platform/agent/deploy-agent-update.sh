#!/bin/bash
# Automated Agent Update Deployment Script
# Version: 1.0.5

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_USER="${1:-root}"
SERVER_HOST="${2}"
AGENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Functions
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_usage() {
    cat <<EOF
Usage: $0 [user] <server-host>

Arguments:
  user         SSH user (default: root)
  server-host  Target server IP or hostname (required)

Examples:
  $0 10.33.7.80                    # Deploy as root
  $0 admin 10.33.7.80              # Deploy as admin user
  $0 root server.example.com       # Deploy to hostname

Environment Variables:
  SSH_KEY      Path to SSH private key (optional)

EOF
}

# Validate arguments
if [[ -z "$SERVER_HOST" ]]; then
    print_error "Server host is required"
    print_usage
    exit 1
fi

# SSH options
SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
if [[ -n "$SSH_KEY" ]]; then
    SSH_OPTS="$SSH_OPTS -i $SSH_KEY"
fi

print_info "=== Agent Update Deployment ==="
print_info "Target: ${SERVER_USER}@${SERVER_HOST}"
print_info "Agent Directory: ${AGENT_DIR}"
echo ""

# Step 1: Build binary
print_info "Step 1/6: Building agent binary..."
cd "$AGENT_DIR"
if ! make build; then
    print_error "Build failed"
    exit 1
fi
print_info "✓ Build successful"
echo ""

# Step 2: Verify binary exists
if [[ ! -f "build/ipquorum-agent" ]]; then
    print_error "Binary not found at build/ipquorum-agent"
    exit 1
fi
print_info "✓ Binary verified: build/ipquorum-agent"
echo ""

# Step 3: Copy binary to server
print_info "Step 2/6: Copying binary to server..."
if ! scp $SSH_OPTS build/ipquorum-agent ${SERVER_USER}@${SERVER_HOST}:/tmp/; then
    print_error "Failed to copy binary to server"
    exit 1
fi
print_info "✓ Binary copied to /tmp/ipquorum-agent"
echo ""

# Step 4: Deploy on server
print_info "Step 3/6: Deploying on server..."
ssh $SSH_OPTS ${SERVER_USER}@${SERVER_HOST} << 'ENDSSH'
set -e

echo "[INFO] Stopping agent service..."
if systemctl is-active --quiet ipquorum-agent; then
    systemctl stop ipquorum-agent
    echo "[INFO] ✓ Agent stopped"
else
    echo "[WARN] Agent was not running"
fi

echo "[INFO] Backing up current binary..."
if [[ -f /usr/local/bin/ipquorum-agent ]]; then
    cp /usr/local/bin/ipquorum-agent /usr/local/bin/ipquorum-agent.backup.$(date +%Y%m%d_%H%M%S)
    echo "[INFO] ✓ Backup created"
else
    echo "[WARN] No existing binary to backup"
fi

echo "[INFO] Installing new binary..."
cp /tmp/ipquorum-agent /usr/local/bin/
chmod +x /usr/local/bin/ipquorum-agent
chown root:root /usr/local/bin/ipquorum-agent
echo "[INFO] ✓ Binary installed"

echo "[INFO] Starting agent service..."
systemctl start ipquorum-agent
sleep 2

if systemctl is-active --quiet ipquorum-agent; then
    echo "[INFO] ✓ Agent started successfully"
else
    echo "[ERROR] Agent failed to start"
    exit 1
fi

echo "[INFO] Cleaning up..."
rm -f /tmp/ipquorum-agent
echo "[INFO] ✓ Cleanup complete"
ENDSSH

if [[ $? -ne 0 ]]; then
    print_error "Deployment failed on server"
    exit 1
fi
print_info "✓ Deployment complete"
echo ""

# Step 5: Verify deployment
print_info "Step 4/6: Verifying deployment..."
ssh $SSH_OPTS ${SERVER_USER}@${SERVER_HOST} << 'ENDSSH'
set -e

echo "[INFO] Checking agent status..."
systemctl status ipquorum-agent --no-pager | head -n 10

echo ""
echo "[INFO] Recent logs:"
journalctl -u ipquorum-agent -n 10 --no-pager
ENDSSH

if [[ $? -ne 0 ]]; then
    print_warn "Verification had issues, but deployment may be successful"
else
    print_info "✓ Verification complete"
fi
echo ""

# Step 6: Test health endpoint
print_info "Step 5/6: Testing health endpoint..."
if ssh $SSH_OPTS ${SERVER_USER}@${SERVER_HOST} 'curl -s http://localhost:8081/health' | grep -q "healthy"; then
    print_info "✓ Health endpoint responding"
else
    print_warn "Health endpoint test failed (agent may still be starting)"
fi
echo ""

# Step 7: Summary
print_info "Step 6/6: Deployment Summary"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
print_info "✓ Agent updated successfully to v1.0.5"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Next Steps:"
echo "  1. Monitor logs: ssh ${SERVER_USER}@${SERVER_HOST} 'journalctl -u ipquorum-agent -f'"
echo "  2. Test API: curl http://${SERVER_HOST}:8081/health"
echo "  3. Run tests: ./test-enhanced-api.sh"
echo ""
echo "Rollback (if needed):"
echo "  ssh ${SERVER_USER}@${SERVER_HOST} 'systemctl stop ipquorum-agent && cp /usr/local/bin/ipquorum-agent.backup.* /usr/local/bin/ipquorum-agent && systemctl start ipquorum-agent'"
echo ""
print_info "Deployment complete! 🎉"


