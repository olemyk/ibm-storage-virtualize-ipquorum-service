# IBM Storage Virtualize IP Quorum Service - Improved Version


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
║  │  IBM Storage Virtualize High Availability Quorum Service               │  ║
║  │  Automated Download • Systemd Integration                              │  ║
║  └────────────────────────────────────────────────────────────────────────┘  ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```


This is an enhanced version of the IP Quorum systemd service with automatic download capability, comprehensive configuration options, and improved security.

## 🆕 What's New in the Improved Version

### Key Features

1. **Automatic Download on Start/Restart**
   - Downloads latest `ip_quorum.jar` before service starts
   - Configurable to use Go, Python, or Bash download tools
   - Optional - can be disabled for manual JAR placement

2. **Flexible Configuration**
   - Centralized configuration file (`/etc/ipquorum/ipquorum.conf`)
   - Environment-based settings
   - Support for multiple download tools

3. **Enhanced Security**
   - Password file authentication (no passwords in config)
   - Automatic backup of existing JAR before download
   - Improved systemd security directives
   - File permission validation

4. **Better Reliability**
   - Automatic backup and restore on download failure
   - Comprehensive logging
   - Retry logic with exponential backoff
   - Health checks before service start

5. **Easy Installation**
   - Automated installation script
   - Automatic Java installation option
   - Firewall configuration
   - User/group creation

## 📋 Prerequisites

### Required

- **Linux with systemd** (RHEL/CentOS 7+, Ubuntu 16.04+, SLES 12+)
- **Java Runtime** (OpenJDK 8, 11, or later)
- **Root access** for installation

### Optional (for automatic download)

- **Network access** to Storage Virtualize cluster (port 7443/HTTPS)
- **One of the download tools**:
  - `ipquorum-download-go` (recommended - fast, single binary)
  - `ipquorum-download.py` (requires Python 3.8+)
  - `ipquorum-restapi-download.sh` (requires curl and jq)

## 🚀 Quick Start Installation

### 1. Run the Installation Script

```bash
# Make the script executable
chmod +x install-ipquorum-service.sh

# Run as root
sudo ./install-ipquorum-service.sh
```

The script will:
- ✅ Check for Java (offer to install if missing)
- ✅ Create service user and group
- ✅ Create required directories
- ✅ Install download script
- ✅ Install configuration file
- ✅ Install systemd service
- ✅ Configure firewall (port 1260/tcp)

### 2. Configure the Service

Edit the configuration file:

```bash
sudo vi /etc/ipquorum/ipquorum.conf
```

**Option A: Enable Automatic Download**

```bash
# Enable download
IPQUORUM_DOWNLOAD_ENABLED=true

# Choose download tool (go, python, or bash)
IPQUORUM_DOWNLOAD_TOOL=go

# Configure API access
API_ENDPOINT=10.33.7.80
VIRTUALIZE_USERNAME=monitor_user

# Create password file
echo 'your_password' | sudo tee /etc/ipquorum/.password > /dev/null
sudo chmod 400 /etc/ipquorum/.password
```

**Option B: Manual JAR Placement**

```bash
# Disable download
IPQUORUM_DOWNLOAD_ENABLED=false

# Manually place JAR file
sudo cp ip_quorum.jar /opt/IBM/ip-quorum/
sudo chown ipquorum:ipquorum /opt/IBM/ip-quorum/ip_quorum.jar
```

### 3. Install Download Tool (if using automatic download)

**Go Binary (Recommended)**

```bash
# Copy the Go binary
sudo cp ipquorum-download-go /usr/local/bin/
sudo chmod 755 /usr/local/bin/ipquorum-download-go

# Verify
/usr/local/bin/ipquorum-download-go version
```

**Python Script**

```bash
# Copy the Python script
sudo cp ipquorum-download.py /usr/local/bin/
sudo chmod 755 /usr/local/bin/ipquorum-download.py

# Verify
/usr/local/bin/ipquorum-download.py --version
```

**Bash Script**

```bash
# Copy the Bash script
sudo cp ipquorum-restapi-download.sh /usr/local/bin/
sudo chmod 755 /usr/local/bin/ipquorum-restapi-download.sh
```

### 4. Start the Service

```bash
# Enable service to start on boot
sudo systemctl enable ipquorum

# Start the service
sudo systemctl start ipquorum

# Check status
sudo systemctl status ipquorum
```

## 📊 Service Management

### Check Service Status

```bash
# View service status
sudo systemctl status ipquorum

# View real-time logs
sudo journalctl -u ipquorum -f

# View download logs
sudo tail -f /opt/IBM/ip-quorum/log/download.log

# View IP Quorum application logs
sudo tail -f /opt/IBM/ip-quorum/log/ip_quorum.log.0
```

### Control the Service

```bash
# Start service
sudo systemctl start ipquorum

# Stop service
sudo systemctl stop ipquorum

# Restart service (will download new JAR if enabled)
sudo systemctl restart ipquorum

# Reload configuration
sudo systemctl daemon-reload
sudo systemctl restart ipquorum
```

### Manual Download Trigger

```bash
# Manually run the download script
sudo /usr/local/bin/ipquorum-download.sh
```

## 🔧 Configuration Reference

### Main Configuration File: `/etc/ipquorum/ipquorum.conf`

```bash
# Basic Configuration
IPQUORUM_DIR=/opt/IBM/ip-quorum
IPQUORUM_JAR=/opt/IBM/ip-quorum/ip_quorum.jar
IPQUORUM_LOG_DIR=/opt/IBM/ip-quorum/log
IPQUORUM_USER=ipquorum
IPQUORUM_GROUP=ipquorum

# Download Configuration
IPQUORUM_DOWNLOAD_ENABLED=false    # true to enable automatic download
IPQUORUM_DOWNLOAD_TOOL=go          # go, python, or bash
IPQUORUM_BACKUP_ENABLED=true       # true to backup before download

# API Configuration
API_ENDPOINT=                      # Storage Virtualize IP/hostname
VIRTUALIZE_USERNAME=               # API username
VIRTUALIZE_PASSWORD_FILE=/etc/ipquorum/.password

# mkquorumapp Configuration (for creating new Quorum apps)
IPQUORUM_MKQUORUMAPP_ENABLED=false # true to create new quorum app
IPQUORUM_PARTNERSYSTEM=            # Partner system name (REQUIRED if mkquorumapp enabled)
IPQUORUM_IP6=false                 # true for IPv6 on local system
IPQUORUM_NOMETADATA=false          # true to disable metadata
IPQUORUM_PARTNERIP6=false          # true for IPv6 on partner system

# TLS/SSL Configuration
IPQUORUM_TLS_VERIFY=false          # true to verify TLS certificates (secure mode)

# Download Tool Paths
DOWNLOAD_TOOL_GO=/usr/local/bin/ipquorum-download-go
DOWNLOAD_TOOL_PYTHON=/usr/local/bin/ipquorum-download.py
DOWNLOAD_TOOL_BASH=/usr/local/bin/ipquorum-restapi-download.sh

# IP Quorum Application Options
IPQUORUM_NAME=                     # Instance name (1-20 chars, A-Z a-z 0-9)
IPQUORUM_DEBUG=false               # Enable debug mode
IPQUORUM_EMIT=false                # Display T3 metadata headers
IPQUORUM_LOG_LOCATION=             # Custom log file path
IPQUORUM_LOG_ROTATION=5            # Number of log files (1-10)
IPQUORUM_LOG_SIZE=5120             # Log file size in KB (1024-10240)

# Java Configuration
JAVA_BIN=/usr/bin/java
JAVA_OPTS=                         # Optional Java options
```

### Password File: `/etc/ipquorum/.password`

The password file must be readable by the `ipquorum` user. There are two recommended approaches:

**Option 1: Group-readable (Recommended for service)**
```bash
# Create password file
echo 'your_password' | sudo tee /etc/ipquorum/.password > /dev/null

# Set permissions: readable by owner and group
sudo chmod 440 /etc/ipquorum/.password
sudo chown root:ipquorum /etc/ipquorum/.password

# Verify
ls -l /etc/ipquorum/.password
# Should show: -r--r----- 1 root ipquorum

# Test readability
sudo -u ipquorum cat /etc/ipquorum/.password

```

**Option 2: User-owned (Alternative)**
```bash
# Create password file
echo 'your_password' | sudo tee /etc/ipquorum/.password > /dev/null

# Set permissions: readable only by ipquorum user
sudo chmod 400 /etc/ipquorum/.password
sudo chown ipquorum:ipquorum /etc/ipquorum/.password

# Verify
ls -l /etc/ipquorum/.password
# Should show: -r-------- 1 ipquorum ipquorum

# Test readability
sudo -u ipquorum cat /etc/ipquorum/.password

```

**Security Considerations:**
- **440 (root:ipquorum)**: Root owns, ipquorum group can read - more secure
- **400 (ipquorum:ipquorum)**: ipquorum user owns and reads - simpler
- **Never use 444 or 644**: Makes password world-readable!

### IP Quorum Application Options

The service supports all IP Quorum Java application command-line options:

#### Instance Name (`IPQUORUM_NAME`)
- **Purpose**: Helps identify the IP quorum app instance in the GUI/lsquorum
- **Format**: 1-20 characters (A-Z, a-z, 0-9 only - no dashes or underscores)
- **Example**: `IPQUORUM_NAME=prodsite1`
- **Command**: `-name prodsite1`

#### Debug Mode (`IPQUORUM_DEBUG`)
- **Purpose**: Run the app in debug mode with verbose messages
- **Values**: `true` or `false` (default: false)
- **Output**: Displays verbose messages on stdout and in log files
- **Example**: `IPQUORUM_DEBUG=true`
- **Command**: `-debug`

#### T3 Metadata (`IPQUORUM_EMIT`)
- **Purpose**: Display T3 metadata header information
- **Values**: `true` or `false` (default: false)
- **Example**: `IPQUORUM_EMIT=true`
- **Command**: `-emit`

#### Log Location (`IPQUORUM_LOG_LOCATION`)
- **Purpose**: Specify custom path for log files
- **Default**: `ip_quorum.log.{timestamp}` in working directory
- **Format**: Full or relative path, allowed characters: [a-z A-Z 0-9 .-_/]
- **Example**: `IPQUORUM_LOG_LOCATION=/opt/IBM/ip-quorum/log/ip_quorum.log`
- **Command**: `-location /opt/IBM/ip-quorum/log/ip_quorum.log`

#### Log Rotation (`IPQUORUM_LOG_ROTATION`)
- **Purpose**: Maximum number of log files to keep
- **Default**: 5
- **Range**: 1-10
- **Example**: `IPQUORUM_LOG_ROTATION=10`
- **Command**: `-rotation 10`

#### Log File Size (`IPQUORUM_LOG_SIZE`)
- **Purpose**: Log file size in kilobytes
- **Default**: 5120 (5MB)
- **Range**: 1024-10240 (1MB-10MB)
- **Example**: `IPQUORUM_LOG_SIZE=10240`
- **Command**: `-size 10240`

#### Example Configuration

```bash
# /etc/ipquorum/ipquorum.conf

# Identify this instance in the GUI
IPQUORUM_NAME=prod-dc1

# Enable debug logging for troubleshooting
IPQUORUM_DEBUG=true

# Custom log location
IPQUORUM_LOG_LOCATION=/opt/IBM/ip-quorum/log/ip_quorum.log

# Keep 10 log files
IPQUORUM_LOG_ROTATION=10

# 10MB log files
IPQUORUM_LOG_SIZE=10240
```

This configuration will start the IP Quorum application with:
```bash
java -jar /opt/IBM/ip-quorum/ip_quorum.jar \
  -name prod-dc1 \
  -debug \
  -location /opt/IBM/ip-quorum/log/ip_quorum.log \
  -rotation 10 \
  -size 10240
```

## 🔒 Security Best Practices

### 1. Password File Security

```bash
# Correct permissions
-r-------- 1 root root /etc/ipquorum/.password

# Set permissions
sudo chmod 400 /etc/ipquorum/.password
sudo chown root:root /etc/ipquorum/.password
```

### 2. Service User

The service runs as a dedicated user (`ipquorum`) with:
- No login shell (`/sbin/nologin`)
- No home directory
- Limited file system access
- No root privileges

### 3. Systemd Security Features

The service includes:
- `PrivateTmp=yes` - Isolated /tmp directory
- `NoNewPrivileges=true` - Cannot gain new privileges
- `ProtectSystem=strict` - Read-only system directories
- `ProtectHome=yes` - No access to home directories
- Resource limits (memory, CPU)

### 4. Network Security

- Use HTTPS (port 7443) for API communication
- Consider using `--secure` flag if you have valid TLS certificates
- Firewall: Only port 1260/tcp needs to be open

## 🔍 Troubleshooting

### Service Won't Start

```bash
# Check service status
sudo systemctl status ipquorum

# View detailed logs
sudo journalctl -u ipquorum -n 50

# Check download logs
sudo cat /opt/IBM/ip-quorum/log/download.log

# Verify JAR file exists
ls -l /opt/IBM/ip-quorum/ip_quorum.jar

# Check Java
java -version
```

### Download Fails

```bash
# Check download logs
sudo tail -f /opt/IBM/ip-quorum/log/download.log

# Verify configuration
sudo cat /etc/ipquorum/ipquorum.conf

# Test download manually
sudo /usr/local/bin/ipquorum-download.sh

# Check API connectivity
curl -k https://YOUR_API_ENDPOINT:7443/rest/v1/auth

# Verify password file
sudo ls -l /etc/ipquorum/.password
```

### Connection Issues

```bash
# Check firewall
sudo firewall-cmd --list-ports
sudo firewall-cmd --add-port=1260/tcp --permanent
sudo firewall-cmd --reload

# Test connectivity to Storage Virtualize
curl -v telnet://YOUR_SV_IP:1260

# Check from Storage Virtualize
# SSH to SV and run:
ping -srcip4 SERVICE_IP IPQUORUM_HOST_IP
```

### Permission Issues

```bash
# Fix ownership
sudo chown -R ipquorum:ipquorum /opt/IBM/ip-quorum
sudo chmod 755 /opt/IBM/ip-quorum
sudo chmod 755 /opt/IBM/ip-quorum/log
sudo chmod 644 /opt/IBM/ip-quorum/ip_quorum.jar

# Fix password file
sudo chmod 400 /etc/ipquorum/.password
sudo chown root:root /etc/ipquorum/.password
```

## 📈 Monitoring

### Check IP Quorum Status from Storage Virtualize

```bash
# SSH to Storage Virtualize
ssh superuser@YOUR_SV_IP

# List quorum devices
lsquorum

# Expected output:
# quorum_index status id name controller_id controller_name active object_type override site_id site_name
# 3            online                                        yes    device      no               hostname/IP
```

### Monitor Service Health

```bash
# Service status
sudo systemctl is-active ipquorum

# Check if process is running
ps aux | grep ip_quorum.jar

# Check network connections
sudo netstat -putan | grep java
sudo lsof -i:1260
```

### Log Rotation

The IP Quorum application automatically rotates logs. Download logs are appended to `/opt/IBM/ip-quorum/log/download.log`.

PS: there is log options in the ip_quorum.jar file also.

To configure log rotation:

```bash
# Create logrotate configuration
sudo vi /etc/logrotate.d/ipquorum

# Add:
/opt/IBM/ip-quorum/log/download.log {
    weekly
    rotate 4
    compress
    missingok
    notifempty
    create 644 ipquorum ipquorum
}
```

## 🔄 Updating the Service

### Update JAR File

**With Automatic Download:**

```bash
# Simply restart the service
sudo systemctl restart ipquorum
```

**Manual Update:**

```bash
# Stop service
sudo systemctl stop ipquorum

# Backup current JAR
sudo cp /opt/IBM/ip-quorum/ip_quorum.jar /opt/IBM/ip-quorum/ip_quorum.jar.backup

# Copy new JAR
sudo cp new_ip_quorum.jar /opt/IBM/ip-quorum/ip_quorum.jar
sudo chown ipquorum:ipquorum /opt/IBM/ip-quorum/ip_quorum.jar

# Start service
sudo systemctl start ipquorum
```

### Update Configuration

```bash
# Edit configuration
sudo vi /etc/ipquorum/ipquorum.conf

# Reload and restart
sudo systemctl daemon-reload
sudo systemctl restart ipquorum
```

### Update Download Tool

```bash
# Replace the binary/script
sudo cp new-ipquorum-download-go /usr/local/bin/ipquorum-download-go
sudo chmod 755 /usr/local/bin/ipquorum-download-go

# Test
/usr/local/bin/ipquorum-download-go version

# Restart service to use new tool
sudo systemctl restart ipquorum
```

## 📚 Additional Resources

- [IBM Storage Virtualize IP Quorum Documentation](https://www.ibm.com/docs/en/flashsystem-7x00/8.6.x?topic=cq-ip-quorum-application-configuration)
- [IP Quorum Requirements](https://www.ibm.com/support/pages/node/7013877)
- [Original Setup Guide](readme-ipquorum-systemd.md)

## 🆚 Comparison: Original vs Improved Service

| Feature | Original Service | Improved Service |
|---------|-----------------|------------------|
| Automatic Download | ❌ No | ✅ Yes (optional) |
| Configuration File | ❌ No | ✅ Yes |
| Multiple Download Tools | ❌ No | ✅ Yes (Go/Python/Bash) |
| Automatic Backup | ❌ No | ✅ Yes |
| Download Logging | ❌ No | ✅ Yes |
| Password File Support | ❌ No | ✅ Yes |
| Installation Script | ❌ No | ✅ Yes |
| Java Auto-Install | ❌ No | ✅ Yes |
| Enhanced Security | ⚠️ Basic | ✅ Advanced |
| Resource Limits | ❌ No | ✅ Yes |
| Firewall Config | ❌ Manual | ✅ Automatic |

## 📝 Example Configurations

### Example 1: Production with Go Binary

```bash
# /etc/ipquorum/ipquorum.conf
IPQUORUM_DOWNLOAD_ENABLED=true
IPQUORUM_DOWNLOAD_TOOL=go
IPQUORUM_BACKUP_ENABLED=true
API_ENDPOINT=10.33.7.80
VIRTUALIZE_USERNAME=monitor_user
VIRTUALIZE_PASSWORD_FILE=/etc/ipquorum/.password
```

### Example 2: Development with Python

```bash
# /etc/ipquorum/ipquorum.conf
IPQUORUM_DOWNLOAD_ENABLED=true
IPQUORUM_DOWNLOAD_TOOL=python
IPQUORUM_BACKUP_ENABLED=true
API_ENDPOINT=sv-dev.example.com
VIRTUALIZE_USERNAME=admin
VIRTUALIZE_PASSWORD_FILE=/etc/ipquorum/.password
```

### Example 3: Manual JAR Management

```bash
# /etc/ipquorum/ipquorum.conf
IPQUORUM_DOWNLOAD_ENABLED=false
# Manually place JAR in /opt/IBM/ip-quorum/
```

### Example 4: Production with mkquorumapp (PBHA Setup)

```bash
# /etc/ipquorum/ipquorum.conf
# Enable download and create new quorum app
IPQUORUM_DOWNLOAD_ENABLED=true
IPQUORUM_DOWNLOAD_TOOL=go
IPQUORUM_BACKUP_ENABLED=true

# API Configuration
API_ENDPOINT=10.33.7.80
VIRTUALIZE_USERNAME=admin_user
VIRTUALIZE_PASSWORD_FILE=/etc/ipquorum/.password

# Create new Quorum app for PBHA
IPQUORUM_MKQUORUMAPP_ENABLED=true
IPQUORUM_PARTNERSYSTEM=svc_cluster02
IPQUORUM_IP6=false
IPQUORUM_NOMETADATA=false
IPQUORUM_PARTNERIP6=false

# Use secure TLS
IPQUORUM_TLS_VERIFY=true
```

**Note**: When `IPQUORUM_MKQUORUMAPP_ENABLED=true`:
- `IPQUORUM_PARTNERSYSTEM` is **REQUIRED** (the remote system name in PBHA)
- User needs **Restricted Administrator** role or higher (not just Monitor)
- This creates a new IP Quorum application on the Storage Virtualize system

### Example 5: IPv6 Environment with Secure TLS

```bash
# /etc/ipquorum/ipquorum.conf
IPQUORUM_DOWNLOAD_ENABLED=true
IPQUORUM_DOWNLOAD_TOOL=go
IPQUORUM_BACKUP_ENABLED=true

# API Configuration
API_ENDPOINT=2001:db8::1
VIRTUALIZE_USERNAME=admin_user
VIRTUALIZE_PASSWORD_FILE=/etc/ipquorum/.password

# IPv6 configuration
IPQUORUM_MKQUORUMAPP_ENABLED=true
IPQUORUM_PARTNERSYSTEM=remote_cluster
IPQUORUM_IP6=true
IPQUORUM_NOMETADATA=false
IPQUORUM_PARTNERIP6=true

# Verify TLS certificates
IPQUORUM_TLS_VERIFY=true
```

## 🤝 Contributing

Improvements and bug reports are welcome! Please submit issues or pull requests to the repository.

## 📄 License

This project follows the same license as the IBM Storage Virtualize IP Quorum application.

---

**Note**: Always test in a non-production environment first before deploying to production systems.