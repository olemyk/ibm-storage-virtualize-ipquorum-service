#!/usr/bin/env bash
set -euo pipefail

# IPQuorum Runner Image Build Script
# Builds the container image with pre-built ipquorum-download-go binary

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Configuration
IMAGE_NAME="${IMAGE_NAME:-ipquorum-runner}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
REGISTRY="${REGISTRY:-localhost}"
PLATFORM="${PLATFORM:-linux/amd64}"
IPQUORUM_DOWNLOADER_VERSION="${IPQUORUM_DOWNLOADER_VERSION:-v3.1.1}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Parse arguments
PUSH_IMAGE=false
NO_CACHE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --push)
            PUSH_IMAGE=true
            shift
            ;;
        --no-cache)
            NO_CACHE=true
            shift
            ;;
        --tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        --registry)
            REGISTRY="$2"
            shift 2
            ;;
        --platform)
            PLATFORM="$2"
            shift 2
            ;;
        --downloader-version)
            IPQUORUM_DOWNLOADER_VERSION="$2"
            shift 2
            ;;
        --help|-h)
            cat <<EOF
Usage: $0 [OPTIONS]

Build IPQuorum Runner container image with pre-built binary.

Options:
  --push              Push image to registry after build
  --no-cache          Build without using cache
  --tag TAG                  Image tag (default: latest)
  --registry REG             Registry prefix (default: localhost)
  --platform PLAT            Target platform (default: linux/amd64)
  --downloader-version VER   ipquorum-download-go version (default: v3.1.1)
  --help, -h                 Show this help message

Examples:
  $0                                           # Build with defaults
  $0 --tag v1.0.0                              # Build with specific tag
  $0 --downloader-version v3.0.9               # Use specific downloader version
  $0 --registry docker.io/myuser               # Build for Docker Hub
  $0 --push --tag v1.0.0                       # Build and push
  $0 --no-cache                                # Force rebuild

Environment Variables:
  IMAGE_NAME                   Image name (default: ipquorum-runner)
  IMAGE_TAG                    Image tag (default: latest)
  REGISTRY                     Registry prefix (default: localhost)
  PLATFORM                     Target platform (default: linux/amd64)
  IPQUORUM_DOWNLOADER_VERSION  ipquorum-download-go version (default: v3.1.1)

Binary Source:
  https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases
EOF
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"

print_info "=========================================="
print_info "Building IPQuorum Runner Image"
print_info "=========================================="
print_info "Image: ${FULL_IMAGE}"
print_info "Platform: ${PLATFORM}"
print_info "Downloader Version: ${IPQUORUM_DOWNLOADER_VERSION}"
print_info "No Cache: ${NO_CACHE}"
print_info "Push: ${PUSH_IMAGE}"
print_info "=========================================="

# Check if Dockerfile exists
if [[ ! -f "Dockerfile" ]]; then
    print_error "Dockerfile not found in current directory"
    exit 1
fi

# Check if entrypoint.sh exists
if [[ ! -f "entrypoint.sh" ]]; then
    print_error "entrypoint.sh not found in current directory"
    exit 1
fi

print_success "All required files found"

# Detect container runtime
if command -v podman &> /dev/null; then
    RUNTIME="podman"
elif command -v docker &> /dev/null; then
    RUNTIME="docker"
else
    print_error "Neither podman nor docker found"
    exit 1
fi

print_info "Using container runtime: ${RUNTIME}"

# Build image
print_info "Building image (this may take a few minutes)..."

BUILD_ARGS=(
    --tag "${FULL_IMAGE}"
    --platform "${PLATFORM}"
    --build-arg "IPQUORUM_DOWNLOADER_VERSION=${IPQUORUM_DOWNLOADER_VERSION}"
)

if [[ "${NO_CACHE}" == "true" ]]; then
    BUILD_ARGS+=(--no-cache)
fi

# Build
if ${RUNTIME} build "${BUILD_ARGS[@]}" .; then
    print_success "Image built successfully"
else
    print_error "Build failed"
    exit 1
fi

# Get image size
IMAGE_SIZE=$(${RUNTIME} images "${FULL_IMAGE}" --format "{{.Size}}" 2>/dev/null || echo "unknown")
print_info "Image size: ${IMAGE_SIZE}"

# Push if requested
if [[ "${PUSH_IMAGE}" == "true" ]]; then
    print_info "Pushing image to registry..."
    
    if ${RUNTIME} push "${FULL_IMAGE}"; then
        print_success "Image pushed successfully"
    else
        print_error "Push failed"
        exit 1
    fi
fi

# Display summary
echo ""
print_success "=========================================="
print_success "Build Complete!"
print_success "=========================================="
echo ""
echo "Image: ${FULL_IMAGE}"
echo "Size:  ${IMAGE_SIZE}"
echo "Downloader Version: ${IPQUORUM_DOWNLOADER_VERSION}"
echo ""
echo "Test the image:"
echo "  ${RUNTIME} run --rm ${FULL_IMAGE} --help"
echo ""
echo "Run with environment variables:"
cat <<'EOF'
  podman run -d \
    --name ipquorum-test \
    -e API_ENDPOINT=10.33.7.80 \
    -e VIRTUALIZE_USERNAME=superuser \
    -e VIRTUALIZE_PASSWORD=password \
    -e PARTNERSYSTEM=cluster02 \
    -e CREATE_QUORUM_APP=false \
    -e DOWNLOAD_JAR=true \
    -p 1260:1260 \
    localhost/ipquorum-runner:latest
EOF
echo ""
echo "Check logs:"
echo "  ${RUNTIME} logs -f ipquorum-test"
echo ""
echo "Stop container:"
echo "  ${RUNTIME} stop ipquorum-test"
echo ""

# Made with Bob
