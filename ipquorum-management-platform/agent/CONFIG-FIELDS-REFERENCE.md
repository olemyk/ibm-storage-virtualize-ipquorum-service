# IPQuorum Instance Configuration Fields Reference

## Overview
This document describes all configuration fields in the instance config file (`/etc/ipquorum/instances/<name>.conf`), indicating which are mandatory, optional, and automatically managed by the agent.

---

## Field Categories

### 🔴 MANDATORY - Must Be Provided
These fields are required for the instance to function.

### 🟡 OPTIONAL - Recommended
These fields enhance functionality or documentation but aren't strictly required.

### 🟢 AUTO-MANAGED - Set by Agent
These fields are automatically set by the agent during instance creation.

### ⚪ SYSTEM DEFAULT - Usually Not Changed
These fields have sensible defaults and rarely need modification.

---

## Configuration Fields

### INSTANCE IDENTIFICATION

#### `INSTANCE_NAME` 🟢 AUTO-MANAGED
- **Description**: Unique identifier for this instance
- **Example**: `prod-cluster01`, `test-system`
- **Set by**: Agent during creation
- **Format**: Alphanumeric, dashes, underscores allowed

#### `IBM_STORAGE_SYSTEM` 🟡 OPTIONAL
- **Description**: Human-readable name of the IBM Storage system (for documentation only)
- **Example**: `"svc_cluster01"`, `"Production-SAN"`, `"DataCenter-A-Storage"`
- **Default**: `"N/A"` (set by agent if not provided)
- **Purpose**: Appears in `ipqm list` output for identification
- **Note**: NOT used for connectivity - purely for human readability

#### `IBM_STORAGE_DESCRIPTION` 🟡 OPTIONAL
- **Description**: Detailed description of the storage system
- **Example**: `"IBM FlashSystem 9200 - Production Site A"`
- **Default**: Empty string
- **Purpose**: Additional documentation

#### `IBM_STORAGE_LOCATION` 🟡 OPTIONAL
- **Description**: Physical location of the storage system
- **Example**: `"Datacenter A, Rack 12"`
- **Default**: Empty string
- **Purpose**: Documentation and asset tracking

---

### BASIC CONFIGURATION

#### `IPQUORUM_DIR` ⚪ SYSTEM DEFAULT
- **Description**: Instance-specific installation directory
- **Default**: `/var/lib/ipquorum/${INSTANCE_NAME}`
- **Purpose**: Stores JAR file and instance data
- **Rarely changed**: Uses variable substitution

#### `IPQUORUM_JAR` ⚪ SYSTEM DEFAULT
- **Description**: Path to IP Quorum JAR file
- **Default**: `${IPQUORUM_DIR}/ip_quorum.jar`
- **Rarely changed**: Derived from IPQUORUM_DIR

#### `IPQUORUM_LOG_DIR` ⚪ SYSTEM DEFAULT
- **Description**: Instance-specific log directory
- **Default**: `/var/log/ipquorum/${INSTANCE_NAME}`
- **Purpose**: Stores service logs
- **Rarely changed**: Uses variable substitution

#### `IPQUORUM_USER` ⚪ SYSTEM DEFAULT
- **Description**: User to run the service
- **Default**: `ipquorum`
- **Rarely changed**: System user created during installation

#### `IPQUORUM_GROUP` ⚪ SYSTEM DEFAULT
- **Description**: Group to run the service
- **Default**: `ipquorum`
- **Rarely changed**: System group created during installation

#### `IPQUORUM_NAME` ⚪ SYSTEM DEFAULT
- **Description**: Name shown in IBM Storage Virtualize "Detected IP quorum Applications"
- **Default**: `${INSTANCE_NAME}` (with dashes/underscores removed)
- **Format**: 1-20 characters, A-Z, a-z, 0-9 only (no dashes or underscores)
- **Example**: `ipquorumsrv1`, `prodquorum`, `dcaquorum`

---

### DOWNLOAD CONFIGURATION

#### `IPQUORUM_DOWNLOAD_ENABLED` ⚪ SYSTEM DEFAULT
- **Description**: Enable automatic JAR download on service start/restart
- **Values**: `true` (enabled) or `false` (disabled)
- **Default**: `true`
- **Purpose**: Automatically downloads/updates ip_quorum.jar

#### `IPQUORUM_DOWNLOAD_TOOL` ⚪ SYSTEM DEFAULT
- **Description**: Download tool to use
- **Values**: `go`, `python`, or `bash`
- **Default**: `go` (recommended)
- **Options**:
  - `go`: Fast, single binary (requires ipquorum-download-go)
  - `python`: Feature-rich (requires Python 3.8+ and ipquorum-download.py)
  - `bash`: Portable (requires curl and jq, uses ipquorum-restapi-download.sh)

#### `DOWNLOAD_TOOL_GO` ⚪ SYSTEM DEFAULT
- **Description**: Path to Go download tool
- **Default**: `/usr/local/bin/ipquorum-download-go`

#### `DOWNLOAD_TOOL_PYTHON` ⚪ SYSTEM DEFAULT
- **Description**: Path to Python download tool
- **Default**: `/usr/local/bin/ipquorum-download.py`

#### `DOWNLOAD_TOOL_BASH` ⚪ SYSTEM DEFAULT
- **Description**: Path to Bash download tool
- **Default**: `/usr/local/bin/ipquorum-restapi-download.sh`

#### `IPQUORUM_BACKUP_ENABLED` ⚪ SYSTEM DEFAULT
- **Description**: Backup existing JAR before downloading new one
- **Values**: `true` or `false`
- **Default**: `true`
- **Backup file**: `ip_quorum.jar.backup`

---

### STORAGE VIRTUALIZE API CONFIGURATION

#### `API_ENDPOINT` 🔴 MANDATORY 🟢 AUTO-MANAGED
- **Description**: IBM Storage Virtualize API endpoint (IP or hostname)
- **Example**: `10.33.7.80` or `sv-cluster.example.com`
- **Set by**: Agent from API request
- **Required for**: Downloading JAR file and creating quorum app

#### `VIRTUALIZE_USERNAME` 🔴 MANDATORY 🟢 AUTO-MANAGED
- **Description**: IBM Storage Virtualize username
- **Example**: `superuser`, `admin`, `monitor_user`
- **Set by**: Agent from API request
- **Required role**:
  - For download only: **Monitor** role
  - For mkquorumapp: **Restricted Administrator** or higher

#### `VIRTUALIZE_PASSWORD_FILE` 🔴 MANDATORY 🟢 AUTO-MANAGED
- **Description**: Path to file containing the password
- **Example**: `/var/lib/ipquorum/.passwords/prod-cluster01.pass`
- **Set by**: Agent (creates file and updates config)
- **Format**: Plain text file, one line, no trailing newline recommended
- **Permissions**: `400` (read-only by owner)
- **Owner**: `ipquorum:ipquorum`
- **Security**: Stored in `/var/lib` for proper SELinux context

#### `API_PORT` ⚪ SYSTEM DEFAULT
- **Description**: IBM Storage Virtualize API port
- **Default**: `7443`
- **Rarely changed**: Standard HTTPS port for SVC/FlashSystem

---

### MKQUORUMAPP CONFIGURATION

#### `IPQUORUM_MKQUORUMAPP_ENABLED` 🟡 OPTIONAL
- **Description**: Create new IP Quorum application on storage system
- **Values**: `true` (create new) or `false` (download existing)
- **Default**: `false`
- **Note**: If `true`, `IPQUORUM_PARTNERSYSTEM` is MANDATORY

#### `IPQUORUM_PARTNERSYSTEM` 🔴 MANDATORY (if mkquorumapp enabled)
- **Description**: Remote system name in PBHA (PowerHA) configuration
- **Example**: `svc_cluster02`, `remote-site-cluster`
- **Required when**: `IPQUORUM_MKQUORUMAPP_ENABLED=true`
- **Purpose**: Identifies the partner system in the cluster

#### `IPQUORUM_IP6` 🟡 OPTIONAL
- **Description**: IPv6 configuration for local system
- **Values**: `true` (IPv6) or `false` (IPv4)
- **Default**: `false`
- **Purpose**: Tells storage system which IP version to use

#### `IPQUORUM_NOMETADATA` 🟡 OPTIONAL
- **Description**: Metadata configuration
- **Values**:
  - `true`: Disable metadata (tie-break only)
  - `false`: Enable metadata (cluster recovery)
- **Default**: `false`
- **Purpose**: Controls whether quorum stores cluster metadata

#### `IPQUORUM_PARTNERIP6` 🟡 OPTIONAL
- **Description**: IPv6 configuration for partner system
- **Values**: `true` (IPv6) or `false` (IPv4)
- **Default**: `false`
- **Purpose**: Tells storage system which IP version partner uses

---

## Agent API Request Mapping

When creating an instance via the agent API, these fields are mapped:

| API Request Field | Config File Field | Status |
|-------------------|-------------------|--------|
| `name` | `INSTANCE_NAME` | 🟢 Auto-set |
| `api_endpoint` | `API_ENDPOINT` | 🟢 Auto-set |
| `username` | `VIRTUALIZE_USERNAME` | 🟢 Auto-set |
| `password` | `VIRTUALIZE_PASSWORD_FILE` | 🟢 Auto-set (creates file) |
| `partnersystem` | `IPQUORUM_PARTNERSYSTEM` | 🟢 Auto-set (if provided) |
| `ip6` | `IPQUORUM_IP6` | 🟢 Auto-set (if provided) |
| `nometadata` | `IPQUORUM_NOMETADATA` | 🟢 Auto-set (if provided) |
| `partnerip6` | `IPQUORUM_PARTNERIP6` | 🟢 Auto-set (if provided) |
| N/A | `IBM_STORAGE_SYSTEM` | 🟢 Set to "N/A" |

---

## Example API Request

### Minimal (Download Only)
```json
{
  "name": "prod-cluster01",
  "api_endpoint": "10.33.7.80",
  "username": "monitor_user",
  "password": "secret123"
}
```

### Full (Create Quorum App)
```json
{
  "name": "prod-cluster01",
  "api_endpoint": "10.33.7.80",
  "username": "admin",
  "password": "secret123",
  "partnersystem": "svc_cluster02",
  "ip6": false,
  "nometadata": false,
  "partnerip6": false
}
```

---

## Manual Configuration

If you need to manually edit the config file:

### Required Steps
1. Edit `/etc/ipquorum/instances/<name>.conf`
2. Update mandatory fields:
   - `API_ENDPOINT`
   - `VIRTUALIZE_USERNAME`
   - `VIRTUALIZE_PASSWORD_FILE`
3. Create password file:
   ```bash
   sudo mkdir -p /var/lib/ipquorum/.passwords
   echo 'your_password' | sudo tee /var/lib/ipquorum/.passwords/<name>.pass > /dev/null
   sudo chmod 400 /var/lib/ipquorum/.passwords/<name>.pass
   sudo chown ipquorum:ipquorum /var/lib/ipquorum/.passwords/<name>.pass
   ```
4. Restart service:
   ```bash
   sudo systemctl restart ipquorum@<name>
   ```

### Optional Enhancements
- Set `IBM_STORAGE_SYSTEM` for better identification in `ipqm list`
- Set `IBM_STORAGE_DESCRIPTION` and `IBM_STORAGE_LOCATION` for documentation
- Enable `IPQUORUM_MKQUORUMAPP_ENABLED` if creating new quorum app
- Set `IPQUORUM_PARTNERSYSTEM` if using mkquorumapp

---

## Summary

### Absolutely Required (Agent Manages)
- ✅ `API_ENDPOINT` - Where to connect
- ✅ `VIRTUALIZE_USERNAME` - Who to authenticate as
- ✅ `VIRTUALIZE_PASSWORD_FILE` - How to authenticate

### Recommended for Documentation
- 📝 `IBM_STORAGE_SYSTEM` - Which system (shows in list)
- 📝 `IBM_STORAGE_DESCRIPTION` - What it is
- 📝 `IBM_STORAGE_LOCATION` - Where it is

### Required for Creating New Quorum App
- ⚙️ `IPQUORUM_MKQUORUMAPP_ENABLED=true`
- ⚙️ `IPQUORUM_PARTNERSYSTEM` - Partner system name

### Everything Else
- 🔧 Has sensible defaults
- 🔧 Rarely needs changing
- 🔧 Uses variable substitution for flexibility