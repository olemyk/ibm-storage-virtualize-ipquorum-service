#!/bin/bash
# Open firewall ports for IPQuorum Management Platform

set -e

echo "Opening firewall ports for IPQuorum Management Platform..."

# Check if firewalld is running
if ! systemctl is-active --quiet firewalld; then
    echo "firewalld is not running. Skipping firewall configuration."
    exit 0
fi

# Open ports
echo "Opening port 3000/tcp (HTTP)..."
sudo firewall-cmd --permanent --add-port=3000/tcp

echo "Opening port 3443/tcp (HTTPS)..."
sudo firewall-cmd --permanent --add-port=3443/tcp

echo "Opening port 8443/tcp (API)..."
sudo firewall-cmd --permanent --add-port=8443/tcp

# Reload firewall
echo "Reloading firewall..."
sudo firewall-cmd --reload

# List open ports
echo ""
echo "Firewall configuration complete. Open ports:"
sudo firewall-cmd --list-ports

echo ""
echo "You can now access the platform:"
echo "  Web Dashboard (HTTPS): https://$(hostname -I | awk '{print $1}'):3443"
echo "  Web Dashboard (HTTP):  http://$(hostname -I | awk '{print $1}'):3000"
echo "  API Server:            https://$(hostname -I | awk '{print $1}'):8443"

# Made with Bob
