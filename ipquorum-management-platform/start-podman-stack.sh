#!/bin/bash
set -e

# Simple Podman Stack Startup Script
# More reliable than podman-compose for development

echo "🚀 Starting IPQuorum Container Management Stack"
echo "================================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
NETWORK_NAME="ipquorum-network"
POSTGRES_CONTAINER="ipquorum-postgres"
AGENT_CONTAINER="ipquorum-agent"
MANAGER_CONTAINER="ipquorum-manager"

# Step 1: Create network
echo "1. Creating network..."
if podman network exists $NETWORK_NAME 2>/dev/null; then
    echo -e "${YELLOW}Network already exists${NC}"
else
    podman network create $NETWORK_NAME
    echo -e "${GREEN}✓ Network created${NC}"
fi

# Step 2: Start PostgreSQL
echo ""
echo "2. Starting PostgreSQL..."
if podman ps -a --format "{{.Names}}" | grep -q "^${POSTGRES_CONTAINER}$"; then
    echo "Removing old container..."
    podman rm -f $POSTGRES_CONTAINER 2>/dev/null || true
fi

podman run -d \
    --name $POSTGRES_CONTAINER \
    --network $NETWORK_NAME \
    -e POSTGRES_DB=ipquorum_platform \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres123 \
    -p 5432:5432 \
    postgres:14-alpine

echo -e "${GREEN}✓ PostgreSQL started${NC}"

# Wait for PostgreSQL
echo "Waiting for PostgreSQL to be ready..."
sleep 5
for i in {1..30}; do
    if podman exec $POSTGRES_CONTAINER pg_isready -U postgres > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
        break
    fi
    echo -n "."
    sleep 1
done

# Step 3: Apply migrations
echo ""
echo "3. Applying database migrations..."
if [ -d "server/migrations" ]; then
    for migration in server/migrations/*.sql; do
        if [ -f "$migration" ]; then
            echo "Applying $(basename $migration)..."
            podman exec -i $POSTGRES_CONTAINER psql -U postgres -d ipquorum_platform < "$migration" 2>/dev/null || true
        fi
    done
    echo -e "${GREEN}✓ Migrations applied${NC}"
else
    echo -e "${YELLOW}⚠ No migrations directory found${NC}"
fi

# Step 4: Start Container Runtime Agent
echo ""
echo "4. Starting Container Runtime Agent..."
if podman ps -a --format "{{.Names}}" | grep -q "^${AGENT_CONTAINER}$"; then
    echo "Removing old container..."
    podman rm -f $AGENT_CONTAINER 2>/dev/null || true
fi

# Check if agent image exists
if ! podman images | grep -q "ipquorum_agent"; then
    echo "Building agent image..."
    cd container-runtime-agent
    podman build -t ipquorum_agent:latest .
    cd ..
fi

# Note: Agent needs access to host Podman socket - running in privileged mode
podman run -d \
    --name $AGENT_CONTAINER \
    --network $NETWORK_NAME \
    --privileged \
    -p 8081:8080 \
    -p 9090:9090 \
    -v /run/user/$(id -u)/podman/podman.sock:/run/podman/podman.sock:rw \
    -v "$(pwd)/container-runtime-agent/config.yaml:/app/config.yaml:ro" \
    -e IPQUORUM_AGENT_SERVER_PORT=8080 \
    -e IPQUORUM_AGENT_RUNTIME_PREFERRED_RUNTIME=podman \
    -e IPQUORUM_AGENT_RUNTIME_PODMAN_SOCKET_PATH=/run/podman/podman.sock \
    -e IPQUORUM_AGENT_AUTH_ENABLED=true \
    -e IPQUORUM_AGENT_AUTH_JWT_SECRET=test-secret-key \
    -e IPQUORUM_AGENT_LOGGING_LEVEL=debug \
    ipquorum_agent:latest

echo -e "${GREEN}✓ Agent started${NC}"

# Step 5: Start Manager Server
echo ""
echo "5. Starting Manager Server..."
if podman ps -a --format "{{.Names}}" | grep -q "^${MANAGER_CONTAINER}$"; then
    echo "Removing old container..."
    podman rm -f $MANAGER_CONTAINER 2>/dev/null || true
fi

# Force rebuild manager image (with CGO for SQLite)
echo "Rebuilding manager image with CGO support..."
cd server
podman build -t ipquorum_manager:latest .
cd ..
echo -e "${GREEN}✓ Manager image rebuilt${NC}"

# Create data directory for SQLite
mkdir -p "$(pwd)/data"

# Generate TLS certificates if they don't exist
if [ ! -f "$(pwd)/tls/server.crt" ]; then
    echo "Generating TLS certificates..."
    ./generate-certs.sh
fi

podman run -d \
    --name $MANAGER_CONTAINER \
    --network $NETWORK_NAME \
    -p 8080:8080 \
    -v "$(pwd)/data:/data:rw" \
    -v "$(pwd)/tls:/etc/ipquorum-platform/tls:ro" \
    -e DATABASE_HOST=$POSTGRES_CONTAINER \
    -e DATABASE_PORT=5432 \
    -e DATABASE_NAME=ipquorum_platform \
    -e DATABASE_USER=postgres \
    -e DATABASE_PASSWORD=postgres123 \
    -e DATABASE_SSLMODE=disable \
    -e JWT_SECRET=test-secret-key \
    -e SERVER_PORT=8080 \
    -e LOG_LEVEL=debug \
    ipquorum_manager:latest

echo -e "${GREEN}✓ Manager started${NC}"

# Step 6: Show status
echo ""
echo "================================================"
echo -e "${GREEN}✅ Stack started successfully!${NC}"
echo ""
echo "Services:"
echo "  PostgreSQL: localhost:5432"
echo "  Agent API:  http://localhost:8081"
echo "  Manager API: http://localhost:8080"
echo "  Metrics:    http://localhost:9090/metrics"
echo ""
echo "Frontend (run separately):"
echo "  cd web && npm run dev"
echo "  Then open: http://localhost:5173"
echo ""
echo "View logs:"
echo "  podman logs -f $POSTGRES_CONTAINER"
echo "  podman logs -f $AGENT_CONTAINER"
echo "  podman logs -f $MANAGER_CONTAINER"
echo ""
echo "Stop all:"
echo "  podman stop $POSTGRES_CONTAINER $AGENT_CONTAINER $MANAGER_CONTAINER"
echo "  podman rm $POSTGRES_CONTAINER $AGENT_CONTAINER $MANAGER_CONTAINER"
echo ""

# Made with Bob
