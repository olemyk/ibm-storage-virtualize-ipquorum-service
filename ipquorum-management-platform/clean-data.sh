#!/usr/bin/env bash
set -euo pipefail

# IP Quorum Management Platform - Data Cleanup Script
# This script cleans the database and logs

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
DATA_DIR="data"
LOGS_DIR="logs"

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

# Parse command line arguments
FORCE="no"
BACKUP="yes"

while [[ $# -gt 0 ]]; do
    case $1 in
        --force|-f)
            FORCE="yes"
            shift
            ;;
        --no-backup)
            BACKUP="no"
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Clean database and logs for IP Quorum Management Platform"
            echo ""
            echo "Options:"
            echo "  --force, -f         Skip confirmation prompt"
            echo "  --no-backup         Don't create database backup"
            echo "  --help, -h          Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                  # Clean with confirmation and backup"
            echo "  $0 --force          # Clean without confirmation"
            echo "  $0 --no-backup      # Clean without creating backup"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}IP Quorum Data Cleanup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Confirmation prompt
if [ "$FORCE" != "yes" ]; then
    print_warning "This will clean the following:"
    echo "  - Database: $DATA_DIR/ipquorum.db"
    echo "  - Logs: $LOGS_DIR/*.log"
    echo ""
    if [ "$BACKUP" = "yes" ]; then
        echo "  A backup of the database will be created before deletion."
    else
        print_warning "No backup will be created (--no-backup specified)"
    fi
    echo ""
    read -p "Are you sure you want to continue? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        print_info "Cleanup cancelled"
        exit 0
    fi
fi

# Stop services
print_info "Stopping services..."
podman-compose -f "$COMPOSE_FILE" down 2>/dev/null || {
    print_warning "Could not stop services (they may not be running)"
}

# Backup database if it exists and backup is enabled
if [ "$BACKUP" = "yes" ] && [ -f "$DATA_DIR/ipquorum.db" ]; then
    backup_file="$DATA_DIR/ipquorum.db.backup.$(date +%Y%m%d_%H%M%S)"
    print_info "Creating backup: $backup_file"
    cp "$DATA_DIR/ipquorum.db" "$backup_file"
    print_success "Backup created"
fi

# Clean database
if [ -f "$DATA_DIR/ipquorum.db" ]; then
    print_info "Removing database..."
    rm -f "$DATA_DIR/ipquorum.db"
    print_success "Database removed"
else
    print_info "No database file found"
fi

# Clean logs
if [ -d "$LOGS_DIR" ]; then
    log_count=$(find "$LOGS_DIR" -name "*.log" 2>/dev/null | wc -l | tr -d ' ')
    if [ "$log_count" -gt 0 ]; then
        print_info "Removing $log_count log file(s)..."
        rm -f "$LOGS_DIR"/*.log 2>/dev/null || true
        print_success "Logs removed"
    else
        print_info "No log files found"
    fi
else
    print_info "No logs directory found"
fi

echo ""
print_success "Cleanup complete!"
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Next Steps:${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Start services with a fresh database:"
echo "  ./start-web-dev.sh"
echo ""
if [ "$BACKUP" = "yes" ] && [ -f "$backup_file" ]; then
    echo "To restore from backup:"
    echo "  cp $backup_file $DATA_DIR/ipquorum.db"
    echo ""
fi

# Made with Bob
