#!/usr/bin/env bash
set -euo pipefail

# IPQuorum Agent Installation Script
# This script installs the IPQuorum agent as a systemd service

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/ipquorum-agent"
SYSTEMD_DIR="/etc/systemd/system"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root"
   exit 1
fi

log_info "Installing IPQuorum Agent..."

# Build the agent binary
log_info "Building agent binary..."
cd "$SCRIPT_DIR"
go build -o ipquorum-agent -ldflags="-s -w" ./cmd/agent

# Install binary
log_info "Installing binary to $INSTALL_DIR..."
install -m 755 ipquorum-agent "$INSTALL_DIR/ipquorum-agent"

# Create configuration directory
log_info "Creating configuration directory..."
mkdir -p "$CONFIG_DIR"

# Install configuration file
if [[ ! -f "$CONFIG_DIR/config.yaml" ]]; then
    log_info "Installing default configuration..."
    cp config.yaml.example "$CONFIG_DIR/config.yaml"
    chmod 600 "$CONFIG_DIR/config.yaml"
    
    # Generate random API key
    API_KEY=$(openssl rand -hex 32)
    sed -i "s/your-secure-api-key-here/$API_KEY/" "$CONFIG_DIR/config.yaml"
    
    log_warn "Generated API key: $API_KEY"
    log_warn "Please save this API key securely!"
else
    log_info "Configuration file already exists, skipping..."
fi

# Setup sudo permissions for ipquorum user
log_info "Setting up sudo permissions for ipquorum user..."
if [[ -f "$SCRIPT_DIR/setup-sudo.sh" ]]; then
    bash "$SCRIPT_DIR/setup-sudo.sh"
else
    log_warn "setup-sudo.sh not found, creating sudoers file manually..."
    cat > /etc/sudoers.d/ipquorum << 'EOF'
# Allow ipquorum user to manage IPQuorum instances without password
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/sed
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/mkdir
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/bash -c *
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/chown
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/chmod
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl start ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl stop ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl status ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl enable ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl disable ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl is-active ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/systemctl is-enabled ipquorum@*
ipquorum ALL=(ALL) NOPASSWD: /usr/bin/rm
EOF
    chmod 0440 /etc/sudoers.d/ipquorum
    
    if visudo -c -f /etc/sudoers.d/ipquorum; then
        log_info "Sudoers file created successfully"
    else
        log_error "Sudoers file validation failed"
        rm -f /etc/sudoers.d/ipquorum
        exit 1
    fi
fi

# Install systemd service
log_info "Installing systemd service..."
cp ipquorum-agent.service "$SYSTEMD_DIR/ipquorum-agent.service"
chmod 644 "$SYSTEMD_DIR/ipquorum-agent.service"

# Reload systemd
log_info "Reloading systemd..."
systemctl daemon-reload

# Enable service
log_info "Enabling ipquorum-agent service..."
systemctl enable ipquorum-agent.service

log_info "Installation complete!"
echo ""
log_info "Next steps:"
echo "  1. Edit configuration: $CONFIG_DIR/config.yaml"
echo "  2. Set the script path in config.yaml"
echo "  3. Start the service: systemctl start ipquorum-agent"
echo "  4. Check status: systemctl status ipquorum-agent"
echo "  5. View logs: journalctl -u ipquorum-agent -f"
echo ""
log_warn "Remember to configure the API key in your management server!"


