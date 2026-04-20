#!/usr/bin/env bash
#
# SELinux Fix Script for IP Quorum Multi-Instance
# This script fixes SELinux contexts for password files
#

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}ERROR: This script must be run as root${NC}" >&2
   echo "Usage: sudo $0 <instance-name>" >&2
   exit 1
fi

INSTANCE_NAME="${1:-}"

if [[ -z "$INSTANCE_NAME" ]]; then
    echo -e "${RED}ERROR: Instance name is required${NC}"
    echo "Usage: sudo $0 <instance-name>"
    echo ""
    echo "Example: sudo $0 svc_cluster01"
    exit 1
fi

echo -e "${BLUE}=== SELinux Fix for Instance: ${INSTANCE_NAME} ===${NC}"
echo ""

# Check if SELinux is enabled
if ! command -v getenforce &>/dev/null; then
    echo -e "${YELLOW}SELinux tools not installed. Skipping SELinux fixes.${NC}"
    exit 0
fi

SELINUX_STATUS=$(getenforce 2>/dev/null || echo "Disabled")
if [[ "$SELINUX_STATUS" == "Disabled" ]]; then
    echo -e "${YELLOW}SELinux is disabled. No fixes needed.${NC}"
    exit 0
fi

echo "SELinux Status: ${SELINUX_STATUS}"
echo ""

OLD_PASSWORD_FILE="/etc/ipquorum/instances/.passwords/${INSTANCE_NAME}.password"
PASSWORD_FILE="/var/lib/ipquorum/.passwords/${INSTANCE_NAME}.password"

# Check if password file exists in old or new location
if [[ ! -f "$PASSWORD_FILE" ]] && [[ ! -f "$OLD_PASSWORD_FILE" ]]; then
    echo -e "${RED}ERROR: Password file not found in either location:${NC}"
    echo "  ${OLD_PASSWORD_FILE}"
    echo "  ${PASSWORD_FILE}"
    exit 1
fi

echo "=== Ensuring password file is in /var/lib (BEST PRACTICE) ==="
echo ""
echo "Password files in /var/lib/ipquorum have correct SELinux context..."

# Create password directory in /var/lib
mkdir -p /var/lib/ipquorum/.passwords
chown ipquorum:ipquorum /var/lib/ipquorum/.passwords
chmod 700 /var/lib/ipquorum/.passwords

# Move/copy password file if needed
if [[ -f "$OLD_PASSWORD_FILE" ]] && [[ ! -f "$PASSWORD_FILE" ]]; then
    echo "Moving password file from /etc to /var/lib..."
    cp "$OLD_PASSWORD_FILE" "$PASSWORD_FILE"
    echo -e "${GREEN}✓ Password file copied to: ${PASSWORD_FILE}${NC}"
elif [[ -f "$PASSWORD_FILE" ]]; then
    echo -e "${GREEN}✓ Password file already in correct location: ${PASSWORD_FILE}${NC}"
fi

# Set correct ownership and permissions
chown ipquorum:ipquorum "$PASSWORD_FILE"
chmod 440 "$PASSWORD_FILE"
echo -e "${GREEN}✓ Ownership and permissions set${NC}"

# Fix SELinux context
restorecon -v "$PASSWORD_FILE"
restorecon -Rv /var/lib/ipquorum/.passwords/

echo ""
echo "Testing access..."
if sudo -u ipquorum cat "$PASSWORD_FILE" &>/dev/null; then
    echo -e "${GREEN}✓ SUCCESS: ipquorum user CAN read the password file${NC}"
    echo ""
    echo -e "${GREEN}=== Fix Complete ===${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Verify config uses correct password file location:"
    echo "     grep PASSWORD_FILE /etc/ipquorum/instances/${INSTANCE_NAME}.conf"
    echo "     (should show: /var/lib/ipquorum/.passwords/${INSTANCE_NAME}.password)"
    echo ""
    echo "  2. Restart the service:"
    echo "     sudo systemctl daemon-reload"
    echo "     sudo systemctl restart ipquorum@${INSTANCE_NAME}.service"
    echo ""
    echo "  3. Verify service status:"
    echo "     sudo systemctl status ipquorum@${INSTANCE_NAME}.service"
    echo ""
    if [[ -f "$OLD_PASSWORD_FILE" ]]; then
        echo "  4. Remove old password file (after verifying service works):"
        echo "     sudo rm ${OLD_PASSWORD_FILE}"
        echo ""
    fi
else
    echo -e "${RED}✗ FAILED: Still cannot read file${NC}"
    echo ""
    echo "=== Alternative Solution: Create SELinux Policy ==="
    echo ""
    echo "Run these commands to create a custom SELinux policy:"
    echo ""
    echo "  # Check for denials"
    echo "  sudo ausearch -m avc -ts recent | grep ipquorum"
    echo ""
    echo "  # Generate policy module"
    echo "  sudo ausearch -m avc -ts recent | audit2allow -M ipquorum_password"
    echo ""
    echo "  # Install policy module"
    echo "  sudo semodule -i ipquorum_password.pp"
    echo ""
    echo "  # Test again"
    echo "  sudo -u ipquorum cat ${NEW_PASSWORD_FILE}"
    echo ""
fi

# Show current contexts
echo ""
echo "=== Current SELinux Contexts ==="
ls -laZ "$PASSWORD_FILE" 2>/dev/null || true
ls -ldZ /var/lib/ipquorum/.passwords/ 2>/dev/null || true
echo ""

# 
