#!/bin/bash
set -euo pipefail

# IPQuorum Container Entrypoint Script
# Handles JAR download and Java application startup

# Configuration from environment variables
API_ENDPOINT="${API_ENDPOINT:-}"
API_PORT="${API_PORT:-}"
VIRTUALIZE_USERNAME="${VIRTUALIZE_USERNAME:-}"
VIRTUALIZE_PASSWORD="${VIRTUALIZE_PASSWORD:-}"
PARTNERSYSTEM="${PARTNERSYSTEM:-}"
IP6="${IP6:-false}"
PARTNERIP6="${PARTNERIP6:-false}"
NOMETADATA="${NOMETADATA:-false}"
INSECURE="${INSECURE:-true}"
DOWNLOAD_JAR="${DOWNLOAD_JAR:-true}"
CREATE_QUORUM_APP="${CREATE_QUORUM_APP:-false}"

JAR_PATH="/opt/ipquorum/jar/ip_quorum.jar"
LOG_DIR="/opt/ipquorum/logs"
DOWNLOADER="/opt/ipquorum/bin/ipquorum-download-go"

# Logging functions
log_info() {
    echo "[$(date -Iseconds)] [INFO] $*"
}

log_success() {
    echo "[$(date -Iseconds)] [SUCCESS] $*"
}

log_warning() {
    echo "[$(date -Iseconds)] [WARNING] $*"
}

log_error() {
    echo "[$(date -Iseconds)] [ERROR] $*" >&2
}

# Print startup banner
log_info "=========================================="
log_info "IPQuorum Container Starting"
log_info "=========================================="
log_info "Configuration:"
log_info "  API_ENDPOINT: ${API_ENDPOINT}"
log_info "  API_PORT: ${API_PORT:-default (7443)}"
log_info "  USERNAME: ${VIRTUALIZE_USERNAME}"
log_info "  PARTNERSYSTEM: ${PARTNERSYSTEM}"
log_info "  DOWNLOAD_JAR: ${DOWNLOAD_JAR}"
log_info "  CREATE_QUORUM_APP: ${CREATE_QUORUM_APP}"
log_info "  INSECURE: ${INSECURE}"
log_info "  IP6: ${IP6}"
log_info "  PARTNERIP6: ${PARTNERIP6}"
log_info "  NOMETADATA: ${NOMETADATA}"
log_info "=========================================="

# Function to download JAR
download_jar() {
    log_info "Downloading ip_quorum.jar..."
    
    # Build arguments array
    local args=(
        --api-endpoint "${API_ENDPOINT}"
        --user "${VIRTUALIZE_USERNAME}"
        --pass "${VIRTUALIZE_PASSWORD}"
        --output "${JAR_PATH}"
    )
    
    # Add optional flags
    if [[ -n "${API_PORT}" ]]; then
        args+=(--api-port "${API_PORT}")
    fi

    if [[ "${INSECURE}" == "true" ]]; then
        args+=(--insecure)
    else
        args+=(--secure)
    fi
    
    # Add mkquorumapp flags
    if [[ "${CREATE_QUORUM_APP}" == "true" ]]; then
        args+=(--mkquorumapp)
        
        # Validate partner system is provided
        if [[ -z "${PARTNERSYSTEM}" ]]; then
            log_error "PARTNERSYSTEM is required when CREATE_QUORUM_APP=true"
            return 1
        fi
        
        args+=(--partnersystem "${PARTNERSYSTEM}")
        args+=(--ip6="${IP6}")
        args+=(--partnerip6="${PARTNERIP6}")
        args+=(--nometadata="${NOMETADATA}")
    else
        args+=(--no-mkquorumapp)
    fi
    
    # Add download flag
    if [[ "${DOWNLOAD_JAR}" == "true" ]]; then
        args+=(--download)
    else
        args+=(--no-download)
    fi
    
    # Log command (mask password)
    log_info "Running: ipquorum-download-go ${args[*]//${VIRTUALIZE_PASSWORD}/***}"
    
    # Run downloader
    if "${DOWNLOADER}" "${args[@]}"; then
        log_success "Download successful"
        return 0
    else
        log_error "Download failed"
        return 1
    fi
}

# Validate required configuration
validate_config() {
    local errors=0
    
    if [[ -z "${API_ENDPOINT}" ]]; then
        log_error "API_ENDPOINT is required"
        errors=$((errors + 1))
    fi
    
    if [[ -z "${VIRTUALIZE_USERNAME}" ]]; then
        log_error "VIRTUALIZE_USERNAME is required"
        errors=$((errors + 1))
    fi
    
    if [[ -z "${VIRTUALIZE_PASSWORD}" ]]; then
        log_error "VIRTUALIZE_PASSWORD is required"
        errors=$((errors + 1))
    fi
    
    if [[ "${CREATE_QUORUM_APP}" == "true" && -z "${PARTNERSYSTEM}" ]]; then
        log_error "PARTNERSYSTEM is required when CREATE_QUORUM_APP=true"
        errors=$((errors + 1))
    fi
    
    if [[ $errors -gt 0 ]]; then
        log_error "Configuration validation failed with $errors error(s)"
        return 1
    fi
    
    log_success "Configuration validated"
    return 0
}

# Check if downloader exists
if [[ ! -x "${DOWNLOADER}" ]]; then
    log_error "ipquorum-download-go not found at ${DOWNLOADER}"
    exit 1
fi

# Validate configuration
if ! validate_config; then
    log_error "Exiting due to configuration errors"
    exit 1
fi

# Check if JAR exists
if [[ ! -f "${JAR_PATH}" ]]; then
    log_info "JAR file not found at ${JAR_PATH}"
    
    if [[ "${DOWNLOAD_JAR}" == "true" ]]; then
        if ! download_jar; then
            log_error "Failed to download JAR file"
            exit 1
        fi
    else
        log_error "JAR file required but DOWNLOAD_JAR=false"
        log_error "Please mount a volume with ip_quorum.jar at ${JAR_PATH}"
        exit 1
    fi
else
    log_info "Using existing JAR file: ${JAR_PATH}"
    
    # Check if we should re-download (force update)
    if [[ "${FORCE_DOWNLOAD:-false}" == "true" ]]; then
        log_warning "FORCE_DOWNLOAD=true, re-downloading JAR..."
        if ! download_jar; then
            log_error "Failed to re-download JAR file"
            exit 1
        fi
    fi
fi

# Validate JAR file
if [[ ! -s "${JAR_PATH}" ]]; then
    log_error "JAR file is empty: ${JAR_PATH}"
    exit 1
fi

JAR_SIZE=$(stat -c%s "${JAR_PATH}" 2>/dev/null || stat -f%z "${JAR_PATH}" 2>/dev/null)
log_success "JAR file validated (size: ${JAR_SIZE} bytes)"

# Verify JAR is a valid ZIP file
if ! unzip -t "${JAR_PATH}" >/dev/null 2>&1; then
    log_error "JAR file appears to be corrupted (not a valid ZIP)"
    exit 1
fi

log_success "JAR file integrity check passed"

# Prepare Java command
JAVA_OPTS="${JAVA_OPTS:--Xmx512m -Xms256m}"
LOG_FILE="${LOG_DIR}/ipquorum-$(date +%Y%m%d-%H%M%S).log"

log_info "=========================================="
log_info "Starting IPQuorum Application"
log_info "=========================================="
log_info "Java options: ${JAVA_OPTS}"
log_info "Log file: ${LOG_FILE}"
log_info "JAR path: ${JAR_PATH}"
log_info "=========================================="

# Trap signals for graceful shutdown
shutdown_handler() {
    local signal=$1
    log_warning "Received ${signal}, initiating graceful shutdown..."
    
    if [[ -n "${JAVA_PID:-}" ]]; then
        log_info "Sending SIGTERM to Java process (PID: ${JAVA_PID})..."
        kill -TERM "${JAVA_PID}" 2>/dev/null || true
        
        # Wait for process to exit (max 30 seconds)
        local wait_time=0
        while kill -0 "${JAVA_PID}" 2>/dev/null && [[ $wait_time -lt 30 ]]; do
            sleep 1
            wait_time=$((wait_time + 1))
        done
        
        if kill -0 "${JAVA_PID}" 2>/dev/null; then
            log_warning "Process did not exit gracefully, forcing shutdown..."
            kill -KILL "${JAVA_PID}" 2>/dev/null || true
        else
            log_success "Process exited gracefully"
        fi
    fi
    
    exit 0
}

trap 'shutdown_handler SIGTERM' TERM
trap 'shutdown_handler SIGINT' INT

# Start Java application
log_info "Executing: java ${JAVA_OPTS} -jar ${JAR_PATH}"

# Run Java in background and capture PID
java ${JAVA_OPTS} -jar "${JAR_PATH}" 2>&1 | tee "${LOG_FILE}" &
JAVA_PID=$!

log_success "IPQuorum started with PID: ${JAVA_PID}"

# Wait for process
wait ${JAVA_PID}
EXIT_CODE=$?

log_info "=========================================="
log_info "IPQuorum exited with code: ${EXIT_CODE}"
log_info "=========================================="

exit ${EXIT_CODE}

# Made with Bob
