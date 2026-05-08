#!/bin/bash
set -e

echo "=== Rebuilding Web Container with Fixed Nginx Config ==="

cd "$(dirname "$0")"

# Stop the web container
echo "Stopping web container..."
podman stop ipquorum-web || true

# Remove the old container
echo "Removing old web container..."
podman rm ipquorum-web || true

# Build new web image locally
echo "Building new web image..."
podman build -f web/Containerfile -t ipquorum-web:local ./web

# Start the web container with the new image
echo "Starting web container with new image..."
podman run -d \
  --name ipquorum-web \
  --network ipquorum-network \
  -p 3000:80 \
  -e VITE_API_URL=http://localhost:3000 \
  ipquorum-web:local

echo ""
echo "=== Web Container Rebuilt Successfully ==="
echo ""
echo "The web dashboard should now work correctly at: http://localhost:3000"
echo "Login with: admin / admin123"
echo ""
echo "To view logs: podman logs -f ipquorum-web"

# Made with Bob
