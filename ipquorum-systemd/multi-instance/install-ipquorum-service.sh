#!/usr/bin/env bash
#
# IBM Storage Virtualize IP Quorum Multi-Instance Installation Script
#
# This script installs the multi-instance systemd service configuration
# for running multiple independent IP Quorum instances on a single host.
#
# Usage: sudo ./install-ipquorum-service.sh
#

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[1;36m'  # Bright cyan for better readability on dark terminals
NC='\033[0m' # No Color

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}ERROR: This script must be run as root${NC}" >&2
   echo "Usage: sudo $0" >&2
   exit 1
fi

# Function to print colored messages
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

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Main installation function
main() {
    print_header "IBM Storage Virtualize IP Quorum Multi-Instance Installer"
    
    # Step 1: Check prerequisites
    print_info "Checking prerequisites..."
    
    if ! command_exists systemctl; then
        print_error "systemd is not available on this system"
        exit 1
    fi
    print_success "systemd is available"
    
    if ! command_exists java; then
        print_warning "Java is not installed"
        read -p "Do you want to install OpenJDK? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "Installing OpenJDK..."
            if command_exists apt-get; then
                apt-get update && apt-get install -y openjdk-11-jre-headless
            elif command_exists yum; then
                yum install -y java-11-openjdk-headless
            elif command_exists dnf; then
                dnf install -y java-11-openjdk-headless
            else
                print_error "Unable to install Java automatically. Please install manually."
                exit 1
            fi
            print_success "Java installed successfully"
        else
            print_warning "Skipping Java installation. You'll need to install it manually."
        fi
    else
        JAVA_VERSION=$(java -version 2>&1 | head -n 1)
        print_success "Java is installed: ${JAVA_VERSION}"
    fi
    
    # Step 2: Create service user and group
    print_header "Creating Service User and Group"
    
    if id "ipquorum" &>/dev/null; then
        print_info "User 'ipquorum' already exists"
    else
        print_info "Creating user 'ipquorum'..."
        useradd -r -s /bin/false -d /var/lib/ipquorum -c "IP Quorum Service User" ipquorum
        print_success "User 'ipquorum' created"
    fi
    
    # Step 3: Create directory structure
    print_header "Creating Directory Structure"
    
    print_info "Creating /etc/ipquorum directories..."
    mkdir -p /etc/ipquorum/instances
    chmod 755 /etc/ipquorum
    chmod 755 /etc/ipquorum/instances
    print_success "Configuration directories created"
    
    print_info "Creating /var/lib/ipquorum directories..."
    mkdir -p /var/lib/ipquorum/.passwords
    chown -R ipquorum:ipquorum /var/lib/ipquorum
    chmod 755 /var/lib/ipquorum
    chmod 700 /var/lib/ipquorum/.passwords
    print_success "Data directory created (including password directory)"
    
    print_info "Creating /var/log/ipquorum directory..."
    mkdir -p /var/log/ipquorum
    chown ipquorum:ipquorum /var/log/ipquorum
    chmod 755 /var/log/ipquorum
    print_success "Log directory created"
    
    # Step 4: Install scripts
    print_header "Installing Scripts"
    
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    
    # Determine the systemd directory location
    # Check if scripts are in same directory as installer or in systemd/ subdirectory
    if [[ -f "${SCRIPT_DIR}/ipquorum-download-multi.sh" ]]; then
        SYSTEMD_DIR="${SCRIPT_DIR}"
    elif [[ -f "${SCRIPT_DIR}/systemd/ipquorum-download-multi.sh" ]]; then
        SYSTEMD_DIR="${SCRIPT_DIR}/systemd"
    else
        print_error "Cannot find systemd scripts. Checked:"
        print_error "  - ${SCRIPT_DIR}/"
        print_error "  - ${SCRIPT_DIR}/systemd/"
        exit 1
    fi
    
    print_info "Using systemd scripts from: ${SYSTEMD_DIR}"
    
    if [[ -f "${SYSTEMD_DIR}/ipquorum-download-multi.sh" ]]; then
        print_info "Installing ipquorum-download-multi.sh..."
        cp "${SYSTEMD_DIR}/ipquorum-download-multi.sh" /usr/local/bin/
        chmod 755 /usr/local/bin/ipquorum-download-multi.sh
        print_success "Download script installed"
    else
        print_error "ipquorum-download-multi.sh not found in ${SYSTEMD_DIR}"
        exit 1
    fi
    
    if [[ -f "${SYSTEMD_DIR}/ipquorum-validate-multi.sh" ]]; then
        print_info "Installing ipquorum-validate-multi.sh..."
        cp "${SYSTEMD_DIR}/ipquorum-validate-multi.sh" /usr/local/bin/
        chmod 755 /usr/local/bin/ipquorum-validate-multi.sh
        print_success "Validation script installed"
    else
        print_error "ipquorum-validate-multi.sh not found in ${SYSTEMD_DIR}"
        exit 1
    fi
    
    if [[ -f "${SYSTEMD_DIR}/ipquorum-start-multi.sh" ]]; then
        print_info "Installing ipquorum-start-multi.sh..."
        cp "${SYSTEMD_DIR}/ipquorum-start-multi.sh" /usr/local/bin/
        chmod 755 /usr/local/bin/ipquorum-start-multi.sh
        print_success "Start script installed"
    else
        print_error "ipquorum-start-multi.sh not found in ${SYSTEMD_DIR}"
        exit 1
    fi
    
    # Step 5: Install systemd template service
    print_header "Installing Systemd Template Service"
    
    if [[ -f "${SYSTEMD_DIR}/ipquorum@.service" ]]; then
        print_info "Installing ipquorum@.service..."
        cp "${SYSTEMD_DIR}/ipquorum@.service" /etc/systemd/system/
        chmod 644 /etc/systemd/system/ipquorum@.service
        print_success "Template service installed"
    else
        print_error "ipquorum@.service not found in ${SYSTEMD_DIR}"
        exit 1
    fi
    
    # Step 6: Reload systemd
    print_info "Reloading systemd daemon..."
    systemctl daemon-reload
    print_success "Systemd daemon reloaded"
    
    # Step 7: Install configuration template
    print_header "Installing Configuration Template"
    
    if [[ -f "${SYSTEMD_DIR}/instance.conf.template" ]]; then
        print_info "Installing configuration template..."
        cp "${SYSTEMD_DIR}/instance.conf.template" /etc/ipquorum/
        chmod 644 /etc/ipquorum/instance.conf.template
        print_success "Configuration template installed"
    else
        print_warning "instance.conf.template not found, skipping"
    fi
    
    # Step 8: Copy example configurations
    if [[ -d "${SYSTEMD_DIR}/examples" ]]; then
        print_info "Copying example configurations..."
        cp -r "${SYSTEMD_DIR}/examples" /etc/ipquorum/
        chmod 644 /etc/ipquorum/examples/*.conf
        print_success "Example configurations copied to /etc/ipquorum/examples/"
    fi
    
    # Step 9: Install download tools
    print_header "IP Quorum Download Tools"
    
    DOWNLOAD_TOOLS_FOUND=0
    
    # Check existing tools
    if [[ -x "/usr/local/bin/ipquorum-download-go" ]]; then
        print_success "Go download tool already installed: /usr/local/bin/ipquorum-download-go"
        DOWNLOAD_TOOLS_FOUND=$((DOWNLOAD_TOOLS_FOUND + 1))
    fi
    
    if [[ -x "/usr/local/bin/ipquorum-download.py" ]]; then
        print_success "Python download tool already installed: /usr/local/bin/ipquorum-download.py"
        DOWNLOAD_TOOLS_FOUND=$((DOWNLOAD_TOOLS_FOUND + 1))
    fi
    
    if [[ -x "/usr/local/bin/ipquorum-restapi-download.sh" ]]; then
        print_success "Bash download tool already installed: /usr/local/bin/ipquorum-restapi-download.sh"
        DOWNLOAD_TOOLS_FOUND=$((DOWNLOAD_TOOLS_FOUND + 1))
    fi
    
    # Offer to install download tools
    if [[ $DOWNLOAD_TOOLS_FOUND -eq 0 ]]; then
        echo ""
        print_warning "No download tools found!"
        print_info "Download tools are required for automatic IP Quorum JAR downloads."
        echo ""
        print_info "Available options:"
        print_info "  1. Download from GitHub (latest release)"
        print_info "  2. Download from custom URL"
        print_info "  3. Install from local directory (Go downloader included with release package)"
        print_info "  4. Skip (manual installation required)"
        echo ""
        
        read -p "Choose an option (1-4): " -r DOWNLOAD_CHOICE
        echo ""
        
        case $DOWNLOAD_CHOICE in
            1)
                print_info "Installing ipquorum-download-go from GitHub..."
                
                # Detect architecture
                ARCH=$(uname -m)
                case $ARCH in
                    x86_64) ARCH="amd64" ;;
                    aarch64|arm64) ARCH="arm64" ;;
                    armv7l) ARCH="armv7" ;;
                    *)
                        print_error "Unsupported architecture: $ARCH"
                        print_info "Please install manually from: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases"
                        ;;
                esac
                
                if [[ -n "$ARCH" ]]; then
                    # Get latest version from GitHub API
                    print_info "Fetching latest version from GitHub..."
                    if command_exists curl; then
                        LATEST_VERSION=$(curl -s https://api.github.com/repos/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
                    elif command_exists wget; then
                        LATEST_VERSION=$(wget -qO- https://api.github.com/repos/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
                    else
                        print_error "Neither curl nor wget found. Cannot download."
                        print_info "Please install curl or wget and try again."
                        LATEST_VERSION=""
                    fi
                    
                    if [[ -z "$LATEST_VERSION" ]]; then
                        print_error "Failed to fetch latest version from GitHub API"
                        print_info "Please install manually from: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases"
                    else
                        DOWNLOAD_URL="https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/v${LATEST_VERSION}/ipquorum-download-go-linux-${ARCH}"
                        
                        print_info "Latest version: v${LATEST_VERSION}"
                        print_info "Downloading from: $DOWNLOAD_URL"
                        
                        if command_exists curl; then
                            curl -fsSL -o /usr/local/bin/ipquorum-download-go "$DOWNLOAD_URL"
                        elif command_exists wget; then
                            wget -q -O /usr/local/bin/ipquorum-download-go "$DOWNLOAD_URL"
                        fi
                        
                        if [[ -f "/usr/local/bin/ipquorum-download-go" ]]; then
                            chmod 755 /usr/local/bin/ipquorum-download-go
                            print_success "ipquorum-download-go v${LATEST_VERSION} installed successfully"
                            DOWNLOAD_TOOLS_FOUND=$((DOWNLOAD_TOOLS_FOUND + 1))
                        else
                            print_error "Download failed"
                        fi
                    fi
                fi
                ;;
            2)
                print_info "Download from custom URL"
                echo ""
                print_info "Examples:"
                print_info "  - Specific version: https://github.com/olemyk/.../releases/download/v1.2.3/ipquorum-download-go-linux-amd64"
                print_info "  - Local server: http://internal-server/downloads/ipquorum-download-go"
                print_info "  - File server: https://files.example.com/ipquorum-download-go-linux-amd64"
                echo ""
                
                read -p "Enter download URL: " -r CUSTOM_URL
                
                if [[ -z "$CUSTOM_URL" ]]; then
                    print_error "No URL provided, skipping"
                else
                    print_info "Downloading from: $CUSTOM_URL"
                    if command_exists curl; then
                        curl -L -o /usr/local/bin/ipquorum-download-go "$DOWNLOAD_URL"
                    elif command_exists wget; then
                        wget -O /usr/local/bin/ipquorum-download-go "$DOWNLOAD_URL"
                    else
                        print_error "Neither curl nor wget found. Cannot download."
                        print_info "Please install curl or wget and try again."
                    fi
                    
                    if [[ -f "/usr/local/bin/ipquorum-download-go" ]]; then
                        chmod 755 /usr/local/bin/ipquorum-download-go
                        print_success "ipquorum-download-go installed successfully"
                        DOWNLOAD_TOOLS_FOUND=$((DOWNLOAD_TOOLS_FOUND + 1))
                    else
                        print_error "Download failed"
                    fi
                fi
                ;;
            3)
                print_info "Looking for download tools in current directory..."
                
                # Detect architecture for selecting the right binary
                ARCH=$(uname -m)
                case $ARCH in
                    x86_64) ARCH_SUFFIX="linux-amd64" ;;
                    aarch64|arm64) ARCH_SUFFIX="linux-arm64" ;;
                    armv7l) ARCH_SUFFIX="linux-armv7" ;;
                    *)
                        print_warning "Unknown architecture: $ARCH, will try all binaries"
                        ARCH_SUFFIX="linux-amd64"
                        ;;
                esac
                
                print_info "Detected architecture: $ARCH (looking for: $ARCH_SUFFIX)"
                
                # Check for Go binary with various names and locations
                local go_found=false
                for go_binary in \
                    "${SCRIPT_DIR}/ipquorum-downloader/ipquorum-download-go-${ARCH_SUFFIX}" \
                    "${SCRIPT_DIR}/ipquorum-downloader/ipquorum-download-go" \
                    "${SCRIPT_DIR}/../ipquorum-downloader/ipquorum-download-go-${ARCH_SUFFIX}" \
                    "${SCRIPT_DIR}/../ipquorum-downloader/ipquorum-download-go" \
                    "${SCRIPT_DIR}/../ipquorum-download-go/ipquorum-download-go" \
                    "${SCRIPT_DIR}/ipquorum-download-go" \
                    "${SCRIPT_DIR}/ipquorum-download-go-${ARCH_SUFFIX}" \
                    "${SCRIPT_DIR}/../ipquorum-download-go-${ARCH_SUFFIX}" \
                    "./ipquorum-downloader/ipquorum-download-go-${ARCH_SUFFIX}" \
                    "./ipquorum-download-go-${ARCH_SUFFIX}" \
                    "./ipquorum-download-go"
                do
                    if [[ -f "$go_binary" ]]; then
                        print_info "Found Go binary: $go_binary"
                        cp "$go_binary" /usr/local/bin/ipquorum-download-go
                        chmod 755 /usr/local/bin/ipquorum-download-go
                        print_success "ipquorum-download-go installed from local directory"
                        DOWNLOAD_TOOLS_FOUND=$((DOWNLOAD_TOOLS_FOUND + 1))
                        go_found=true
                        break
                    fi
                done
                
                if [[ "$go_found" == "false" ]]; then
                    print_warning "Go binary not found in local directories"
                fi
                
                # Check for Python script
                if [[ -f "${SCRIPT_DIR}/../ipquorum-download/python-version/ipquorum-restapi-download.py" ]]; then
                    print_info "Found Python script, installing..."
                    cp "${SCRIPT_DIR}/../ipquorum-download/python-version/ipquorum-restapi-download.py" /usr/local/bin/ipquorum-download.py
                    chmod 755 /usr/local/bin/ipquorum-download.py
                    print_success "ipquorum-download.py installed from local directory"
                    DOWNLOAD_TOOLS_FOUND=$((DOWNLOAD_TOOLS_FOUND + 1))
                fi
                
                # Check for Bash script
                if [[ -f "${SCRIPT_DIR}/../ipquorum-download/bash-version/ipquorum-restapi-download.sh" ]]; then
                    print_info "Found Bash script, installing..."
                    cp "${SCRIPT_DIR}/../ipquorum-download/bash-version/ipquorum-restapi-download.sh" /usr/local/bin/
                    chmod 755 /usr/local/bin/ipquorum-restapi-download.sh
                    print_success "ipquorum-restapi-download.sh installed from local directory"
                    DOWNLOAD_TOOLS_FOUND=$((DOWNLOAD_TOOLS_FOUND + 1))
                fi
                
                if [[ $DOWNLOAD_TOOLS_FOUND -eq 0 ]]; then
                    print_warning "No download tools found in local directory"
                fi
                ;;
            4)
                print_info "Skipping download tool installation"
                print_warning "You will need to install a download tool manually or disable automatic downloads"
                ;;
            *)
                print_warning "Invalid option, skipping download tool installation"
                ;;
        esac
    fi
    
    if [[ $DOWNLOAD_TOOLS_FOUND -gt 0 ]]; then
        echo ""
        print_success "Download tools available: $DOWNLOAD_TOOLS_FOUND"
    else
        echo ""
        print_warning "No download tools installed"
        print_info "To install manually:"
        print_info "  Go: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases"
        print_info "  Or set IPQUORUM_DOWNLOAD_ENABLED=false in instance configs"
    fi
    
    # Step 10: Configure firewall (optional)
    print_header "Firewall Configuration"
    
    if command_exists firewall-cmd; then
        read -p "Do you want to open port 1260/tcp in firewall? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "Opening port 1260/tcp..."
            firewall-cmd --permanent --add-port=1260/tcp
            firewall-cmd --reload
            print_success "Firewall configured"
        fi
    elif command_exists ufw; then
        read -p "Do you want to open port 1260/tcp in firewall? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "Opening port 1260/tcp..."
            ufw allow 1260/tcp
            print_success "Firewall configured"
        fi
    else
        print_info "No firewall detected or firewall management skipped"
    fi
    
    # Install instance manager
    print_header "Installing Instance Manager"
    
    if [[ -f "${SYSTEMD_DIR}/ipquorum-instance-manager.sh" ]]; then
        print_info "Installing ipquorum-instance-manager.sh..."
        cp "${SYSTEMD_DIR}/ipquorum-instance-manager.sh" /usr/local/bin/
        chmod 755 /usr/local/bin/ipquorum-instance-manager.sh
        print_success "Instance manager installed: /usr/local/bin/ipquorum-instance-manager.sh"
    else
        print_warning "ipquorum-instance-manager.sh not found, skipping"
    fi
    
    # Final summary
    print_header "Installation Complete!"
    
    echo ""
    echo -e "${GREEN}✓ Multi-instance IP Quorum service installed successfully!${NC}"
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}Quick Start Guide${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
    echo -e "${YELLOW}Manual Configuration${NC}"
    echo ""
    echo "1. Create instance configuration:"
    echo "   sudo cp /etc/ipquorum/examples/system1.conf /etc/ipquorum/instances/myinstance.conf"
    echo "   sudo vi /etc/ipquorum/instances/myinstance.conf"
    echo ""
    echo "2. Create password file:"
    echo "   sudo mkdir -p /var/lib/ipquorum/.passwords"
    echo "   echo 'your_password' | sudo tee /var/lib/ipquorum/.passwords/myinstance.password"
    echo "   sudo chmod 440 /var/lib/ipquorum/.passwords/myinstance.password"
    echo "   sudo chown ipquorum:ipquorum /var/lib/ipquorum/.passwords/myinstance.password"
    echo ""
    echo "3. Create instance directories:"
    echo "   sudo mkdir -p /var/lib/ipquorum/myinstance"
    echo "   sudo mkdir -p /var/log/ipquorum/myinstance"
    echo "   sudo chown -R ipquorum:ipquorum /var/lib/ipquorum/myinstance"
    echo "   sudo chown -R ipquorum:ipquorum /var/log/ipquorum/myinstance"
    echo ""
    echo "4. Enable and start the instance:"
    echo "   sudo systemctl enable ipquorum@myinstance.service"
    echo "   sudo systemctl start ipquorum@myinstance.service"
    echo ""
    echo "5. Check instance status:"
    echo "   sudo systemctl status ipquorum@myinstance.service"
    echo "   sudo journalctl -u ipquorum@myinstance.service -f"
    echo ""
    echo "Example configurations are available in: /etc/ipquorum/examples/"
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}RECOMMENDED: Instance Manager (Interactive)${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "Create a new instance with interactive configuration:"
    echo -e "  ${BLUE}sudo /usr/local/bin/ipquorum-instance-manager.sh create svc_cluster01${NC}"
    echo ""
    echo "The instance manager will prompt you for:"
    echo "  • Description"
    echo "  • API endpoint and credentials"
    echo "  • Download and mkquorumapp settings"
    echo "  • Partner system configuration (if needed)"
    echo ""
    echo "Then manage your instance:"
    echo -e "  ${BLUE}sudo /usr/local/bin/ipquorum-instance-manager.sh enable svc_cluster01${NC}"
    echo -e "  ${BLUE}sudo /usr/local/bin/ipquorum-instance-manager.sh start svc_cluster01${NC}"
    echo -e "  ${BLUE}sudo /usr/local/bin/ipquorum-instance-manager.sh status svc_cluster01${NC}"
    echo -e "  ${BLUE}sudo /usr/local/bin/ipquorum-instance-manager.sh logs svc_cluster01${NC}"
    echo ""
    echo "List all instances:"
    echo -e "  ${BLUE}sudo /usr/local/bin/ipquorum-instance-manager.sh list${NC}"
    echo ""
    echo -e "${YELLOW}TIP: Create a permanent alias for easier access:${NC}"
    echo "  echo \"alias ipqm='sudo /usr/local/bin/ipquorum-instance-manager.sh'\" >> ~/.bashrc"
    echo "  source ~/.bashrc"
    echo ""
    echo -e "Then use: ${BLUE}ipqm create svc_cluster01${NC}"
    echo ""
    echo "For more information, see the documentation in /etc/ipquorum/"
    echo ""
}

# Run main function
main "$@"




