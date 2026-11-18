#!/usr/bin/env bash
set -euo pipefail

# Deploy IP Quorum Management Platform to Linux server for testing
# Usage: ./deploy-to-linux.sh [--with-agent]

# Configuration
REMOTE_USER="packer"
REMOTE_HOST="10.33.3.215"
REMOTE_DIR="~/ipquorum-test"
WITH_AGENT=false

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
while [[ $# -gt 0 ]]; do
    case $1 in
        --with-agent)
            WITH_AGENT=true
            shift
            ;;
        --help|-h)
            cat <<EOF
Usage: $0 [OPTIONS]

Deploy IP Quorum Management Platform to Linux server for testing.

Options:
  --with-agent    Include container runtime agent files
  --help, -h      Show this help message

Examples:
  $0                  # Deploy core services only
  $0 --with-agent     # Deploy with agent support

Remote Server: ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}
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

print_info "Deploying to ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}"
echo ""

# Test SSH connection
print_info "Testing SSH connection..."
if ! ssh -o ConnectTimeout=5 "${REMOTE_USER}@${REMOTE_HOST}" "echo 'Connection successful'" >/dev/null 2>&1; then
    print_error "Cannot connect to ${REMOTE_USER}@${REMOTE_HOST}"
    print_info "Please ensure:"
    print_info "  1. SSH key is configured"
    print_info "  2. Server is reachable"
    print_info "  3. User has proper permissions"
    exit 1
fi
print_success "SSH connection successful"

# Cleanup existing deployment
print_info "Cleaning up existing deployment on remote server..."
ssh "${REMOTE_USER}@${REMOTE_HOST}" bash <<'CLEANUP_SCRIPT'
set -e

REMOTE_DIR="~/ipquorum-test"
cd "$REMOTE_DIR" 2>/dev/null || { echo "Directory does not exist, skipping cleanup"; exit 0; }

echo "  → Stopping containers..."
if command -v podman-compose >/dev/null 2>&1; then
    podman-compose -f podman-compose.dev.yml down 2>/dev/null || true
fi

echo "  → Removing ipquorum containers..."
podman ps -a --filter "name=ipquorum" --format "{{.ID}}" | xargs -r podman rm -f 2>/dev/null || true

echo "  → Removing ipquorum images..."
podman images --filter "reference=localhost/ipquorum*" --format "{{.ID}}" | xargs -r podman rmi -f 2>/dev/null || true
podman images --filter "reference=ipquorum*" --format "{{.ID}}" | xargs -r podman rmi -f 2>/dev/null || true

echo "  → Removing ipquorum networks..."
podman network ls --filter "name=ipquorum" --format "{{.Name}}" | xargs -r podman network rm 2>/dev/null || true

echo "  → Cleaning directory contents (keeping deploy script)..."
# Remove everything except the deploy script itself
find "$REMOTE_DIR" -mindepth 1 -maxdepth 1 ! -name 'deploy-to-linux.sh' -exec rm -rf {} + 2>/dev/null || true

echo "  → Cleanup complete"
CLEANUP_SCRIPT

print_success "Remote cleanup complete"

# Create remote directory structure
print_info "Creating remote directory structure..."
ssh "${REMOTE_USER}@${REMOTE_HOST}" "mkdir -p ${REMOTE_DIR}/{server,web,container-runtime-agent,ipquorum-runner,certs,data,logs,scripts}"
print_success "Remote directory created"

# Generate SSL certificates if they don't exist
print_info "Generating SSL certificates for nginx..."
ssh "${REMOTE_USER}@${REMOTE_HOST}" bash <<'CERT_SCRIPT'
set -e
# Expand ~ to actual home directory
REMOTE_DIR="/home/packer/ipquorum-test"
cd "${REMOTE_DIR}"

if [ ! -f certs/server.crt ] || [ ! -f certs/server.key ]; then
    echo "  → Generating self-signed SSL certificate..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout certs/server.key \
        -out certs/server.crt \
        -subj '/CN=localhost' 2>/dev/null
    chmod 644 certs/server.crt certs/server.key
    echo "  → SSL certificates generated"
else
    echo "  → SSL certificates already exist"
fi
CERT_SCRIPT
print_success "SSL certificates ready"

# Create file list
print_info "Preparing file list..."
FILES_TO_COPY=(
    "podman-compose.dev.yml"
    "start-web-dev.sh"
    "clean-data.sh"
    "test-healthchecks.sh"
    ".env.example"
    "PODMAN-COMPOSE-DEV-README.md"
    "WEB-COMPOSE-DEPLOYMENT-SUMMARY.md"
    "PODMAN-HEALTHCHECK-WORKAROUND.md"
    "BUGFIX-SUMMARY.md"
    "CONTAINER-CREATION-WORKFLOW.md"
    "AGENT-ARCHITECTURE-CLARIFICATION.md"
    "SYSTEMD-NOT-SUPPORTED-IN-CONTAINER.md"
)

DIRS_TO_COPY=(
    "server"
    "web"
    "ipquorum-runner"
)

if [ "$WITH_AGENT" = true ]; then
    DIRS_TO_COPY+=("container-runtime-agent")
    print_info "Including agent files"
fi

print_info "Including IPQuorum runner files"

# Copy files
print_info "Copying files to remote server..."
for file in "${FILES_TO_COPY[@]}"; do
    if [ -f "$file" ]; then
        print_info "  Copying $file..."
        scp -q "$file" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"
    else
        print_warning "  Skipping $file (not found)"
    fi
done

# Copy directories
for dir in "${DIRS_TO_COPY[@]}"; do
    if [ -d "$dir" ]; then
        print_info "  Copying $dir/..."
        rsync -az --exclude='node_modules' --exclude='dist' --exclude='build' --exclude='.git' \
            "$dir/" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/$dir/"
    else
        print_warning "  Skipping $dir/ (not found)"
    fi
done

print_success "Files copied successfully"

# Create setup script on remote server
print_info "Creating setup script on remote server..."
ssh "${REMOTE_USER}@${REMOTE_HOST}" "cat > ${REMOTE_DIR}/setup.sh" <<'SETUP_SCRIPT'
#!/usr/bin/env bash
set -euo pipefail

echo "Setting up IP Quorum Management Platform..."

# Make scripts executable
chmod +x start-web-dev.sh clean-data.sh test-healthchecks.sh

# Make IPQuorum runner scripts executable
if [ -d ipquorum-runner ]; then
    chmod +x ipquorum-runner/*.sh 2>/dev/null || true
fi

# Create .env if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cp .env.example .env
    echo "JWT_SECRET=$(openssl rand -hex 32)" >> .env
    # Set Buildah format to docker for HEALTHCHECK support
    echo "BUILDAH_FORMAT=docker" >> .env
fi

# Check for podman
if ! command -v podman &> /dev/null; then
    echo "ERROR: Podman is not installed"
    echo "Please install podman first:"
    echo "  sudo dnf install -y podman podman-compose"
    exit 1
fi

# Check for podman-compose
if ! command -v podman-compose &> /dev/null; then
    echo "ERROR: podman-compose is not installed"
    echo "Please install podman-compose:"
    echo "  sudo dnf install -y podman-compose"
    echo "  # or"
    echo "  pip3 install podman-compose"
    exit 1
fi

# Check for Go (needed for ipquorum-download-go in runner)
if ! command -v go &> /dev/null; then
    echo "WARNING: Go is not installed"
    echo "Go is required to build the IPQuorum runner image"
    echo "Install with: sudo dnf install -y golang"
fi

echo ""
echo "Setup complete!"
echo ""
echo "To build IPQuorum runner image:"
echo "  cd ipquorum-runner && ./build.sh"
echo ""
echo "To start the platform:"
echo "  ./start-web-dev.sh"
echo ""
echo "To start with agent:"
echo "  ./start-web-dev.sh --agent"
echo ""
echo "To start with monitoring:"
echo "  ./start-web-dev.sh --monitoring"
echo ""
echo "To verify health checks after deployment:"
echo "  ./test-healthchecks.sh"
echo ""
echo "For health check troubleshooting, see:"
echo "  PODMAN-HEALTHCHECK-WORKAROUND.md"
echo ""
SETUP_SCRIPT

ssh "${REMOTE_USER}@${REMOTE_HOST}" "chmod +x ${REMOTE_DIR}/setup.sh"
print_success "Setup script created"

# Display next steps
echo ""
print_success "Deployment complete!"
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Next Steps:${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "1. SSH to the server:"
echo "   ssh ${REMOTE_USER}@${REMOTE_HOST}"
echo ""
echo "2. Navigate to the directory:"
echo "   cd ${REMOTE_DIR}"
echo ""
echo "3. Run the setup script:"
echo "   ./setup.sh"
echo ""
echo "4. Start the platform:"
echo "   ./start-web-dev.sh"
echo ""
if [ "$WITH_AGENT" = true ]; then
    echo "5. Start with agent (Linux only):"
    echo "   ./start-web-dev.sh --agent"
    echo ""
fi
echo "Access the platform:"
echo "  Web Dashboard: https://${REMOTE_HOST}:3443"
echo "  API Server:    https://${REMOTE_HOST}:8443"
if [ "$WITH_AGENT" = true ]; then
    echo "  Agent API:     http://${REMOTE_HOST}:8081"
    echo "  Agent Metrics: http://${REMOTE_HOST}:9090"
fi
echo ""

# Made with Bob
