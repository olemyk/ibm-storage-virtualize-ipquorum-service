#!/usr/bin/env bash
#
# Create Release Package Script
# This script creates a release package with all necessary files
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# Check if version is provided
if [ $# -eq 0 ]; then
    log_error "Version number required"
    echo "Usage: $0 <version>"
    echo "Example: $0 2.0.0"
    exit 1
fi

VERSION="$1"
RELEASE_DIR="ipquorum-service-${VERSION}"
TARBALL="ipquorum-service-${VERSION}.tar.gz"

# Detect repository root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

log_info "Repository root: ${REPO_ROOT}"
log_info "Creating release package for version ${VERSION}"

# Change to repository root
cd "${REPO_ROOT}"

# Clean up any existing release directory
if [ -d "${RELEASE_DIR}" ]; then
    log_warn "Removing existing release directory"
    rm -rf "${RELEASE_DIR}"
fi

# Create release directory structure
log_info "Creating directory structure"
mkdir -p "${RELEASE_DIR}/ipquorum-downloader"
mkdir -p "${RELEASE_DIR}/systemd"
mkdir -p "${RELEASE_DIR}/docs"

# Check if Go directory exists
if [ ! -d "ipquorum-download-go" ]; then
    log_error "ipquorum-download-go directory not found"
    log_error "Please run this script from the repository root or ensure the directory exists"
    exit 1
fi

# Build Go binaries for all platforms
log_info "Building Go binaries for all platforms"
cd ipquorum-download-go

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

for platform in "${PLATFORMS[@]}"; do
    GOOS="${platform%/*}"
    GOARCH="${platform#*/}"
    OUTPUT="../${RELEASE_DIR}/ipquorum-downloader/ipquorum-download-go-${GOOS}-${GOARCH}"
    
    log_info "Building for ${GOOS}/${GOARCH}"
    
    GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=0 go build \
        -ldflags "-s -w -X main.version=${VERSION} -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
        -o "${OUTPUT}" \
        main.go
    
    chmod +x "${OUTPUT}"
    log_info "✓ Built ${OUTPUT}"
done

cd ..

# Copy multi-instance system files
log_info "Copying multi-instance system files"
mkdir -p "${RELEASE_DIR}/systemd"
cp ipquorum-systemd/multi-instance/ipquorum@.service "${RELEASE_DIR}/systemd/"
cp ipquorum-systemd/multi-instance/ipquorum-download-multi.sh "${RELEASE_DIR}/systemd/"
cp ipquorum-systemd/multi-instance/ipquorum-start-multi.sh "${RELEASE_DIR}/systemd/"
cp ipquorum-systemd/multi-instance/ipquorum-validate-multi.sh "${RELEASE_DIR}/systemd/"
cp ipquorum-systemd/multi-instance/install-ipquorum-service.sh "${RELEASE_DIR}/"
cp ipquorum-systemd/multi-instance/ipquorum-instance-manager.sh "${RELEASE_DIR}/systemd/"
cp ipquorum-systemd/multi-instance/instance.conf.template "${RELEASE_DIR}/systemd/"

# Copy multi-instance examples
mkdir -p "${RELEASE_DIR}/systemd/examples"
cp ipquorum-systemd/multi-instance/examples/*.conf "${RELEASE_DIR}/systemd/examples/"

# Copy multi-instance troubleshooting tools
mkdir -p "${RELEASE_DIR}/systemd/troubleshoot-tools"
cp ipquorum-systemd/multi-instance/troubleshoot-tools/*.sh "${RELEASE_DIR}/systemd/troubleshoot-tools/"

# Make scripts executable
chmod +x "${RELEASE_DIR}/install-ipquorum-service.sh"
chmod +x "${RELEASE_DIR}/systemd/"*.sh
chmod +x "${RELEASE_DIR}/systemd/troubleshoot-tools/"*.sh

# Copy documentation
log_info "Copying documentation"
cp README.md "${RELEASE_DIR}/" 2>/dev/null || log_warn "README.md not found"
cp ipquorum-systemd/multi-instance/README.md "${RELEASE_DIR}/docs/"
cp ipquorum-systemd/multi-instance/README-ADVANCED.md "${RELEASE_DIR}/docs/"
cp ipquorum-systemd/multi-instance/QUICK-REFERENCE.md "${RELEASE_DIR}/docs/"
cp ipquorum-systemd/multi-instance/IMPLEMENTATION-SUMMARY.md "${RELEASE_DIR}/docs/"
cp ipquorum-systemd/multi-instance/RESOURCE-LIMITS.md "${RELEASE_DIR}/docs/" 2>/dev/null || true

# Create LICENSE file if it doesn't exist
if [ ! -f LICENSE ]; then
    log_warn "LICENSE file not found, creating placeholder"
    cat > "${RELEASE_DIR}/LICENSE" <<EOF
IBM Storage Virtualize IP Quorum Service

See IBM Storage Virtualize documentation for license terms.
EOF
else
    cp LICENSE "${RELEASE_DIR}/"
fi

# Create VERSION file
echo "${VERSION}" > "${RELEASE_DIR}/VERSION"

# Create INSTALL.txt with quick start instructions
log_info "Creating installation instructions"
cat > "${RELEASE_DIR}/INSTALL.txt" <<'EOF'
IBM Storage Virtualize IP Quorum Service - Multi-Instance Installation

QUICK START
===========

Run multiple independent IP Quorum instances on a single host:

1. Run the installation script:
   sudo ./install-ipquorum-service.sh

2. Create your first instance (interactive):
   sudo ./systemd/ipquorum-instance-manager.sh create myinstance

3. Follow the interactive prompts to configure:
   - Description for documentation
   - API endpoint (IP or hostname)
   - Username and password
   - Download settings
   - mkquorumapp configuration (if needed)

4. Enable and start the instance:
   sudo systemctl enable --now ipquorum@myinstance.service

5. Check status:
   sudo systemctl status ipquorum@myinstance.service

6. View logs:
   sudo journalctl -u ipquorum@myinstance.service -f

MANAGING INSTANCES
==================

List all instances:
   sudo systemctl list-units 'ipquorum@*'

Start/stop/restart instance:
   sudo systemctl start ipquorum@myinstance.service
   sudo systemctl stop ipquorum@myinstance.service
   sudo systemctl restart ipquorum@myinstance.service

Using instance manager:
   sudo ./systemd/ipquorum-instance-manager.sh list
   sudo ./systemd/ipquorum-instance-manager.sh info myinstance
   sudo ./systemd/ipquorum-instance-manager.sh validate myinstance
   sudo ./systemd/ipquorum-instance-manager.sh logs myinstance

DOCUMENTATION
=============

- Quick Start Guide: docs/README.md
- Advanced Guide: docs/README-ADVANCED.md
- Quick Reference: docs/QUICK-REFERENCE.md
- Implementation Details: docs/IMPLEMENTATION-SUMMARY.md

BINARIES
========

The ipquorum-downloader/ directory contains pre-built binaries for:
- Linux x86_64 (amd64)
- Linux ARM64
- macOS Intel (amd64)
- macOS Apple Silicon (arm64)

The installation script will automatically select the correct binary for your platform.

SUPPORT
=======

For issues and questions, please refer to the documentation or contact support.
EOF

# Create tarball
log_info "Creating tarball"
tar -czf "${TARBALL}" "${RELEASE_DIR}"

# Generate checksums
log_info "Generating checksums"
sha256sum "${TARBALL}" > checksums.txt
sha256sum "${RELEASE_DIR}/ipquorum-downloader/"* >> checksums.txt

# Display summary
log_info "Release package created successfully!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Release Package: ${TARBALL}"
echo "  Size: $(du -h "${TARBALL}" | cut -f1)"
echo "  Checksums: checksums.txt"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Contents:"
tree -L 2 "${RELEASE_DIR}" 2>/dev/null || find "${RELEASE_DIR}" -type f | head -20
echo ""
echo "To test the package:"
echo "  tar -xzf ${TARBALL}"
echo "  cd ${RELEASE_DIR}"
echo "  sudo ./install-ipquorum-service.sh"
echo ""
echo "To create a GitHub release:"
echo "  git tag v${VERSION}"
echo "  git push origin v${VERSION}"
echo ""


