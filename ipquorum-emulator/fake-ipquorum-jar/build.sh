#!/usr/bin/env bash
# Builds the fake ip_quorum.jar template (no quorum.properties inside; the
# mock-flashsystem adds a per-cluster one when the JAR is downloaded).
#
# Usage: build.sh [output-jar]   (default: build/fake-ip_quorum.jar)
# Requires a JDK (javac, jar). Targets Java 17, the runner image's JRE.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="${1:-${SCRIPT_DIR}/build/fake-ip_quorum.jar}"
CLASSES="$(mktemp -d)"
trap 'rm -rf "${CLASSES}"' EXIT

mkdir -p "$(dirname "${OUT}")"
OUT="$(cd "$(dirname "${OUT}")" && pwd)/$(basename "${OUT}")"

javac --release 17 -Xlint:all -Werror -d "${CLASSES}" "${SCRIPT_DIR}"/src/ipquorum/*.java
jar --create --file "${OUT}" --main-class ipquorum.FakeIpQuorum -C "${CLASSES}" .
echo "Built ${OUT}"
