#!/usr/bin/env bash
#
# IBM Storage Virtualize IP Quorum Service Installation Script
# This script installs and configures the IP Quorum systemd service
#
# Usage: sudo ./install-ipquorum-service.sh [options]
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
INSTALL_DIR="/opt/IBM/ip-quorum"
CONFIG_DIR="/etc/ipquorum"
SYSTEMD_DIR="/etc/systemd/system"
BIN_DIR="/usr/local/bin"
SERVICE_NAME="ipquorum"
IPQUORUM_USER="ipquorum"
IPQUORUM_GROUP="ipquorum"
DOWNLOAD_TOOL="go"
ENABLE_DOWNLOAD="false"
AUTO_INSTALL_JAVA="false"

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$*${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root"
        exit 1
    fi
}

detect_package_manager() {
    if command -v dnf &> /dev/null; then
        echo "dnf"
    elif command -v yum &> /dev/null; then
        echo "yum"
    elif command -v apt-get &> /dev/null; then
        echo "apt"
    elif command -v zypper &> /dev/null; then
        echo "zypper"
    else
        echo "unknown"
    fi
}

install_java() {
    local pkg_manager=$(detect_package_manager)
    
    print_info "Detected package manager: $pkg_manager"
    
    case "$pkg_manager" in
        dnf|yum)
            print_info "Installing Java using $pkg_manager..."
            $pkg_manager install -y java-11-openjdk-headless || \
            $pkg_manager install -y java-1.8.0-openjdk-headless
            ;;
        apt)
            print_info "Installing Java using apt..."
            apt-get update
            apt-get install -y openjdk-11-jre-headless || \
            apt-get install -y openjdk-8-jre-headless
            ;;
        zypper)
            print_info "Installing Java using zypper..."
            zypper install -y java-11-openjdk-headless || \
            zypper install -y java-1_8_0-openjdk-headless
            ;;
        *)
            print_error "Unknown package manager. Please install Java manually."
            print_info "Supported Java versions: OpenJDK 8, 11, or later"
            return 1
            ;;
    esac
    
    if command -v java &> /dev/null; then
        print_success "Java installed successfully"
        java -version
        return 0
    else
        print_error "Java installation failed"
        return 1
    fi
}

check_java() {
    print_info "Checking for Java..."
    
    if command -v java &> /dev/null; then
        local java_version=$(java -version 2>&1 | head -n 1)
        print_success "Java is installed: $java_version"
        return 0
    else
        print_warning "Java is not installed"
        
        # Ask user if they want to install Java
        echo ""
        read -p "Would you like to install Java now? (y/n): " -n 1 -r
        echo ""
        
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            if install_java; then
                return 0
            else
                print_error "Failed to install Java"
                return 1
            fi
        else
            print_error "Java is required to run IP Quorum service"
            print_info "Please install Java manually and run this script again"
            print_info "Supported versions: OpenJDK 8, 11, or later"
            return 1
        fi
    fi
}

check_dependencies() {
    print_info "Checking dependencies..."
    
    local missing_deps=()
    
    # Check for Java (with installation option)
    if ! check_java; then
        exit 1
    fi
    
    # Check for systemctl
    if ! command -v systemctl &> /dev/null; then
        missing_deps+=("systemd")
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        print_info "Please install missing dependencies and try again"
        exit 1
    fi
    
    print_success "All dependencies satisfied"
}

create_user() {
    print_info "Creating service user and group..."
    
    # Create group if it doesn't exist
    if ! getent group "$IPQUORUM_GROUP" >/dev/null 2>&1; then
        groupadd -r "$IPQUORUM_GROUP"
        print_success "Created group: $IPQUORUM_GROUP"
    else
        print_info "Group already exists: $IPQUORUM_GROUP"
    fi
    
    # Create user if it doesn't exist
    if ! getent passwd "$IPQUORUM_USER" >/dev/null 2>&1; then
        useradd -r -g "$IPQUORUM_GROUP" -s /sbin/nologin \
            -c "IP Quorum service account" -d / "$IPQUORUM_USER"
        print_success "Created user: $IPQUORUM_USER"
    else
        print_info "User already exists: $IPQUORUM_USER"
    fi
}

create_directories() {
    print_info "Creating directories..."
    
    # Create installation directory
    mkdir -p "$INSTALL_DIR/log"
    chown -R "$IPQUORUM_USER:$IPQUORUM_GROUP" "$INSTALL_DIR"
    chmod 755 "$INSTALL_DIR"
    chmod 755 "$INSTALL_DIR/log"
    print_success "Created: $INSTALL_DIR"
    
    # Create download log file with proper permissions
    touch "$INSTALL_DIR/log/download.log"
    chown "$IPQUORUM_USER:$IPQUORUM_GROUP" "$INSTALL_DIR/log/download.log"
    chmod 644 "$INSTALL_DIR/log/download.log"
    print_success "Created download log file with proper permissions"
    
    # Create configuration directory
    mkdir -p "$CONFIG_DIR"
    chmod 755 "$CONFIG_DIR"
    print_success "Created: $CONFIG_DIR"
}

install_download_script() {
    print_info "Installing download script..."
    
    local source_script="$SCRIPT_DIR/ipquorum-download.sh"
    local dest_script="$BIN_DIR/ipquorum-download.sh"
    
    if [[ ! -f "$source_script" ]]; then
        print_error "Download script not found: $source_script"
        exit 1
    fi
    
    cp "$source_script" "$dest_script"
    chmod 755 "$dest_script"
    print_success "Installed: $dest_script"
}

install_start_script() {
    print_info "Installing start script..."
    
    local source_script="$SCRIPT_DIR/ipquorum-start.sh"
    local dest_script="$BIN_DIR/ipquorum-start.sh"
    
    if [[ ! -f "$source_script" ]]; then
        print_error "Start script not found: $source_script"
        exit 1
    fi
    
    cp "$source_script" "$dest_script"
    chmod 755 "$dest_script"
    print_success "Installed: $dest_script"
}

install_config() {
    print_info "Installing configuration file..."
    
    local source_config="$SCRIPT_DIR/ipquorum.conf"
    local dest_config="$CONFIG_DIR/ipquorum.conf"
    
    if [[ ! -f "$source_config" ]]; then
        print_error "Configuration file not found: $source_config"
        exit 1
    fi
    
    if [[ -f "$dest_config" ]]; then
        print_warning "Configuration file already exists: $dest_config"
        print_info "Creating backup: ${dest_config}.backup"
        cp "$dest_config" "${dest_config}.backup"
    fi
    
    cp "$source_config" "$dest_config"
    chmod 644 "$dest_config"
    print_success "Installed: $dest_config"
}

install_service() {
    print_info "Installing systemd service..."
    
    local source_service="$SCRIPT_DIR/ibm-virtualize-ipquorum-improved.service"
    local dest_service="$SYSTEMD_DIR/${SERVICE_NAME}.service"
    
    if [[ ! -f "$source_service" ]]; then
        print_error "Service file not found: $source_service"
        exit 1
    fi
    
    if [[ -f "$dest_service" ]]; then
        print_warning "Service file already exists: $dest_service"
        print_info "Creating backup: ${dest_service}.backup"
        cp "$dest_service" "${dest_service}.backup"
    fi
    
    # Copy and customize the service file with actual paths
    print_info "Customizing service file with installation paths..."
    sed -e "s|WorkingDirectory=/opt/IBM/ip-quorum/log|WorkingDirectory=${INSTALL_DIR}/log|g" \
        -e "s|User=ipquorum|User=${IPQUORUM_USER}|g" \
        -e "s|Group=ipquorum|Group=${IPQUORUM_GROUP}|g" \
        -e "s|ReadWritePaths=/opt/IBM/ip-quorum|ReadWritePaths=${INSTALL_DIR}|g" \
        "$source_service" > "$dest_service"
    
    chmod 644 "$dest_service"
    print_success "Installed: $dest_service"
    
    # Reload systemd
    systemctl daemon-reload
    print_success "Systemd daemon reloaded"
}

configure_firewall() {
    print_info "Configuring firewall..."
    
    if command -v firewall-cmd &> /dev/null; then
        if firewall-cmd --state &> /dev/null; then
            firewall-cmd --add-port=1260/tcp --permanent
            firewall-cmd --reload
            print_success "Firewall configured (port 1260/tcp)"
        else
            print_warning "Firewalld is not running"
        fi
    elif command -v ufw &> /dev/null; then
        ufw allow 1260/tcp
        print_success "Firewall configured (port 1260/tcp)"
    else
        print_warning "No supported firewall found, skipping firewall configuration"
        print_info "Please manually open port 1260/tcp if you have a firewall"
    fi
}

print_next_steps() {
    print_header "Installation Complete!"
    
    echo "Next steps:"
    echo ""
    echo "1. Configure the service:"
    echo "   Edit: $CONFIG_DIR/ipquorum.conf"
    echo ""
    echo "2. If enabling automatic download, configure API access:"
    echo "   - Set API_ENDPOINT (e.g., 10.33.7.80)"
    echo "   - Set VIRTUALIZE_USERNAME"
    echo "   - Create password file (choose one option):"
    echo ""
    echo "   Option A (Recommended): Group-readable"
    echo "     echo 'your_password' | sudo tee $CONFIG_DIR/.password > /dev/null"
    echo "     sudo chmod 440 $CONFIG_DIR/.password"
    echo "     sudo chown root:$IPQUORUM_GROUP $CONFIG_DIR/.password"
    echo ""
    echo "   Option B (Alternative): User-owned"
    echo "     echo 'your_password' | sudo tee $CONFIG_DIR/.password > /dev/null"
    echo "     sudo chmod 400 $CONFIG_DIR/.password"
    echo "     sudo chown $IPQUORUM_USER:$IPQUORUM_GROUP $CONFIG_DIR/.password"
    echo ""
    echo "   - Set IPQUORUM_DOWNLOAD_ENABLED=true in $CONFIG_DIR/ipquorum.conf"
    echo ""
    echo "3. If NOT using automatic download:"
    echo "   - Manually place ip_quorum.jar in: $INSTALL_DIR/"
    echo "   - Set ownership: chown $IPQUORUM_USER:$IPQUORUM_GROUP $INSTALL_DIR/ip_quorum.jar"
    echo ""
    echo "4. Install a download tool (if using automatic download):"
    echo "   - Go binary (recommended): Copy ipquorum-download-go to $BIN_DIR/"
    echo "   - Python script: Copy ipquorum-download.py to $BIN_DIR/"
    echo "   - Bash script: Copy ipquorum-restapi-download.sh to $BIN_DIR/"
    echo ""
    echo "5. Enable and start the service:"
    echo "   systemctl enable $SERVICE_NAME"
    echo "   systemctl start $SERVICE_NAME"
    echo ""
    echo "6. Check service status:"
    echo "   systemctl status $SERVICE_NAME"
    echo "   journalctl -u $SERVICE_NAME -f"
    echo ""
    echo "7. View download logs:"
    echo "   tail -f $INSTALL_DIR/log/download.log"
    echo ""
}

main() {
    print_header "IBM Storage Virtualize IP Quorum Service Installer"
    
    check_root
    check_dependencies
    create_user
    create_directories
    install_download_script
    install_start_script
    install_config
    install_service
    configure_firewall
    print_next_steps
}

# Run main function
main "$@"


