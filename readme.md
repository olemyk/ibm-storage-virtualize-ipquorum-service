

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
║  │  IBM Storage Virtualize - Quorum Service for High Availability         │  ║
║  │  Automated Download • Systemd Integration                              │  ║
║  └────────────────────────────────────────────────────────────────────────┘  ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

# Automated IP Quorum Service for IBM Storage Virtualize

**Automate your IP Quorum deployment** with intelligent installation, automatic JAR downloads, and multi-instance management for IBM Storage Virtualize (FlashSystem & SVC).

## ✨ Key Features

- 🚀 **Fully Automated Installation** - One command setup with intelligent configuration
- 🔄 **Multi-Instance Support** - Run multiple IP Quorum instances on a single host
- 📥 **Automatic JAR Download** - Go-powered download tool (no dependencies!)
- 🎯 **Instance Manager** - Easy create, configure, start, stop, and monitor instances
- 🔒 **Production-Ready Security** - Systemd hardening, SELinux support, secure credentials


---

## 🎯 What is IP Quorum?

### Split-Brain Prevention Concept

```
    ┌──────────────────────────────────────────────────────┐
    │                                                      │
    │   Site A Flashsystem        Site B  Flashsystem      │
    │   ┌────┐                    ┌────┐                   │
    │   │ ██ │ ←─────────────────→│ ██ │                   │
    │   └────┘  ISL Link Failure  └────┘                   │
    │      ↓                         ↓                     │
    │      │                         │                     │
    │      └────────→ ┌────┐ ←───────┘                     │
    │                 │ Q  │  IP Quorum                    │
    │                 └────┘  Decides!                     │
    │                                                      │
    │        Prevents Split-Brain Scenarios                │
    │                                                      │
    └──────────────────────────────────────────────────────┘
```

**IP Quorum** is a tie-breaker service that prevents split-brain scenarios in IBM Storage Virtualize clusters. When exactly half of the nodes become unavailable due to a Link or SAN Switch fault, the IP Quorum application determines which nodes can continue processing host operations, ensuring data integrity and preventing independent I/O processing by both halves of the system.

**Powers:** IBM Storage FlashSystem and IBM SAN Volume Controller (SVC)

---

## 🚀 Quick Start - Automated Installation

### One-Command Installation

```bash
# Download latest release (automatically gets the latest version)
LATEST_VERSION=$(curl -s https://api.github.com/repos/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
curl -fsSL "https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/v${LATEST_VERSION}/ipquorum-service-${LATEST_VERSION}.tar.gz" -o ipquorum-service-${LATEST_VERSION}.tar.gz
tar -xzf ipquorum-service-${LATEST_VERSION}.tar.gz
cd ipquorum-service-${LATEST_VERSION}
sudo ./install-ipquorum-service.sh

# Or download specific version (e.g., 2.0.5)
curl -fsSL https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/v2.0.5/ipquorum-service-2.0.5.tar.gz -o ipquorum-service-2.0.5.tar.gz
tar -xzf ipquorum-service-2.0.5.tar.gz
cd ipquorum-service-2.0.5
sudo ./install-ipquorum-service.sh
```

The installer will:
1. ✅ Install systemd service files
2. ✅ Download Go binary - Included in Release Package (or use Python/Bash) 
3. ✅ Create directory structure
4. ✅ Set up instance manager
5. ✅ Configure permissions and security

### Create Your First Instance

```bash
# Create instance interactively (recommended)
sudo /usr/local/bin/ipquorum-instance-manager.sh create svc_cluster01

# Or use the alias (after adding to ~/.bashrc)
echo "alias ipqm='sudo /usr/local/bin/ipquorum-instance-manager.sh'" >> ~/.bashrc
source ~/.bashrc
```
```bash
ipqm create svc_cluster01

# Prompts:
# 1. IP Quorum Name (shown in IBM Storage Virtualize):
# 2. IBM Storage System name (optional)
# 3. Description (optional)
# 4. Location (optional)
# 5. Enable automatic JAR download? (yes/no) [yes]
# 6. API endpoint (if download enabled)
# 7. Username (if download enabled)
# 8. Password (if download enabled)
# 9. Create new quorum app? (yes/no) [yes] 
# 10. Partner system name (if mkquorumapp enabled)
# 11. IPv6 settings (if mkquorumapp enabled)
```

```bash
# Start the instance
 ipqm enable svc_cluster01
 ipqm start svc_cluster01
```

```bash
# Check status
 ipqm status svc_cluster01

    [packer@rhel94 ipquorum-service-2.0.4]$ ipqm status svc_cluster01
    ● ipquorum@svc_cluster01.service - IBM Storage Virtualize IP Quorum Service (svc_cluster01)
        Loaded: loaded (/etc/systemd/system/ipquorum@.service; enabled; preset: disabled)
        Active: active (running) since Wed 2026-04-22 11:57:04 CEST; 2h 38min ago
        CGroup: /system.slice/system-ipquorum.slice/ipquorum@svc_cluster01.service
                ├─35678 bash /usr/local/bin/ipquorum-start-multi.sh svc_cluster01
                └─35683 /usr/bin/java -jar /var/lib/ipquorum/svc_cluster01/ip_quorum.jar -name rhelvm01 -location /var/log/ipquorum/svc_cluster01/ip_quorum.log

    Apr 22 11:57:08 rhel94 ipquorum-svc_cluster01[35683]: Waiting for UID
    Apr 22 11:57:08 rhel94 ipquorum-svc_cluster01[35683]: Waiting for UI
```

```bash
# Check info about instance
ipqm info svc_cluster01
╔════════════════════════════════════════════════════════════════╗
║  Instance Information: svc_cluster01
╚════════════════════════════════════════════════════════════════╝

General:
  Instance Name:        svc_cluster01
  IP Quorum Name:       ipquorumsrv (shown in IBM Storage Virtualize)
  IBM Storage System:   SVC Prod
  Description:          SVC SV2
  Location:             OSLO-RACK1

Connection:
  API Endpoint:         10.33.7.80
  Username:             superuser
  TLS Verify:           false

Configuration:
  Download Enabled:     true
  Download Tool:        go
  mkquorumapp Enabled:  true
  Partner System:       svc_cluster02

Paths:
  Configuration:        /etc/ipquorum/instances/svc_cluster01.conf
  Password File:        /var/lib/ipquorum/.passwords/svc_cluster01.password
  Data Directory:       /var/lib/ipquorum/svc_cluster01
  Log Directory:        /var/log/ipquorum/svc_cluster01
  JAR File:             /var/lib/ipquorum/svc_cluster01/ip_quorum.jar

Service Status:
  Status:               Running
  Enabled:              Yes

```


```bash
# Check logs for instance
 ipqm logs svc_cluster01

[packer@rhel94 ipquorum-service-2.0.4]$ ipqm logs svc_cluster01
=== Systemd Journal Logs ===
Apr 22 11:56:59 rhel94 ipquorum-svc_cluster01[35667]: 2026-04-22 11:56:59 - INFO - Get Token, please wait...
Apr 22 11:56:59 rhel94 ipquorum-svc_cluster01[35667]: 2026-04-22 11:56:59 - INFO - Auth attempt 1/

=== Download Logs ===
[2026-04-22 11:49:25] [svc_cluster01] [INFO] === IP Quorum Download Script Started for Instance: svc_cluster01 ===
[2026-04-22 11:49:25] [svc_cluster01] [INFO] IBM Storage System: svc_cluster01-prod
[2026-04-22 11:49:25] [svc_cluster01] [INFO] mkquorumapp is enabled, validating configuration...
[2026-04-22 11:49:25] [svc_cluster01] [INFO] mkquorumapp configuration validated:

=== IP Quorum Application Logs ===
==> /var/log/ipquorum/svc_cluster01/ip_quorum.log.17768513779750.0 <==
2026-04-22 11:55:18:692 10.33.7.90 [13] FINE: >Msg [protocol=1, sequence=25, 


```

**That's it!** Your IP Quorum service is running and will automatically:
- Download the latest JAR file from your Storage Virtualize system
- Start on boot
- Restart on failure
- Log to centralized location

---
## 📺 Visual Guide - IP Quorum Service installation demo

![Animated demonstration showing the complete IP Quorum Service installation process including download, configuration, and service startup](ipquorum-service-install.gif)
*Complete installation and configuration demonstration*

---

## 📋 Prerequisites

### Required
- **Linux host** (RHEL/CentOS/Rocky/AlmaLinux 8+, Ubuntu 20.04+)
- **Java Runtime** (OpenJDK 8-17) OpenJDK 11 in installed automatically
- **Network connectivity** to IBM Storage Virtualize:
  - Port **1260/TCP** (IP Quorum communication)
  - Port **7443/HTTPS** (REST API for automatic download)
- **Valid credentials**:
  - **Monitor** role (Have permission for downloading JAR)
  - **Restricted Administrator** role (have permission to download and creating New Quorum App ipquorum.jar)

### Optional (Installed Automatically)
- **Go binary** (recommended - no dependencies, fast)
- **Python 3.8+** or **Bash** (alternative download tools)

---

## 🌟 Multi-Instance Architecture

Run **multiple independent IP Quorum instances** on a single host - perfect for:
- Multiple storage systems
- Different data centers
- Production and DR environments
- PBHA (PowerHA) configurations

### Architecture Overview

```
┌────────────────────────────────────────────────────────────────────┐
│                         Linux Host (RHEL/Ubuntu)                   │
│                                                                    │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Systemd Service Manager                  │   │
│  │                                                             │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │   │
│  │  │ ipquorum@    │  │ ipquorum@    │  │ ipquorum@    │       │   │
│  │  │ system1      │  │ system2      │  │ datacenter-a │       │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘       │   │
│  └─────────┼──────────────────┼──────────────────┼─────────────┘   │
│            │                  │                  │                 │
│  ┌─────────▼──────────────────▼──────────────────▼─────────────┐   │
│  │              Instance-Specific Resources                    │   │
│  │                                                             │   │
│  │  /etc/ipquorum/instances/                                   │   │
│  │  ├── system1.conf          ← Configuration                  │   │
│  │  ├── system2.conf                                           │   │
│  │  └── datacenter-a.conf                                      │   │
│  │                                                             │   │
│  │  /var/lib/ipquorum/                                         │   │
│  │  ├── system1/                                               │   │
│  │  │   ├── ip_quorum.jar     ← JAR file                       │   │
│  │  │   ├── .password         ← Credentials                    │   │
│  │  │   └── logs/             ← Log files                      │   │
│  │  ├── system2/                                               │   │
│  │  └── datacenter-a/                                          │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                    │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │              Management Tools                                 │ │
│  │                                                               │ │
│  │  /usr/local/bin/                                              │ │
│  │  ├── ipquorum-instance-manager.sh  ← CLI management           │ │
│  │  ├── ipquorum-start-multi.sh       ← Startup script           │ │
│  │  ├── ipquorum-download-multi.sh    ← Download script          │ │
│  │  └── ipquorum-validate-multi.sh    ← Validation script        │ │
│  └───────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────┘
         │                    │                    │
         │ HTTPS:7443         │ HTTPS:7443         │ HTTPS:7443
         │ IPQuorum 1260.     │ IPQuorum 1260      │ IPQuorum 1260
         ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ IBM Storage     │  │ IBM Storage     │  │ IBM Storage     │
│ Virtualize      │  │ Virtualize      │  │ Virtualize      │
│ System 1        │  │ System 2        │  │ Datacenter A    │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

### Instance Management

```bash
# List all instances
 ipqm list

# Create multiple instances
 ipqm create datacenter-a
 ipqm create datacenter-b
 ipqm create production

# Manage instances
 ipqm start datacenter-a
 ipqm stop datacenter-b
 ipqm restart production
 ipqm status datacenter-a
 ipqm logs datacenter-a -f

# Enable/disable auto-start
 ipqm enable datacenter-a
 ipqm disable datacenter-b

# Delete instance
 ipqm delete old-system
```

Each instance has:
- ✅ **Isolated configuration** (`/etc/ipquorum/instances/<name>.conf`)
- ✅ **Separate credentials** (`/var/lib/ipquorum/<name>/.password`)
- ✅ **Independent JAR file** (`/var/lib/ipquorum/<name>/ip_quorum.jar`)
- ✅ **Dedicated logs** (`/var/log/ipquorum/<name>/`)
- ✅ **Systemd service** (`ipquorum@<name>.service`)

📖 **[Complete Multi-Instance Guide](ipquorum-systemd/multi-instance/README.md)**

---

## 📥 Automatic JAR Download - Go-Powered

The installer includes a **Go-based download tool** that automatically fetches the IP Quorum JAR from your Storage Virtualize system via REST API.

### Why Go?

| Feature | Go (Default) | Python | Bash |
|---------|-------------|--------|------|
| **Dependencies** | ✅ None | ⚠️ Python 3.8+ | ⚠️ curl, jq |
| **Performance** | ✅ <10ms startup | ✅ ~100ms | ✅ Fast |
| **Memory** | ✅ <20MB | ⚠️ ~50MB | ✅ Low |
| **Cross-Platform** | ✅ Linux/macOS/Windows | ✅ Yes | ⚠️ Linux/macOS |
| **Single Binary** | ✅ Yes | ❌ No | ❌ No |

### Download Tool Features

- ✅ **Automatic authentication** with retry logic
- ✅ **Create Quorum Apps** for PBHA configurations
- ✅ **Secure password handling** (prompt, file, or environment)
- ✅ **TLS configuration** (insecure by default, secure option available)
- ✅ **Smart error handling** with detailed diagnostics

### Manual Download of IPQuorum Jar file from IBM Storage Virtualize - Example

```bash
# Interactive (most secure)
ipquorum-download \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-prompt \
  --mkquorumapp --partnersystem svc_cluster02 \
  --download

# Automated with password file
echo 'password' > ~/.ipquorum_pass && chmod 400 ~/.ipquorum_pass
ipquorum-download \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-file ~/.ipquorum_pass \
  --download
```

📖 **[Go Download Tool Documentation](ipquorum-download-go/README.md)**

---

## 🔧 Manual Configuration of IPQuorum Service config file

### Instance Configuration File

Each instance has a configuration file at `/etc/ipquorum/instances/<name>.conf`:

```bash
# Instance identification
INSTANCE_NAME=svc_cluster01
IBM_STORAGE_SYSTEM="Production-SAN"  # Optional, for documentation

# API Configuration
API_ENDPOINT=10.33.7.80
API_PORT=7443
VIRTUALIZE_USERNAME=superuser

# Download settings
DOWNLOADIPQ=yes
DOWNLOAD_TOOL=go  # or python, bash

# Quorum App creation (for PBHA)
mkquorumapp=yes
partnersystem=svc_cluster02
ip6=false
partnerip6=false
nometadata=false

# Security
insecure_api=--insecure  # Use --secure for production with valid certs
```

### Password Management

```bash
# Set password securely
echo 'your_password' | sudo tee /var/lib/ipquorum/svc_cluster01/.password > /dev/null
sudo chmod 400 /var/lib/ipquorum/svc_cluster01/.password
sudo chown root:root /var/lib/ipquorum/svc_cluster01/.password
```

---

## 📊 Monitoring & Troubleshooting

### Check Service Status

```bash
# Instance status
 ipqm status svc_cluster01

# Info status
 ipqm info svc_cluster01

# Systemd status
 sudo systemctl status ipquorum@svc_cluster01

# View logs
 ipqm logs svc_cluster01 -f
 sudo journalctl -u ipquorum@svc_cluster01 -f

# Check from Storage Virtualize
ssh superuser@YOUR_SV_IP
lsquorum
```

### Troubleshooting Tools

```bash
# Validate instance configuration
sudo /usr/local/bin/ipquorum-validate-multi.sh svc_cluster01

# Fix permissions
sudo /usr/local/bin/fix-permissions.sh svc_cluster01

# Diagnose SELinux issues
sudo /usr/local/bin/diagnose-selinux.sh svc_cluster01

# Fix SELinux contexts
sudo /usr/local/bin/fix-selinux.sh svc_cluster01
```

### Common Issues

**Connection refused:**
```bash
# Check firewall
sudo firewall-cmd --list-ports
sudo firewall-cmd --add-port=1260/tcp --permanent
sudo firewall-cmd --reload
```

**Authentication failed:**
```bash
# Validate the config
ipqm validate svc_cluster01
# Verify credentials
cat /var/lib/ipquorum/svc_cluster01/.password
# Check user role on Storage Virtualize
ssh superuser@YOUR_SV_IP "lsuser superuser"
```

**JAR download fails:**
```bash
# Test API connectivity
curl -k https://YOUR_API_ENDPOINT:7443/rest/v1/auth
# Check download tool
ipquorum-download --help
```

---

## 🔒 Security Best Practices

### Password Security
- ✅ Use `--pass-prompt` for interactive password entry
- ✅ Use password files with `chmod 400` for automation
- ❌ Never use `--pass` with password in command line
- ✅ Store passwords in `/var/lib/ipquorum/<instance>/.password`

### TLS Configuration
- ✅ Use `--secure` flag in production with valid certificates
- ⚠️ `--insecure` is default for self-signed certificates
- ✅ Verify certificates when possible

### Service Hardening
- ✅ Runs as dedicated `ipquorum` user (not root)
- ✅ Systemd security features enabled
- ✅ SELinux contexts properly configured
- ✅ File permissions restricted (400 for passwords, 600 for configs)

---

## 📚 Documentation

### Deployment Guides
- **[Multi-Instance Service](ipquorum-systemd/multi-instance/README.md)** - Complete guide (recommended)
- **[Deployment Guide](DEPLOYMENT-GUIDE.md)** - Distribution and packaging options
- **[Architecture Overview](ipquorum-systemd/multi-instance/ARCHITECTURE.md)** - System design and components

### Download Tools
- **[Go Download Tool](ipquorum-download-go/README.md)** - Recommended, single binary
- **[Python Download Tool](ipquorum-download/python-version/README.md)** - Feature-rich alternative
- **[Bash Download Script](ipquorum-download/bash-version/README.md)** - Lightweight option

### Additional Resources
- **[Security Guide](ipquorum-download/python-version/SECURITY-GUIDE.md)** - Best practices
- **[Changelog](ipquorum-systemd/CHANGELOG.md)** - Version history
- **[IBM Documentation](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=quorum-ip-application)** - Official docs

---

## 🎯 Use Cases

### PBHA Storage System
```bash
# Simple setup for one storage system
 ipqm create production
# Configure and start
 ipqm enable production
 ipqm start production
```

### Multiple Storage Systems in PBHA
```bash
# Create instances for each system
 ipqm create datacenter-a
 ipqm create datacenter-b
 ipqm create dr-site

# Start all instances
sudo systemctl start ipquorum@{datacenter-a,datacenter-b,dr-site}
```

### Storage System PBHA Manual Configuration
```bash
# Create Quorum App for partner system
 ipqm create svc_cluster01
# Edit config to set partnersystem=svc_cluster02
sudo vi /etc/ipquorum/instances/svc_cluster01.conf
# Start with mkquorumapp enabled
 ipqm start svc_cluster01
```

---

## 🤝 Contributing

Contributions are welcome! Please submit issues or pull requests to improve this repository.

---

## 👤 Maintainer

**Ole Kristian Myklebust**

---

## 📄 License

See LICENSE file for details.

---

**Made with ❤️ for IBM Storage Virtualize**

*Automate your IP Quorum deployment today!*
