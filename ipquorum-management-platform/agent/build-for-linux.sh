#!/bin/bash
# Build agent for Linux from Mac
set -euo pipefail

echo "Building ipquorum-agent for Linux (amd64)..."

# Set environment for cross-compilation
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

# Get version info
VERSION="${VERSION:-1.0.1}"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build
go build \
  -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
  -o build/ipquorum-agent-linux-amd64 \
  ./cmd/agent

echo "✅ Build complete: build/ipquorum-agent-linux-amd64"
echo ""
echo "To deploy to RHEL server:"
echo "  scp build/ipquorum-agent-linux-amd64 packer@rhel94:/tmp/"
echo "  ssh packer@rhel94"
echo "  sudo systemctl stop ipquorum-agent"
echo "  sudo cp /tmp/ipquorum-agent-linux-amd64 /usr/local/bin/ipquorum-agent"
echo "  sudo chmod +x /usr/local/bin/ipquorum-agent"
echo "  sudo systemctl start ipquorum-agent"
echo "  sudo systemctl status ipquorum-agent"


