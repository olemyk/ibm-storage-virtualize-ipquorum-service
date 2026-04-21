# Changelog - IBM Storage Virtualize IP Quorum Systemd Service

All notable changes to the IP Quorum systemd service will be documented in this file.

## [2.0.4] - 2026-04-21 - Go Binary Standalone Releases

### 🚀 New Features

#### Standalone Go Binary Downloads
- **Go binaries as separate release assets** - Download individual binaries without the full tarball
- **Direct binary downloads** - Each platform binary available as standalone asset
- **Individual SHA256 checksums** - Each binary has its own `.sha256` checksum file
- **Automatic version detection** - Installer uses GitHub API to fetch latest version

#### Installer Improvements
- **GitHub API integration** - Automatically detects and downloads latest version
- **Improved error handling** - Better feedback when downloads fail
- **Version-specific URLs** - Uses `/releases/download/v{VERSION}/` instead of `/latest/download/`
- **Fallback support** - Works with both curl and wget

### 🔧 Technical Changes

#### Release Workflow
- Go binaries now uploaded as individual release assets
- Binaries included in both tarball and as standalone downloads
- Individual checksum files generated for each binary
- Improved asset organization in releases

#### Download URLs
**Old (broken):**
```bash
# This returns 404 because filename includes version
https://github.com/.../releases/latest/download/ipquorum-download-go-linux-amd64
```

**New (working):**
```bash
# Get latest version via API
LATEST_VERSION=$(curl -s https://api.github.com/repos/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')

# Download with version-specific URL
curl -fsSL "https://github.com/.../releases/download/v${LATEST_VERSION}/ipquorum-download-go-linux-amd64" -o ipquorum-download-go
```

### 📦 Available Binaries

Each release now includes standalone downloads for:
- `ipquorum-download-go-linux-amd64` (+ `.sha256`)
- `ipquorum-download-go-linux-arm64` (+ `.sha256`)
- `ipquorum-download-go-darwin-amd64` (+ `.sha256`)
- `ipquorum-download-go-darwin-arm64` (+ `.sha256`)

### 💡 Usage Examples

**Download latest binary automatically:**
```bash
LATEST_VERSION=$(curl -s https://api.github.com/repos/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
curl -fsSL "https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/v${LATEST_VERSION}/ipquorum-download-go-linux-amd64" -o ipquorum-download-go
chmod +x ipquorum-download-go
sudo mv ipquorum-download-go /usr/local/bin/
```

**Verify checksum:**
```bash
curl -fsSL "https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/v${LATEST_VERSION}/ipquorum-download-go-linux-amd64.sha256" -o ipquorum-download-go.sha256
sha256sum -c ipquorum-download-go.sha256
```

### ⚠️ Breaking Changes

None - This is an enhancement release that adds new download options.

---

## [2.0.3] - 2026-04-21 - Release Download Instructions Fix

### 🔧 Improvements

#### Download Instructions
- **Fixed "latest" download URL** - GitHub doesn't support `/releases/latest/download/` with version-specific filenames
- **Added GitHub API method** - Uses GitHub API to automatically get latest version number
- **Updated README.md** - Shows both automatic latest download and specific version download
- **Updated release body** - Uses `curl -fsSL` and shows actual version in Quick Start

#### GitHub Release Workflow
- **Dynamic version substitution** - Release body now shows actual version number (e.g., 2.0.3)
- **Improved heredoc syntax** - Fixed shell script generation for release body
- **Added ARCHITECTURE.md link** - Included in release documentation links

### 📚 Documentation Updates

**New download methods:**
```bash
# Automatic latest version
LATEST_VERSION=$(curl -s https://api.github.com/repos/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
curl -fsSL "https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/v${LATEST_VERSION}/ipquorum-service-${LATEST_VERSION}.tar.gz" -o ipquorum-service-${LATEST_VERSION}.tar.gz

# Or specific version
curl -fsSL https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/v2.0.3/ipquorum-service-2.0.3.tar.gz -o ipquorum-service-2.0.3.tar.gz
```

### ⚠️ Breaking Changes

None - This is a documentation fix release.

---

## [2.0.2] - 2026-04-21 - Documentation and Release Workflow Fixes

### 📚 Documentation

#### New Documentation
- **ARCHITECTURE.md** - Comprehensive system architecture documentation
  - System architecture diagrams
  - Component descriptions
  - Data flow explanations
  - Security architecture
  - Resource management details
  - Monitoring and logging guide
  - High availability considerations
  - Disaster recovery procedures
  - Troubleshooting guide

### 🔧 Improvements

#### GitHub Release Workflow
- **Fixed release filename** - Removed extra "v" prefix from tarball name
  - Old: `ipquorum-service-v2.0.1.tar.gz`
  - New: `ipquorum-service-2.0.1.tar.gz`
- **Updated download URLs** - Corrected Quick Start instructions in release body
- **Added ARCHITECTURE.md** to release package docs
- **Improved version handling** - Better variable usage in workflow

### 📦 Release Package Updates

**New files in release:**
- `docs/ARCHITECTURE.md` - System architecture documentation

**Fixed download instructions:**
```bash
# Now works correctly
VERSION="2.0.2"
wget https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/v${VERSION}/ipquorum-service-${VERSION}.tar.gz
```

### ⚠️ Breaking Changes

None - This is a documentation and workflow fix release.

---

## [2.0.1] - 2026-04-20 - Major Multi-Instance Architecture Update

### 🎉 Major Features

#### Complete Multi-Instance Architecture
- **Multi-instance systemd service** - Run multiple independent IP Quorum instances on a single host
  - Template-based systemd unit (`ipquorum@.service`)
  - Instance-specific configurations in `/etc/ipquorum/instances/`
  - Isolated JAR files, logs, and credentials per instance
  - Independent lifecycle management (start/stop/restart per instance)
- **Instance Manager CLI** - Comprehensive management tool (`ipquorum-instance-manager.sh`)
  - Interactive instance creation with guided prompts
  - List all instances with status and configuration
  - Start, stop, restart, enable, disable instances
  - View logs and status for specific instances
  - Delete instances with safety confirmations
- **Automated Installation** - Complete rewrite of installer
  - Intelligent Go binary detection (multiple naming patterns)
  - Local and remote download options for Go binaries
  - Custom URL support for enterprise environments
  - Automatic directory structure creation
  - SELinux context configuration
  - Systemd service installation and reload

#### Enhanced Documentation
- **Complete README overhaul** with ASCII banner logo
  - Multi-instance architecture emphasis
  - Split-brain prevention diagram
  - Automation-first approach
  - Clear separation of service vs. download tools
- **New DEPLOYMENT-GUIDE.md** - Comprehensive deployment documentation
  - Component architecture explanation
  - Service startup flow diagrams
  - Directory structure documentation
  - Multiple use case examples (PBHA, multi-system, DR)
- **RELEASE-INSTRUCTIONS.md** - GitHub release workflow documentation
  - Go binary release process
  - Version naming conventions (`go-v*.*.*`)
  - Integration with main service package

#### Security & Licensing
- **Apache 2.0 License** - Added open source license to project root
- **Enhanced security** - Improved credential handling
  - Password files per instance in `/var/lib/ipquorum/<instance>/`
  - Secure file permissions (400/600)
  - SELinux context support
  - No passwords in configuration files

### 🔧 Improvements

#### Installer Enhancements
- **Fixed ANSI color code rendering** - Added `-e` flag to echo commands (line 515)
- **Go binary auto-detection** - Searches for multiple naming patterns:
  - `ipquorum-download-go`
  - `ipquorum-download-go-linux-amd64`
  - `ipquorum-download-go-linux-arm64`
  - Architecture-specific variants
- **Script path detection** - Automatically finds scripts in `systemd/` subdirectory
- **Local binary installation** - Improved detection in `ipquorum-downloader/` directory
- **Custom URL support** - Download Go binaries from custom URLs
- **Better error messages** - Clear guidance on missing dependencies

#### Instance Manager Improvements
- **Clarified prompts** - Updated IBM_STORAGE_SYSTEM prompt to state it's for documentation only
- **Hidden placeholder values** - List command shows "N/A" instead of `<HOSTNAME_OR_IP>`
- **Full path usage** - Resolves sudo PATH issues with absolute paths
- **Permanent alias suggestions** - Shows how to add aliases to shell profile
- **Interactive validation** - Validates configuration during creation
- **Status integration** - Shows systemd status in list output

#### Configuration Updates
- **Updated instance.conf.template** - Clarified IBM_STORAGE_SYSTEM field (lines 23-28)
  - "OPTIONAL - for documentation/identification only"
  - "Not used for connectivity - API_ENDPOINT is used instead"
- **Example configurations** - Updated all examples with clear comments
- **Validation improvements** - Better detection of placeholder values

### 📦 New Files & Structure

#### Core Service Files
- `ipquorum-systemd/multi-instance/ipquorum@.service` - Template systemd unit
- `ipquorum-systemd/multi-instance/ipquorum-instance-manager.sh` - Instance management CLI
- `ipquorum-systemd/multi-instance/install-ipquorum-service.sh` - Automated installer
- `ipquorum-systemd/multi-instance/instance.conf.template` - Configuration template

#### Supporting Scripts
- `ipquorum-systemd/multi-instance/ipquorum-start-multi.sh` - Instance startup script
- `ipquorum-systemd/multi-instance/ipquorum-download-multi.sh` - Download script
- `ipquorum-systemd/multi-instance/ipquorum-validate-multi.sh` - Validation script

#### Troubleshooting Tools
- `ipquorum-systemd/multi-instance/troubleshoot-tools/diagnose-selinux.sh`
- `ipquorum-systemd/multi-instance/troubleshoot-tools/fix-selinux.sh`
- `ipquorum-systemd/multi-instance/troubleshoot-tools/fix-permissions.sh`

#### Example Configurations
- `ipquorum-systemd/multi-instance/examples/system1.conf`
- `ipquorum-systemd/multi-instance/examples/datacenter-a.conf`
- `ipquorum-systemd/multi-instance/examples/partition-prod.conf`

#### Documentation
- `LICENSE` - Apache 2.0 license
- `readme.md` - Complete rewrite with multi-instance focus
- `ipquorum-systemd/multi-instance/DEPLOYMENT-GUIDE.md`
- `ipquorum-download-go/RELEASE-INSTRUCTIONS.md`

#### Archive Structure
- `archive/` - Deprecated single-instance components
- `archive/README.md` - Deprecation notice and migration guide
- `ipquorum-systemd/DEPRECATION-NOTICE.md`

### 🐛 Bug Fixes

- **Fixed ANSI color codes** - Colors now render properly in installer output
- **Fixed sudo PATH issues** - Scripts use absolute paths
- **Fixed Go binary detection** - Handles multiple naming patterns
- **Fixed script path detection** - Works from any directory
- **Fixed placeholder validation** - Better detection of unconfigured values
- **Fixed SELinux contexts** - Proper labeling for all directories

### 📊 Architecture Changes

#### Directory Structure
```
/etc/ipquorum/
├── instances/
│   ├── system1.conf
│   ├── system2.conf
│   └── datacenter-a.conf
└── ipquorum@.service (symlink)

/var/lib/ipquorum/
├── system1/
│   ├── ip_quorum.jar
│   ├── .password
│   └── logs/
├── system2/
│   └── ...
└── datacenter-a/
    └── ...

/usr/local/bin/
├── ipquorum-instance-manager.sh
├── ipquorum-start-multi.sh
├── ipquorum-download-multi.sh
└── ipquorum-validate-multi.sh
```

#### Service Management
```bash
# Old (single instance)
systemctl start ipquorum

# New (multi-instance)
systemctl start ipquorum@system1
systemctl start ipquorum@system2
systemctl start ipquorum@datacenter-a

# Or use instance manager
ipquorum-instance-manager.sh start system1
```

### 🚀 Use Cases

#### Multiple Storage Systems
Run separate IP Quorum instances for different IBM Storage Virtualize clusters:
```bash
ipquorum-instance-manager.sh create svc_cluster01
ipquorum-instance-manager.sh create svc_cluster02
ipquorum-instance-manager.sh create flashsystem_prod
```

#### Partner-Based High Availability (PBHA)
Configure quorum for stretched cluster configurations:
```bash
ipquorum-instance-manager.sh create datacenter-a
ipquorum-instance-manager.sh create datacenter-b
```

#### Production + DR Environments
Separate instances for production and disaster recovery:
```bash
ipquorum-instance-manager.sh create prod-primary
ipquorum-instance-manager.sh create dr-secondary
```

### 🔄 Migration from v2.0.0

#### Automated Migration
The new installer preserves existing single-instance configurations:
1. Detects existing `/etc/ipquorum/ipquorum.conf`
2. Creates backup in `/etc/ipquorum/backup/`
3. Migrates to multi-instance as `default` instance
4. Preserves all settings and credentials

#### Manual Migration
```bash
# 1. Install new multi-instance system
sudo ./install-ipquorum-service.sh

# 2. Create instance from old config
sudo ipquorum-instance-manager.sh create system1

# 3. Stop old service
sudo systemctl stop ipquorum-old
sudo systemctl disable ipquorum-old

# 4. Start new instance
sudo ipquorum-instance-manager.sh start system1
```

### ⚠️ Breaking Changes

- **Configuration location changed** - Now in `/etc/ipquorum/instances/`
- **Service name changed** - Use `ipquorum@<instance>` instead of `ipquorum`
- **JAR location changed** - Now in `/var/lib/ipquorum/<instance>/`
- **Log location changed** - Now in `/var/lib/ipquorum/<instance>/logs/`
- **Password file location changed** - Now in `/var/lib/ipquorum/<instance>/`

### 📈 Performance & Reliability

- **Isolated instances** - Failures in one instance don't affect others
- **Independent restarts** - Restart policies per instance
- **Resource limits** - CPU and memory limits per instance
- **Better logging** - Separate logs per instance
- **Health monitoring** - Status checks per instance

### 🔮 Future Enhancements

Planned for v2.1.0:
- [ ] Web UI for instance management
- [ ] Prometheus metrics per instance
- [ ] Automated health checks with alerting
- [ ] Ansible playbook for deployment
- [ ] Container/Kubernetes support
- [ ] Backup and restore functionality
- [ ] Configuration validation tool
- [ ] Performance monitoring dashboard

### 🤝 Contributing

This release includes contributions and testing feedback from production deployments on RHEL 9.4.

### 📄 License

This project is now licensed under Apache License 2.0. See LICENSE file for details.

---

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