# Agent API Enhancement Plan

## Objective
Enhance the agent API to support all configuration fields that the interactive bash script (`ipquorum-instance-manager.sh`) collects, achieving 100% feature parity.

---

## Current State

### Current API Request Model (server.go lines 146-150)
```go
var req struct {
    Name        string `json:"name" binding:"required"`
    APIEndpoint string `json:"api_endpoint" binding:"required"`
    Username    string `json:"username" binding:"required"`
    Password    string `json:"password" binding:"required"`
    // Missing fields...
}
```

### Bash Script Collects (lines 235-338)
```bash
# Documentation fields
storage_system=""           # IBM_STORAGE_SYSTEM
description=""              # IBM_STORAGE_DESCRIPTION  
storage_location=""         # IBM_STORAGE_LOCATION
ipquorum_name=""           # IPQUORUM_NAME

# Download configuration
enable_download="true"      # IPQUORUM_DOWNLOAD_ENABLED

# mkquorumapp configuration
enable_mkquorumapp="false"  # IPQUORUM_MKQUORUMAPP_ENABLED
partnersystem=""           # IPQUORUM_PARTNERSYSTEM (MANDATORY if mkquorumapp=true)
ip6="false"                # IPQUORUM_IP6
partnerip6="false"         # IPQUORUM_PARTNERIP6
nometadata="false"         # IPQUORUM_NOMETADATA
```

---

## Required Changes

### 1. Update API Request Model

**File**: `ipquorum-management-platform/agent/internal/api/server.go`

**Current** (lines 146-150):
```go
var req struct {
    Name        string `json:"name" binding:"required"`
    APIEndpoint string `json:"api_endpoint" binding:"required"`
    Username    string `json:"username" binding:"required"`
    Password    string `json:"password" binding:"required"`
}
```

**New** (Enhanced):
```go
var req struct {
    // Required fields
    Name        string `json:"name" binding:"required"`
    APIEndpoint string `json:"api_endpoint" binding:"required"`
    Username    string `json:"username" binding:"required"`
    Password    string `json:"password" binding:"required"`
    
    // Documentation fields (optional)
    StorageSystem      string `json:"storage_system"`       // IBM_STORAGE_SYSTEM
    StorageDescription string `json:"storage_description"`  // IBM_STORAGE_DESCRIPTION
    StorageLocation    string `json:"storage_location"`     // IBM_STORAGE_LOCATION
    IPQuorumName       string `json:"ipquorum_name"`        // IPQUORUM_NAME
    
    // Download configuration (optional, defaults to true)
    DownloadEnabled *bool `json:"download_enabled"`  // IPQUORUM_DOWNLOAD_ENABLED
    
    // mkquorumapp configuration (optional)
    MkQuorumAppEnabled *bool  `json:"mkquorumapp_enabled"`  // IPQUORUM_MKQUORUMAPP_ENABLED
    PartnerSystem      string `json:"partnersystem"`        // IPQUORUM_PARTNERSYSTEM
    IP6                *bool  `json:"ip6"`                  // IPQUORUM_IP6
    PartnerIP6         *bool  `json:"partnerip6"`           // IPQUORUM_PARTNERIP6
    NoMetadata         *bool  `json:"nometadata"`           // IPQUORUM_NOMETADATA
}
```

**Note**: Use pointers for boolean fields to distinguish between "not provided" (nil) and "false" (explicit false).

### 2. Add Validation Logic

**After binding**, add validation:
```go
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}

// Validate: partnersystem is mandatory when mkquorumapp_enabled is true
if req.MkQuorumAppEnabled != nil && *req.MkQuorumAppEnabled {
    if req.PartnerSystem == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "partnersystem is required when mkquorumapp_enabled is true",
        })
        return
    }
}
```

### 3. Update Executor CreateInstance Method

**File**: `ipquorum-management-platform/agent/internal/executor/executor.go`

**Current signature** (line 72):
```go
func (e *Executor) CreateInstance(ctx context.Context, req CreateInstanceRequest) (string, error)
```

**Update CreateInstanceRequest struct**:
```go
type CreateInstanceRequest struct {
    Name               string
    APIEndpoint        string
    Username           string
    Password           string
    StorageSystem      string
    StorageDescription string
    StorageLocation    string
    IPQuorumName       string
    DownloadEnabled    bool
    MkQuorumAppEnabled bool
    PartnerSystem      string
    IP6                bool
    PartnerIP6         bool
    NoMetadata         bool
}
```

### 4. Update updateInstanceConfig Method

**File**: `ipquorum-management-platform/agent/internal/executor/executor.go`

**Current signature** (line 88):
```go
func (e *Executor) updateInstanceConfig(ctx context.Context, name, apiEndpoint, username, password string) error
```

**New signature**:
```go
func (e *Executor) updateInstanceConfig(ctx context.Context, req CreateInstanceRequest) error
```

**Add sed commands for new fields**:
```go
commands := [][]string{
    // Existing commands
    {"sudo", "sed", "-i", fmt.Sprintf("s|^API_ENDPOINT=<API_ENDPOINT>|API_ENDPOINT=%s|g", req.APIEndpoint), configPath},
    {"sudo", "sed", "-i", fmt.Sprintf("s|^VIRTUALIZE_USERNAME=<USERNAME>|VIRTUALIZE_USERNAME=%s|g", req.Username), configPath},
    
    // New: Documentation fields
    {"sudo", "sed", "-i", fmt.Sprintf("s|^IBM_STORAGE_SYSTEM=\"<HOSTNAME_OR_IP>\"|IBM_STORAGE_SYSTEM=\"%s\"|g", req.StorageSystem), configPath},
    {"sudo", "sed", "-i", fmt.Sprintf("s|^IBM_STORAGE_DESCRIPTION=\"\"|IBM_STORAGE_DESCRIPTION=\"%s\"|g", req.StorageDescription), configPath},
    {"sudo", "sed", "-i", fmt.Sprintf("s|^IBM_STORAGE_LOCATION=\"\"|IBM_STORAGE_LOCATION=\"%s\"|g", req.StorageLocation), configPath},
    
    // New: IP Quorum name
    {"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_NAME=.*|IPQUORUM_NAME=%s|g", req.IPQuorumName), configPath},
    
    // New: Download configuration
    {"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_DOWNLOAD_ENABLED=.*|IPQUORUM_DOWNLOAD_ENABLED=%t|g", req.DownloadEnabled), configPath},
    
    // New: mkquorumapp configuration
    {"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_MKQUORUMAPP_ENABLED=.*|IPQUORUM_MKQUORUMAPP_ENABLED=%t|g", req.MkQuorumAppEnabled), configPath},
    {"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_PARTNERSYSTEM=.*|IPQUORUM_PARTNERSYSTEM=%s|g", req.PartnerSystem), configPath},
    {"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_IP6=.*|IPQUORUM_IP6=%t|g", req.IP6), configPath},
    {"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_PARTNERIP6=.*|IPQUORUM_PARTNERIP6=%t|g", req.PartnerIP6), configPath},
    {"sudo", "sed", "-i", fmt.Sprintf("s|^IPQUORUM_NOMETADATA=.*|IPQUORUM_NOMETADATA=%t|g", req.NoMetadata), configPath},
}
```

### 5. Update Handler to Pass Full Request

**File**: `ipquorum-management-platform/agent/internal/api/server.go`

**Current** (after line 150):
```go
output, err := s.executor.CreateInstance(c.Request.Context(), executor.CreateInstanceRequest{
    Name:        req.Name,
    APIEndpoint: req.APIEndpoint,
    Username:    req.Username,
    Password:    req.Password,
})
```

**New**:
```go
// Set defaults for optional fields
downloadEnabled := true
if req.DownloadEnabled != nil {
    downloadEnabled = *req.DownloadEnabled
}

mkquorumappEnabled := false
if req.MkQuorumAppEnabled != nil {
    mkquorumappEnabled = *req.MkQuorumAppEnabled
}

ip6 := false
if req.IP6 != nil {
    ip6 = *req.IP6
}

partnerip6 := false
if req.PartnerIP6 != nil {
    partnerip6 = *req.PartnerIP6
}

nometadata := false
if req.NoMetadata != nil {
    nometadata = *req.NoMetadata
}

// Set default storage system if not provided
storageSystem := req.StorageSystem
if storageSystem == "" {
    storageSystem = "N/A"
}

// Set default IPQuorum name if not provided (sanitize instance name)
ipquorumName := req.IPQuorumName
if ipquorumName == "" {
    // Remove dashes and underscores, truncate to 20 chars
    ipquorumName = strings.Map(func(r rune) rune {
        if r == '-' || r == '_' {
            return -1
        }
        return r
    }, req.Name)
    if len(ipquorumName) > 20 {
        ipquorumName = ipquorumName[:20]
    }
}

output, err := s.executor.CreateInstance(c.Request.Context(), executor.CreateInstanceRequest{
    Name:               req.Name,
    APIEndpoint:        req.APIEndpoint,
    Username:           req.Username,
    Password:           req.Password,
    StorageSystem:      storageSystem,
    StorageDescription: req.StorageDescription,
    StorageLocation:    req.StorageLocation,
    IPQuorumName:       ipquorumName,
    DownloadEnabled:    downloadEnabled,
    MkQuorumAppEnabled: mkquorumappEnabled,
    PartnerSystem:      req.PartnerSystem,
    IP6:                ip6,
    PartnerIP6:         partnerip6,
    NoMetadata:         nometadata,
})
```

---

## Example API Requests

### Minimal (Download Only)
```json
{
  "name": "prod-cluster01",
  "api_endpoint": "10.33.7.80",
  "username": "monitor_user",
  "password": "secret123"
}
```

### With Documentation
```json
{
  "name": "prod-cluster01",
  "api_endpoint": "10.33.7.80",
  "username": "superuser",
  "password": "secret123",
  "storage_system": "svc_cluster01",
  "storage_description": "IBM FlashSystem 9200 - Production Site A",
  "storage_location": "Datacenter A, Rack 12",
  "ipquorum_name": "prodquorum01"
}
```

### Full (Create Quorum App)
```json
{
  "name": "prod-cluster01",
  "api_endpoint": "10.33.7.80",
  "username": "admin",
  "password": "secret123",
  "storage_system": "svc_cluster01",
  "storage_description": "IBM FlashSystem 9200 - Production Site A",
  "storage_location": "Datacenter A, Rack 12",
  "download_enabled": true,
  "mkquorumapp_enabled": true,
  "partnersystem": "svc_cluster02",
  "ip6": false,
  "partnerip6": false,
  "nometadata": false
}
```

---

## Testing Checklist

- [ ] Minimal request (only required fields)
- [ ] Request with documentation fields
- [ ] Request with mkquorumapp_enabled=true and partnersystem
- [ ] Request with mkquorumapp_enabled=true but NO partnersystem (should fail validation)
- [ ] Request with all optional fields
- [ ] Verify config file has correct values for all fields
- [ ] Verify defaults are applied correctly

---

## Implementation Order

1. ✅ Update CreateInstanceRequest struct in executor.go
2. ✅ Update updateInstanceConfig signature and implementation
3. ✅ Update API request model in server.go
4. ✅ Add validation logic in handler
5. ✅ Update handler to pass full request with defaults
6. ✅ Build and test
7. ✅ Update documentation

---

## Files to Modify

1. `ipquorum-management-platform/agent/internal/executor/executor.go`
   - Update CreateInstanceRequest struct
   - Update updateInstanceConfig method signature
   - Add sed commands for new fields

2. `ipquorum-management-platform/agent/internal/api/server.go`
   - Update request struct in handleCreateInstance
   - Add validation logic
   - Update executor call with full request

3. `ipquorum-management-platform/agent/CONFIG-FIELDS-REFERENCE.md`
   - Update API Request Mapping table
   - Add new example requests

---

## Notes

- Use pointer types (`*bool`) for optional boolean fields to distinguish nil from false
- Apply sensible defaults in the handler before calling executor
- Validate mandatory relationships (e.g., partnersystem required when mkquorumapp_enabled=true)
- Sanitize IPQuorumName (remove dashes/underscores, max 20 chars)
- Set StorageSystem to "N/A" if not provided (matches bash script behavior)