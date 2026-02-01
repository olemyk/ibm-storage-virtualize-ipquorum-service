# Changelog - IBM Storage Virtualize IP Quorum Systemd Service

All notable changes to the IP Quorum systemd service will be documented in this file.

## [2.0.0] - 2026-02-01 - Initial Release with Automation

### ✨ New Features

#### mkquorumapp Configuration Support
- **Create new IP Quorum applications** during download
  - `IPQUORUM_MKQUORUMAPP_ENABLED` - Enable/disable mkquorumapp call
  - `IPQUORUM_PARTNERSYSTEM` - Partner system name (REQUIRED when enabled)
  - `IPQUORUM_IP6` - IPv6 support for local system
  - `IPQUORUM_NOMETADATA` - Metadata configuration
  - `IPQUORUM_PARTNERIP6` - IPv6 support for partner system
- **Automatic validation** ensures partnersystem is set when mkquorumapp is enabled
- **Support for PBHA setups** - Create quorum apps for Partner-Based High Availability
- **Works with all download tools** - Go, Python, and Bash versions

#### TLS/SSL Configuration
- **TLS certificate verification control**
  - `IPQUORUM_TLS_VERIFY=true` - Verify TLS certificates (secure mode)
  - `IPQUORUM_TLS_VERIFY=false` - Skip verification (insecure mode, default)
- **Flexible security** - Choose based on your environment

### 🔧 Improvements

#### Enhanced Download Script
- **Configuration validation** for mkquorumapp parameters
- **Boolean value validation** for all true/false settings
- **Detailed logging** of mkquorumapp configuration
- **Error messages** guide users to fix configuration issues

#### Updated Documentation
- **New configuration examples** for PBHA setups
- **IPv6 environment examples** with secure TLS
- **Clear explanations** of when to use mkquorumapp
- **Permission requirements** documented (Restricted Administrator role needed)

### 📝 Configuration Changes

New variables added to `/etc/ipquorum/ipquorum.conf`:
```bash
# mkquorumapp Configuration
IPQUORUM_MKQUORUMAPP_ENABLED=false
IPQUORUM_PARTNERSYSTEM=
IPQUORUM_IP6=false
IPQUORUM_NOMETADATA=false
IPQUORUM_PARTNERIP6=false

# TLS/SSL Configuration
IPQUORUM_TLS_VERIFY=false
```

### ⚠️ Important Notes

- **mkquorumapp requires higher privileges**: User needs Restricted Administrator role or higher (not just Monitor role)
- **Partnersystem is mandatory**: When `IPQUORUM_MKQUORUMAPP_ENABLED=true`, you must set `IPQUORUM_PARTNERSYSTEM`
- **Backward compatible**: Existing configurations continue to work (defaults to disabled)

## [2.0.0] - 2026-01-30 - Improved Version

### 🎉 Major Features Added

#### Automatic Download Capability
- **Pre-start download script** (`ipquorum-download.sh`)
  - Downloads `ip_quorum.jar` before service starts
  - Configurable to use Go, Python, or Bash download tools
  - Optional - can be disabled for manual JAR placement
  - Automatic backup before download
  - Restore backup on download failure
  - Comprehensive logging to `/opt/IBM/ip-quorum/log/download.log`

#### Centralized Configuration
- **Configuration file** (`/etc/ipquorum/ipquorum.conf`)
  - All settings in one place
  - Environment-based configuration
  - Support for multiple download tools
  - Secure password file authentication
  - Extensive inline documentation

#### Enhanced Security
- **Password file support** (`/etc/ipquorum/.password`)
  - No passwords in configuration files
  - Automatic permission validation (400/600)
  - Secure storage recommendations
- **Improved systemd security directives**
  - `ProtectSystem=strict` - Read-only system directories
  - `ProtectHome=yes` - No access to home directories
  - `NoNewPrivileges=true` - Cannot gain new privileges
  - `PrivateTmp=yes` - Isolated /tmp directory
  - `ReadWritePaths` - Limited write access
  - `ReadOnlyPaths` - Protected configuration

#### Installation Automation
- **Installation script** (`install-ipquorum-service.sh`)
  - Automated setup process
  - Automatic Java detection and installation
  - User and group creation
  - Directory structure creation
  - Firewall configuration (firewalld/ufw)
  - Service installation and configuration
  - Backup of existing configurations

#### Resource Management
- **Resource limits**
  - Memory limit: 512MB
  - CPU quota: 50%
  - File descriptor limit: 65536
- **Restart policy**
  - Automatic restart on failure
  - 10-second restart delay
  - Rate limiting (5 restarts per 5 minutes)

#### Logging Improvements
- **Structured logging**
  - Systemd journal integration
  - Separate download logs
  - Timestamp-based log entries
  - Log level indicators (INFO, WARN, ERROR)

### 📝 Configuration Options

#### Download Configuration
- `IPQUORUM_DOWNLOAD_ENABLED` - Enable/disable automatic download
- `IPQUORUM_DOWNLOAD_TOOL` - Choose download tool (go/python/bash)
- `IPQUORUM_BACKUP_ENABLED` - Enable/disable automatic backup
- `DOWNLOAD_TOOL_GO` - Path to Go binary
- `DOWNLOAD_TOOL_PYTHON` - Path to Python script
- `DOWNLOAD_TOOL_BASH` - Path to Bash script

#### API Configuration
- `API_ENDPOINT` - Storage Virtualize IP/hostname
- `VIRTUALIZE_USERNAME` - API username
- `VIRTUALIZE_PASSWORD_FILE` - Password file location

#### Service Configuration
- `IPQUORUM_DIR` - Installation directory
- `IPQUORUM_JAR` - JAR file location
- `IPQUORUM_LOG_DIR` - Log directory
- `IPQUORUM_USER` - Service user
- `IPQUORUM_GROUP` - Service group
- `JAVA_BIN` - Java binary location
- `JAVA_OPTS` - Additional Java options

### 🔧 Technical Improvements

#### Download Script Features
- Multi-tool support (Go, Python, Bash)
- Automatic tool detection and validation
- Password file permission checking
- File size validation
- Automatic ownership and permission setting
- Comprehensive error handling
- Backup and restore functionality

#### Service File Enhancements
- Environment file integration
- Pre-start hook for downloads
- Enhanced security directives
- Resource limits
- Better restart policy
- Systemd journal logging

#### Installation Script Features
- Root privilege checking
- Package manager detection (dnf/yum/apt/zypper)
- Automatic Java installation
- Dependency validation
- User/group creation
- Directory structure setup
- Firewall configuration
- Backup of existing files

### 📚 Documentation

#### New Documentation Files
- `README-IMPROVED-SERVICE.md` - Comprehensive guide
  - Quick start installation
  - Configuration reference
  - Security best practices
  - Troubleshooting guide
  - Monitoring instructions
  - Update procedures
  - Example configurations
- `CHANGELOG.md` - This file
- Inline documentation in all scripts

#### Documentation Sections
- Prerequisites and requirements
- Installation instructions
- Configuration examples
- Service management commands
- Security guidelines
- Troubleshooting procedures
- Monitoring and logging
- Update procedures
- Comparison with original version

### 🔒 Security Enhancements

#### Authentication
- Password file-based authentication
- No passwords in configuration files
- Automatic permission validation
- Secure file storage recommendations

#### Service Isolation
- Dedicated service user (no root)
- No login shell
- Limited file system access
- Protected system directories
- Isolated temporary directory
- No privilege escalation

#### Network Security
- HTTPS communication (port 7443)
- TLS certificate validation options
- Minimal port exposure (1260/tcp)

### 🐛 Bug Fixes
- Fixed potential race conditions in service startup
- Improved error handling in download process
- Better validation of configuration parameters
- Enhanced logging for troubleshooting

### ⚡ Performance Improvements
- Faster startup with pre-validation
- Efficient download with retry logic
- Resource limits prevent runaway processes
- Optimized logging

### 🔄 Backward Compatibility
- Original service file preserved (`ibm-virtualize-ipquorum.service`)
- New service file separate (`ibm-virtualize-ipquorum-improved.service`)
- Can run both versions side-by-side (different service names)
- Migration path documented

### 📊 Comparison with Original Version

| Feature | Original (v1.0) | Improved (v2.0) |
|---------|----------------|-----------------|
| Automatic Download | ❌ | ✅ |
| Configuration File | ❌ | ✅ |
| Multiple Download Tools | ❌ | ✅ |
| Automatic Backup | ❌ | ✅ |
| Download Logging | ❌ | ✅ |
| Password File Support | ❌ | ✅ |
| Installation Script | ❌ | ✅ |
| Java Auto-Install | ❌ | ✅ |
| Enhanced Security | ⚠️ Basic | ✅ Advanced |
| Resource Limits | ❌ | ✅ |
| Firewall Config | ❌ Manual | ✅ Automatic |
| Comprehensive Docs | ⚠️ Basic | ✅ Extensive |

### 🎯 Use Cases

#### Production Environment
- Automatic download with Go binary (fastest)
- Password file authentication
- Automatic backups enabled
- Resource limits enforced
- Enhanced security enabled

#### Development Environment
- Flexible download tool selection
- Easy configuration changes
- Detailed logging for debugging
- Quick restart for testing

#### Air-Gapped Environment
- Disable automatic download
- Manual JAR placement
- No external dependencies
- Full offline operation

### 🚀 Migration Guide

#### From Original to Improved Version

1. **Install improved version**
   ```bash
   sudo ./install-ipquorum-service.sh
   ```

2. **Stop original service**
   ```bash
   sudo systemctl stop ipquorum-original
   sudo systemctl disable ipquorum-original
   ```

3. **Configure new service**
   ```bash
   sudo vi /etc/ipquorum/ipquorum.conf
   ```

4. **Copy existing JAR (if not using download)**
   ```bash
   sudo cp /old/path/ip_quorum.jar /opt/IBM/ip-quorum/
   ```

5. **Start new service**
   ```bash
   sudo systemctl enable ipquorum
   sudo systemctl start ipquorum
   ```

### 📦 Files Included

#### Service Files
- `ibm-virtualize-ipquorum-improved.service` - Systemd service unit
- `ipquorum-download.sh` - Pre-start download script
- `ipquorum.conf` - Configuration file template

#### Installation
- `install-ipquorum-service.sh` - Automated installation script

#### Documentation
- `README-IMPROVED-SERVICE.md` - Comprehensive guide
- `CHANGELOG.md` - This file
- `readme-ipquorum-systemd.md` - Original documentation (preserved)

### 🔮 Future Enhancements

Planned for future versions:
- [ ] Scheduled download checks (daily/weekly)
- [ ] Health check endpoint
- [ ] Prometheus metrics export
- [ ] Email notifications on failures
- [ ] Web UI for configuration
- [ ] Ansible role integration
- [ ] Container/Kubernetes deployment
- [ ] High availability configuration
- [ ] Automatic version checking
- [ ] Configuration validation tool

### 🤝 Contributing

Contributions are welcome! Please:
1. Test changes in non-production environment
2. Update documentation
3. Follow existing code style
4. Add entries to CHANGELOG

### 📄 License

This project follows the same license as the IBM Storage Virtualize IP Quorum application.

---

## [1.0.0] - Original Version

### Initial Release
- Basic systemd service file
- Manual JAR placement
- Simple configuration
- Basic security directives
- Manual installation process

### Features
- Service runs as dedicated user
- Automatic restart on failure
- Basic file system protection
- Working directory configuration
- Systemd integration

### Limitations
- No automatic download
- No centralized configuration
- Manual setup required
- Limited security features
- Basic documentation
- No installation automation

---

**Note**: Version 2.0.0 is a major upgrade with breaking changes in configuration. Please review the migration guide before upgrading production systems.