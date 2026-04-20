#!/usr/bin/env bash
#
# SELinux Diagnostic Script for IP Quorum
#

set -euo pipefail

INSTANCE_NAME="${1:-svc_cluster01}"
PASSWORD_FILE="/var/lib/ipquorum/.passwords/${INSTANCE_NAME}.password"

echo "=== SELinux Diagnostic for Instance: ${INSTANCE_NAME} ==="
echo ""

# Check if SELinux is enabled
echo "1. SELinux Status:"
if command -v getenforce &>/dev/null; then
    getenforce
    sestatus 2>/dev/null || echo "  (sestatus not available)"
else
    echo "  SELinux tools not installed"
fi
echo ""

# Check file context
echo "2. Password File Context:"
ls -laZ "${PASSWORD_FILE}" 2>/dev/null || echo "  File not found: ${PASSWORD_FILE}"
echo ""

# Check parent directory context
echo "3. Parent Directory Context:"
ls -ldZ /var/lib/ipquorum/.passwords/ 2>/dev/null || echo "  Directory not found"
echo ""

# Check if ipquorum user can read
echo "4. Access Test (as ipquorum user):"
if sudo -u ipquorum cat "${PASSWORD_FILE}" &>/dev/null; then
    echo "  ✓ SUCCESS: ipquorum user CAN read the file"
else
    echo "  ✗ FAILED: ipquorum user CANNOT read the file"
fi
echo ""

# Check AVC denials
echo "5. Recent SELinux Denials (last 10):"
if command -v ausearch &>/dev/null; then
    sudo ausearch -m avc -ts recent 2>/dev/null | grep -i ipquorum | tail -10 || echo "  No recent denials found"
else
    echo "  ausearch not available"
fi
echo ""

# Check file permissions
echo "6. File Permissions:"
stat "${PASSWORD_FILE}" 2>/dev/null || echo "  File not found"
echo ""

# Suggest fixes
echo "=== Suggested Fixes ==="
echo ""
echo "Option 1: Set correct SELinux context (RECOMMENDED):"
echo "  sudo semanage fcontext -a -t etc_t '${PASSWORD_FILE}'"
echo "  sudo restorecon -v '${PASSWORD_FILE}'"
echo ""
echo "Option 3: Create SELinux policy module (advanced):"
echo "  sudo ausearch -m avc -ts recent | audit2allow -M ipquorum_password"
echo "  sudo semodule -i ipquorum_password.pp"
echo ""
echo "Option 4: Temporarily set SELinux to permissive (testing only):"
echo "  sudo setenforce 0"
echo "  # Test if it works, then re-enable: sudo setenforce 1"
echo ""

# 
