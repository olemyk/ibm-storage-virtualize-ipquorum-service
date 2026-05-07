#!/bin/bash
# Manual fix script for instance configuration
# Usage: sudo ./fix-instance-config.sh <instance_name> <api_endpoint> <username> <password>

set -euo pipefail

if [[ $# -ne 4 ]]; then
    echo "Usage: $0 <instance_name> <api_endpoint> <username> <password>"
    echo "Example: $0 test_instance03 10.33.7.80 superuser mypassword"
    exit 1
fi

INSTANCE_NAME="$1"
API_ENDPOINT="$2"
USERNAME="$3"
PASSWORD="$4"

CONFIG_FILE="/etc/ipquorum/instances/${INSTANCE_NAME}.conf"
PASSWORD_DIR="/var/lib/ipquorum/.passwords"
PASSWORD_FILE="${PASSWORD_DIR}/${INSTANCE_NAME}.pass"

echo "Fixing configuration for instance: ${INSTANCE_NAME}"

# Check if config file exists
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "Error: Config file not found: $CONFIG_FILE"
    exit 1
fi

# Update config file
echo "Updating config file..."
sed -i "s|API_ENDPOINT=<API_ENDPOINT>|API_ENDPOINT=\"${API_ENDPOINT}\"|g" "$CONFIG_FILE"
sed -i "s|VIRTUALIZE_USERNAME=<USERNAME>|VIRTUALIZE_USERNAME=\"${USERNAME}\"|g" "$CONFIG_FILE"
sed -i "s|DOWNLOADIPQ=\"no\"|DOWNLOADIPQ=\"yes\"|g" "$CONFIG_FILE"
sed -i "s|IBM_STORAGE_SYSTEM=\"<HOSTNAME_OR_IP>\"|IBM_STORAGE_SYSTEM=\"N/A\"|g" "$CONFIG_FILE"

# Create password directory if it doesn't exist
if [[ ! -d "$PASSWORD_DIR" ]]; then
    echo "Creating password directory..."
    mkdir -p "$PASSWORD_DIR"
    chown ipquorum:ipquorum "$PASSWORD_DIR"
    chmod 700 "$PASSWORD_DIR"
fi

# Write password to file
echo "Creating password file..."
echo "$PASSWORD" > "$PASSWORD_FILE"
chmod 400 "$PASSWORD_FILE"
chown ipquorum:ipquorum "$PASSWORD_FILE"

# Update config to use password file
sed -i "s|PASSWORD_FILE=\"\"|PASSWORD_FILE=\"${PASSWORD_FILE}\"|g" "$CONFIG_FILE"

echo "Configuration updated successfully!"
echo ""
echo "Config file: $CONFIG_FILE"
echo "Password file: $PASSWORD_FILE"
echo ""
echo "You can now start the instance:"
echo "  sudo systemctl start ipquorum@${INSTANCE_NAME}"
echo "  sudo systemctl status ipquorum@${INSTANCE_NAME}"


