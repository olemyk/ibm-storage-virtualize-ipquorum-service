#!/usr/bin/env bash
#
# IBM Storage Virtualize IP Quorum Download Script (Multi-Instance)
# This script is called by the systemd service to download ip_quorum.jar
# before starting the IP Quorum service.
#
# Usage: ipquorum-download-multi.sh <instance-name>
#
# Configuration is read from /etc/ipquorum/instances/<instance-name>.conf
#

set -euo pipefail

# Get instance name from command line argument
INSTANCE_NAME="${1:-}"

if [[ -z "$INSTANCE_NAME" ]]; then
    echo "ERROR: Instance name is required" >&2
    echo "Usage: $0 <instance-name>" >&2
    exit 1
fi

# Source instance-specific configuration
INSTANCE_CONF="/etc/ipquorum/instances/${INSTANCE_NAME}.conf"

if [[ ! -f "$INSTANCE_CONF" ]]; then
    echo "ERROR: Instance configuration not found: ${INSTANCE_CONF}" >&2
    exit 1
fi

# Source the configuration
source "$INSTANCE_CONF"

# Default configuration (if not set in instance config)
IPQUORUM_DIR="${IPQUORUM_DIR:-/var/lib/ipquorum/${INSTANCE_NAME}}"
IPQUORUM_JAR="${IPQUORUM_JAR:-${IPQUORUM_DIR}/ip_quorum.jar}"
IPQUORUM_LOG_DIR="${IPQUORUM_LOG_DIR:-/var/log/ipquorum/${INSTANCE_NAME}}"
IPQUORUM_DOWNLOAD_ENABLED="${IPQUORUM_DOWNLOAD_ENABLED:-false}"
IPQUORUM_DOWNLOAD_TOOL="${IPQUORUM_DOWNLOAD_TOOL:-go}"
IPQUORUM_BACKUP_ENABLED="${IPQUORUM_BACKUP_ENABLED:-true}"

# API Configuration
API_ENDPOINT="${API_ENDPOINT:-}"
VIRTUALIZE_USERNAME="${VIRTUALIZE_USERNAME:-}"
VIRTUALIZE_PASSWORD_FILE="${VIRTUALIZE_PASSWORD_FILE:-/etc/ipquorum/instances/.passwords/${INSTANCE_NAME}.password}"

# mkquorumapp Configuration
IPQUORUM_MKQUORUMAPP_ENABLED="${IPQUORUM_MKQUORUMAPP_ENABLED:-false}"
IPQUORUM_PARTNERSYSTEM="${IPQUORUM_PARTNERSYSTEM:-}"
IPQUORUM_IP6="${IPQUORUM_IP6:-false}"
IPQUORUM_NOMETADATA="${IPQUORUM_NOMETADATA:-false}"
IPQUORUM_PARTNERIP6="${IPQUORUM_PARTNERIP6:-false}"

# TLS Configuration
IPQUORUM_TLS_VERIFY="${IPQUORUM_TLS_VERIFY:-false}"

# Download tool paths
DOWNLOAD_TOOL_GO="${DOWNLOAD_TOOL_GO:-/usr/local/bin/ipquorum-download-go}"
DOWNLOAD_TOOL_PYTHON="${DOWNLOAD_TOOL_PYTHON:-/usr/local/bin/ipquorum-download.py}"
DOWNLOAD_TOOL_BASH="${DOWNLOAD_TOOL_BASH:-/usr/local/bin/ipquorum-restapi-download.sh}"

# Logging
LOG_FILE="${IPQUORUM_LOG_DIR}/download.log"

# Ensure log directory exists
mkdir -p "$IPQUORUM_LOG_DIR" 2>/dev/null || true

# Function to log messages
log() {
    local level="$1"
    shift
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] [${INSTANCE_NAME}] [$level] $*"
    
    # Try to write to log file, fallback to stdout only if it fails
    if echo "$msg" >> "$LOG_FILE" 2>/dev/null; then
        echo "$msg"
    else
        # If we can't write to log file, just output to stdout
        echo "$msg"
        # Try to create log file with proper permissions on first failure
        if [[ ! -f "$LOG_FILE" ]]; then
            touch "$LOG_FILE" 2>/dev/null || true
        fi
    fi
}

# Function to backup existing JAR file
backup_jar() {
    if [[ -f "$IPQUORUM_JAR" ]]; then
        local backup_file="${IPQUORUM_JAR}.backup"
        log "INFO" "Backing up existing JAR file to ${backup_file}"
        cp -f "$IPQUORUM_JAR" "$backup_file"
        log "INFO" "Backup completed successfully"
    else
        log "INFO" "No existing JAR file to backup"
    fi
}

# Function to restore backup if download fails
restore_backup() {
    local backup_file="${IPQUORUM_JAR}.backup"
    if [[ -f "$backup_file" ]]; then
        log "WARN" "Restoring backup JAR file"
        cp -f "$backup_file" "$IPQUORUM_JAR"
        log "INFO" "Backup restored successfully"
    fi
}

# Function to download using Go binary
download_with_go() {
    log "INFO" "Downloading using Go binary: ${DOWNLOAD_TOOL_GO}"
    
    # Check if the specified path exists and is executable
    if [[ ! -x "$DOWNLOAD_TOOL_GO" ]]; then
        log "WARN" "Go download tool not found at: ${DOWNLOAD_TOOL_GO}"
        
        # Try to find alternative Go binary names in common locations
        local alt_paths=(
            "/usr/local/bin/ipquorum-download-go-linux-amd64"
            "./ipquorum-download-go-linux-amd64"
            "./ipquorum-download-go"
            "/usr/local/bin/ipquorum-download-go"
        )
        
        local found=false
        for alt_path in "${alt_paths[@]}"; do
            if [[ -x "$alt_path" ]]; then
                log "INFO" "Found alternative Go binary at: ${alt_path}"
                DOWNLOAD_TOOL_GO="$alt_path"
                found=true
                break
            fi
        done
        
        if [[ "$found" == "false" ]]; then
            log "ERROR" "Go download tool not found or not executable"
            log "ERROR" "Tried: ${DOWNLOAD_TOOL_GO} and alternatives"
            log "ERROR" "Please install the Go downloader or set DOWNLOAD_TOOL_GO in config"
            return 1
        fi
    fi
    
    # Build command with base options
    local cmd_args=(
        "--api-endpoint" "$API_ENDPOINT"
        "--user" "$VIRTUALIZE_USERNAME"
        "--pass-file" "$VIRTUALIZE_PASSWORD_FILE"
        "--output" "$IPQUORUM_JAR"
        "--download"
    )
    
    # Add mkquorumapp options if enabled
    if [[ "${IPQUORUM_MKQUORUMAPP_ENABLED}" == "true" ]]; then
        log "INFO" "mkquorumapp enabled with partnersystem: ${IPQUORUM_PARTNERSYSTEM}"
        cmd_args+=("--mkquorumapp")
        cmd_args+=("--partnersystem" "$IPQUORUM_PARTNERSYSTEM")
        cmd_args+=("--ip6=$IPQUORUM_IP6")
        cmd_args+=("--nometadata=$IPQUORUM_NOMETADATA")
        cmd_args+=("--partnerip6=$IPQUORUM_PARTNERIP6")
    else
        cmd_args+=("--no-mkquorumapp")
    fi
    
    # Add TLS verification option
    if [[ "${IPQUORUM_TLS_VERIFY}" == "true" ]]; then
        cmd_args+=("--secure")
    else
        cmd_args+=("--insecure")
    fi
    
    "$DOWNLOAD_TOOL_GO" "${cmd_args[@]}"
}

# Function to download using Python script
download_with_python() {
    log "INFO" "Downloading using Python script: ${DOWNLOAD_TOOL_PYTHON}"
    
    if [[ ! -x "$DOWNLOAD_TOOL_PYTHON" ]]; then
        log "ERROR" "Python download tool not found or not executable: ${DOWNLOAD_TOOL_PYTHON}"
        return 1
    fi
    
    # Build command with base options
    local cmd_args=(
        "--api-endpoint" "$API_ENDPOINT"
        "--user" "$VIRTUALIZE_USERNAME"
        "--pass-file" "$VIRTUALIZE_PASSWORD_FILE"
        "--output" "$IPQUORUM_JAR"
        "--download"
    )
    
    # Add mkquorumapp options if enabled
    if [[ "${IPQUORUM_MKQUORUMAPP_ENABLED}" == "true" ]]; then
        log "INFO" "mkquorumapp enabled with partnersystem: ${IPQUORUM_PARTNERSYSTEM}"
        cmd_args+=("--mkquorumapp")
        cmd_args+=("--partnersystem" "$IPQUORUM_PARTNERSYSTEM")
        cmd_args+=("--ip6=$IPQUORUM_IP6")
        cmd_args+=("--nometadata=$IPQUORUM_NOMETADATA")
        cmd_args+=("--partnerip6=$IPQUORUM_PARTNERIP6")
    else
        cmd_args+=("--no-mkquorumapp")
    fi
    
    # Add TLS verification option
    if [[ "${IPQUORUM_TLS_VERIFY}" == "true" ]]; then
        cmd_args+=("--secure")
    else
        cmd_args+=("--insecure")
    fi
    
    "$DOWNLOAD_TOOL_PYTHON" "${cmd_args[@]}"
}

# Function to download using Bash script
download_with_bash() {
    log "INFO" "Downloading using Bash script: ${DOWNLOAD_TOOL_BASH}"
    
    if [[ ! -x "$DOWNLOAD_TOOL_BASH" ]]; then
        log "ERROR" "Bash download tool not found or not executable: ${DOWNLOAD_TOOL_BASH}"
        return 1
    fi
    
    export API_ENDPOINT
    export VIRTUALIZE_USERNAME
    export IPQ_OUTPUT_FILE="$IPQUORUM_JAR"
    
    # Read password from file
    if [[ -f "$VIRTUALIZE_PASSWORD_FILE" ]]; then
        export VIRTUALIZE_PASSWORD=$(cat "$VIRTUALIZE_PASSWORD_FILE")
    else
        log "ERROR" "Password file not found: ${VIRTUALIZE_PASSWORD_FILE}"
        return 1
    fi
    
    # Build command with base options
    local cmd_args=("--download")
    
    # Add mkquorumapp options if enabled
    if [[ "${IPQUORUM_MKQUORUMAPP_ENABLED}" == "true" ]]; then
        log "INFO" "mkquorumapp enabled with partnersystem: ${IPQUORUM_PARTNERSYSTEM}"
        cmd_args+=("--mkquorumapp")
        cmd_args+=("--partnersystem" "$IPQUORUM_PARTNERSYSTEM")
        cmd_args+=("--ip6=$IPQUORUM_IP6")
        cmd_args+=("--nometadata=$IPQUORUM_NOMETADATA")
        cmd_args+=("--partnerip6=$IPQUORUM_PARTNERIP6")
    else
        cmd_args+=("--no-mkquorumapp")
    fi
    
    # Add TLS verification option
    if [[ "${IPQUORUM_TLS_VERIFY}" == "true" ]]; then
        cmd_args+=("--secure")
    else
        cmd_args+=("--insecure")
    fi
    
    "$DOWNLOAD_TOOL_BASH" "${cmd_args[@]}"
}

# Main execution
main() {
    log "INFO" "=== IP Quorum Download Script Started for Instance: ${INSTANCE_NAME} ==="
    log "INFO" "IBM Storage System: ${IBM_STORAGE_SYSTEM:-Unknown}"
    
    # Check if download is enabled
    if [[ "${IPQUORUM_DOWNLOAD_ENABLED}" != "true" ]]; then
        log "INFO" "Download is disabled (IPQUORUM_DOWNLOAD_ENABLED=${IPQUORUM_DOWNLOAD_ENABLED})"
        log "INFO" "Checking if JAR file exists..."
        
        if [[ ! -f "$IPQUORUM_JAR" ]]; then
            log "ERROR" "JAR file not found and download is disabled: ${IPQUORUM_JAR}"
            log "ERROR" "Please enable download or manually place the JAR file"
            exit 1
        fi
        
        log "INFO" "JAR file exists: ${IPQUORUM_JAR}"
        log "INFO" "=== Download Script Completed (Skipped) ==="
        exit 0
    fi
    
    # Validate configuration
    if [[ -z "$API_ENDPOINT" ]]; then
        log "ERROR" "API_ENDPOINT is not configured in ${INSTANCE_CONF}"
        exit 1
    fi
    
    if [[ -z "$VIRTUALIZE_USERNAME" ]]; then
        log "ERROR" "VIRTUALIZE_USERNAME is not configured in ${INSTANCE_CONF}"
        exit 1
    fi
    
    if [[ ! -f "$VIRTUALIZE_PASSWORD_FILE" ]]; then
        log "ERROR" "Password file not found: ${VIRTUALIZE_PASSWORD_FILE}"
        exit 1
    fi
    
    # Check if we can read the password file
    if [[ ! -r "$VIRTUALIZE_PASSWORD_FILE" ]]; then
        log "ERROR" "Cannot read password file: ${VIRTUALIZE_PASSWORD_FILE}"
        log "ERROR" "File exists but current user $(whoami) doesn't have read permission"
        log "ERROR" "Fix with: sudo chown ipquorum:ipquorum ${VIRTUALIZE_PASSWORD_FILE}"
        log "ERROR" "Or: sudo chmod 440 ${VIRTUALIZE_PASSWORD_FILE} && sudo chgrp ipquorum ${VIRTUALIZE_PASSWORD_FILE}"
        exit 1
    fi
    
    # Check password file permissions
    local perms=$(stat -c '%a' "$VIRTUALIZE_PASSWORD_FILE" 2>/dev/null || stat -f '%A' "$VIRTUALIZE_PASSWORD_FILE" 2>/dev/null)
    if [[ "$perms" != "400" && "$perms" != "440" && "$perms" != "600" ]]; then
        log "WARN" "Password file has unusual permissions: ${perms}"
        log "WARN" "Recommended: 440 (readable by owner and group) or 400 (readable by owner only)"
    fi
    
    # Validate mkquorumapp configuration
    if [[ "${IPQUORUM_MKQUORUMAPP_ENABLED}" == "true" ]]; then
        log "INFO" "mkquorumapp is enabled, validating configuration..."
        
        if [[ -z "${IPQUORUM_PARTNERSYSTEM}" ]]; then
            log "ERROR" "IPQUORUM_PARTNERSYSTEM is required when IPQUORUM_MKQUORUMAPP_ENABLED=true"
            log "ERROR" "Please set IPQUORUM_PARTNERSYSTEM in ${INSTANCE_CONF}"
            exit 1
        fi
        
        # Validate boolean values
        for var in IPQUORUM_IP6 IPQUORUM_NOMETADATA IPQUORUM_PARTNERIP6; do
            val="${!var}"
            if [[ "$val" != "true" && "$val" != "false" ]]; then
                log "ERROR" "${var} must be 'true' or 'false' (got: '${val}')"
                exit 1
            fi
        done
        
        log "INFO" "mkquorumapp configuration validated:"
        log "INFO" "  - Partner System: ${IPQUORUM_PARTNERSYSTEM}"
        log "INFO" "  - IPv6: ${IPQUORUM_IP6}"
        log "INFO" "  - No Metadata: ${IPQUORUM_NOMETADATA}"
        log "INFO" "  - Partner IPv6: ${IPQUORUM_PARTNERIP6}"
    fi
    
    # Validate TLS configuration
    if [[ "${IPQUORUM_TLS_VERIFY}" != "true" && "${IPQUORUM_TLS_VERIFY}" != "false" ]]; then
        log "ERROR" "IPQUORUM_TLS_VERIFY must be 'true' or 'false' (got: '${IPQUORUM_TLS_VERIFY}')"
        exit 1
    fi
    log "INFO" "TLS verification: ${IPQUORUM_TLS_VERIFY}"
    
    # Ensure instance directory exists
    mkdir -p "$IPQUORUM_DIR" 2>/dev/null || true
    
    # Create backup if enabled
    if [[ "${IPQUORUM_BACKUP_ENABLED}" == "true" ]]; then
        backup_jar
    fi
    
    # Download based on selected tool
    log "INFO" "Starting download using tool: ${IPQUORUM_DOWNLOAD_TOOL}"
    
    case "${IPQUORUM_DOWNLOAD_TOOL}" in
        go)
            if download_with_go; then
                log "INFO" "Download completed successfully using Go binary"
            else
                log "ERROR" "Download failed using Go binary"
                restore_backup
                exit 1
            fi
            ;;
        python)
            if download_with_python; then
                log "INFO" "Download completed successfully using Python script"
            else
                log "ERROR" "Download failed using Python script"
                restore_backup
                exit 1
            fi
            ;;
        bash)
            if download_with_bash; then
                log "INFO" "Download completed successfully using Bash script"
            else
                log "ERROR" "Download failed using Bash script"
                restore_backup
                exit 1
            fi
            ;;
        *)
            log "ERROR" "Unknown download tool: ${IPQUORUM_DOWNLOAD_TOOL}"
            log "ERROR" "Valid options: go, python, bash"
            exit 1
            ;;
    esac
    
    # Verify downloaded file
    if [[ ! -f "$IPQUORUM_JAR" ]]; then
        log "ERROR" "Download completed but JAR file not found: ${IPQUORUM_JAR}"
        restore_backup
        exit 1
    fi
    
    local file_size=$(stat -c '%s' "$IPQUORUM_JAR" 2>/dev/null || stat -f '%z' "$IPQUORUM_JAR" 2>/dev/null)
    log "INFO" "Downloaded JAR file size: ${file_size} bytes"
    
    if [[ "$file_size" -lt 1000 ]]; then
        log "ERROR" "Downloaded file is too small (${file_size} bytes), likely invalid"
        restore_backup
        exit 1
    fi
    
    # Set proper ownership and permissions
    if [[ -n "${IPQUORUM_USER:-}" ]]; then
        chown "${IPQUORUM_USER}:${IPQUORUM_GROUP:-$IPQUORUM_USER}" "$IPQUORUM_JAR" 2>/dev/null || true
    fi
    chmod 644 "$IPQUORUM_JAR"
    
    log "INFO" "=== Download Script Completed Successfully for Instance: ${INSTANCE_NAME} ==="
    exit 0
}

# Run main function
main "$@"

# 

# 
