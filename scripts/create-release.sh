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

log_info "Creating release package for version ${VERSION}"

# Clean up any existing release directory
if [ -d "${RELEASE_DIR}" ]; then
    log_warn "Removing existing release directory"
    rm -rf "${RELEASE_DIR}"
fi

# Create release directory structure
log_info "Creating directory structure"
mkdir -p "${RELEASE_DIR}/bin"
mkdir -p "${RELEASE_DIR}/systemd"
mkdir -p "${RELEASE_DIR}/docs"

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
    OUTPUT="../${RELEASE_DIR}/bin/ipquorum-download-go-${GOOS}-${GOARCH}"
    
    log_info "Building for ${GOOS}/${GOARCH}"
    
    GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=0 go build \
        -ldflags "-s -w -X main.version=${VERSION} -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
        -o "${OUTPUT}" \
        main.go
    
    chmod +x "${OUTPUT}"
    log_info "✓ Built ${OUTPUT}"
done

cd ..

# Copy systemd service files
log_info "Copying systemd service files"
cp ipquorum-systemd/ibm-virtualize-ipquorum-improved.service "${RELEASE_DIR}/systemd/"
cp ipquorum-systemd/ipquorum.conf "${RELEASE_DIR}/systemd/"
cp ipquorum-systemd/ipquorum-download.sh "${RELEASE_DIR}/systemd/"
cp ipquorum-systemd/ipquorum-start.sh "${RELEASE_DIR}/systemd/"
cp ipquorum-systemd/install-ipquorum-service.sh "${RELEASE_DIR}/"

# Make scripts executable
chmod +x "${RELEASE_DIR}/install-ipquorum-service.sh"
chmod +x "${RELEASE_DIR}/systemd/"*.sh

# Copy documentation
log_info "Copying documentation"
cp README.md "${RELEASE_DIR}/" 2>/dev/null || log_warn "README.md not found"
cp ipquorum-systemd/README-IMPROVED-SERVICE.md "${RELEASE_DIR}/docs/"
cp ipquorum-systemd/ARCHITECTURE.md "${RELEASE_DIR}/docs/"
cp ipquorum-systemd/CHANGELOG.md "${RELEASE_DIR}/docs/"
cp ipquorum-systemd/LOGO.md "${RELEASE_DIR}/docs/"
cp DEPLOYMENT-GUIDE.md "${RELEASE_DIR}/docs/" 2>/dev/null || log_warn "DEPLOYMENT-GUIDE.md not found"

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
IBM Storage Virtualize IP Quorum Service - Installation Instructions

QUICK START
===========

1. Run the installation script:
   sudo ./install-ipquorum-service.sh

2. Configure the service:
   sudo vi /etc/ipquorum/ipquorum.conf

3. Set the password:
   echo 'your_password' | sudo tee /etc/ipquorum/.password > /dev/null
   sudo chmod 440 /etc/ipquorum/.password
   sudo chown root:ipquorum /etc/ipquorum/.password

4. Start the service:
   sudo systemctl start ibm-virtualize-ipquorum
   sudo systemctl enable ibm-virtualize-ipquorum

5. Check status:
   sudo systemctl status ibm-virtualize-ipquorum

DOCUMENTATION
=============

- Installation Guide: docs/README-IMPROVED-SERVICE.md
- Architecture: docs/ARCHITECTURE.md
- Deployment Guide: docs/DEPLOYMENT-GUIDE.md
- Changelog: docs/CHANGELOG.md

BINARIES
========

The bin/ directory contains pre-built binaries for:
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
sha256sum "${RELEASE_DIR}/bin/"* >> checksums.txt

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

# Made with help from Bob
