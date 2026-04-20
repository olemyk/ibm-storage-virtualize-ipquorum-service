#!/usr/bin/env bash
#
# Fix Permissions Script for IP Quorum Multi-Instance
# This script fixes common permission issues
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

echo -e "${BLUE}=== Fixing Permissions for Instance: ${INSTANCE_NAME} ===${NC}"
echo ""

# Check if ipquorum user exists
if ! id "ipquorum" &>/dev/null; then
    echo -e "${RED}ERROR: User 'ipquorum' does not exist${NC}"
    echo "Creating user..."
    useradd -r -s /bin/false -d /var/lib/ipquorum -c "IP Quorum Service User" ipquorum
    echo -e "${GREEN}✓ User 'ipquorum' created${NC}"
fi

# Fix base directories
echo "Fixing base directory permissions..."
mkdir -p /etc/ipquorum/instances
chmod 755 /etc/ipquorum
chmod 755 /etc/ipquorum/instances
echo -e "${GREEN}✓ Configuration directories fixed${NC}"

# Fix password directory in /var/lib (best practice for SELinux)
echo "Fixing password directory..."
mkdir -p /var/lib/ipquorum/.passwords
chown ipquorum:ipquorum /var/lib/ipquorum/.passwords
chmod 700 /var/lib/ipquorum/.passwords
echo -e "${GREEN}✓ Password directory: /var/lib/ipquorum/.passwords${NC}"

# Fix instance data directory
echo "Fixing instance data directory..."
mkdir -p "/var/lib/ipquorum/${INSTANCE_NAME}"
chown -R ipquorum:ipquorum "/var/lib/ipquorum/${INSTANCE_NAME}"
chmod 755 "/var/lib/ipquorum/${INSTANCE_NAME}"
echo -e "${GREEN}✓ Data directory: /var/lib/ipquorum/${INSTANCE_NAME}${NC}"

# Fix instance log directory
echo "Fixing instance log directory..."
mkdir -p "/var/log/ipquorum/${INSTANCE_NAME}"
chown -R ipquorum:ipquorum "/var/log/ipquorum/${INSTANCE_NAME}"
chmod 755 "/var/log/ipquorum/${INSTANCE_NAME}"
echo -e "${GREEN}✓ Log directory: /var/log/ipquorum/${INSTANCE_NAME}${NC}"

# Check password file
PASSWORD_FILE="/var/lib/ipquorum/.passwords/${INSTANCE_NAME}.password"
if [[ ! -f "$PASSWORD_FILE" ]]; then
    echo -e "${YELLOW}⚠ Password file not found: ${PASSWORD_FILE}${NC}"
    echo "Creating empty password file..."
    touch "$PASSWORD_FILE"
    chown ipquorum:ipquorum "$PASSWORD_FILE"
    chmod 440 "$PASSWORD_FILE"
    echo -e "${YELLOW}⚠ Please add your password to: ${PASSWORD_FILE}${NC}"
    echo "  Example: echo 'your_password' | sudo tee ${PASSWORD_FILE}"
else
    echo "Fixing password file permissions..."
    chown ipquorum:ipquorum "$PASSWORD_FILE"
    chmod 440 "$PASSWORD_FILE"
    echo -e "${GREEN}✓ Password file: ${PASSWORD_FILE}${NC}"
fi

# Fix SELinux context if SELinux is enabled
if command -v getenforce &>/dev/null && [[ "$(getenforce 2>/dev/null)" != "Disabled" ]]; then
    echo "Fixing SELinux contexts..."
    chcon -t etc_t "$PASSWORD_FILE" 2>/dev/null || true
    restorecon -v "$PASSWORD_FILE" 2>/dev/null || true
    restorecon -Rv "/var/lib/ipquorum/${INSTANCE_NAME}" 2>/dev/null || true
    restorecon -Rv "/var/log/ipquorum/${INSTANCE_NAME}" 2>/dev/null || true
    echo -e "${GREEN}✓ SELinux contexts fixed${NC}"
fi

# Check configuration file
CONFIG_FILE="/etc/ipquorum/instances/${INSTANCE_NAME}.conf"
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo -e "${YELLOW}⚠ Configuration file not found: ${CONFIG_FILE}${NC}"
else
    chmod 644 "$CONFIG_FILE"
    echo -e "${GREEN}✓ Configuration file: ${CONFIG_FILE}${NC}"
fi

# Fix scripts permissions
echo "Fixing script permissions..."
chmod 755 /usr/local/bin/ipquorum-download-multi.sh 2>/dev/null || true
chmod 755 /usr/local/bin/ipquorum-start-multi.sh 2>/dev/null || true
echo -e "${GREEN}✓ Script permissions fixed${NC}"

echo ""
echo -e "${GREEN}=== Permission Fix Complete ===${NC}"
echo ""
echo "Summary:"
echo "  Instance: ${INSTANCE_NAME}"
echo "  Data Dir: /var/lib/ipquorum/${INSTANCE_NAME} (owner: ipquorum:ipquorum)"
echo "  Log Dir:  /var/log/ipquorum/${INSTANCE_NAME} (owner: ipquorum:ipquorum)"
echo "  Password: ${PASSWORD_FILE} (permissions: 440, owner: ipquorum:ipquorum)"
echo ""

# Check if password file is empty
if [[ -f "$PASSWORD_FILE" ]]; then
    if [[ ! -s "$PASSWORD_FILE" ]]; then
        echo -e "${YELLOW}⚠ WARNING: Password file is empty!${NC}"
        echo "  Add password: echo 'your_password' | sudo tee ${PASSWORD_FILE}"
        echo ""
    fi
fi

echo "Next steps:"
echo "  1. Verify password file has content: sudo cat ${PASSWORD_FILE}"
echo "  2. Reload systemd: sudo systemctl daemon-reload"
echo "  3. Restart instance: sudo systemctl restart ipquorum@${INSTANCE_NAME}.service"
echo "  4. Check status: sudo systemctl status ipquorum@${INSTANCE_NAME}.service"
echo ""

# 
