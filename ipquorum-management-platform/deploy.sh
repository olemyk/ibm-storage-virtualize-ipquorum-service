#!/bin/bash
#
# IP Quorum Management Platform - Automated Deployment Script
# 
# This script automates the deployment of the IP Quorum Management Platform
# using pre-built container images from GitHub Container Registry.
#
# Usage: ./deploy.sh [options]
#
# Options:
#   --docker          Use Docker (default if available)
#   --podman          Use Podman
#   --version VERSION Specify version to deploy (default: 3.0.2)
#   --monitoring      Enable monitoring stack (Prometheus + Grafana)
#   --update          Update existing deployment
#   --stop            Stop all services
#   --status          Show deployment status
#   --logs            Show service logs
#   --help            Show this help message
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
VERSION="${VERSION:-3.0.6}"
CONTAINER_RUNTIME=""
COMPOSE_CMD=""
ENABLE_MONITORING=false
ACTION="deploy"
DEPLOYMENT_DIR="/opt/ipquorum-platform"
COMPOSE_FILE="docker-compose.prod.yml"

# Image names
MANAGER_IMAGE="ghcr.io/olemyk/ipquorum-manager"
WEB_IMAGE="ghcr.io/olemyk/ipquorum-web"

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

# Function to show help
show_help() {
    cat << EOF
IP Quorum Management Platform - Deployment Script

Usage: $0 [options]

Options:
  --docker          Use Docker (default if available)
  --podman          Use Podman
  --version VERSION Specify version to deploy (default: v3.0.0)
  --monitoring      Enable monitoring stack (Prometheus + Grafana)
  --update          Update existing deployment
  --stop            Stop all services
  --status          Show deployment status
  --logs            Show service logs
  --help            Show this help message

Examples:
  # Deploy with Docker
  $0 --docker

  # Deploy with Podman and monitoring
  $0 --podman --monitoring

  # Deploy specific version
  $0 --version v3.1.0

  # Update existing deployment
  $0 --update

  # Check deployment status
  $0 --status

  # View logs
  $0 --logs

EOF
    exit 0
}

# Function to detect container runtime
detect_container_runtime() {
    print_info "Detecting container runtime..."
    
    if command -v docker &> /dev/null; then
        if docker ps &> /dev/null; then
            CONTAINER_RUNTIME="docker"
            if docker compose version &> /dev/null 2>&1; then
                COMPOSE_CMD="docker compose"
            elif command -v docker-compose &> /dev/null; then
                COMPOSE_CMD="docker-compose"
            else
                print_error "Docker Compose not found. Please install docker-compose or docker compose plugin."
                exit 1
            fi
            print_success "Using Docker with $COMPOSE_CMD"
            return 0
        fi
    fi
    
    if command -v podman &> /dev/null; then
        CONTAINER_RUNTIME="podman"
        if command -v podman-compose &> /dev/null; then
            COMPOSE_CMD="podman-compose"
            print_success "Using Podman with podman-compose"
            return 0
        else
            print_error "podman-compose not found. Please install podman-compose."
            exit 1
        fi
    fi
    
    print_error "No container runtime found. Please install Docker or Podman."
    exit 1
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check if running as root for system-wide installation
    if [[ "$DEPLOYMENT_DIR" == "/opt/"* ]] && [[ $EUID -ne 0 ]]; then
        print_warning "System-wide installation requires root privileges."
        print_info "Run with sudo or change DEPLOYMENT_DIR to a user directory."
        exit 1
    fi
    
    # Check disk space (minimum 10GB)
    local available_space=$(df -BG "$DEPLOYMENT_DIR" 2>/dev/null | awk 'NR==2 {print $4}' | sed 's/G//')
    if [[ -n "$available_space" ]] && [[ $available_space -lt 10 ]]; then
        print_warning "Low disk space: ${available_space}GB available. Minimum 10GB recommended."
    fi
    
    # Check memory (minimum 4GB)
    local total_mem=$(free -g | awk 'NR==2 {print $2}')
    if [[ $total_mem -lt 4 ]]; then
        print_warning "Low memory: ${total_mem}GB available. Minimum 4GB recommended."
    fi
    
    print_success "Prerequisites check completed"
}

# Function to create directory structure
create_directories() {
    print_info "Creating directory structure..."
    
    mkdir -p "$DEPLOYMENT_DIR"/{data,logs,config,scripts,backups,tls}
    
    # Set appropriate permissions
    if [[ $EUID -eq 0 ]]; then
        chown -R 1000:1000 "$DEPLOYMENT_DIR"/{data,logs}
        
        # Fix SELinux context for RHEL/CentOS/Fedora
        if command -v chcon &> /dev/null && [[ -f /etc/selinux/config ]]; then
            print_info "Applying SELinux context for container volumes..."
            chcon -Rt svirt_sandbox_file_t "$DEPLOYMENT_DIR"/{data,logs,tls} 2>/dev/null || \
                print_warning "Failed to set SELinux context. You may need to run: sudo chcon -Rt svirt_sandbox_file_t $DEPLOYMENT_DIR/{data,logs,tls}"
        fi
    fi
    
    print_success "Directory structure created at $DEPLOYMENT_DIR"
}

# Function to download configuration files
download_configs() {
    print_info "Downloading configuration files..."
    
    cd "$DEPLOYMENT_DIR"
    
    # Download docker-compose file
    if [[ ! -f "$COMPOSE_FILE" ]] || [[ "$ACTION" == "update" ]]; then
        print_info "Downloading $COMPOSE_FILE..."
        curl -fsSL -o "$COMPOSE_FILE" \
            "https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-ipquorum-service/Major-update-to-the-ipquorum-/ipquorum-management-platform/$COMPOSE_FILE" \
            || print_warning "Failed to download $COMPOSE_FILE. Using existing file if available."
    fi
    
    # Download .env template if not exists
    if [[ ! -f ".env" ]]; then
        print_info "Downloading .env template..."
        curl -fsSL -o ".env.tmp" \
            "https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-ipquorum-service/Major-update-to-the-ipquorum-/ipquorum-management-platform/.env.example" \
            || print_warning "Failed to download .env template."
        
        # Strip comments and empty lines for podman-compose compatibility
        if [[ -f ".env.tmp" ]]; then
            print_info "Creating clean .env file (removing comments for podman-compose)..."
            grep -v '^#' .env.tmp | grep -v '^$' | grep '=' > .env
            rm -f .env.tmp
        fi
        
        # Generate secure secrets
        if command -v openssl &> /dev/null; then
            print_info "Generating secure secrets..."
            JWT_SECRET=$(openssl rand -base64 32)
            GRAFANA_PASSWORD=$(openssl rand -base64 16)
            
            sed -i "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
            sed -i "s|GRAFANA_ADMIN_PASSWORD=.*|GRAFANA_ADMIN_PASSWORD=$GRAFANA_PASSWORD|" .env
            
            print_success "Secure secrets generated"
        fi
    else
        print_info "Using existing .env file"
    fi
    
    print_success "Configuration files ready"
}

# Function to generate TLS certificates
generate_tls_certificates() {
    print_info "Checking TLS certificates..."
    
    local tls_dir="$DEPLOYMENT_DIR/tls"
    local cert_file="$tls_dir/server.crt"
    local key_file="$tls_dir/server.key"
    
    if [[ -f "$cert_file" ]] && [[ -f "$key_file" ]]; then
        print_info "TLS certificates already exist"
        return 0
    fi
    
    if ! command -v openssl &> /dev/null; then
        print_warning "OpenSSL not found. Skipping TLS certificate generation."
        print_warning "The server will fail to start without TLS certificates."
        return 1
    fi
    
    print_info "Generating self-signed TLS certificates..."
    
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$key_file" \
        -out "$cert_file" \
        -subj "/C=US/ST=State/L=City/O=IPQuorum/CN=localhost" \
        -addext "subjectAltName=DNS:localhost,DNS:ipquorum-server,IP:127.0.0.1" \
        2>/dev/null
    
    if [[ $? -eq 0 ]]; then
        chmod 600 "$key_file"
        chmod 644 "$cert_file"
        
        if [[ $EUID -eq 0 ]]; then
            chown 1000:1000 "$key_file" "$cert_file"
        fi
        
        print_success "TLS certificates generated successfully"
    else
        print_error "Failed to generate TLS certificates"
        return 1
    fi
}

# Function to update image versions in compose file
update_image_versions() {
    print_info "Using image versions from docker-compose file..."
    
    cd "$DEPLOYMENT_DIR"
    
    # Note: We don't modify the docker-compose file anymore to avoid YAML syntax issues
    # The docker-compose.prod.yml file already has the correct image versions
    # Users can manually edit the file if they need different versions
    
    print_success "Image versions ready"
}

# Function to pull container images
pull_images() {
    print_info "Pulling container images..."
    
    cd "$DEPLOYMENT_DIR"
    
    # Pull images
    if [[ "$CONTAINER_RUNTIME" == "docker" ]]; then
        docker pull "$MANAGER_IMAGE:$VERSION" || {
            print_error "Failed to pull manager image"
            exit 1
        }
        docker pull "$WEB_IMAGE:$VERSION" || {
            print_error "Failed to pull web image"
            exit 1
        }
    else
        podman pull "$MANAGER_IMAGE:$VERSION" || {
            print_error "Failed to pull manager image"
            exit 1
        }
        podman pull "$WEB_IMAGE:$VERSION" || {
            print_error "Failed to pull web image"
            exit 1
        }
    fi
    
    print_success "Container images pulled successfully"
}

# Function to backup existing data
backup_data() {
    if [[ -f "$DEPLOYMENT_DIR/data/ipquorum.db" ]]; then
        print_info "Backing up existing database..."
        
        local backup_file="$DEPLOYMENT_DIR/backups/ipquorum.db.backup-$(date +%Y%m%d-%H%M%S)"
        cp "$DEPLOYMENT_DIR/data/ipquorum.db" "$backup_file"
        
        print_success "Database backed up to $backup_file"
    fi
}

# Function to start services
start_services() {
    print_info "Starting services..."
    
    cd "$DEPLOYMENT_DIR"
    
    local compose_args="-f $COMPOSE_FILE"
    
    if [[ "$ENABLE_MONITORING" == true ]]; then
        compose_args="$compose_args --profile monitoring"
        print_info "Monitoring stack enabled"
    fi
    
    # Start services
    $COMPOSE_CMD $compose_args up -d
    
    print_success "Services started"
}

# Function to wait for services to be healthy
wait_for_services() {
    print_info "Waiting for services to be healthy..."
    
    local max_attempts=30
    local attempt=0
    
    while [[ $attempt -lt $max_attempts ]]; do
        if curl -f http://localhost:8443/health &> /dev/null; then
            print_success "API server is healthy"
            break
        fi
        
        attempt=$((attempt + 1))
        echo -n "."
        sleep 2
    done
    
    echo ""
    
    if [[ $attempt -eq $max_attempts ]]; then
        print_warning "API server health check timed out. Check logs with: $0 --logs"
    fi
    
    # Check web dashboard
    if curl -f http://localhost:3000/ &> /dev/null; then
        print_success "Web dashboard is accessible"
    else
        print_warning "Web dashboard not accessible yet. It may still be starting."
    fi
}

# Function to show deployment status
show_status() {
    print_info "Deployment Status:"
    echo ""
    
    cd "$DEPLOYMENT_DIR"
    
    # Show running containers
    $COMPOSE_CMD -f "$COMPOSE_FILE" ps
    
    echo ""
    print_info "Service URLs:"
    echo "  Web Dashboard: http://localhost:3000"
    echo "  API Server:    http://localhost:8443"
    echo "  Prometheus:    http://localhost:9090 (if monitoring enabled)"
    echo "  Grafana:       http://localhost:3001 (if monitoring enabled)"
    echo ""
    
    # Show resource usage
    if [[ "$CONTAINER_RUNTIME" == "docker" ]]; then
        print_info "Resource Usage:"
        docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" \
            $(docker ps --filter "label=com.ipquorum.service" -q) 2>/dev/null || true
    fi
}

# Function to show logs
show_logs() {
    cd "$DEPLOYMENT_DIR"
    
    print_info "Showing service logs (Ctrl+C to exit)..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" logs -f
}

# Function to stop services
stop_services() {
    print_info "Stopping services..."
    
    cd "$DEPLOYMENT_DIR"
    
    $COMPOSE_CMD -f "$COMPOSE_FILE" down
    
    print_success "Services stopped"
}

# Function to validate deployment
validate_deployment() {
    print_info "Validating deployment..."
    
    local errors=0
    
    # Check API health
    if ! curl -f http://localhost:8443/health &> /dev/null; then
        print_error "API server health check failed"
        errors=$((errors + 1))
    else
        print_success "API server is healthy"
    fi
    
    # Check web dashboard
    if ! curl -f http://localhost:3000/ &> /dev/null; then
        print_error "Web dashboard is not accessible"
        errors=$((errors + 1))
    else
        print_success "Web dashboard is accessible"
    fi
    
    # Check database
    if [[ ! -f "$DEPLOYMENT_DIR/data/ipquorum.db" ]]; then
        print_warning "Database file not found (will be created on first use)"
    else
        print_success "Database file exists"
    fi
    
    if [[ $errors -eq 0 ]]; then
        print_success "Deployment validation passed"
        return 0
    else
        print_error "Deployment validation failed with $errors error(s)"
        return 1
    fi
}

# Function to show post-deployment instructions
show_instructions() {
    cat << EOF

${GREEN}╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║  IP Quorum Management Platform Deployed Successfully!         ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝${NC}

${BLUE}Access Information:${NC}
  Web Dashboard: http://localhost:3000
  API Server:    http://localhost:8443
  
${BLUE}Default Credentials:${NC}
  Username: admin
  Password: admin123
  
${YELLOW}⚠️  IMPORTANT: Change the default password immediately!${NC}

${BLUE}Next Steps:${NC}
  1. Access the web dashboard at http://localhost:3000
  2. Login with default credentials
  3. Navigate to Settings → Security and change password
  4. Add your first server in Servers → Add Server
  5. Create your first IP Quorum instance

${BLUE}Useful Commands:${NC}
  Check status:  $0 --status
  View logs:     $0 --logs
  Stop services: $0 --stop
  Update:        $0 --update

${BLUE}Documentation:${NC}
  Deployment Guide: $DEPLOYMENT_DIR/DEPLOYMENT-GUIDE.md
  Docker Guide:     $DEPLOYMENT_DIR/DOCKER-DEPLOYMENT.md
  Podman Guide:     $DEPLOYMENT_DIR/PODMAN-DEPLOYMENT.md

${BLUE}Support:${NC}
  GitHub: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service
  Issues: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/issues

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --docker)
            CONTAINER_RUNTIME="docker"
            shift
            ;;
        --podman)
            CONTAINER_RUNTIME="podman"
            shift
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --monitoring)
            ENABLE_MONITORING=true
            shift
            ;;
        --update)
            ACTION="update"
            shift
            ;;
        --stop)
            ACTION="stop"
            shift
            ;;
        --status)
            ACTION="status"
            shift
            ;;
        --logs)
            ACTION="logs"
            shift
            ;;
        --help)
            show_help
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            ;;
    esac
done

# Main execution
main() {
    echo ""
    print_info "IP Quorum Management Platform - Deployment Script"
    print_info "Version: $VERSION"
    echo ""
    
    # Handle different actions
    case $ACTION in
        status)
            detect_container_runtime
            show_status
            exit 0
            ;;
        logs)
            detect_container_runtime
            show_logs
            exit 0
            ;;
        stop)
            detect_container_runtime
            stop_services
            exit 0
            ;;
    esac
    
    # Deployment or update
    detect_container_runtime
    check_prerequisites
    create_directories
    download_configs
    generate_tls_certificates
    update_image_versions
    
    if [[ "$ACTION" == "update" ]]; then
        backup_data
    fi
    
    pull_images
    start_services
    wait_for_services
    
    echo ""
    
    if validate_deployment; then
        show_instructions
    else
        print_error "Deployment validation failed. Check logs with: $0 --logs"
        exit 1
    fi
}

# Run main function
main

# Made with Bob
