# IPQuorum Containers Feature - Direct Agent API Access

## Overview

Added a new feature to create IPQuorum containers directly via the agent API, bypassing the systemd instance requirement. This solves the problem where containerized management servers cannot manage systemd instances.

## Problem Solved

**Original Issue**: The management server container cannot create systemd instances because:
- No bash scripts in container
- No systemd in Alpine Linux
- Non-root user cannot manage systemd
- Scripts directory `/opt/ipquorum/scripts` doesn't exist

**Solution**: New "IPQuorum Containers" feature that creates containers directly via agent API without requiring instances.

## New Features

### 1. IPQuorum Containers Page

**Route**: `/ipquorum-containers`

**Features**:
- Landing page explaining direct container creation
- "Create IPQuorum Container" button
- Links to documentation
- Info about bypassing instance requirement

### 2. IPQuorum Container Creation Form

**Route**: `/ipquorum-containers/new`

**Component**: `IPQuorumContainerForm` (already existed, now accessible)

**Features**:
- Full IPQuorum configuration (no instance needed)
- Storage Virtualize connection settings
- Quorum app configuration
- Network settings
- Download options
- Advanced container options

**Fields**:
- Container Name
- Agent ID (select from registered agents)
- API Endpoint (SVC system IP)
- Username/Password
- Partner System
- IPv6 options
- Port configuration
- Java options

### 3. Navigation

Added "IPQuorum Containers" link in sidebar navigation with Box icon.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│         Management Web UI                            │
│                                                       │
│  New Route: /ipquorum-containers/new                │
│  Component: IPQuorumContainerForm                    │
│                                                       │
│  Sends request to: /api/ipquorum/containers         │
└────────────────────┬────────────────────────────────┘
                     │
                     │ HTTP POST
                     ↓
┌─────────────────────────────────────────────────────┐
│         Management Server API                        │
│                                                       │
│  Endpoint: POST /api/ipquorum/containers            │
│  Handler: Creates container via agent client        │
└────────────────────┬────────────────────────────────┘
                     │
                     │ Agent API Call
                     ↓
┌─────────────────────────────────────────────────────┐
│         Container Runtime Agent                      │
│                                                       │
│  Endpoint: POST /api/v1/containers                  │
│  Action: Creates IPQuorum container with config     │
└─────────────────────────────────────────────────────┘
```

## Usage Workflow

### Step 1: Register Agent Server

1. Go to **Servers** page
2. Click "Add Server"
3. Fill in:
   - Hostname: `agentserver`
   - IP Address: `ipquorum-agent`
   - Port: `8081`
   - API Key: `dev-secret-change-in-production`
4. Submit

### Step 2: Create IPQuorum Container

1. Go to **IPQuorum Containers** page (new!)
2. Click "Create IPQuorum Container"
3. Fill in the form:

   **Basic Configuration:**
   - Container Name: `ipquorum-prod-01`
   - Agent ID: Select `agentserver`

   **Storage Virtualize Connection:**
   - API Endpoint: `10.33.7.81`
   - Username: `superuser`
   - Password: `your-password`
   - Insecure TLS: ✓ (if using self-signed certs)

   **Quorum App Configuration:**
   - Create Quorum App: ✓ (if needed)
   - Partner System: `svc_cluster02`
   - IPv6 options as needed

   **Network Configuration:**
   - Port: `1260` (default)

   **Download Options:**
   - Download JAR: ✓

4. Click "Create Container"

### Step 3: Monitor Container

The container will be created directly on the agent and will appear in the regular Containers list.

## API Endpoints

### Frontend API (`/api/ipquorum/containers`)

Already implemented in `web/src/api/ipquorum.ts`:

```typescript
// Create IPQuorum container
POST /api/ipquorum/containers
Body: CreateIPQuorumContainerRequest

// List IPQuorum containers
GET /api/ipquorum/containers

// Get container details
GET /api/ipquorum/containers/:id

// Container operations
POST /api/ipquorum/containers/:id/start
POST /api/ipquorum/containers/:id/stop
POST /api/ipquorum/containers/:id/restart
DELETE /api/ipquorum/containers/:id

// Logs and stats
GET /api/ipquorum/containers/:id/logs
GET /api/ipquorum/containers/:id/stats
```

### Agent API (`/api/v1/containers`)

Container runtime agent endpoints:

```
POST /api/v1/containers          - Create container
GET /api/v1/containers           - List containers
GET /api/v1/containers/:id       - Inspect container
POST /api/v1/containers/:id/start    - Start container
POST /api/v1/containers/:id/stop     - Stop container
POST /api/v1/containers/:id/restart  - Restart container
DELETE /api/v1/containers/:id    - Remove container
GET /api/v1/containers/:id/logs  - Get logs
GET /api/v1/containers/:id/stats - Get stats
```

## Files Modified

### Frontend

1. **`web/src/pages/IPQuorumContainers.tsx`** (NEW)
   - Landing page for IPQuorum containers
   - Explains direct creation workflow

2. **`web/src/components/IPQuorumContainerForm.tsx`** (EXISTING)
   - Already had full IPQuorum configuration form
   - Now accessible via new route

3. **`web/src/App.tsx`**
   - Added routes for `/ipquorum-containers` and `/ipquorum-containers/new`
   - Imported IPQuorumContainerForm component

4. **`web/src/components/Layout.tsx`**
   - Added "IPQuorum Containers" navigation link
   - Added Box icon import

5. **`web/src/pages/InstanceForm.tsx`**
   - Fixed: Added `deployment_type: 'systemd'` field

### Backend

No backend changes needed - the `/api/ipquorum/containers` endpoints already exist!

### Deployment

6. **`deploy-to-linux.sh`**
   - Added new documentation files to deployment

## Documentation Created

1. **`SYSTEMD-NOT-SUPPORTED-IN-CONTAINER.md`**
   - Explains why systemd doesn't work in containers
   - Three solution options
   - Verification commands

2. **`AGENT-ARCHITECTURE-CLARIFICATION.md`**
   - Architecture mismatch explanation
   - Two agent concepts (systemd vs container)
   - Correct workflows

3. **`CONTAINER-CREATION-WORKFLOW.md`**
   - Step-by-step workflow guide
   - Architecture diagrams
   - Troubleshooting tips

4. **`IPQUORUM-CONTAINERS-FEATURE.md`** (this file)
   - Complete feature documentation

## Benefits

### ✅ Advantages

1. **No Instance Required**: Create containers directly without systemd instances
2. **Container-Native**: Perfect for containerized deployments
3. **Simpler Workflow**: One-step container creation
4. **Agent API Direct**: Uses agent API as intended
5. **Full Configuration**: All IPQuorum options available
6. **Existing Components**: Reuses existing form and API

### ⚠️ Limitations

1. **No Instance Management**: Containers aren't linked to instances
2. **Container List**: Currently shows in regular Containers page (future: separate list)
3. **Agent Required**: Must have registered agent server
4. **Manual Config**: Must enter all IPQuorum settings manually

## Testing

### Deploy and Test

```bash
# From local machine
cd ipquorum-management-platform
./deploy-to-linux.sh

# On RHEL server
ssh packer@10.33.3.215
cd ~/ipquorum-test
./setup.sh
./start-web-dev.sh

# Wait for containers to start
sleep 60
./test-healthchecks.sh
```

### Access the Feature

1. Open browser: `https://10.33.3.215:3443`
2. Login with credentials
3. Click "IPQuorum Containers" in sidebar
4. Click "Create IPQuorum Container"
5. Fill in form and submit

### Verify Container

```bash
# Check container was created
podman exec ipquorum-agent podman ps

# Check container logs
podman exec ipquorum-agent podman logs <container-name>

# Check IPQuorum is running
curl http://ipquorum-agent:1260
```

## Future Enhancements

### Phase 1: Container List
- Dedicated list view for IPQuorum containers
- Filter by agent
- Quick actions (start/stop/restart)

### Phase 2: Templates
- Save IPQuorum configurations as templates
- Quick deploy from templates
- Share templates between users

### Phase 3: Bulk Operations
- Create multiple containers at once
- Bulk start/stop/restart
- Container groups

### Phase 4: Monitoring
- Real-time container stats
- IPQuorum health checks
- Alert on failures

## Comparison: Old vs New Workflow

### Old Workflow (Doesn't Work in Containers)

```
1. Servers → Register agent ✓
2. Instances → Create instance ✗ (fails - no systemd)
3. Containers → Create container ✗ (can't proceed)
```

### New Workflow (Works!)

```
1. Servers → Register agent ✓
2. IPQuorum Containers → Create container ✓ (direct!)
3. Done! ✓
```

## Summary

The new IPQuorum Containers feature provides a **container-native way** to deploy IPQuorum applications without requiring systemd instances. This is the **correct solution** for containerized management server deployments.

**Key Point**: The management server container was never meant to manage systemd instances. It's designed to manage containers via agents. This feature makes that clear and provides the proper workflow.

---

**Made with Bob** - IPQuorum Management Platform