# IBM Storage Virtualize IP Quorum - Multi-Instance Service

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║     ██╗██████╗       ██████╗ ██╗   ██╗ ██████╗ ██████╗ ██╗   ██╗███╗   ███╗  ║
║     ██║██╔══██╗     ██╔═══██╗██║   ██║██╔═══██╗██╔══██╗██║   ██║████╗ ████║  ║
║     ██║██████╔╝     ██║   ██║██║   ██║██║   ██║██████╔╝██║   ██║██╔████╔██║  ║
║     ██║██╔═══╝      ██║▄▄ ██║██║   ██║██║   ██║██╔══██╗██║   ██║██║╚██╔╝██║  ║
║     ██║██║          ╚██████╔╝╚██████╔╝╚██████╔╝██║  ██║╚██████╔╝██║ ╚═╝ ██║  ║
║     ╚═╝╚═╝           ╚══▀▀═╝  ╚═════╝  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝     ╚═╝  ║
║                                                                              ║
║  ┌────────────────────────────────────────────────────────────────────────┐  ║
║  │  Multi-Instance Service Configuration                                  │  ║
║  │  Run Multiple Independent IP Quorum Instances on a Single Host        │  ║
║  └────────────────────────────────────────────────────────────────────────┘  ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Management](#management)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

---

## 🎯 Overview

The multi-instance IP Quorum service configuration enables running **multiple independent IP Quorum instances** on a single host, where each instance connects to a different IBM Storage Virtualize System, partition, or IP address.

### Key Features

✅ **Multiple Independent Instances**: Each instance runs with its own configuration, process, and resources
✅ **Isolated Configurations**: Per-instance API endpoints, credentials, and settings
✅ **Separate Working Directories**: `/var/lib/ipquorum/<instance>/`
✅ **Isolated Logging**: `/var/log/ipquorum/<instance>/`
✅ **Systemd Template Units**: Easy instance management with `ipquorum@<name>.service`
✅ **Systemd Aliases**: Descriptive names for easy identification
✅ **Resource Limits**: Per-instance CPU and memory quotas
✅ **Security**: Isolated credentials and file permissions

### Use Cases

- **Multiple IBM Storage Systems**: Connect to different FlashSystem or SVC clusters
- **Multiple Partitions**: Separate quorum instances for different storage partitions
- **Development/Testing**: Run test and production instances on the same host
- **Geographic Distribution**: Connect to storage systems in different datacenters
- **PBHA Configurations**: Support multiple PowerHA setups

---

## 🏗️ Architecture

### Directory Structure

```
/etc/ipquorum/
├── instances/                          # Instance configurations
│   ├── system1.conf                    # Instance: system1
│   ├── system2.conf                    # Instance: system2
│   ├── datacenter-a.conf               # Instance: datacenter-a
│   └── .passwords/                     # Password files (secure)
│       ├── system1.password            # chmod 400/440
│       ├── system2.password
│       └── datacenter-a.password
├── instance.conf.template              # Configuration template
└── examples/                           # Example configurations
    ├── system1.conf
    ├── datacenter-a.conf
    └── partition-prod.conf

/var/lib/ipquorum/                      # Runtime data
├── system1/                            # Instance: system1
│   ├── ip_quorum.jar                   # Downloaded JAR
│   └── ip_quorum.jar.backup            # Backup
├── system2/
│   └── ip_quorum.jar
└── datacenter-a/
    └── ip_quorum.jar

/var/log/ipquorum/                      # Logs
├── system1/                            # Instance: system1
│   ├── download.log                    # Download logs
│   └── ip_quorum.log.1                 # App logs (rotated)
├── system2/
│   └── download.log
└── datacenter-a/
    └── download.log

/etc/systemd/system/
├── ipquorum@.service                   # Template service
└── multi-user.target.wants/
    ├── ipquorum@system1.service -> ../ipquorum@.service
    ├── ipquorum@system2.service -> ../ipquorum@.service
    └── ipquorum@datacenter-a.service -> ../ipquorum@.service
```

### Systemd Template Unit

The system uses systemd's **template unit** feature (`@` specifier):

- **Template**: `ipquorum@.service`
- **Instances**: `ipquorum@system1.service`, `ipquorum@datacenter-a.service`, etc.
- **Instance Name**: `%i` in the service file (e.g., "system1", "datacenter-a")

---

## 🚀 Installation

### Prerequisites

- **Linux with systemd** (RHEL/CentOS 7+, Ubuntu 16.04+, SLES 12+)
- **Java Runtime** (OpenJDK 8, 11, or later)
- **Root access** for installation
- **Network access** to IBM Storage Virtualize clusters (port 7443/HTTPS)
- **Download tool** (at least one):
  - `ipquorum-download-go` (recommended - fast, single binary)
  - `ipquorum-download.py` (requires Python 3.8+)
  - `ipquorum-restapi-download.sh` (requires curl and jq)

### Installation Steps

1. **Run the installation script**:

```bash
cd ipquorum-systemd/multi-instance
chmod +x install-multi-instance.sh
sudo ./install-multi-instance.sh
```

The script will:
- ✅ Check for Java (offer to install if missing)
- ✅ Create service user and group (`ipquorum`)
- ✅ Create required directories
- ✅ Install scripts (`ipquorum-download-multi.sh`, `ipquorum-start-multi.sh`)
- ✅ Install systemd template service (`ipquorum@.service`)
- ✅ Install configuration template
- ✅ Copy example configurations
- ✅ Configure firewall (optional)

2. **Verify installation**:

```bash
# Check systemd template
systemctl cat ipquorum@.service

# Check directories
ls -la /etc/ipquorum/instances/
ls -la /var/lib/ipquorum/
ls -la /var/log/ipquorum/

# Check scripts
ls -la /usr/local/bin/ipquorum-*-multi.sh
```

---

## 🎯 Quick Start

### Create Your First Instance

1. **Create instance configuration**:

```bash
# Copy template
sudo cp /etc/ipquorum/instance.conf.template /etc/ipquorum/instances/myinstance.conf

# Edit configuration
sudo vi /etc/ipquorum/instances/myinstance.conf
```

2. **Configure required settings**:

```bash
# Minimum required settings:
INSTANCE_NAME=myinstance
IBM_STORAGE_SYSTEM="IBM FlashSystem 9200 - Production"
API_ENDPOINT=10.33.7.80
VIRTUALIZE_USERNAME=monitor_user
```

3. **Create password file**:

```bash
# Create password file
echo 'your_password' | sudo tee /etc/ipquorum/instances/.passwords/myinstance.password > /dev/null

# Set secure permissions
sudo chmod 400 /etc/ipquorum/instances/.passwords/myinstance.password
sudo chown root:ipquorum /etc/ipquorum/instances/.passwords/myinstance.password
```

4. **Create instance directories**:

```bash
sudo mkdir -p /var/lib/ipquorum/myinstance
sudo mkdir -p /var/log/ipquorum/myinstance
sudo chown -R ipquorum:ipquorum /var/lib/ipquorum/myinstance
sudo chown -R ipquorum:ipquorum /var/log/ipquorum/myinstance
```

5. **Enable and start the instance**:

```bash
# Enable (start on boot)
sudo systemctl enable ipquorum@myinstance.service

# Start now
sudo systemctl start ipquorum@myinstance.service

# Check status
sudo systemctl status ipquorum@myinstance.service
```

6. **View logs**:

```bash
# Systemd journal
sudo journalctl -u ipquorum@myinstance.service -f

# Download log
sudo tail -f /var/log/ipquorum/myinstance/download.log

# Application log
sudo tail -f /var/log/ipquorum/myinstance/ip_quorum.log.1
```

---

## ⚙️ Configuration

### Instance Configuration File

Each instance has its own configuration file: `/etc/ipquorum/instances/<instance-name>.conf`

#### Required Settings

```bash
# Instance identification
INSTANCE_NAME=myinstance
IBM_STORAGE_SYSTEM="Description of your IBM Storage System"

# API connection
API_ENDPOINT=10.33.7.80                    # IP or hostname
VIRTUALIZE_USERNAME=monitor_user           # Username
VIRTUALIZE_PASSWORD_FILE=/etc/ipquorum/instances/.passwords/myinstance.password

# Paths
IPQUORUM_DIR=/var/lib/ipquorum/myinstance
IPQUORUM_JAR=${IPQUORUM_DIR}/ip_quorum.jar
IPQUORUM_LOG_DIR=/var/log/ipquorum/myinstance
```

#### Optional Settings

```bash
# Download configuration
IPQUORUM_DOWNLOAD_ENABLED=true             # Enable automatic download
IPQUORUM_DOWNLOAD_TOOL=go                  # go, python, or bash
IPQUORUM_BACKUP_ENABLED=true               # Backup before download

# mkquorumapp (for PBHA)
IPQUORUM_MKQUORUMAPP_ENABLED=false         # Create new quorum app
IPQUORUM_PARTNERSYSTEM=svc_cluster_remote  # Partner system name
IPQUORUM_IP6=false                         # IPv6 support
IPQUORUM_NOMETADATA=false                  # Metadata configuration
IPQUORUM_PARTNERIP6=false                  # Partner IPv6

# TLS/SSL
IPQUORUM_TLS_VERIFY=false                  # Strict TLS verification

# IP Quorum application
IPQUORUM_NAME=myinstance                   # Instance name (A-Z, a-z, 0-9)
IPQUORUM_DEBUG=false                       # Debug mode
IPQUORUM_LOG_ROTATION=5                    # Log rotation count (1-10)
IPQUORUM_LOG_SIZE=5120                     # Log size in KB (1024-10240)

# Resource limits
IPQUORUM_MEMORY_MAX=512M                   # Maximum memory
IPQUORUM_CPU_QUOTA=50%                     # CPU quota
```

### Password File Security

**IMPORTANT**: Password files must have restrictive permissions!

```bash
# Recommended: Group-readable (for systemd service)
sudo chmod 440 /etc/ipquorum/instances/.passwords/myinstance.password
sudo chown root:ipquorum /etc/ipquorum/instances/.passwords/myinstance.password

# Alternative: Owner-only readable
sudo chmod 400 /etc/ipquorum/instances/.passwords/myinstance.password
sudo chown ipquorum:ipquorum /etc/ipquorum/instances/.passwords/myinstance.password
```

**Never use world-readable permissions (444, 644)!**

---

## 🔧 Management

### Using systemctl (Direct)

```bash
# Start instance
sudo systemctl start ipquorum@myinstance.service

# Stop instance
sudo systemctl stop ipquorum@myinstance.service

# Restart instance
sudo systemctl restart ipquorum@myinstance.service

# Enable (start on boot)
sudo systemctl enable ipquorum@myinstance.service

# Disable (don't start on boot)
sudo systemctl disable ipquorum@myinstance.service

# Check status
sudo systemctl status ipquorum@myinstance.service

# View logs
sudo journalctl -u ipquorum@myinstance.service -f

# View logs (last 100 lines)
sudo journalctl -u ipquorum@myinstance.service -n 100
```

### Using Instance Manager (Recommended)

The `ipquorum-instance-manager.sh` script provides a convenient interface:

```bash
# Install the manager
sudo cp ipquorum-instance-manager.sh /usr/local/bin/
sudo chmod 755 /usr/local/bin/ipquorum-instance-manager.sh

# Create alias (optional)
alias ipqm='sudo /usr/local/bin/ipquorum-instance-manager.sh'
```

#### Manager Commands

```bash
# List all instances
ipqm list

# Create new instance
ipqm create myinstance

# Show instance information
ipqm info myinstance

# Validate configuration
ipqm validate myinstance

# Start/stop/restart
ipqm start myinstance
ipqm stop myinstance
ipqm restart myinstance

# Enable/disable
ipqm enable myinstance
ipqm disable myinstance

# View logs
ipqm logs myinstance
ipqm logs myinstance 100    # Last 100 lines

# Check status
ipqm status myinstance
ipqm status                 # All instances

# Delete instance
ipqm delete myinstance
```

### Multi-Instance Operations

```bash
# Start multiple instances
sudo systemctl start ipquorum@system1 ipquorum@system2 ipquorum@datacenter-a

# Stop all instances
sudo systemctl stop 'ipquorum@*'

# Status of all instances
sudo systemctl status 'ipquorum@*'

# List all running instances
sudo systemctl list-units 'ipquorum@*'

# Enable multiple instances
sudo systemctl enable ipquorum@{system1,system2,datacenter-a}.service
```

### Using Systemd Aliases

Each instance has an alias for easier identification:

```bash
# Access via alias
sudo systemctl status ipquorum-myinstance.service
sudo systemctl restart ipquorum-datacenter-a.service
```

---

## 📚 Examples

### Example 1: Simple Production System

**Scenario**: Single IBM FlashSystem, download-only

```bash
# Configuration: /etc/ipquorum/instances/prod-fs9200.conf
INSTANCE_NAME=prod-fs9200
IBM_STORAGE_SYSTEM="IBM FlashSystem 9200 - Production"
IBM_STORAGE_LOCATION="Datacenter A, Rack 12"

API_ENDPOINT=10.33.7.80
VIRTUALIZE_USERNAME=monitor_user
VIRTUALIZE_PASSWORD_FILE=/etc/ipquorum/instances/.passwords/prod-fs9200.password

IPQUORUM_DOWNLOAD_ENABLED=true
IPQUORUM_DOWNLOAD_TOOL=go
IPQUORUM_MKQUORUMAPP_ENABLED=false
IPQUORUM_TLS_VERIFY=false

IPQUORUM_NAME=prodfs9200
IPQUORUM_DEBUG=false
IPQUORUM_MEMORY_MAX=512M
IPQUORUM_CPU_QUOTA=50%
```

**Setup**:

```bash
# Create configuration
sudo cp /etc/ipquorum/examples/system1.conf /etc/ipquorum/instances/prod-fs9200.conf
sudo vi /etc/ipquorum/instances/prod-fs9200.conf

# Create password
echo 'your_password' | sudo tee /etc/ipquorum/instances/.passwords/prod-fs9200.password
sudo chmod 400 /etc/ipquorum/instances/.passwords/prod-fs9200.password

# Create directories
sudo mkdir -p /var/lib/ipquorum/prod-fs9200 /var/log/ipquorum/prod-fs9200
sudo chown -R ipquorum:ipquorum /var/lib/ipquorum/prod-fs9200 /var/log/ipquorum/prod-fs9200

# Enable and start
sudo systemctl enable --now ipquorum@prod-fs9200.service
```

### Example 2: PBHA Configuration

**Scenario**: IBM SVC with PowerHA, create quorum app

```bash
# Configuration: /etc/ipquorum/instances/pbha-svc.conf
INSTANCE_NAME=pbha-svc
IBM_STORAGE_SYSTEM="IBM SVC Cluster - PBHA Configuration"

API_ENDPOINT=192.168.10.50
VIRTUALIZE_USERNAME=admin
VIRTUALIZE_PASSWORD_FILE=/etc/ipquorum/instances/.passwords/pbha-svc.password

IPQUORUM_DOWNLOAD_ENABLED=true
IPQUORUM_DOWNLOAD_TOOL=go

# mkquorumapp enabled for PBHA
IPQUORUM_MKQUORUMAPP_ENABLED=true
IPQUORUM_PARTNERSYSTEM=svc_cluster_remote
IPQUORUM_IP6=false
IPQUORUM_NOMETADATA=false
IPQUORUM_PARTNERIP6=false

IPQUORUM_TLS_VERIFY=false
IPQUORUM_NAME=pbhasvc
IPQUORUM_DEBUG=false

# Higher resources for production PBHA
IPQUORUM_MEMORY_MAX=1G
IPQUORUM_CPU_QUOTA=75%
```

### Example 3: Multiple Datacenters

**Scenario**: Three storage systems in different locations

```bash
# Datacenter A
sudo cp /etc/ipquorum/instance.conf.template /etc/ipquorum/instances/dc-a.conf
# Edit: API_ENDPOINT=10.10.1.50, IPQUORUM_NAME=dcA

# Datacenter B
sudo cp /etc/ipquorum/instance.conf.template /etc/ipquorum/instances/dc-b.conf
# Edit: API_ENDPOINT=10.20.1.50, IPQUORUM_NAME=dcB

# Datacenter C
sudo cp /etc/ipquorum/instance.conf.template /etc/ipquorum/instances/dc-c.conf
# Edit: API_ENDPOINT=10.30.1.50, IPQUORUM_NAME=dcC

# Create passwords for all
for dc in dc-a dc-b dc-c; do
    echo "password_${dc}" | sudo tee /etc/ipquorum/instances/.passwords/${dc}.password
    sudo chmod 400 /etc/ipquorum/instances/.passwords/${dc}.password
done

# Create directories for all
for dc in dc-a dc-b dc-c; do
    sudo mkdir -p /var/lib/ipquorum/${dc} /var/log/ipquorum/${dc}
    sudo chown -R ipquorum:ipquorum /var/lib/ipquorum/${dc} /var/log/ipquorum/${dc}
done

# Enable and start all
sudo systemctl enable --now ipquorum@{dc-a,dc-b,dc-c}.service

# Check status
sudo systemctl status 'ipquorum@dc-*'
```

---

## 🔍 Troubleshooting

### Common Issues

#### 1. Instance Won't Start

**Check configuration**:
```bash
sudo /usr/local/bin/ipquorum-instance-manager.sh validate myinstance
```

**Check logs**:
```bash
sudo journalctl -u ipquorum@myinstance.service -n 50
sudo tail -f /var/log/ipquorum/myinstance/download.log
```

**Common causes**:
- Missing or incorrect API_ENDPOINT
- Invalid credentials
- Password file not readable
- Missing directories

#### 2. Download Fails

**Check password file**:
```bash
# Verify file exists and is readable
sudo ls -la /etc/ipquorum/instances/.passwords/myinstance.password

# Check permissions (should be 400 or 440)
stat /etc/ipquorum/instances/.passwords/myinstance.password

# Test reading as ipquorum user
sudo -u ipquorum cat /etc/ipquorum/instances/.passwords/myinstance.password
```

**Check network connectivity**:
```bash
# Test API endpoint
curl -k https://10.33.7.80:7443/rest/v1/auth

# Check firewall
sudo firewall-cmd --list-all
```

#### 3. Permission Denied Errors

**Fix directory permissions**:
```bash
sudo chown -R ipquorum:ipquorum /var/lib/ipquorum/myinstance
sudo chown -R ipquorum:ipquorum /var/log/ipquorum/myinstance
sudo chmod 755 /var/lib/ipquorum/myinstance
sudo chmod 755 /var/log/ipquorum/myinstance
```

**Fix password file permissions**:
```bash
sudo chmod 440 /etc/ipquorum/instances/.passwords/myinstance.password
sudo chown root:ipquorum /etc/ipquorum/instances/.passwords/myinstance.password
```

#### 4. Instance Name Conflicts

**Check for duplicate names**:
```bash
# List all instances
sudo /usr/local/bin/ipquorum-instance-manager.sh list

# Check systemd
systemctl list-units 'ipquorum@*'
```

#### 5. Java Not Found

**Install Java**:
```bash
# RHEL/CentOS
sudo yum install -y java-11-openjdk-headless

# Ubuntu/Debian
sudo apt-get install -y openjdk-11-jre-headless

# Verify
java -version
```

### Debug Mode

Enable debug mode for detailed logging:

```bash
# Edit instance configuration
sudo vi /etc/ipquorum/instances/myinstance.conf

# Set debug mode
IPQUORUM_DEBUG=true

# Restart instance
sudo systemctl restart ipquorum@myinstance.service

# View detailed logs
sudo journalctl -u ipquorum@myinstance.service -f
```

### Health Checks

```bash
# Check if instance is running
sudo systemctl is-active ipquorum@myinstance.service

# Check if instance is enabled
sudo systemctl is-enabled ipquorum@myinstance.service

# Check JAR file
ls -lh /var/lib/ipquorum/myinstance/ip_quorum.jar

# Check process
ps aux | grep "ip_quorum.jar" | grep myinstance

# Check network connections
sudo netstat -tlnp | grep 1260
```

---

## 🎯 Best Practices

### Security

1. **Password Files**:
   - Use restrictive permissions (400 or 440)
   - Never commit to version control
   - Rotate passwords regularly
   - Use different passwords for each instance

2. **TLS/SSL**:
   - Use `IPQUORUM_TLS_VERIFY=true` in production with valid certificates
   - Use `IPQUORUM_TLS_VERIFY=false` only for self-signed certificates

3. **User Permissions**:
   - Run service as dedicated `ipquorum` user
   - Never run as root
   - Use systemd security directives

### Configuration Management

1. **Naming Conventions**:
   - Use descriptive instance names (e.g., `prod-dc-a`, `test-partition-1`)
   - Avoid special characters (use only letters, numbers, hyphens, underscores)
   - Keep names short and meaningful

2. **Documentation**:
   - Document each instance in `IBM_STORAGE_SYSTEM` field
   - Include location in `IBM_STORAGE_LOCATION` field
   - Maintain a central inventory of instances

3. **Version Control**:
   - Store configuration templates in version control
   - **Never** commit password files
   - Use `.gitignore` for sensitive files

### Monitoring

1. **Systemd Status**:
   ```bash
   # Monitor all instances
   watch -n 5 'systemctl status "ipquorum@*" --no-pager'
   ```

2. **Log Monitoring**:
   ```bash
   # Centralized logging
   sudo journalctl -u 'ipquorum@*' -f
   ```

3. **Automated Alerts**:
   - Configure systemd email notifications
   - Use monitoring tools (Nagios, Zabbix, Prometheus)
   - Set up log aggregation (ELK, Splunk)

### Resource Management

1. **Memory Limits**:
   - Set appropriate `IPQUORUM_MEMORY_MAX` per instance
   - Monitor memory usage: `systemctl status ipquorum@myinstance`

2. **CPU Quotas**:
   - Set `IPQUORUM_CPU_QUOTA` to prevent resource starvation
   - Adjust based on workload

3. **Log Rotation**:
   - Configure `IPQUORUM_LOG_ROTATION` and `IPQUORUM_LOG_SIZE`
   - Monitor disk space: `df -h /var/log/ipquorum/`

### Backup and Recovery

1. **Configuration Backup**:
   ```bash
   # Backup all configurations
   sudo tar -czf ipquorum-configs-$(date +%Y%m%d).tar.gz /etc/ipquorum/instances/
   ```

2. **JAR File Backup**:
   - Enable `IPQUORUM_BACKUP_ENABLED=true`
   - Backup files are created automatically before downloads

3. **Disaster Recovery**:
   - Document recovery procedures
   - Test recovery process regularly
   - Keep offline backups

---

## 📞 Support

### Documentation

- **IBM Documentation**: https://www.ibm.com/docs/en/flashsystem-7x00/8.6.x
- **IP Quorum Info**: https://www.ibm.com/support/pages/node/7013877
- **Local Documentation**: `/etc/ipquorum/README-ADVANCED.md`

### Logs

- **Systemd Journal**: `sudo journalctl -u ipquorum@<instance>.service`
- **Download Logs**: `/var/log/ipquorum/<instance>/download.log`
- **Application Logs**: `/var/log/ipquorum/<instance>/ip_quorum.log.*`

### Getting Help

1. Check logs for error messages
2. Validate configuration: `ipqm validate <instance>`
3. Review this documentation
4. Check IBM support documentation
5. Contact IBM support if needed

---

## 📝 License

This multi-instance configuration system is provided as-is for use with IBM Storage Virtualize systems.

---

**Made with ❤️ and systemd**

*Last updated: 2026-04-16*