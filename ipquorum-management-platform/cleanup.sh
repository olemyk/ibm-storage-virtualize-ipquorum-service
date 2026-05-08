#!/bin/bash
#
# IP Quorum Management Platform - Complete Cleanup Script
# 
# This script removes all containers, images, volumes, networks, and data
# associated with the IP Quorum Management Platform deployment.
#
# Usage: ./cleanup.sh [options]
#
# Options:
#   --keep-data       Keep data directory (database and logs)
#   --keep-images     Keep container images (only remove containers)
#   --docker          Use Docker
#   --podman          Use Podman
#   --force           Skip confirmation prompts
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
CONTAINER_RUNTIME=""
COMPOSE_CMD=""
KEEP_DATA=false
KEEP_IMAGES=false
FORCE=false
DEPLOYMENT_DIR="/opt/ipquorum-platform"

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
IP Quorum Management Platform - Cleanup Script

Usage: $0 [options]

Options:
  --keep-data       Keep data directory (database and logs)
  --keep-images     Keep container images (only remove containers)
  --docker          Use Docker
  --podman          Use Podman
  --force           Skip confirmation prompts
  --help            Show this help message

Examples:
  # Complete cleanup (removes everything)
  $0 --podman

  # Remove containers but keep data and images
  $0 --podman --keep-data --keep-images

  # Force cleanup without prompts
  $0 --podman --force

EOF
    exit 0
}

# Function to detect container runtime
detect_container_runtime() {
    print_info "Detecting container runtime..."
    
    if command -v docker &> /dev/null; then
        if docker ps &> /dev/null 2>&1; then
            CONTAINER_RUNTIME="docker"
            if docker compose version &> /dev/null 2>&1; then
                COMPOSE_CMD="docker compose"
            elif command -v docker-compose &> /dev/null; then
                COMPOSE_CMD="docker-compose"
            fi
            print_success "Using Docker"
            return 0
        fi
    fi
    
    if command -v podman &> /dev/null; then
        CONTAINER_RUNTIME="podman"
        if command -v podman-compose &> /dev/null; then
            COMPOSE_CMD="podman-compose"
        fi
        print_success "Using Podman"
        return 0
    fi
    
    print_error "No container runtime found"
    exit 1
}

# Function to confirm action
confirm_action() {
    if [[ "$FORCE" == true ]]; then
        return 0
    fi
    
    local message="$1"
    echo -e "${YELLOW}${message}${NC}"
    read -p "Are you sure? (yes/no): " response
    case "$response" in
        yes|YES|y|Y) return 0 ;;
        *) return 1 ;;
    esac
}

# Function to stop and remove containers
cleanup_containers() {
    print_info "Stopping and removing containers..."
    
    if [[ -f "$DEPLOYMENT_DIR/docker-compose.prod.yml" ]]; then
        cd "$DEPLOYMENT_DIR"
        $COMPOSE_CMD -f docker-compose.prod.yml down -v 2>/dev/null || true
        print_success "Containers stopped and removed via compose"
    fi
    
    # Remove individual containers if they exist
    local containers=(
        "ipquorum-server"
        "ipquorum-web"
        "ipquorum-prometheus"
        "ipquorum-grafana"
        "ipquorum-alertmanager"
    )
    
    for container in "${containers[@]}"; do
        if $CONTAINER_RUNTIME ps -a --format "{{.Names}}" | grep -q "^${container}$"; then
            print_info "Removing container: $container"
            $CONTAINER_RUNTIME stop "$container" 2>/dev/null || true
            $CONTAINER_RUNTIME rm -f "$container" 2>/dev/null || true
        fi
    done
    
    print_success "All containers removed"
}

# Function to remove container images
cleanup_images() {
    if [[ "$KEEP_IMAGES" == true ]]; then
        print_info "Skipping image removal (--keep-images specified)"
        return 0
    fi
    
    print_info "Removing container images..."
    
    local images=(
        "ghcr.io/olemyk/ipquorum-manager"
        "ghcr.io/olemyk/ipquorum-web"
        "ipquorum-manager"
        "ipquorum-web"
        "ipquorum-web:local"
    )
    
    for image in "${images[@]}"; do
        # Remove all tags of this image
        $CONTAINER_RUNTIME images --format "{{.Repository}}:{{.Tag}}" | grep "^${image}" | while read -r img; do
            print_info "Removing image: $img"
            $CONTAINER_RUNTIME rmi -f "$img" 2>/dev/null || true
        done
    done
    
    print_success "Container images removed"
}

# Function to remove volumes
cleanup_volumes() {
    print_info "Removing volumes..."
    
    local volumes=(
        "ipquorum_data"
        "ipquorum_logs"
        "ipquorum_prometheus_data"
        "ipquorum_grafana_data"
    )
    
    for volume in "${volumes[@]}"; do
        if $CONTAINER_RUNTIME volume ls --format "{{.Name}}" | grep -q "^${volume}$"; then
            print_info "Removing volume: $volume"
            $CONTAINER_RUNTIME volume rm -f "$volume" 2>/dev/null || true
        fi
    done
    
    print_success "Volumes removed"
}

# Function to remove networks
cleanup_networks() {
    print_info "Removing networks..."
    
    local networks=(
        "ipquorum-network"
        "ipquorum_default"
    )
    
    for network in "${networks[@]}"; do
        if $CONTAINER_RUNTIME network ls --format "{{.Name}}" | grep -q "^${network}$"; then
            print_info "Removing network: $network"
            $CONTAINER_RUNTIME network rm "$network" 2>/dev/null || true
        fi
    done
    
    print_success "Networks removed"
}

# Function to remove pods (Podman only)
cleanup_pods() {
    if [[ "$CONTAINER_RUNTIME" != "podman" ]]; then
        return 0
    fi
    
    print_info "Removing pods..."
    
    podman pod ls --format "{{.Name}}" | grep "ipquorum" | while read -r pod; do
        print_info "Removing pod: $pod"
        podman pod rm -f "$pod" 2>/dev/null || true
    done
    
    print_success "Pods removed"
}

# Function to remove data directories
cleanup_data() {
    if [[ "$KEEP_DATA" == true ]]; then
        print_info "Skipping data removal (--keep-data specified)"
        return 0
    fi
    
    if [[ ! -d "$DEPLOYMENT_DIR" ]]; then
        print_info "Deployment directory not found: $DEPLOYMENT_DIR"
        return 0
    fi
    
    print_warning "This will permanently delete all data, logs, and configuration!"
    if ! confirm_action "Remove deployment directory: $DEPLOYMENT_DIR"; then
        print_info "Skipping data removal"
        return 0
    fi
    
    print_info "Removing deployment directory..."
    
    if [[ $EUID -eq 0 ]]; then
        rm -rf "$DEPLOYMENT_DIR"
    else
        sudo rm -rf "$DEPLOYMENT_DIR"
    fi
    
    print_success "Deployment directory removed"
}

# Function to show cleanup summary
show_summary() {
    cat << EOF

${GREEN}╔════════════════════════════════════════════════════════════════╗
║                                                                ║
║  IP Quorum Management Platform - Cleanup Complete!            ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝${NC}

${BLUE}Cleanup Summary:${NC}
  ✓ Containers stopped and removed
  ✓ Volumes removed
  ✓ Networks removed
$([ "$CONTAINER_RUNTIME" == "podman" ] && echo "  ✓ Pods removed")
$([ "$KEEP_IMAGES" == false ] && echo "  ✓ Container images removed")
$([ "$KEEP_DATA" == false ] && echo "  ✓ Data directory removed")

${BLUE}What was kept:${NC}
$([ "$KEEP_IMAGES" == true ] && echo "  • Container images (--keep-images)")
$([ "$KEEP_DATA" == true ] && echo "  • Data directory: $DEPLOYMENT_DIR (--keep-data)")

${BLUE}To redeploy:${NC}
  curl -fsSL https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-ipquorum-service/Major-update-to-the-ipquorum-/ipquorum-management-platform/deploy.sh -o deploy.sh
  chmod +x deploy.sh
  sudo ./deploy.sh --podman

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
        --keep-data)
            KEEP_DATA=true
            shift
            ;;
        --keep-images)
            KEEP_IMAGES=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --help|-h)
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
    print_info "IP Quorum Management Platform - Cleanup Script"
    echo ""
    
    # Detect container runtime if not specified
    if [[ -z "$CONTAINER_RUNTIME" ]]; then
        detect_container_runtime
    fi
    
    # Show warning
    print_warning "This script will remove IP Quorum Management Platform components"
    echo ""
    echo "  Containers: YES"
    echo "  Volumes:    YES"
    echo "  Networks:   YES"
    echo "  Pods:       $([ "$CONTAINER_RUNTIME" == "podman" ] && echo "YES" || echo "N/A")"
    echo "  Images:     $([ "$KEEP_IMAGES" == false ] && echo "YES" || echo "NO (--keep-images)")"
    echo "  Data:       $([ "$KEEP_DATA" == false ] && echo "YES" || echo "NO (--keep-data)")"
    echo ""
    
    if ! confirm_action "Proceed with cleanup?"; then
        print_info "Cleanup cancelled"
        exit 0
    fi
    
    echo ""
    
    # Execute cleanup steps
    cleanup_containers
    cleanup_volumes
    cleanup_networks
    cleanup_pods
    cleanup_images
    cleanup_data
    
    echo ""
    show_summary
}

# Run main function
main

# Made with Bob
