#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}IPQuorum Health Check Verification${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to print status
print_status() {
    local status=$1
    local message=$2
    if [[ "$status" == "OK" ]]; then
        echo -e "${GREEN}✓${NC} $message"
    elif [[ "$status" == "WARN" ]]; then
        echo -e "${YELLOW}⚠${NC} $message"
    else
        echo -e "${RED}✗${NC} $message"
    fi
}

# Check if containers are running
echo -e "${BLUE}1. Checking container status...${NC}"
if ! podman ps --format '{{.Names}}' | grep -q "ipquorum-server"; then
    print_status "FAIL" "ipquorum-server container not running"
    echo "Run: podman-compose -f podman-compose.dev.yml up -d"
    exit 1
fi
print_status "OK" "ipquorum-server container is running"

if ! podman ps --format '{{.Names}}' | grep -q "ipquorum-agent"; then
    print_status "FAIL" "ipquorum-agent container not running"
    echo "Run: podman-compose -f podman-compose.dev.yml up -d"
    exit 1
fi
print_status "OK" "ipquorum-agent container is running"
echo ""

# Check health status
echo -e "${BLUE}2. Checking health status...${NC}"
server_health=$(podman inspect ipquorum-server --format '{{.State.Health.Status}}' 2>/dev/null || echo "none")
agent_health=$(podman inspect ipquorum-agent --format '{{.State.Health.Status}}' 2>/dev/null || echo "none")

if [[ "$server_health" == "healthy" ]]; then
    print_status "OK" "Server health: $server_health"
elif [[ "$server_health" == "starting" ]]; then
    print_status "WARN" "Server health: $server_health (waiting for health checks...)"
else
    print_status "FAIL" "Server health: $server_health"
fi

if [[ "$agent_health" == "healthy" ]]; then
    print_status "OK" "Agent health: $agent_health"
elif [[ "$agent_health" == "starting" ]]; then
    print_status "WARN" "Agent health: $agent_health (waiting for health checks...)"
else
    print_status "FAIL" "Agent health: $agent_health"
fi
echo ""

# Check health check configuration
echo -e "${BLUE}3. Checking health check configuration...${NC}"
server_healthcheck=$(podman inspect ipquorum-server --format '{{.Config.Healthcheck.Test}}' 2>/dev/null || echo "none")
agent_healthcheck=$(podman inspect ipquorum-agent --format '{{.Config.Healthcheck.Test}}' 2>/dev/null || echo "none")

if [[ "$server_healthcheck" != "none" && "$server_healthcheck" != "[]" ]]; then
    print_status "OK" "Server healthcheck configured: $server_healthcheck"
else
    print_status "FAIL" "Server healthcheck not configured"
fi

if [[ "$agent_healthcheck" != "none" && "$agent_healthcheck" != "[]" ]]; then
    print_status "OK" "Agent healthcheck configured: $agent_healthcheck"
else
    print_status "FAIL" "Agent healthcheck not configured"
fi
echo ""

# Show recent health check logs
echo -e "${BLUE}4. Recent health check logs...${NC}"
echo -e "${YELLOW}Server:${NC}"
podman inspect ipquorum-server --format '{{range .State.Health.Log}}{{.Start}} - Exit: {{.ExitCode}} - {{.Output}}{{end}}' 2>/dev/null | tail -n 3 || echo "No health logs available"
echo ""
echo -e "${YELLOW}Agent:${NC}"
podman inspect ipquorum-agent --format '{{range .State.Health.Log}}{{.Start}} - Exit: {{.ExitCode}} - {{.Output}}{{end}}' 2>/dev/null | tail -n 3 || echo "No health logs available"
echo ""

# Manual health check test
echo -e "${BLUE}5. Manual health check test...${NC}"
echo -e "${YELLOW}Testing server health endpoint...${NC}"
if podman exec ipquorum-server wget --no-verbose --tries=1 -O /dev/null http://localhost:8080/health 2>&1; then
    print_status "OK" "Server health endpoint responds correctly"
else
    print_status "FAIL" "Server health endpoint failed"
fi

echo -e "${YELLOW}Testing agent health endpoint...${NC}"
if podman exec ipquorum-agent wget --no-verbose --tries=1 -O /dev/null http://localhost:8081/health 2>&1; then
    print_status "OK" "Agent health endpoint responds correctly"
else
    print_status "FAIL" "Agent health endpoint failed"
fi
echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Summary${NC}"
echo -e "${BLUE}========================================${NC}"
podman ps --format 'table {{.Names}}\t{{.Status}}' | grep -E "ipquorum-(server|agent)|NAMES"
echo ""

# Final verdict
if [[ "$server_health" == "healthy" && "$agent_health" == "healthy" ]]; then
    echo -e "${GREEN}✓ All health checks passing!${NC}"
    exit 0
elif [[ "$server_health" == "starting" || "$agent_health" == "starting" ]]; then
    echo -e "${YELLOW}⚠ Health checks still starting. Wait 30-60 seconds and run again.${NC}"
    echo "Run: ./test-healthchecks.sh"
    exit 0
else
    echo -e "${RED}✗ Health checks failing. Review logs above.${NC}"
    echo ""
    echo "Troubleshooting steps:"
    echo "1. Check container logs: podman logs ipquorum-server"
    echo "2. Check health logs: podman inspect ipquorum-server --format '{{json .State.Health}}' | jq"
    echo "3. Rebuild containers: podman-compose -f podman-compose.dev.yml up -d --build"
    echo "4. See PODMAN-HEALTHCHECK-WORKAROUND.md for details"
    exit 1
fi

# Made with Bob
