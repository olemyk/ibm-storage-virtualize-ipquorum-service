# IP Quorum Multi-Instance Architecture

## Overview

The IP Quorum Multi-Instance Service is designed to run multiple independent IP Quorum instances on a single host, each connecting to a different IBM Storage Virtualize system. This architecture enables centralized quorum management for multiple storage clusters while maintaining complete isolation between instances.

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Linux Host (RHEL/Ubuntu)                     │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Systemd Service Manager                    │   │
│  │                                                               │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │   │
│  │  │ ipquorum@    │  │ ipquorum@    │  │ ipquorum@    │      │   │
│  │  │ system1      │  │ system2      │  │ datacenter-a │      │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │   │
│  └─────────┼──────────────────┼──────────────────┼─────────────┘   │
│            │                  │                  │                   │
│  ┌─────────▼──────────────────▼──────────────────▼─────────────┐   │
│  │              Instance-Specific Resources                      │   │
│  │                                                               │   │
│  │  /etc/ipquorum/instances/                                    │   │
│  │  ├── system1.conf          ← Configuration                   │   │
│  │  ├── system2.conf                                            │   │
│  │  └── datacenter-a.conf                                       │   │
│  │                                                               │   │
│  │  /var/lib/ipquorum/                                          │   │
│  │  ├── system1/                                                │   │
│  │  │   ├── ip_quorum.jar     ← JAR file                        │   │
│  │  │   ├── .password         ← Credentials                     │   │
│  │  │   └── logs/             ← Log files                       │   │
│  │  ├── system2/                                                │   │
│  │  └── datacenter-a/                                           │   │
│  └───────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────────┐   │
│  │              Management Tools                                  │   │
│  │                                                               │   │
│  │  /usr/local/bin/                                             │   │
│  │  ├── ipquorum-instance-manager.sh  ← CLI management          │   │
│  │  ├── ipquorum-start-multi.sh       ← Startup script          │   │
│  │  ├── ipquorum-download-multi.sh    ← Download script         │   │
│  │  └── ipquorum-validate-multi.sh    ← Validation script       │   │
│  └───────────────────────────────────────────────────────────────┘   │
└───────────────────────────────────────────────────────────────────────┘
         │                    │                    │
         │ HTTPS:7443         │ HTTPS:7443         │ HTTPS:7443
         ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ IBM Storage     │  │ IBM Storage     │  │ IBM Storage     │
│ Virtualize      │  │ Virtualize      │  │ Virtualize      │
│ System 1        │  │ System 2        │  │ Datacenter A    │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

---

## Core Components

### 1. Systemd Template Unit (`ipquorum@.service`)

**Location**: `/etc/systemd/system/ipquorum@.service`

**Purpose**: Template service file that enables multiple independent instances

**Key Features**:
- Instance parameterization using `%i` (instance name)
- Per-instance configuration loading
- Independent lifecycle management
- Resource limits and security hardening
- Automatic restart on failure

**Example Usage**:
```bash
systemctl start ipquorum@system1.service
systemctl start ipquorum@system2.service
systemctl start ipquorum@datacenter-a.service
```

### 2. Instance Configuration System

**Location**: `/etc/ipquorum/instances/<instance>.conf`

**Purpose**: Store instance-specific settings

**Configuration Parameters**:
```bash
# System Identification
IBM_STORAGE_SYSTEM="svc_cluster01"  # Documentation only
INSTANCE_DESCRIPTION="Production SVC Cluster"

# API Connection
API_ENDPOINT="10.33.7.80"           # Storage system IP/hostname
API_PORT="7443"                      # REST API port
VIRTUALIZE_USERNAME="superuser"     # API username

# Download Settings
IPQUORUM_DOWNLOAD_ENABLED="true"    # Auto-download JAR
IPQUORUM_DOWNLOAD_TOOL="go"         # go|python|bash
IPQUORUM_BACKUP_ENABLED="true"      # Backup before download

# mkquorumapp Settings (Optional)
IPQUORUM_MKQUORUMAPP_ENABLED="false"
IPQUORUM_PARTNERSYSTEM=""           # Required if mkquorumapp enabled
IPQUORUM_IP6="false"
IPQUORUM_NOMETADATA="false"
IPQUORUM_PARTNERIP6="false"

# Security
IPQUORUM_TLS_VERIFY="false"         # TLS certificate verification
```

### 3. Instance Data Directory

**Location**: `/var/lib/ipquorum/<instance>/`

**Structure**:
```
/var/lib/ipquorum/<instance>/
├── ip_quorum.jar          # Java application
├── .password              # Encrypted credentials (400 permissions)
└── logs/                  # Instance-specific logs
    ├── ipquorum.log
    ├── download.log
    └── startup.log
```

**Ownership**: `ipquorum:ipquorum` (dedicated service user)

**Permissions**:
- Directories: `755`
- JAR file: `644`
- Password file: `400`
- Log files: `644`

### 4. Management Scripts

#### Instance Manager (`ipquorum-instance-manager.sh`)

**Purpose**: Central CLI for instance lifecycle management

**Commands**:
```bash
# Create new instance
ipquorum-instance-manager.sh create <instance>

# List all instances
ipquorum-instance-manager.sh list

# Start/stop/restart instance
ipquorum-instance-manager.sh start <instance>
ipquorum-instance-manager.sh stop <instance>
ipquorum-instance-manager.sh restart <instance>

# Enable/disable autostart
ipquorum-instance-manager.sh enable <instance>
ipquorum-instance-manager.sh disable <instance>

# View status and logs
ipquorum-instance-manager.sh status <instance>
ipquorum-instance-manager.sh logs <instance>

# Delete instance
ipquorum-instance-manager.sh delete <instance>
```

#### Startup Script (`ipquorum-start-multi.sh`)

**Purpose**: Initialize and start IP Quorum Java process

**Responsibilities**:
- Load instance configuration
- Validate environment
- Set Java options
- Start IP Quorum application
- Handle errors and logging

#### Download Script (`ipquorum-download-multi.sh`)

**Purpose**: Automatically download `ip_quorum.jar` from storage system

**Features**:
- Multi-tool support (Go/Python/Bash)
- Automatic backup before download
- Restore on failure
- Checksum verification
- Retry logic with exponential backoff

#### Validation Script (`ipquorum-validate-multi.sh`)

**Purpose**: Pre-flight checks before service start

**Checks**:
- Configuration file exists and is readable
- Required parameters are set
- JAR file exists
- Password file exists and has correct permissions
- Network connectivity to storage system
- Java runtime available

---

## Data Flow

### Instance Startup Sequence

```
1. systemctl start ipquorum@system1.service
   │
   ├─→ 2. Systemd loads ipquorum@.service template
   │      - Substitutes %i with "system1"
   │      - Loads /etc/ipquorum/instances/system1.conf
   │
   ├─→ 3. ExecStartPre: ipquorum-validate-multi.sh
   │      - Validates configuration
   │      - Checks prerequisites
   │      - Fails fast if issues found
   │
   ├─→ 4. ExecStartPre: ipquorum-download-multi.sh (if enabled)
   │      - Downloads ip_quorum.jar from storage system
   │      - Creates backup of existing JAR
   │      - Verifies download integrity
   │      - Restores backup on failure
   │
   ├─→ 5. ExecStart: ipquorum-start-multi.sh
   │      - Loads instance configuration
   │      - Sets Java options
   │      - Starts IP Quorum application
   │      - Monitors process
   │
   └─→ 6. IP Quorum Application Running
          - Listens on port 1260/tcp
          - Maintains connection to storage system
          - Provides quorum services
```

### Download Flow (Optional)

```
1. ipquorum-download-multi.sh invoked
   │
   ├─→ 2. Load instance configuration
   │      - Read /etc/ipquorum/instances/<instance>.conf
   │      - Validate IPQUORUM_DOWNLOAD_ENABLED=true
   │
   ├─→ 3. Select download tool
   │      - Go binary (preferred, no dependencies)
   │      - Python script (requires Python + requests)
   │      - Bash script (requires curl + jq)
   │
   ├─→ 4. Backup existing JAR (if enabled)
   │      - Copy ip_quorum.jar to ip_quorum.jar.backup
   │      - Timestamp backup
   │
   ├─→ 5. Authenticate to storage system
   │      - POST /rest/v1/auth
   │      - Obtain X-Auth-Token
   │      - Handle rate limiting (429)
   │
   ├─→ 6. Download JAR file
   │      - POST /rest/v1/download
   │      - Save to /var/lib/ipquorum/<instance>/ip_quorum.jar
   │      - Verify file size > 0
   │
   ├─→ 7. Set ownership and permissions
   │      - chown ipquorum:ipquorum
   │      - chmod 644
   │
   └─→ 8. Success or restore backup
          - On success: Remove backup
          - On failure: Restore from backup
```

---

## Security Architecture

### Process Isolation

**Service User**: `ipquorum` (non-login, dedicated)
- No shell access
- Limited file system access
- No privilege escalation

**Systemd Security Directives**:
```ini
# File System Protection
ProtectSystem=strict              # Read-only /usr, /boot, /efi
ProtectHome=yes                   # No access to /home
ReadWritePaths=/var/lib/ipquorum  # Only writable directory
ReadOnlyPaths=/etc/ipquorum       # Read-only config

# Process Restrictions
NoNewPrivileges=true              # Cannot gain privileges
PrivateTmp=yes                    # Isolated /tmp
ProtectKernelTunables=yes         # No /proc/sys access
ProtectControlGroups=yes          # No cgroup access
RestrictRealtime=yes              # No realtime scheduling

# Network
RestrictAddressFamilies=AF_INET AF_INET6  # Only IPv4/IPv6
```

### Credential Management

**Password Storage**:
- Location: `/var/lib/ipquorum/<instance>/.password`
- Permissions: `400` (owner read-only)
- Owner: `ipquorum:ipquorum`
- Never stored in configuration files
- Never logged or displayed

**API Authentication**:
- HTTPS only (port 7443)
- Token-based authentication
- Tokens expire after session
- Optional TLS certificate verification

### SELinux Support

**Custom Policy** (if needed):
```bash
# Allow IP Quorum to bind to port 1260
semanage port -a -t ipquorum_port_t -p tcp 1260

# Allow network connections
setsebool -P ipquorum_can_network_connect 1
```

**File Contexts**:
```bash
semanage fcontext -a -t ipquorum_exec_t "/usr/local/bin/ipquorum-.*"
semanage fcontext -a -t ipquorum_var_lib_t "/var/lib/ipquorum(/.*)?"
semanage fcontext -a -t ipquorum_etc_t "/etc/ipquorum(/.*)?"
restorecon -Rv /usr/local/bin /var/lib/ipquorum /etc/ipquorum
```

---

## Resource Management

### Per-Instance Limits

**CPU**:
```ini
CPUQuota=50%                      # Max 50% of one CPU core
CPUAccounting=yes                 # Enable CPU accounting
```

**Memory**:
```ini
MemoryLimit=512M                  # Max 512MB RAM
MemoryAccounting=yes              # Enable memory accounting
```

**File Descriptors**:
```ini
LimitNOFILE=65536                 # Max open files
```

### Restart Policy

```ini
Restart=on-failure                # Restart on crash
RestartSec=10                     # Wait 10s before restart
StartLimitInterval=300            # 5 minute window
StartLimitBurst=5                 # Max 5 restarts in window
```

---

## Monitoring and Logging

### Systemd Journal Integration

**View logs**:
```bash
# All instances
journalctl -u 'ipquorum@*'

# Specific instance
journalctl -u ipquorum@system1.service

# Follow logs
journalctl -u ipquorum@system1.service -f

# Last 100 lines
journalctl -u ipquorum@system1.service -n 100
```

### Instance-Specific Logs

**Location**: `/var/lib/ipquorum/<instance>/logs/`

**Log Files**:
- `ipquorum.log` - Application logs
- `download.log` - Download script logs
- `startup.log` - Startup script logs

### Status Monitoring

```bash
# System-wide status
systemctl list-units 'ipquorum@*'

# Instance status
systemctl status ipquorum@system1.service

# Instance manager status
ipquorum-instance-manager.sh status system1
```

---

## High Availability Considerations

### Multiple Instances on Single Host

**Benefits**:
- Centralized quorum management
- Reduced hardware costs
- Simplified administration
- Consistent configuration

**Considerations**:
- Host becomes single point of failure
- Resource contention between instances
- Network bandwidth sharing
- Requires adequate CPU/memory

**Best Practices**:
- Use dedicated quorum host
- Monitor resource usage
- Set appropriate resource limits
- Implement host-level HA (optional)

### Network Requirements

**Connectivity**:
- Stable network to all storage systems
- Low latency (<10ms recommended)
- Redundant network paths (recommended)
- Firewall rules for port 7443 (outbound) and 1260 (inbound)

**Bandwidth**:
- Minimal bandwidth required
- Primarily heartbeat traffic
- Occasional JAR downloads (~10-50MB)

---

## Scalability

### Instance Limits

**Theoretical**: Unlimited instances per host

**Practical Limits**:
- **CPU**: ~20-30 instances per modern CPU
- **Memory**: ~512MB per instance (10-20 instances on 16GB host)
- **Network**: Depends on network capacity
- **File Descriptors**: 65536 per instance

**Recommended**:
- **Small deployments**: 5-10 instances
- **Medium deployments**: 10-20 instances
- **Large deployments**: 20-30 instances
- **Beyond 30**: Consider multiple quorum hosts

### Performance Tuning

**Java Options** (per instance):
```bash
JAVA_OPTS="-Xms256m -Xmx512m -XX:+UseG1GC"
```

**Systemd Tuning**:
```ini
# Increase file descriptor limit
LimitNOFILE=131072

# Adjust CPU quota
CPUQuota=100%  # For high-priority instances
```

---

## Disaster Recovery

### Backup Strategy

**Configuration Backup**:
```bash
# Backup all instance configs
tar -czf ipquorum-configs-$(date +%Y%m%d).tar.gz /etc/ipquorum/instances/
```

**JAR Backup**:
- Automatic backup before each download (if enabled)
- Manual backup: `cp ip_quorum.jar ip_quorum.jar.backup`

**Password Backup**:
- Securely store password files
- Use encrypted backup storage
- Maintain offline copies

### Recovery Procedures

**Instance Recovery**:
```bash
# Restore configuration
cp backup/system1.conf /etc/ipquorum/instances/

# Restore JAR
cp backup/ip_quorum.jar /var/lib/ipquorum/system1/

# Restore password
cp backup/.password /var/lib/ipquorum/system1/
chmod 400 /var/lib/ipquorum/system1/.password

# Restart instance
systemctl restart ipquorum@system1.service
```

**Complete System Recovery**:
```bash
# Reinstall service
sudo ./install-ipquorum-service.sh

# Restore all configurations
tar -xzf ipquorum-configs-backup.tar.gz -C /

# Restore all instance data
tar -xzf ipquorum-data-backup.tar.gz -C /

# Fix permissions
chown -R ipquorum:ipquorum /var/lib/ipquorum
chmod 400 /var/lib/ipquorum/*/.password

# Start all instances
for instance in $(ls /etc/ipquorum/instances/*.conf | xargs -n1 basename | sed 's/.conf//'); do
    systemctl enable --now ipquorum@${instance}.service
done
```

---

## Troubleshooting

### Common Issues

**Instance won't start**:
```bash
# Check configuration
ipquorum-validate-multi.sh system1

# Check logs
journalctl -u ipquorum@system1.service -n 50

# Check permissions
ls -la /var/lib/ipquorum/system1/
```

**Download fails**:
```bash
# Check connectivity
curl -k https://<storage-ip>:7443/rest/v1/auth

# Check credentials
cat /var/lib/ipquorum/system1/.password

# Manual download
ipquorum-download-multi.sh system1
```

**SELinux denials**:
```bash
# Check denials
ausearch -m avc -ts recent

# Fix contexts
restorecon -Rv /var/lib/ipquorum /etc/ipquorum

# Use troubleshooting tools
/usr/local/bin/diagnose-selinux.sh
/usr/local/bin/fix-selinux.sh
```

---

## Future Enhancements

### Planned Features

- **Web UI**: Browser-based instance management
- **Metrics Export**: Prometheus/Grafana integration
- **Health Checks**: Automated health monitoring
- **Alerting**: Email/Slack notifications
- **Ansible Role**: Automated deployment
- **Container Support**: Docker/Podman images
- **Kubernetes**: Helm charts for K8s deployment

---

## References

- [IBM Storage Virtualize Documentation](https://www.ibm.com/docs/en/flashsystem)
- [Systemd Documentation](https://www.freedesktop.org/software/systemd/man/)
- [SELinux User Guide](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/9/html/using_selinux/)
- [Project Repository](https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service)

---

**Document Version**: 2.0.1  
**Last Updated**: 2026-04-21  
**Author**: Ole Kristian Myklebust