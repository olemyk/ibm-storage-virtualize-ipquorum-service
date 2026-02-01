#!/usr/bin/env bash
#
# IBM Storage Virtualize IP Quorum Start Script
# This script builds the Java command with IP Quorum application options
# and starts the IP Quorum service
#

set -euo pipefail

# Source configuration
if [[ -f /etc/ipquorum/ipquorum.conf ]]; then
    source /etc/ipquorum/ipquorum.conf
fi

# Default values if not set in config
JAVA_BIN="${JAVA_BIN:-/usr/bin/java}"
JAVA_OPTS="${JAVA_OPTS:-}"
IPQUORUM_JAR="${IPQUORUM_JAR:-/opt/IBM/ip-quorum/ip_quorum.jar}"
IPQUORUM_NAME="${IPQUORUM_NAME:-}"
IPQUORUM_DEBUG="${IPQUORUM_DEBUG:-false}"
IPQUORUM_EMIT="${IPQUORUM_EMIT:-false}"
IPQUORUM_LOG_LOCATION="${IPQUORUM_LOG_LOCATION:-}"
IPQUORUM_LOG_ROTATION="${IPQUORUM_LOG_ROTATION:-5}"
IPQUORUM_LOG_SIZE="${IPQUORUM_LOG_SIZE:-5120}"

# Build IP Quorum application arguments
IPQUORUM_ARGS=()

# Add -name option if specified
if [[ -n "$IPQUORUM_NAME" ]]; then
    # Validate name: 1-20 characters, A-Z, a-z, 0-9 only
    if [[ "$IPQUORUM_NAME" =~ ^[A-Za-z0-9]{1,20}$ ]]; then
        IPQUORUM_ARGS+=("-name" "$IPQUORUM_NAME")
    else
        echo "WARNING: Invalid IPQUORUM_NAME '$IPQUORUM_NAME'. Must be 1-20 characters (A-Z, a-z, 0-9 only). Ignoring." >&2
    fi
fi

# Add -debug option if enabled
if [[ "$IPQUORUM_DEBUG" == "true" ]]; then
    IPQUORUM_ARGS+=("-debug")
fi

# Add -emit option if enabled
if [[ "$IPQUORUM_EMIT" == "true" ]]; then
    IPQUORUM_ARGS+=("-emit")
fi

# Add -location option if specified
if [[ -n "$IPQUORUM_LOG_LOCATION" ]]; then
    # Validate location: allowed characters [a-z A-Z 0-9 .-_/]
    if [[ "$IPQUORUM_LOG_LOCATION" =~ ^[a-zA-Z0-9._/-]+$ ]]; then
        IPQUORUM_ARGS+=("-location" "$IPQUORUM_LOG_LOCATION")
    else
        echo "WARNING: Invalid IPQUORUM_LOG_LOCATION '$IPQUORUM_LOG_LOCATION'. Allowed characters: [a-z A-Z 0-9 .-_/]. Ignoring." >&2
    fi
fi

# Add -rotation option if not default
if [[ "$IPQUORUM_LOG_ROTATION" != "5" ]]; then
    # Validate rotation: 1-10
    if [[ "$IPQUORUM_LOG_ROTATION" =~ ^[1-9]$|^10$ ]]; then
        IPQUORUM_ARGS+=("-rotation" "$IPQUORUM_LOG_ROTATION")
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
    else
        echo "WARNING: Invalid IPQUORUM_LOG_SIZE '$IPQUORUM_LOG_SIZE'. Must be 1024-10240. Using default (5120)." >&2
    fi
fi

# Build and execute the Java command
# shellcheck disable=SC2086
exec "$JAVA_BIN" $JAVA_OPTS -jar "$IPQUORUM_JAR" "${IPQUORUM_ARGS[@]}"

# Made with help from Bob
