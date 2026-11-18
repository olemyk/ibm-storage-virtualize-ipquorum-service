#!/usr/bin/env bash
set -euo pipefail

# IP Quorum Management Platform - Web Development Stack Startup Script
# This script starts the development stack with nginx and HTTPS support

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="podman-compose.dev.yml"
TLS_DIR="certs"
DATA_DIR="data"
LOGS_DIR="logs"
SCRIPTS_DIR="scripts"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}IP Quorum Web Development Stack${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Load environment variables from .env if it exists
if [ -f .env ]; then
    print_info "Loading environment variables from .env"
    set -a
    source .env
    set +a
fi

# Set default BUILDAH_FORMAT if not set
export BUILDAH_FORMAT="${BUILDAH_FORMAT:-docker}"

# Check if podman-compose is installed
if ! command -v podman-compose &> /dev/null; then
    print_error "podman-compose is not installed"
    echo "Install it with: pip3 install podman-compose"
    exit 1
fi

# Create required directories
print_info "Creating required directories..."
mkdir -p "$DATA_DIR" "$LOGS_DIR" "$SCRIPTS_DIR"

# Set ownership to UID 1000 (container appuser) for data directory
# This allows the container to write to the database
if command -v chown >/dev/null 2>&1; then
    chown -R 1000:1000 "$DATA_DIR" 2>/dev/null || {
        print_warning "Could not set ownership on $DATA_DIR (may need sudo)"
        print_info "Run: sudo chown -R 1000:1000 $DATA_DIR"
    }
fi

print_success "Directories created"

# Check and generate TLS certificates if needed
if [ ! -f "$TLS_DIR/server.crt" ] || [ ! -f "$TLS_DIR/server.key" ]; then
    print_warning "TLS certificates not found. Generating self-signed certificates..."
    mkdir -p "$TLS_DIR"
    
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$TLS_DIR/server.key" \
        -out "$TLS_DIR/server.crt" \
        -subj "/C=US/ST=State/L=City/O=IPQuorum/CN=localhost" \
        2>/dev/null
    
    chmod 600 "$TLS_DIR/server.key"
    chmod 644 "$TLS_DIR/server.crt"
    
    print_success "Self-signed certificates generated in $TLS_DIR/"
    print_warning "Browser will show security warnings for self-signed certificates"
else
    print_success "TLS certificates found"
fi

# Parse command line arguments first to set PROFILE
PROFILE=""
DETACH="-d"
BUILD_IMAGES="yes"
REBUILD_IMAGES="no"
CLEAN_DATA="no"
BUILD_IPQUORUM_RUNNER="no"

while [[ $# -gt 0 ]]; do
    case $1 in
        --monitoring)
            if [ -n "$PROFILE" ]; then
                PROFILE="$PROFILE --profile monitoring"
            else
                PROFILE="--profile monitoring"
            fi
            shift
            ;;
        --agent)
            if [ -n "$PROFILE" ]; then
                PROFILE="$PROFILE --profile agent"
            else
                PROFILE="--profile agent"
            fi
            shift
            ;;
        --foreground|-f)
            DETACH=""
            shift
            ;;
        --no-build)
            BUILD_IMAGES="no"
            shift
            ;;
        --rebuild)
            REBUILD_IMAGES="yes"
            BUILD_IMAGES="yes"
            shift
            ;;
        --clean)
            CLEAN_DATA="yes"
            shift
            ;;
        --build-ipquorum-runner)
            BUILD_IPQUORUM_RUNNER="yes"
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --monitoring              Start with monitoring stack (Prometheus + Grafana)"
            echo "  --agent                   Start with container runtime agent (Podman)"
            echo "                            Note: Automatically generates API key if not configured"
            echo "  --foreground, -f          Run in foreground (don't detach)"
            echo "  --no-build                Skip building images (use existing)"
            echo "  --rebuild                 Force rebuild images with --no-cache (for code updates)"
            echo "  --build-ipquorum-runner   Build IPQuorum runner image before starting"
            echo "  --clean                   Clean database and logs before starting"
            echo "  --help, -h                Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                                # Build and start core services"
            echo "  $0 --monitoring                   # Start with monitoring"
            echo "  $0 --agent                        # Start with container agent + API key"
            echo "  $0 --rebuild --agent              # Force rebuild with new code + start agent"
            echo "  $0 --build-ipquorum-runner        # Build IPQuorum runner + start services"
            echo "  $0 --foreground                   # Run in foreground"
            echo "  $0 --no-build                     # Skip build, use existing images"
            echo "  $0 --clean                        # Clean data and start fresh"
            echo "  $0 --monitoring --agent           # Start with monitoring and agent"
            echo ""
            echo "Rebuild vs Build:"
            echo "  --rebuild    Force rebuild with --no-cache (use after code changes)"
            echo "  (default)    Normal build with cache (faster, but may miss code changes)"
            echo "  --no-build   Skip build entirely (use existing images)"
            echo ""
            echo "IPQuorum Runner:"
            echo "  The IPQuorum runner image allows you to run IBM Storage Virtualize"
            echo "  IP Quorum applications in containers. Build it with --build-ipquorum-runner"
            echo "  or manually: cd ipquorum-runner && ./build.sh"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Check and configure firewall if needed
check_firewall() {
    # Only check on Linux systems with firewalld
    if ! command -v firewall-cmd &> /dev/null; then
        return 0
    fi
    
    # Check if firewalld is running
    if ! systemctl is-active --quiet firewalld 2>/dev/null; then
        return 0
    fi
    
    # Check if ports are already open
    local ports_needed=("3000/tcp" "3443/tcp" "8443/tcp")
    
    # Add agent ports if agent profile is enabled
    if [[ "$PROFILE" == *"agent"* ]]; then
        ports_needed+=("8081/tcp" "9090/tcp")
    fi
    
    local ports_missing=()
    
    for port in "${ports_needed[@]}"; do
        if ! sudo firewall-cmd --list-ports 2>/dev/null | grep -q "$port"; then
            ports_missing+=("$port")
        fi
    done
    
    if [ ${#ports_missing[@]} -eq 0 ]; then
        print_success "Firewall ports already configured"
        return 0
    fi
    
    print_warning "Firewall detected - some ports need to be opened: ${ports_missing[*]}"
    echo ""
    echo "The following ports need to be opened for remote access:"
    echo "  - 3000/tcp (HTTP web dashboard)"
    echo "  - 3443/tcp (HTTPS web dashboard)"
    echo "  - 8443/tcp (API server)"
    if [[ "$PROFILE" == *"agent"* ]]; then
        echo "  - 8081/tcp (Container agent API)"
        echo "  - 9090/tcp (Container agent metrics)"
    fi
    echo ""
    
    read -p "Would you like to open these ports now? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Opening firewall ports..."
        
        for port in "${ports_missing[@]}"; do
            print_info "Opening port $port..."
            sudo firewall-cmd --permanent --add-port="$port" || {
                print_error "Failed to open port $port"
                return 1
            }
        done
        
        print_info "Reloading firewall..."
        sudo firewall-cmd --reload || {
            print_error "Failed to reload firewall"
            return 1
        }
        
        print_success "Firewall configured successfully"
        echo ""
        print_info "Open ports:"
        sudo firewall-cmd --list-ports
        echo ""
    else
        print_warning "Skipping firewall configuration"
        print_info "You can configure it later by running: ./open-firewall-ports.sh"
    fi
}

check_firewall

# Function to fix Podman socket permissions for agent
fix_podman_socket_permissions() {
    local socket_path="/run/user/$(id -u)/podman/podman.sock"
    
    # Only run if agent profile is enabled
    if [[ "$PROFILE" != *"agent"* ]]; then
        return 0
    fi
    
    # Check if socket exists
    if [ ! -S "$socket_path" ]; then
        print_warning "Podman socket not found at $socket_path"
        print_info "Make sure Podman is running: systemctl --user start podman.socket"
        return 1
    fi
    
    # Check current permissions
    local current_perms=$(stat -c '%a' "$socket_path" 2>/dev/null || stat -f '%A' "$socket_path" 2>/dev/null)
    
    if [ "$current_perms" = "666" ]; then
        print_success "Podman socket permissions already configured"
        return 0
    fi
    
    print_info "Fixing Podman socket permissions for container agent..."
    chmod 0666 "$socket_path" || {
        print_error "Failed to change socket permissions"
        print_info "Run manually: chmod 0666 $socket_path"
        return 1
    }
    
    print_success "Podman socket permissions fixed (666)"
    print_info "Socket: $socket_path"
}

# Function to update nginx DNS resolver for Podman
fix_nginx_dns_resolver() {
    local nginx_conf="web/nginx.conf"
    
    # Check if nginx.conf exists
    if [ ! -f "$nginx_conf" ]; then
        print_warning "nginx.conf not found at $nginx_conf"
        return 0
    fi
    
    # Check if already using Podman DNS (10.89.0.1)
    if grep -q "resolver 10.89.0.1" "$nginx_conf"; then
        print_success "Nginx DNS resolver already configured for Podman"
        return 0
    fi
    
    # Check if using Docker DNS (127.0.0.11)
    if grep -q "resolver 127.0.0.11" "$nginx_conf"; then
        print_info "Updating nginx DNS resolver for Podman..."
        
        # Backup original
        cp "$nginx_conf" "${nginx_conf}.backup.$(date +%Y%m%d_%H%M%S)"
        
        # Update DNS resolver
        sed -i.tmp 's/resolver 127\.0\.0\.11/resolver 10.89.0.1/g' "$nginx_conf"
        rm -f "${nginx_conf}.tmp"
        
        print_success "Nginx DNS resolver updated to 10.89.0.1 (Podman)"
        print_info "Backup saved: ${nginx_conf}.backup.*"
    fi
}

# Apply fixes before starting services
print_info "Applying deployment fixes..."
fix_nginx_dns_resolver
if [[ "$PROFILE" == *"agent"* ]]; then
    fix_podman_socket_permissions
fi

# Check if .env file exists
if [ ! -f ".env" ]; then
    print_warning ".env file not found. Creating from .env.example..."
    if [ -f ".env.example" ]; then
        cp .env.example .env
        print_success ".env file created"
        print_warning "Please review and update .env file with your settings"
    else
        print_warning "No .env.example found. Using default values"
    fi
fi

# Function to generate and configure API key for agent
setup_agent_api_key() {
    print_info "Checking agent API key configuration..."
    
    # Check if AGENT_API_KEY is already set in .env
    if [ -f ".env" ] && grep -q "^AGENT_API_KEY=.\+" .env; then
        local existing_key=$(grep "^AGENT_API_KEY=" .env | cut -d'=' -f2)
        if [ -n "$existing_key" ]; then
            print_success "Agent API key already configured"
            echo ""
            print_info "Your API Key: $existing_key"
            echo ""
            print_info "Use this key when registering servers in the web UI"
            return 0
        fi
    fi
    
    # Generate new API key
    print_warning "Agent API key not found. Generating new key..."
    local api_key=$(openssl rand -base64 32 | tr -d '\n')
    
    # Update .env file
    if [ -f ".env" ]; then
        # Backup existing .env
        cp ".env" ".env.backup.$(date +%Y%m%d_%H%M%S)"
        
        # Update or add AGENT_API_KEY
        if grep -q "^AGENT_API_KEY=" .env; then
            sed -i.tmp "s|^AGENT_API_KEY=.*|AGENT_API_KEY=$api_key|" .env
            rm -f .env.tmp
        else
            echo "" >> .env
            echo "# Agent API Key (generated $(date))" >> .env
            echo "AGENT_API_KEY=$api_key" >> .env
        fi
        
        print_success "Agent API key generated and saved to .env"
    else
        # Create new .env with API key
        cat > .env <<EOF
# IP Quorum Management Platform - Environment Configuration
# Generated: $(date)

# Agent API Key
AGENT_API_KEY=$api_key

# JWT Secret
JWT_SECRET=$(openssl rand -base64 32 | tr -d '\n')

# Build format
BUILDAH_FORMAT=docker
EOF
        print_success "Created .env with agent API key"
    fi
    
    # Display the API key
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}Agent API Key Generated:${NC}"
    echo ""
    echo "  $api_key"
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    print_warning "IMPORTANT: Save this API key securely!"
    print_info "You will need this key to register servers in the web UI"
    echo ""
    
    # Save to secure note file
    local note_file=".api-key-$(date +%Y%m%d).txt"
    cat > "$note_file" <<EOF
IP Quorum Management Platform - API Key
Generated: $(date)

API Key: $api_key

This key is required for:
- Registering agents in the web UI
- Agent-to-server authentication

Keep this file secure and delete after saving the key to your password manager.
EOF
    chmod 600 "$note_file"
    print_success "API key also saved to: $note_file (chmod 600)"
    echo ""
}

# Command line arguments already parsed above before check_firewall()

# Clean data if requested
if [ "$CLEAN_DATA" = "yes" ]; then
    print_warning "Cleaning database and logs..."
    
    # Stop services first
    print_info "Stopping services..."
    podman-compose -f "$COMPOSE_FILE" $PROFILE down 2>/dev/null || true
    
    # Remove ipquorum networks
    print_info "Removing ipquorum networks..."
    podman network ls --filter "name=ipquorum" --format "{{.Name}}" | xargs -r podman network rm 2>/dev/null || true
    
    # Backup old data
    if [ -f "$DATA_DIR/ipquorum.db" ]; then
        backup_file="$DATA_DIR/ipquorum.db.backup.$(date +%Y%m%d_%H%M%S)"
        print_info "Backing up database to: $backup_file"
        cp "$DATA_DIR/ipquorum.db" "$backup_file" 2>/dev/null || {
            print_warning "Could not backup database (permission denied)"
            print_info "Run: sudo chown -R $(id -u):$(id -g) $DATA_DIR"
        }
    fi
    
    # Clean database
    if [ -f "$DATA_DIR/ipquorum.db" ]; then
        rm -f "$DATA_DIR/ipquorum.db"
        print_success "Database cleaned"
    fi
    
    # Clean logs
    if [ -d "$LOGS_DIR" ] && [ "$(ls -A $LOGS_DIR 2>/dev/null)" ]; then
        rm -f "$LOGS_DIR"/*.log 2>/dev/null || true
        print_success "Logs cleaned"
    fi
    
    # Recreate directories with proper permissions after cleanup
    print_info "Recreating directories..."
    mkdir -p "$DATA_DIR" "$LOGS_DIR" "$SCRIPTS_DIR"
    chmod 755 "$DATA_DIR" "$LOGS_DIR" "$SCRIPTS_DIR"
    
    # Set ownership to UID 1000 (container appuser) for data directory
    # This allows the container to write to the database
    if command -v chown >/dev/null 2>&1; then
        chown -R 1000:1000 "$DATA_DIR" 2>/dev/null || {
            print_warning "Could not set ownership on $DATA_DIR (may need sudo)"
            print_info "Run: sudo chown -R 1000:1000 $DATA_DIR"
        }
    fi
    
    print_success "Cleanup complete"
fi

# Setup agent API key if --agent flag is used
if [[ "$PROFILE" == *"agent"* ]]; then
    setup_agent_api_key
fi

# Build IPQuorum runner image if requested
if [ "$BUILD_IPQUORUM_RUNNER" = "yes" ]; then
    if [ -d "ipquorum-runner" ]; then
        print_info "Building IPQuorum runner image..."
        echo ""
        
        # Check if build script exists
        if [ -f "ipquorum-runner/build.sh" ]; then
            # Make it executable
            chmod +x ipquorum-runner/build.sh
            
            # Run the build script
            (cd ipquorum-runner && ./build.sh) || {
                print_error "Failed to build IPQuorum runner image"
                print_info "You can build it manually later: cd ipquorum-runner && ./build.sh"
                echo ""
            }
            
            print_success "IPQuorum runner image built successfully"
            echo ""
        else
            print_warning "IPQuorum runner build script not found"
            print_info "Expected: ipquorum-runner/build.sh"
        fi
    else
        print_warning "IPQuorum runner directory not found"
        print_info "Skipping IPQuorum runner build"
    fi
fi

# Build images if requested
if [ "$BUILD_IMAGES" = "yes" ]; then
    if [ "$REBUILD_IMAGES" = "yes" ]; then
        print_warning "Force rebuilding images with --no-cache (this will take longer)..."
        print_info "This ensures all code changes are included in the build"
        print_info "Using BUILDAH_FORMAT=${BUILDAH_FORMAT}"
        
        # Stop services first to avoid conflicts
        print_info "Stopping services..."
        podman-compose -f "$COMPOSE_FILE" $PROFILE down 2>/dev/null || true
        
        # Remove old images to force complete rebuild
        print_info "Removing old images..."
        podman images --filter "reference=localhost/ipquorum*" --format "{{.ID}}" | xargs -r podman rmi -f 2>/dev/null || true
        
        # Build with --no-cache
        podman-compose -f "$COMPOSE_FILE" $PROFILE build --no-cache || {
            print_error "Failed to rebuild images"
            exit 1
        }
        print_success "Images rebuilt successfully with --no-cache"
    else
        print_info "Building images (this may take a few minutes)..."
        print_info "Using BUILDAH_FORMAT=${BUILDAH_FORMAT}"
        # BUILDAH_FORMAT is already exported from .env or set to default 'docker'
        podman-compose -f "$COMPOSE_FILE" $PROFILE build || {
            print_error "Failed to build images"
            exit 1
        }
        print_success "Images built successfully"
    fi
fi

# Start services
print_info "Starting services..."
echo ""

if [ -n "$PROFILE" ]; then
    print_info "Starting with monitoring stack enabled"
fi

if [ -z "$DETACH" ]; then
    print_info "Running in foreground mode (Ctrl+C to stop)"
fi

podman-compose -f "$COMPOSE_FILE" $PROFILE up $DETACH

if [ -n "$DETACH" ]; then
    echo ""
    print_success "Services started successfully!"
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Service Endpoints:${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "${GREEN}Web Dashboard:${NC}"
    echo "  HTTP:  http://localhost:3000 (redirects to HTTPS)"
    echo "  HTTPS: https://localhost:3443"
    echo ""
    echo -e "${GREEN}API Server:${NC}"
    echo "  HTTPS: https://localhost:8443"
    echo ""
    echo -e "${GREEN}Container Agent:${NC}"
    echo "  API:     http://localhost:8081"
    echo "  Metrics: http://localhost:9090"
    echo ""
    
    if [[ "$PROFILE" == *"monitoring"* ]]; then
        echo -e "${GREEN}Monitoring:${NC}"
        echo "  Prometheus: http://localhost:9091"
        echo "  Grafana:    http://localhost:3001 (admin/admin)"
        echo ""
    fi
    
    if [[ "$PROFILE" == *"agent"* ]]; then
        echo -e "${GREEN}Container Agent:${NC}"
        echo "  API:     http://localhost:8081"
        echo "  Metrics: http://localhost:9090"
        echo ""
        print_warning "Note: Agent requires Podman socket at /run/podman/podman.sock"
        print_warning "      On macOS, the agent will fail - it requires Linux with native Podman"
        print_warning "      The agent is designed for production Linux environments"
    fi
    
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Useful Commands:${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo "View logs:"
    echo "  podman-compose -f $COMPOSE_FILE logs -f"
    echo ""
    echo "Check status:"
    echo "  podman-compose -f $COMPOSE_FILE ps"
    echo ""
    echo "Stop services:"
    echo "  podman-compose -f $COMPOSE_FILE down"
    echo ""
    echo "Restart services:"
    echo "  podman-compose -f $COMPOSE_FILE restart"
    echo ""
    echo "Build IPQuorum runner:"
    echo "  cd ipquorum-runner && ./build.sh"
    echo ""
    echo "Test IPQuorum runner:"
    echo "  podman run --rm localhost/ipquorum-runner:latest --help"
    echo ""
    
    print_warning "Note: Browser will show security warnings for self-signed certificates"
    print_info "Accept the certificate to proceed"
    
    # Check if IPQuorum runner image exists
    if podman images localhost/ipquorum-runner:latest --format "{{.Repository}}" 2>/dev/null | grep -q "ipquorum-runner"; then
        echo ""
        print_success "IPQuorum runner image is available"
        print_info "You can now create IPQuorum containers from the web UI"
    else
        echo ""
        print_warning "IPQuorum runner image not found"
        print_info "Build it with: ./start-web-dev.sh --build-ipquorum-runner"
        print_info "Or manually: cd ipquorum-runner && ./build.sh"
    fi
fi

# Made with Bob
