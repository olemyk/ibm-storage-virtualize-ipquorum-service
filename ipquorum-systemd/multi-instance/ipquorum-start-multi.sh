#!/usr/bin/env bash
#
# IBM Storage Virtualize IP Quorum Start Script (Multi-Instance)
# This script builds the Java command with IP Quorum application options
# and starts the IP Quorum service for a specific instance.
#
# Usage: ipquorum-start-multi.sh <instance-name>
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
JAVA_OPTS="${JAVA_OPTS:-}"
IPQUORUM_JAR="${IPQUORUM_JAR:-/var/lib/ipquorum/${INSTANCE_NAME}/ip_quorum.jar}"
IPQUORUM_NAME="${IPQUORUM_NAME:-${INSTANCE_NAME}}"
IPQUORUM_DEBUG="${IPQUORUM_DEBUG:-false}"
IPQUORUM_EMIT="${IPQUORUM_EMIT:-false}"
IPQUORUM_LOG_LOCATION="${IPQUORUM_LOG_LOCATION:-}"
IPQUORUM_LOG_ROTATION="${IPQUORUM_LOG_ROTATION:-5}"
IPQUORUM_LOG_SIZE="${IPQUORUM_LOG_SIZE:-5120}"

# Verify JAR file exists
if [[ ! -f "$IPQUORUM_JAR" ]]; then
    echo "ERROR: IP Quorum JAR file not found: ${IPQUORUM_JAR}" >&2
    echo "Please ensure the download script completed successfully" >&2
    exit 1
fi

# Verify Java is available
if [[ ! -x "$JAVA_BIN" ]]; then
    echo "ERROR: Java binary not found or not executable: ${JAVA_BIN}" >&2
    exit 1
fi

# Log startup information
echo "Starting IP Quorum instance: ${INSTANCE_NAME}"
echo "IBM Storage System: ${IBM_STORAGE_SYSTEM:-Unknown}"
echo "JAR file: ${IPQUORUM_JAR}"
echo "Java: ${JAVA_BIN}"

# Build IP Quorum application arguments
IPQUORUM_ARGS=()

# Sanitize IPQUORUM_NAME: remove dashes and underscores (not allowed by IP Quorum)
# IP Quorum only allows A-Z, a-z, 0-9 (1-20 characters)
SANITIZED_NAME=$(echo "$IPQUORUM_NAME" | tr -d '_-')

# Add -name option if specified
if [[ -n "$SANITIZED_NAME" ]]; then
    # Validate name: 1-20 characters, A-Z, a-z, 0-9 only
    if [[ "$SANITIZED_NAME" =~ ^[A-Za-z0-9]{1,20}$ ]]; then
        IPQUORUM_ARGS+=("-name" "$SANITIZED_NAME")
        echo "IP Quorum name: ${SANITIZED_NAME}"
    else
        echo "WARNING: Invalid IPQUORUM_NAME '$SANITIZED_NAME'. Must be 1-20 characters (A-Z, a-z, 0-9 only). Ignoring." >&2
    fi
fi

# Add -debug option if enabled
if [[ "$IPQUORUM_DEBUG" == "true" ]]; then
    IPQUORUM_ARGS+=("-debug")
    echo "Debug mode: enabled"
fi

# Add -emit option if enabled
if [[ "$IPQUORUM_EMIT" == "true" ]]; then
    IPQUORUM_ARGS+=("-emit")
    echo "T3 metadata emit: enabled"
fi

# Add -location option if specified
if [[ -n "$IPQUORUM_LOG_LOCATION" ]]; then
    # Validate location: allowed characters [a-z A-Z 0-9 .-_/]
    if [[ "$IPQUORUM_LOG_LOCATION" =~ ^[a-zA-Z0-9._/-]+$ ]]; then
        IPQUORUM_ARGS+=("-location" "$IPQUORUM_LOG_LOCATION")
        echo "Log location: ${IPQUORUM_LOG_LOCATION}"
    else
        echo "WARNING: Invalid IPQUORUM_LOG_LOCATION '$IPQUORUM_LOG_LOCATION'. Allowed characters: [a-z A-Z 0-9 .-_/]. Ignoring." >&2
    fi
fi

# Add -rotation option if not default
if [[ "$IPQUORUM_LOG_ROTATION" != "5" ]]; then
    # Validate rotation: 1-10
    if [[ "$IPQUORUM_LOG_ROTATION" =~ ^[1-9]$|^10$ ]]; then
        IPQUORUM_ARGS+=("-rotation" "$IPQUORUM_LOG_ROTATION")
        echo "Log rotation: ${IPQUORUM_LOG_ROTATION}"
    else
        echo "WARNING: Invalid IPQUORUM_LOG_ROTATION '$IPQUORUM_LOG_ROTATION'. Must be 1-10. Using default (5)." >&2
    fi
fi

# Add -size option if not default
if [[ "$IPQUORUM_LOG_SIZE" != "5120" ]]; then
    # Validate size: 1024-10240
    if [[ "$IPQUORUM_LOG_SIZE" =~ ^[0-9]+$ ]] && \
       [[ "$IPQUORUM_LOG_SIZE" -ge 1024 ]] && \
       [[ "$IPQUORUM_LOG_SIZE" -le 10240 ]]; then
        IPQUORUM_ARGS+=("-size" "$IPQUORUM_LOG_SIZE")
        echo "Log size: ${IPQUORUM_LOG_SIZE} KB"
    else
        echo "WARNING: Invalid IPQUORUM_LOG_SIZE '$IPQUORUM_LOG_SIZE'. Must be 1024-10240. Using default (5120)." >&2
    fi
fi

# Display final command (for debugging)
echo "Starting IP Quorum with arguments: ${IPQUORUM_ARGS[*]}"
echo "=========================================="

# Add retry logic for initial connection issues
MAX_START_RETRIES="${MAX_START_RETRIES:-3}"
RETRY_DELAY="${RETRY_DELAY:-5}"

for attempt in $(seq 1 "$MAX_START_RETRIES"); do
    echo "Start attempt ${attempt}/${MAX_START_RETRIES}..."
    
    # Build and execute the Java command
    # shellcheck disable=SC2086
    if "$JAVA_BIN" $JAVA_OPTS -jar "$IPQUORUM_JAR" "${IPQUORUM_ARGS[@]}"; then
        # If Java exits cleanly (exit code 0), don't retry
        echo "IP Quorum exited cleanly"
        exit 0
    else
        EXIT_CODE=$?
        echo "IP Quorum exited with code: ${EXIT_CODE}"
        
        # If this is not the last attempt, wait and retry
        if [[ $attempt -lt $MAX_START_RETRIES ]]; then
            echo "Waiting ${RETRY_DELAY} seconds before retry..."
            sleep "$RETRY_DELAY"
        else
            echo "ERROR: Failed to start IP Quorum after ${MAX_START_RETRIES} attempts"
            exit "$EXIT_CODE"
        fi
    fi
done

# This should never be reached, but just in case
exit 1




