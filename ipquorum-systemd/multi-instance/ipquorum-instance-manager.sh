#!/usr/bin/env bash
#
# IBM Storage Virtualize IP Quorum Instance Manager
#
# This script helps manage multiple IP Quorum service instances
#
# Usage: ipquorum-instance-manager.sh <command> [options]
#

set -euo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration paths
INSTANCES_DIR="/etc/ipquorum/instances"
DATA_DIR="/var/lib/ipquorum"
PASSWORDS_DIR="${DATA_DIR}/.passwords"
LOG_DIR="/var/log/ipquorum"
TEMPLATE_FILE="/etc/ipquorum/instance.conf.template"

# Functions
print_usage() {
    cat <<EOF
IBM Storage Virtualize IP Quorum Instance Manager

Usage: $(basename "$0") <command> [options]

Commands:
  create <name> [--non-interactive]  Create a new instance (interactive by default)
  delete <name>                      Delete an instance (stops service and removes config)
  list                               List all configured instances
  status [name]                      Show status of instance(s)
  start <name>                       Start an instance
  stop <name>                        Stop an instance
  restart <name>                     Restart an instance
  enable <name>                      Enable instance to start on boot
  disable <name>                     Disable instance from starting on boot
  logs <name> [lines]                Show logs for an instance (default: 50 lines)
  validate <name>                    Validate instance configuration
  info <name>                        Show detailed information about an instance
  update-password <name>             Update the password for an instance

Examples:
  # Interactive instance creation (recommended)
  $(basename "$0") create svc_cluster01
  
  # Non-interactive (manual configuration required)
  $(basename "$0") create svc_cluster01 --non-interactive
  
  # Manage instances
  $(basename "$0") list
  $(basename "$0") status svc_cluster01
  $(basename "$0") start svc_cluster01
  $(basename "$0") logs svc_cluster01 100
  $(basename "$0") validate svc_cluster01
  
  # Update password
  $(basename "$0") update-password svc_cluster01

EOF
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

# Check if running as root for certain operations
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This operation requires root privileges"
        echo "Please run with sudo: sudo $0 $*"
        exit 1
    fi
}

# Validate instance name
validate_instance_name() {
    local name="$1"
    
    if [[ ! "$name" =~ ^[a-zA-Z0-9_-]+$ ]]; then
        print_error "Invalid instance name: $name"
        echo "Instance name must contain only letters, numbers, hyphens, and underscores"
        return 1
    fi
    
    return 0
}

# Check if instance exists
instance_exists() {
    local name="$1"
    [[ -f "${INSTANCES_DIR}/${name}.conf" ]]
}

# Create new instance
cmd_create() {
    check_root
    
    local name="${1:-}"
    local interactive="${2:-yes}"
    
    if [[ -z "$name" ]]; then
        print_error "Instance name is required"
        echo "Usage: $0 create <instance-name> [--non-interactive]"
        exit 1
    fi
    
    # Check for non-interactive flag
    if [[ "$interactive" == "--non-interactive" ]]; then
        interactive="no"
    fi
    
    validate_instance_name "$name" || exit 1
    
    if instance_exists "$name"; then
        print_error "Instance '$name' already exists"
        exit 1
    fi
    
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}Creating IP Quorum Instance: ${name}${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
    
    # Interactive configuration
    local description storage_system storage_location ipquorum_name api_endpoint username password
    local enable_download="true" enable_mkquorumapp="false"
    local partnersystem ip6="false" partnerip6="false" nometadata="false"
    
    if [[ "$interactive" == "yes" ]]; then
        # Basic information
        echo "╔════════════════════════════════════════════════════════════════╗"
        echo "║  Instance Configuration"
        echo "╚════════════════════════════════════════════════════════════════╝"
        echo ""
        
        # IP Quorum name (shown in IBM Storage Virtualize)
        echo "IP Quorum Name (shown in IBM Storage Virtualize):"
        echo "  - This name appears in 'Detected IP quorum Applications' on the storage system"
        echo "  - Must be 1-20 characters (A-Z, a-z, 0-9 only - no dashes or underscores)"
        echo "  - Example: 'ipquorumsrv1', 'prodquorum', 'dcaquorum'"
        read -p "IP Quorum name [${name}]: " ipquorum_name
        ipquorum_name="${ipquorum_name:-${name}}"
        
        # Validate and sanitize IP Quorum name
        ipquorum_name=$(echo "$ipquorum_name" | tr -d '_-')
        if [[ ! "$ipquorum_name" =~ ^[A-Za-z0-9]{1,20}$ ]]; then
            print_warning "Invalid IP Quorum name. Using sanitized instance name."
            ipquorum_name=$(echo "${name}" | tr -d '_-')
        fi
        echo "  → Will use: ${ipquorum_name}"
        echo ""
        
        # Storage system identification (optional, for documentation)
        read -p "IBM Storage System name (e.g., 'svc_cluster01', 'Production-SAN') [optional]: " storage_system
        storage_system="${storage_system:-}"
        
        read -p "Description (e.g., 'IBM FlashSystem 7600 - Production Site A') [optional]: " description
        description="${description:-}"
        
        read -p "Location (e.g., 'Datacenter A, Rack 12') [optional]: " storage_location
        storage_location="${storage_location:-}"
        
        echo ""
        
        # Download configuration
        read -p "Enable automatic JAR download? (yes/no) [yes]: " enable_download_input
        enable_download=$(echo "${enable_download_input:-yes}" | tr '[:upper:]' '[:lower:]')
        if [[ "$enable_download" =~ ^(yes|y|true|1)$ ]]; then
            enable_download="true"
            
            # API configuration
            read -p "IBM Storage Virtualize Cluster/API endpoint (IP or hostname): " api_endpoint
            while [[ -z "$api_endpoint" ]]; do
                print_error "API endpoint is required for automatic download"
                read -p "IBM Storage API endpoint (IP or hostname): " api_endpoint
            done
            
            read -p "Username (Monitor role for download, Admin for mkquorumapp): " username
            while [[ -z "$username" ]]; do
                print_error "Username is required"
                read -p "Username: " username
            done
            
            read -sp "Password: " password
            echo ""
            while [[ -z "$password" ]]; do
                print_error "Password is required"
                read -sp "Password: " password
                echo ""
            done
            
            # mkquorumapp configuration
            read -p "Create new quorum app (mkquorumapp)? (yes/no) [yes]: " mkquorum_input
            mkquorum_input=$(echo "${mkquorum_input:-yes}" | tr '[:upper:]' '[:lower:]')
            if [[ "$mkquorum_input" =~ ^(yes|y|true|1)$ ]]; then
                enable_mkquorumapp="true"
                
                read -p "Partner system name (remote system in PBHA): " partnersystem
                while [[ -z "$partnersystem" ]]; do
                    print_error "Partner system is required for mkquorumapp"
                    read -p "Partner system name: " partnersystem
                done
                
                read -p "Use IPv6 for local system? (yes/no) [no]: " ip6_input
                ip6=$(echo "${ip6_input:-no}" | tr '[:upper:]' '[:lower:]')
                [[ "$ip6" =~ ^(yes|y|true|1)$ ]] && ip6="true" || ip6="false"
                
                read -p "Use IPv6 for partner system? (yes/no) [no]: " partnerip6_input
                partnerip6=$(echo "${partnerip6_input:-no}" | tr '[:upper:]' '[:lower:]')
                [[ "$partnerip6" =~ ^(yes|y|true|1)$ ]] && partnerip6="true" || partnerip6="false"
                
                read -p "Disable metadata? (yes/no) [no]: " nometadata_input
                nometadata=$(echo "${nometadata_input:-no}" | tr '[:upper:]' '[:lower:]')
                [[ "$nometadata" =~ ^(yes|y|true|1)$ ]] && nometadata="true" || nometadata="false"
            fi
        else
            enable_download="false"
        fi
    fi
    
    # Create configuration from template
    if [[ ! -f "$TEMPLATE_FILE" ]]; then
        print_error "Template file not found: $TEMPLATE_FILE"
        exit 1
    fi
    
    print_info "Creating configuration file..."
    local conf_file="${INSTANCES_DIR}/${name}.conf"
    cp "$TEMPLATE_FILE" "$conf_file"
    
    # Replace placeholders
    sed -i "s/<INSTANCE_NAME>/${name}/g" "$conf_file"
    sed -i "s/<DATE>/$(date '+%Y-%m-%d')/g" "$conf_file"
    
    # Update documentation fields (set during interactive mode or leave empty)
    if [[ -n "$storage_system" ]]; then
        sed -i "s|IBM_STORAGE_SYSTEM=\"<HOSTNAME_OR_IP>\"|IBM_STORAGE_SYSTEM=\"${storage_system}\"|g" "$conf_file"
    fi
    if [[ -n "$description" ]]; then
        sed -i "s|IBM_STORAGE_DESCRIPTION=\"\"|IBM_STORAGE_DESCRIPTION=\"${description}\"|g" "$conf_file"
    fi
    if [[ -n "$storage_location" ]]; then
        sed -i "s|IBM_STORAGE_LOCATION=\"\"|IBM_STORAGE_LOCATION=\"${storage_location}\"|g" "$conf_file"
    fi
    if [[ -n "$ipquorum_name" ]]; then
        sed -i "s|IPQUORUM_NAME=\${INSTANCE_NAME}|IPQUORUM_NAME=${ipquorum_name}|g" "$conf_file"
    fi
    
    # Update configuration based on interactive input
    if [[ "$interactive" == "yes" && "$enable_download" == "true" ]]; then
        sed -i "s|<API_ENDPOINT>|${api_endpoint}|g" "$conf_file"
        sed -i "s|<USERNAME>|${username}|g" "$conf_file"
        sed -i "s|IPQUORUM_DOWNLOAD_ENABLED=.*|IPQUORUM_DOWNLOAD_ENABLED=true|g" "$conf_file"
        sed -i "s|IPQUORUM_MKQUORUMAPP_ENABLED=.*|IPQUORUM_MKQUORUMAPP_ENABLED=${enable_mkquorumapp}|g" "$conf_file"
        
        if [[ "$enable_mkquorumapp" == "true" ]]; then
            sed -i "s|IPQUORUM_PARTNERSYSTEM=.*|IPQUORUM_PARTNERSYSTEM=${partnersystem}|g" "$conf_file"
            sed -i "s|IPQUORUM_IP6=.*|IPQUORUM_IP6=${ip6}|g" "$conf_file"
            sed -i "s|IPQUORUM_PARTNERIP6=.*|IPQUORUM_PARTNERIP6=${partnerip6}|g" "$conf_file"
            sed -i "s|IPQUORUM_NOMETADATA=.*|IPQUORUM_NOMETADATA=${nometadata}|g" "$conf_file"
        fi
    fi
    
    print_success "Configuration file created: $conf_file"
    
    # Create password file
    print_info "Creating password file..."
    mkdir -p "${PASSWORDS_DIR}"
    chown ipquorum:ipquorum "${PASSWORDS_DIR}"
    chmod 700 "${PASSWORDS_DIR}"
    
    local pass_file="${PASSWORDS_DIR}/${name}.password"
    if [[ "$interactive" == "yes" && -n "$password" ]]; then
        echo -n "$password" > "$pass_file"
        chmod 440 "$pass_file"
        chown ipquorum:ipquorum "$pass_file"
        print_success "Password file created and configured: $pass_file"
    else
        touch "$pass_file"
        chmod 440 "$pass_file"
        chown ipquorum:ipquorum "$pass_file"
        print_success "Password file created: $pass_file"
        print_warning "Please add your password: echo 'your_password' | sudo tee $pass_file"
    fi
    
    # Create instance directories
    print_info "Creating instance directories..."
    mkdir -p "${DATA_DIR}/${name}"
    mkdir -p "${LOG_DIR}/${name}"
    chown -R ipquorum:ipquorum "${DATA_DIR}/${name}"
    chown -R ipquorum:ipquorum "${LOG_DIR}/${name}"
    print_success "Instance directories created"
    
    # Fix SELinux contexts if SELinux is enabled
    if command -v restorecon &>/dev/null && [[ "$(getenforce 2>/dev/null)" != "Disabled" ]]; then
        print_info "Applying SELinux contexts..."
        restorecon -Rv "${PASSWORDS_DIR}" &>/dev/null || true
        restorecon -Rv "${DATA_DIR}/${name}" &>/dev/null || true
        restorecon -Rv "${LOG_DIR}/${name}" &>/dev/null || true
        print_success "SELinux contexts applied"
    fi
    
    echo ""
    print_success "Instance '$name' created successfully!"
    echo ""
    print_info "Next steps:"
    if [[ "$interactive" == "yes" && "$enable_download" == "true" ]]; then
        echo "1. Review configuration: sudo cat $conf_file"
        echo "2. Enable instance: sudo $0 enable $name"
        echo "3. Start instance: sudo $0 start $name"
        echo "4. Check status: sudo $0 status $name"
        echo "5. View logs: sudo $0 logs $name"
    else
        echo "1. Edit configuration: sudo vi $conf_file"
        echo "2. Set password: echo 'your_password' | sudo tee $pass_file"
        echo "3. Enable instance: sudo $0 enable $name"
        echo "4. Start instance: sudo $0 start $name"
    fi
}

# Delete instance
cmd_delete() {
    check_root
    
    local name="${1:-}"
    
    if [[ -z "$name" ]]; then
        print_error "Instance name is required"
        echo "Usage: $0 delete <instance-name>"
        exit 1
    fi
    
    if ! instance_exists "$name"; then
        print_error "Instance '$name' does not exist"
        exit 1
    fi
    
    # Confirm deletion
    read -p "Are you sure you want to delete instance '$name'? (yes/no): " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        print_info "Deletion cancelled"
        exit 0
    fi
    
    print_info "Deleting instance: $name"
    
    # Stop and disable service
    if systemctl is-active --quiet "ipquorum@${name}.service"; then
        systemctl stop "ipquorum@${name}.service"
        print_success "Service stopped"
    fi
    
    if systemctl is-enabled --quiet "ipquorum@${name}.service" 2>/dev/null; then
        systemctl disable "ipquorum@${name}.service"
        print_success "Service disabled"
    fi
    
    # Remove configuration
    rm -f "${INSTANCES_DIR}/${name}.conf"
    print_success "Configuration removed"
    
    # Remove password file
    rm -f "${PASSWORDS_DIR}/${name}.password"
    print_success "Password file removed"
    
    # Ask about data and logs
    read -p "Delete instance data and logs? (yes/no): " -r
    if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        rm -rf "${DATA_DIR}/${name}"
        rm -rf "${LOG_DIR}/${name}"
        print_success "Data and logs removed"
    else
        print_info "Data and logs preserved in ${DATA_DIR}/${name} and ${LOG_DIR}/${name}"
    fi
    
    print_success "Instance '$name' deleted"
}

# List instances
cmd_list() {
    print_info "Configured IP Quorum Instances:"
    echo ""
    
    if [[ ! -d "$INSTANCES_DIR" ]]; then
        print_warning "No instances directory found"
        return
    fi
    
    local count=0
    printf "%-20s %-15s %-30s %-20s\n" "INSTANCE" "STATUS" "IBM STORAGE SYSTEM" "API ENDPOINT"
    printf "%-20s %-15s %-30s %-20s\n" "--------" "------" "------------------" "------------"
    
    for conf in "${INSTANCES_DIR}"/*.conf; do
        if [[ -f "$conf" && "$conf" != *".template" ]]; then
            local name=$(basename "$conf" .conf)
            local status="unknown"
            local storage_system=""
            local api_endpoint=""
            
            # Get status
            if systemctl is-active --quiet "ipquorum@${name}.service"; then
                status="${GREEN}running${NC}"
            elif systemctl is-enabled --quiet "ipquorum@${name}.service" 2>/dev/null; then
                status="${YELLOW}stopped${NC}"
            else
                status="${RED}disabled${NC}"
            fi
            
            # Get info from config
            if [[ -f "$conf" ]]; then
                storage_system=$(grep "^IBM_STORAGE_SYSTEM=" "$conf" | cut -d'"' -f2 | cut -c1-28)
                api_endpoint=$(grep "^API_ENDPOINT=" "$conf" | cut -d'=' -f2)
                
                # Hide placeholder values - show actual system name or N/A
                if [[ "$storage_system" == "<HOSTNAME_OR_IP>" || "$storage_system" == "" ]]; then
                    storage_system="N/A"
                fi
            fi
            
            printf "%-20s %-24s %-30s %-20s\n" "$name" "$(echo -e "$status")" "$storage_system" "$api_endpoint"
            count=$((count + 1))
        fi
    done
    
    echo ""
    print_info "Total instances: $count"
}

# Show status
cmd_status() {
    local name="${1:-}"
    
    if [[ -z "$name" ]]; then
        # Show all instances
        systemctl status 'ipquorum@*' --no-pager || true
    else
        if ! instance_exists "$name"; then
            print_error "Instance '$name' does not exist"
            exit 1
        fi
        systemctl status "ipquorum@${name}.service" --no-pager
    fi
}

# Start instance
cmd_start() {
    check_root
    
    local name="${1:-}"
    
    if [[ -z "$name" ]]; then
        print_error "Instance name is required"
        exit 1
    fi
    
    if ! instance_exists "$name"; then
        print_error "Instance '$name' does not exist"
        exit 1
    fi
    
    systemctl start "ipquorum@${name}.service"
    print_success "Instance '$name' started"
}

# Stop instance
cmd_stop() {
    check_root
    
    local name="${1:-}"
    
    if [[ -z "$name" ]]; then
        print_error "Instance name is required"
        exit 1
    fi
    
    if ! instance_exists "$name"; then
        print_error "Instance '$name' does not exist"
        exit 1
    fi
    
    systemctl stop "ipquorum@${name}.service"
    print_success "Instance '$name' stopped"
}

# Restart instance
cmd_restart() {
    check_root
    
    local name="${1:-}"
    
    if [[ -z "$name" ]]; then
        print_error "Instance name is required"
        exit 1
    fi
    
    if ! instance_exists "$name"; then
        print_error "Instance '$name' does not exist"
        exit 1
    fi
    
    systemctl restart "ipquorum@${name}.service"
    print_success "Instance '$name' restarted"
}

# Enable instance
cmd_enable() {
    check_root
    
    local name="${1:-}"
    
    if [[ -z "$name" ]]; then
        print_error "Instance name is required"
        exit 1
    fi
    
    if ! instance_exists "$name"; then
        print_error "Instance '$name' does not exist"
        exit 1
    fi
    
    systemctl enable "ipquorum@${name}.service"
    print_success "Instance '$name' enabled (will start on boot)"
}

# Disable instance
cmd_disable() {
    check_root
    
    local name="${1:-}"
    
    if [[ -z "$name" ]]; then
        print_error "Instance name is required"
        exit 1
    fi
    
    if ! instance_exists "$name"; then
        print_error "Instance '$name' does not exist"
        exit 1
    fi
    
    systemctl disable "ipquorum@${name}.service"
    print_success "Instance '$name' disabled (will not start on boot)"
}

# Show logs
cmd_logs() {
    local name="${1:-}"
    local lines="${2:-50}"
    
    if [[ -z "$name" ]]; then
        print_error "Instance name is required"
        exit 1
    fi
    
    if ! instance_exists "$name"; then
        print_error "Instance '$name' does not exist"
        exit 1
    fi
    
    echo "=== Systemd Journal Logs ==="
    journalctl -u "ipquorum@${name}.service" -n "$lines" --no-pager
    
    echo ""
    echo "=== Download Logs ==="
    if [[ -f "${LOG_DIR}/${name}/download.log" ]]; then
        tail -n "$lines" "${LOG_DIR}/${name}/download.log"
    else
        print_warning "No download log found"
    fi
    
    echo ""
    echo "=== IP Quorum Application Logs ==="
    if ls "${LOG_DIR}/${name}"/ip_quorum.log.* 1> /dev/null 2>&1; then
        tail -n "$lines" "${LOG_DIR}/${name}"/ip_quorum.log.*
    else
        print_warning "No application logs found"
    fi
}

# Validate configuration
cmd_validate() {
    local name="${1:-}"
    
    if [[ -z "$name" ]]; then
        print_error "Instance name is required"
        exit 1
    fi
    
    if ! instance_exists "$name"; then
        print_error "Instance '$name' does not exist"
        exit 1
    fi
    
    local conf_file="${INSTANCES_DIR}/${name}.conf"
    print_info "Validating instance: $name"
    
    # Source configuration
    source "$conf_file"
    
    local errors=0
    
    # Check required variables
    if [[ -z "${API_ENDPOINT:-}" ]]; then
        print_error "API_ENDPOINT is not set"
        errors=$((errors + 1))
    else
        print_success "API_ENDPOINT: $API_ENDPOINT"
    fi
    
    if [[ -z "${VIRTUALIZE_USERNAME:-}" ]]; then
        print_error "VIRTUALIZE_USERNAME is not set"
        errors=$((errors + 1))
    else
        print_success "VIRTUALIZE_USERNAME: $VIRTUALIZE_USERNAME"
    fi
    
    if [[ -z "${VIRTUALIZE_PASSWORD_FILE:-}" ]]; then
        print_error "VIRTUALIZE_PASSWORD_FILE is not set"
        errors=$((errors + 1))
    elif [[ ! -f "${VIRTUALIZE_PASSWORD_FILE}" ]]; then
        print_error "Password file not found: ${VIRTUALIZE_PASSWORD_FILE}"
        errors=$((errors + 1))
    elif [[ ! -r "${VIRTUALIZE_PASSWORD_FILE}" ]]; then
        print_error "Password file not readable: ${VIRTUALIZE_PASSWORD_FILE}"
        errors=$((errors + 1))
    else
        print_success "Password file exists and is readable"
    fi
    
    # Check directories
    if [[ ! -d "${DATA_DIR}/${name}" ]]; then
        print_warning "Data directory not found: ${DATA_DIR}/${name}"
    else
        print_success "Data directory exists"
    fi
    
    if [[ ! -d "${LOG_DIR}/${name}" ]]; then
        print_warning "Log directory not found: ${LOG_DIR}/${name}"
    else
        print_success "Log directory exists"
    fi
    
    if [[ $errors -eq 0 ]]; then
        print_success "Configuration is valid"
        return 0
    else
        print_error "Configuration has $errors error(s)"
        return 1
    fi
}

# Show instance info
cmd_info() {
    local name="${1:-}"
    
    if [[ -z "$name" ]]; then
        print_error "Instance name is required"
        exit 1
    fi
    
    if ! instance_exists "$name"; then
        print_error "Instance '$name' does not exist"
        exit 1
    fi
    
    local conf_file="${INSTANCES_DIR}/${name}.conf"
    source "$conf_file"
    
    echo ""
    echo -e "${CYAN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}  Instance Information: ${GREEN}${name}${NC}"
    echo -e "${CYAN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    echo -e "${BLUE}General:${NC}"
    echo "  Instance Name:        $name"
    echo "  IP Quorum Name:       ${IPQUORUM_NAME:-Not set} (shown in IBM Storage Virtualize)"
    echo "  IBM Storage System:   ${IBM_STORAGE_SYSTEM:-Not set}"
    echo "  Description:          ${IBM_STORAGE_DESCRIPTION:-Not set}"
    echo "  Location:             ${IBM_STORAGE_LOCATION:-Not set}"
    echo ""
    
    echo -e "${BLUE}Connection:${NC}"
    echo "  API Endpoint:         ${API_ENDPOINT:-Not set}"
    echo "  Username:             ${VIRTUALIZE_USERNAME:-Not set}"
    echo "  TLS Verify:           ${IPQUORUM_TLS_VERIFY:-false}"
    echo ""
    
    echo -e "${BLUE}Configuration:${NC}"
    echo "  Download Enabled:     ${IPQUORUM_DOWNLOAD_ENABLED:-false}"
    echo "  Download Tool:        ${IPQUORUM_DOWNLOAD_TOOL:-go}"
    echo "  mkquorumapp Enabled:  ${IPQUORUM_MKQUORUMAPP_ENABLED:-false}"
    if [[ "${IPQUORUM_MKQUORUMAPP_ENABLED:-false}" == "true" ]]; then
        echo "  Partner System:       ${IPQUORUM_PARTNERSYSTEM:-Not set}"
    fi
    echo ""
    
    echo -e "${BLUE}Paths:${NC}"
    echo "  Configuration:        $conf_file"
    echo "  Password File:        ${VIRTUALIZE_PASSWORD_FILE:-Not set}"
    echo "  Data Directory:       ${DATA_DIR}/${name}"
    echo "  Log Directory:        ${LOG_DIR}/${name}"
    echo "  JAR File:             ${IPQUORUM_JAR:-Not set}"
    echo ""
    
    echo -e "${BLUE}Service Status:${NC}"
    if systemctl is-active --quiet "ipquorum@${name}.service"; then
        echo -e "  Status:               ${GREEN}Running${NC}"
    else
        echo -e "  Status:               ${RED}Stopped${NC}"
    fi
    
    if systemctl is-enabled --quiet "ipquorum@${name}.service" 2>/dev/null; then
        echo -e "  Enabled:              ${GREEN}Yes${NC}"
    else
        echo -e "  Enabled:              ${RED}No${NC}"
    fi
    echo ""
}

# Update password for an instance
cmd_update_password() {
    check_root
    
    local name="${1:-}"
    
    if [[ -z "$name" ]]; then
        print_error "Instance name is required"
        exit 1
    fi
    
    if ! instance_exists "$name"; then
        print_error "Instance '$name' does not exist"
        exit 1
    fi
    
    local conf_file="${INSTANCES_DIR}/${name}.conf"
    source "$conf_file"
    
    local password_file="${PASSWORDS_DIR}/${name}.password"
    
    echo ""
    echo -e "${CYAN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}  Update Password for Instance: ${GREEN}${name}${NC}"
    echo -e "${CYAN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    print_info "Current username: ${VIRTUALIZE_USERNAME:-Not set}"
    echo ""
    
    # Prompt for new password (hidden input)
    local new_password
    local confirm_password
    
    while true; do
        echo -n "Enter new password: "
        read -rs new_password
        echo ""
        
        if [[ -z "$new_password" ]]; then
            print_error "Password cannot be empty"
            continue
        fi
        
        echo -n "Confirm new password: "
        read -rs confirm_password
        echo ""
        
        if [[ "$new_password" != "$confirm_password" ]]; then
            print_error "Passwords do not match. Please try again."
            echo ""
            continue
        fi
        
        break
    done
    
    # Create password file with secure permissions
    echo "$new_password" > "$password_file"
    chmod 400 "$password_file"
    chown ipquorum:ipquorum "$password_file" 2>/dev/null || true
    
    print_success "Password updated successfully"
    echo ""
    
    # Ask if user wants to restart the service
    local restart_choice
    read -p "Restart the service to apply changes? (yes/no) [yes]: " restart_choice
    restart_choice=$(echo "${restart_choice:-yes}" | tr '[:upper:]' '[:lower:]')
    
    if [[ "$restart_choice" == "yes" || "$restart_choice" == "y" ]]; then
        echo ""
        print_info "Restarting service..."
        systemctl restart "ipquorum@${name}.service"
        
        # Wait a moment and check status
        sleep 2
        if systemctl is-active --quiet "ipquorum@${name}.service"; then
            print_success "Service restarted successfully"
        else
            print_warning "Service may have failed to start. Check logs with: sudo ipquorum logs ${name}"
        fi
    else
        print_info "Password updated but service not restarted"
        print_info "Restart manually with: sudo systemctl restart ipquorum@${name}.service"
    fi
    
    echo ""
}

# Main command dispatcher
main() {
    local command="${1:-}"
    
    if [[ -z "$command" ]]; then
        print_usage
        exit 1
    fi
    
    shift || true
    
    case "$command" in
        create)
            cmd_create "$@"
            ;;
        delete)
            cmd_delete "$@"
            ;;
        list)
            cmd_list "$@"
            ;;
        status)
            cmd_status "$@"
            ;;
        start)
            cmd_start "$@"
            ;;
        stop)
            cmd_stop "$@"
            ;;
        restart)
            cmd_restart "$@"
            ;;
        enable)
            cmd_enable "$@"
            ;;
        disable)
            cmd_disable "$@"
            ;;
        logs)
            cmd_logs "$@"
            ;;
        validate)
            cmd_validate "$@"
            ;;
        info)
            cmd_info "$@"
            ;;
        update-password)
            cmd_update_password "$@"
            ;;
        help|--help|-h)
            print_usage
            ;;
        *)
            print_error "Unknown command: $command"
            print_usage
            exit 1
            ;;
    esac
}

# Run main
main "$@"

# 

# 
