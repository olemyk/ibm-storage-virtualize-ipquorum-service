#!/bin/bash
# Sync sudo fix files to RHEL server

set -euo pipefail

REMOTE_HOST="${1:-packer@10.33.7.94}"
REMOTE_DIR="/home/packer/agent"

echo "Syncing sudo fix files to ${REMOTE_HOST}:${REMOTE_DIR}..."

# Sync only the new/modified files
rsync -avz --progress \
  setup-sudo.sh \
  SUDO-FIX.md \
  CRITICAL-SUDO-ISSUE.md \
  install-agent.sh \
  "${REMOTE_HOST}:${REMOTE_DIR}/"

echo ""
echo "✓ Files synced successfully!"
echo ""
echo "Next steps on RHEL server:"
echo "  1. ssh ${REMOTE_HOST}"
echo "  2. cd ${REMOTE_DIR}"
echo "  3. chmod +x setup-sudo.sh"
echo "  4. sudo bash setup-sudo.sh"
echo "  5. sudo systemctl restart ipquorum-agent"
echo "  6. Test instance creation"


