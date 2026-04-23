#!/usr/bin/env bash
#
# IBM Storage Virtualize IP Quorum Validation Script (Multi-Instance)
# This script validates the environment before starting IP Quorum
#
# Usage: ipquorum-validate-multi.sh <instance-name>
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

# Default values if not set in config
JAVA_BIN="${JAVA_BIN:-/usr/bin/java}"
IPQUORUM_JAR="${IPQUORUM_JAR:-/var/lib/ipquorum/${INSTANCE_NAME}/ip_quorum.jar}"
IBM_STORAGE_SYSTEM="${IBM_STORAGE_SYSTEM:-}"

echo "Validating IP Quorum instance: ${INSTANCE_NAME}"

# Check 1: Java binary
if [[ ! -x "$JAVA_BIN" ]]; then
    echo "ERROR: Java binary not found or not executable: ${JAVA_BIN}" >&2
    exit 1
fi
echo "✓ Java binary found: ${JAVA_BIN}"

# Check 2: Java version
JAVA_VERSION=$("$JAVA_BIN" -version 2>&1 | head -n1 | awk -F '"' '{print $2}')
echo "✓ Java version: ${JAVA_VERSION}"

# Check 3: JAR file
if [[ ! -f "$IPQUORUM_JAR" ]]; then
    echo "WARNING: IP Quorum JAR file not found: ${IPQUORUM_JAR}" >&2
    echo "Attempting to download..." >&2
    
    # Try to run download script
    if [[ -x "/usr/local/bin/ipquorum-download-multi.sh" ]]; then
        echo "Running download script..." >&2
        if /usr/local/bin/ipquorum-download-multi.sh "${INSTANCE_NAME}"; then
            echo "✓ Download completed successfully" >&2
        else
            echo "ERROR: Download failed. Please run manually:" >&2
            echo "  sudo /usr/local/bin/ipquorum-download-multi.sh ${INSTANCE_NAME}" >&2
            exit 1
        fi
    else
        echo "ERROR: Download script not found at /usr/local/bin/ipquorum-download-multi.sh" >&2
        echo "Please download the JAR file manually or run:" >&2
        echo "  sudo /usr/local/bin/ipquorum-download-multi.sh ${INSTANCE_NAME}" >&2
        exit 1
    fi
    
    # Verify download succeeded
    if [[ ! -f "$IPQUORUM_JAR" ]]; then
        echo "ERROR: JAR file still not found after download attempt" >&2
        exit 1
    fi
fi
echo "✓ JAR file found: ${IPQUORUM_JAR}"

# Check 4: JAR file size (should be > 100KB)
JAR_SIZE=$(stat -c%s "$IPQUORUM_JAR" 2>/dev/null || stat -f%z "$IPQUORUM_JAR" 2>/dev/null || echo "0")
if [[ "$JAR_SIZE" -lt 102400 ]]; then
    echo "ERROR: JAR file is too small (${JAR_SIZE} bytes). It may be corrupted." >&2
    exit 1
fi
echo "✓ JAR file size: ${JAR_SIZE} bytes"

# Check 5: Working directory
WORK_DIR="/var/lib/ipquorum/${INSTANCE_NAME}"
if [[ ! -d "$WORK_DIR" ]]; then
    echo "ERROR: Working directory not found: ${WORK_DIR}" >&2
    exit 1
fi
echo "✓ Working directory: ${WORK_DIR}"

# Check 6: Log directory
LOG_DIR="/var/log/ipquorum/${INSTANCE_NAME}"
if [[ ! -d "$LOG_DIR" ]]; then
    echo "WARNING: Log directory not found: ${LOG_DIR}" >&2
    echo "Creating log directory..." >&2
    mkdir -p "$LOG_DIR"
    chown ipquorum:ipquorum "$LOG_DIR" 2>/dev/null || true
fi
echo "✓ Log directory: ${LOG_DIR}"

# Check 7: Network connectivity to IBM Storage System (if configured)
if [[ -n "$IBM_STORAGE_SYSTEM" ]]; then
    echo "Testing connectivity to IBM Storage System: ${IBM_STORAGE_SYSTEM}"
    
    # Try to resolve hostname
    if ! getent hosts "$IBM_STORAGE_SYSTEM" >/dev/null 2>&1; then
        echo "WARNING: Cannot resolve hostname: ${IBM_STORAGE_SYSTEM}" >&2
        echo "The IP Quorum service may fail to connect." >&2
    else
        echo "✓ Hostname resolved: ${IBM_STORAGE_SYSTEM}"
    fi
    
    # Try to ping (with timeout)
    if ping -c 1 -W 2 "$IBM_STORAGE_SYSTEM" >/dev/null 2>&1; then
        echo "✓ Network connectivity: ${IBM_STORAGE_SYSTEM} is reachable"
    else
        echo "WARNING: Cannot ping ${IBM_STORAGE_SYSTEM}" >&2
        echo "This may be normal if ICMP is blocked, but verify network connectivity." >&2
    fi
fi

# Check 7b: Port connectivity to storage system (if API endpoint is configured)
if [[ -n "${IPQUORUM_API_ENDPOINT:-}" ]]; then
    echo ""
    echo "Checking port connectivity to storage system: ${IPQUORUM_API_ENDPOINT}"
    
    # Check REST API port (7443)
    echo -n "  Checking REST API port (7443)... "
    if timeout 5 bash -c "cat < /dev/null > /dev/tcp/${IPQUORUM_API_ENDPOINT}/7443" 2>/dev/null; then
        echo "✓ REACHABLE"
    else
        echo "⚠ UNREACHABLE"
        echo "    WARNING: REST API port (7443) is not reachable" >&2
        echo "    This may affect automatic JAR downloads and mkquorumapp functionality" >&2
    fi
    
    # Check IP Quorum port (1260)
    echo -n "  Checking IP Quorum port (1260)... "
    if timeout 5 bash -c "cat < /dev/null > /dev/tcp/${IPQUORUM_API_ENDPOINT}/1260" 2>/dev/null; then
        echo "✓ REACHABLE"
    else
        echo "⚠ UNREACHABLE"
        echo "    WARNING: IP Quorum port (1260) is not reachable" >&2
        echo "    The IP Quorum service may not start successfully" >&2
        echo "    This is normal if IP Quorum is not yet configured on the storage system" >&2
    fi
fi

# Check 8: Port availability (check if another instance is using the same port)
# IP Quorum typically uses ports in the 1100-1200 range
# We'll check if any Java process is already running for this instance
if pgrep -f "ip_quorum.jar.*${INSTANCE_NAME}" >/dev/null 2>&1; then
    echo "WARNING: IP Quorum process already running for instance: ${INSTANCE_NAME}" >&2
    echo "This may cause startup issues. Check running processes:" >&2
    pgrep -fa "ip_quorum.jar.*${INSTANCE_NAME}" || true
fi

echo ""
echo "✓ Validation complete for instance: ${INSTANCE_NAME}"
echo "Ready to start IP Quorum service"
exit 0

# 
