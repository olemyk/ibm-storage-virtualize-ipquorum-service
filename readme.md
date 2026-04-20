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
║  │  Automated Installation • Multi-Instance Support • Go-Powered Download │  ║
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
- 📊 **Complete Observability** - Centralized logging, status monitoring, health checks

---

## 🎯 What is IP Quorum?

### Split-Brain Prevention Concept

```
    ┌──────────────────────────────────────────────────────┐
    │                                                       │
    │   Site A                    Site B                   │
    │   ┌────┐                    ┌────┐                   │
    │   │ ██ │ ←─────────────────→│ ██ │                   │
    │   └────┘    Link Failure    └────┘                   │
    │      ↓                         ↓                      │
    │      │                         │                      │
    │      └────────→ ┌────┐ ←───────┘                     │
    │                 │ Q  │  IP Quorum                    │
    │                 └────┘  Decides!                     │
    │                                                       │
    │        Prevents Split-Brain Scenarios                │
    │                                                       │
    └──────────────────────────────────────────────────────┘
```

**IP Quorum** is a tie-breaker service that prevents split-brain scenarios in IBM Storage Virtualize clusters. When exactly half of the nodes become unavailable due to a SAN fault, the IP Quorum application determines which nodes can continue processing host operations, ensuring data integrity and preventing independent I/O processing by both halves of the system.

**Powers:** IBM Storage FlashSystem and IBM SAN Volume Controller (SVC)

---

## 🚀 Quick Start - Automated Installation

### One-Command Installation

```bash
# Download and run the automated installer
curl -fsSL https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest/download/ipquorum-service-latest.tar.gz | tar -xz
cd ipquorum-service-*
sudo ./install-ipquorum-service.sh
```

The installer will:
1. ✅ Install systemd service files
2. ✅ Download Go binary (or use Python/Bash)
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
sudo ipqm create svc_cluster01

# Start the instance
sudo ipqm enable svc_cluster01
sudo ipqm start svc_cluster01

# Check status
sudo ipqm status svc_cluster01
sudo ipqm logs svc_cluster01
```

**That's it!** Your IP Quorum service is running and will automatically:
- Download the latest JAR file from your Storage Virtualize system
- Start on boot
- Restart on failure
- Log to centralized location

---

## 📋 Prerequisites

### Required
- **Linux host** (RHEL/CentOS/Rocky/AlmaLinux 8+, Ubuntu 20.04+)
- **Java Runtime** (OpenJDK 8, 11, 14, or later)
- **Network connectivity** to IBM Storage Virtualize:
  - Port **1260/TCP** (IP Quorum communication)
  - Port **7443/HTTPS** (REST API for automatic download)
- **Valid credentials**:
  - **Monitor** role (for downloading JAR)
  - **Restricted Administrator** role (for creating Quorum Apps)

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

<img src="ipquorum-systemd/images/ipquorum-smal.png" alt="IP Quorum Architecture" style="width:600px;"/>

### Instance Management

```bash
# List all instances
sudo ipqm list

# Create multiple instances
sudo ipqm create datacenter-a
sudo ipqm create datacenter-b
sudo ipqm create production

# Manage instances
sudo ipqm start datacenter-a
sudo ipqm stop datacenter-b
sudo ipqm restart production
sudo ipqm status datacenter-a
sudo ipqm logs datacenter-a -f

# Enable/disable auto-start
sudo ipqm enable datacenter-a
sudo ipqm disable datacenter-b

# Delete instance
sudo ipqm delete old-system
```

Each instance has:
- ✅ **Isolated configuration** (`/etc/ipquorum/instances/<name>.conf`)
- ✅ **Separate credentials** (`/var/lib/ipquorum/<name>/.password`)
- ✅ **Independent JAR file** (`/var/lib/ipquorum/<name>/ip_quorum.jar`)
- ✅ **Dedicated logs** (`/var/log/ipquorum/<name>/`)
- ✅ **Systemd service** (`ipquorum@<name>.service`)

📖 **[Complete Multi-Instance Guide](ipquorum-systemd/multi-instance/README-MULTI-INSTANCE.md)**

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

### Manual Download Example

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

## 🔧 Configuration

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
sudo ipqm status svc_cluster01

# Systemd status
sudo systemctl status ipquorum@svc_cluster01

# View logs
sudo ipqm logs svc_cluster01 -f
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
- **[Multi-Instance Service](ipquorum-systemd/multi-instance/README-MULTI-INSTANCE.md)** - Complete guide (recommended)
- **[Deployment Guide](DEPLOYMENT-GUIDE.md)** - Distribution and packaging options
- **[Architecture Overview](ipquorum-systemd/ARCHITECTURE.md)** - System design and components

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

### Single Storage System
```bash
# Simple setup for one storage system
sudo ipqm create production
# Configure and start
sudo ipqm enable production
sudo ipqm start production
```

### Multiple Storage Systems
```bash
# Create instances for each system
sudo ipqm create datacenter-a
sudo ipqm create datacenter-b
sudo ipqm create dr-site

# Start all instances
sudo systemctl start ipquorum@{datacenter-a,datacenter-b,dr-site}
```

### PBHA (PowerHA) Configuration
```bash
# Create Quorum App for partner system
sudo ipqm create svc_cluster01
# Edit config to set partnersystem=svc_cluster02
sudo vi /etc/ipquorum/instances/svc_cluster01.conf
# Start with mkquorumapp enabled
sudo ipqm start svc_cluster01
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
