#!/bin/bash
set -euo pipefail

echo "=== Reset Admin Password Script ==="
echo ""
echo "This script will:"
echo "1. Stop the management server"
echo "2. Delete the database"
echo "3. Restart the server (which will recreate the database with default admin user)"
echo ""
echo "Default credentials after reset:"
echo "  Username: admin"
echo "  Password: changeme"
echo ""
read -p "Continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

echo ""
echo "Step 1: Stopping management server..."
podman-compose -f podman-compose.dev.yml stop ipquorum-server

echo "Step 2: Removing database..."
sudo rm -f data/ipquorum.db

echo "Step 3: Starting management server..."
podman-compose -f podman-compose.dev.yml start ipquorum-server

echo ""
echo "Waiting for server to initialize..."
sleep 5

echo ""
echo "Testing login..."
RESPONSE=$(curl -k -s -X POST https://localhost:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"changeme"}')

if echo "$RESPONSE" | jq -e '.token' >/dev/null 2>&1; then
    echo "✓ Login successful!"
    echo ""
    echo "You can now login with:"
    echo "  Username: admin"
    echo "  Password: changeme"
    echo ""
    echo "⚠️  IMPORTANT: Change the default password immediately!"
else
    echo "✗ Login failed. Response:"
    echo "$RESPONSE" | jq '.'
    echo ""
    echo "Please check the server logs:"
    echo "  podman logs ipquorum-server"
fi

# Made with Bob
