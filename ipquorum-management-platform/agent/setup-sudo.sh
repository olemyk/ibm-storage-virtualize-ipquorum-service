#!/bin/bash
# Setup sudo permissions for ipquorum user
# This script must be run as root

set -euo pipefail

echo "Setting up sudo permissions for ipquorum user..."

# Create sudoers file for ipquorum user
cat > /etc/sudoers.d/ipquorum << 'EOF'
# Allow ipquorum user to manage IPQuorum instances without password
# Required for config file updates and password file management

# Allow sed for config file updates
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/sed

# Allow mkdir for directory creation
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/mkdir

# Allow bash for password file creation with proper ownership
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/bash -c *

# Allow chown for setting file ownership
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/chown

# Allow chmod for setting file permissions
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/chmod

# Allow systemctl for service management
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl start ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl stop ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl status ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl enable ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl disable ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl is-active ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl is-enabled ipquorum@*

# Allow rm for cleanup operations
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/rm
EOF

# Set proper permissions on sudoers file
chmod 0440 /etc/sudoers.d/ipquorum

# Validate sudoers file
if visudo -c -f /etc/sudoers.d/ipquorum; then
    echo "✓ Sudoers file created and validated successfully"
else
    echo "✗ ERROR: Sudoers file validation failed"
    rm -f /etc/sudoers.d/ipquorum
    exit 1
fi

# Test sudo access
echo "Testing sudo access for ipquorum user..."
if sudo -u ipquorum sudo -n true 2>/dev/null; then
    echo "✓ Passwordless sudo is working for ipquorum user"
else
    echo "✗ WARNING: Passwordless sudo test failed"
    exit 1
fi

echo ""
echo "Setup complete! The ipquorum user now has the following sudo permissions:"
echo "  - sed (for config file updates)"
echo "  - mkdir (for directory creation)"
echo "  - bash -c (for password file creation)"
echo "  - chown/chmod (for file ownership/permissions)"
echo "  - systemctl (for service management)"
echo "  - rm (for cleanup operations)"
echo ""
echo "You can verify with: sudo -l -U ipquorum"


