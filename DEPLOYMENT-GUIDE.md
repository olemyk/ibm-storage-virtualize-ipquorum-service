# IP Quorum Service - Deployment & Distribution Guide

This guide outlines best practices for packaging, distributing, and deploying the **automated multi-instance IP Quorum service** to customers.

---

## 🎯 What is This?

The IP Quorum Service is a **systemd-based service management system** that:
- ✅ Installs and manages **Java runtime** for running IP Quorum
- ✅ Creates **systemd services** to run the `ip_quorum.jar` file
- ✅ Supports **multiple independent instances** on a single host
- ✅ Provides **instance manager** for easy lifecycle management
- ✅ Includes **optional automatic JAR download** via Go/Python/Bash tools

### Components

**1. IP Quorum Service (This Project)**
- Systemd template service (`ipquorum@.service`)
- Instance manager script
- Configuration management
- Java runtime management
- Runs the `ip_quorum.jar` file from IBM Storage Virtualize

**2. Download Tools (Optional Helpers)**
- **Go binary** (`ipquorum-download-go`) - Downloads JAR via REST API
- **Python script** - Alternative download tool
- **Bash script** - Lightweight download option

The download tools are **optional helpers** - you can also manually download the JAR file from your IBM Storage Virtualize system.

---

## 📦 Recommended Distribution: GitHub Releases

**Best for**: Version control, easy updates, wide accessibility

### Release Package Structure

```
ipquorum-service-2.0.5.tar.gz
├── install-ipquorum-service.sh          # Main installer (installs systemd service)
├── systemd/                             # Systemd service files
│   ├── ipquorum@.service                # Template service (runs ip_quorum.jar)
│   ├── ipquorum-instance-manager.sh     # Instance manager
│   ├── ipquorum-start-multi.sh          # Start script (launches Java)
│   ├── ipquorum-download-multi.sh       # Download wrapper script
│   ├── ipquorum-validate-multi.sh       # Validation script
│   ├── instance.conf.template           # Config template
│   ├── examples/                        # Example configs
│   └── troubleshoot-tools/              # Diagnostic tools
├── ipquorum-downloader/                 # Optional download tools
│   ├── ipquorum-download-go-linux-amd64 # Go binary (downloads JAR)
│   ├── ipquorum-download-go-linux-arm64
│   ├── ipquorum-download-go-darwin-amd64
│   └── ipquorum-download-go-darwin-arm64
├── README.md                            # Main documentation
├── DEPLOYMENT-GUIDE.md                  # This file
└── checksums.txt                        # SHA256 checksums
```

### Creating a Release

Use the automated release script:

```bash
# Create release package
./scripts/create-release.sh 2.0.5

# This creates:
# - ipquorum-service-2.0.5.tar.gz (complete package)
# - Individual Go binaries for each platform
# - checksums.txt with SHA256 hashes
```

### Customer Installation

**One-line install:**
```bash
curl -fsSL https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest/download/ipquorum-service-latest.tar.gz | tar -xz
cd ipquorum-service-*
sudo ./install-ipquorum-service.sh
```

**Manual download:**
```bash
# Download release
wget https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/v2.0.5/ipquorum-service-2.0.5.tar.gz

# Verify checksum
sha256sum -c checksums.txt

# Extract and install
tar -xzf ipquorum-service-2.0.5.tar.gz
cd ipquorum-service-2.0.5
sudo ./install-ipquorum-service.sh
```

---

## 🚀 What the Installer Does

The `install-ipquorum-service.sh` script:

### 1. Installs Systemd Service Files
- Copies `ipquorum@.service` template to `/etc/systemd/system/`
- Installs management scripts to `/usr/local/bin/`
- Creates directory structure for instances

### 2. Optionally Installs Download Tool
Offers choice of:
- **Go binary** (recommended - single file, no dependencies)
- **Python script** (requires Python 3.8+)
- **Bash script** (requires curl, jq)
- **Skip** (manually download JAR files)

### 3. Creates Directory Structure

```
/etc/ipquorum/
├── instances/                    # Instance configurations
│   └── <instance>.conf          # Config for each instance
└── .instances/                   # Instance metadata

/var/lib/ipquorum/
└── <instance>/
    ├── ip_quorum.jar            # JAR file (from IBM Storage Virtualize)
    └── .password                # Encrypted password

/var/log/ipquorum/
└── <instance>/                  # Instance logs

/usr/local/bin/
├── ipquorum-instance-manager.sh # Instance manager
├── ipquorum-download-multi.sh   # Download wrapper
├── ipquorum-start-multi.sh      # Start script (launches Java + JAR)
├── ipquorum-validate-multi.sh   # Validation
└── ipquorum-download            # Optional: Go binary or python/bash script

/etc/systemd/system/
└── ipquorum@.service            # Template service (runs Java)
```

---

## 🎯 Post-Installation Workflow

### For Customers

**1. Create First Instance:**
```bash
sudo /usr/local/bin/ipquorum-instance-manager.sh create svc_cluster01
```

This creates:
- Configuration file: `/etc/ipquorum/instances/svc_cluster01.conf`
- Data directory: `/var/lib/ipquorum/svc_cluster01/`
- Log directory: `/var/log/ipquorum/svc_cluster01/`

**2. Configure Instance:**
```bash
sudo vi /etc/ipquorum/instances/svc_cluster01.conf
```

Set:
- `API_ENDPOINT` - Your IBM Storage Virtualize IP/hostname
- `VIRTUALIZE_USERNAME` - Username for API access
- `DOWNLOADIPQ` - yes (to auto-download JAR) or no (manual)

**3. Set Password:**
```bash
echo 'your_password' | sudo tee /var/lib/ipquorum/svc_cluster01/.password > /dev/null
sudo chmod 400 /var/lib/ipquorum/svc_cluster01/.password
```

**4. Start Instance:**
```bash
# The service will:
# - Download ip_quorum.jar (if DOWNLOADIPQ=yes)
# - Launch Java with the JAR file
# - Connect to IBM Storage Virtualize

sudo /usr/local/bin/ipquorum-instance-manager.sh enable svc_cluster01
sudo /usr/local/bin/ipquorum-instance-manager.sh start svc_cluster01
```

**5. Verify:**
```bash
# Check systemd service status
sudo /usr/local/bin/ipquorum-instance-manager.sh status svc_cluster01

# View logs
sudo /usr/local/bin/ipquorum-instance-manager.sh logs svc_cluster01

# Check from IBM Storage Virtualize
ssh superuser@YOUR_SV_IP
lsquorum
```

### Optional: Create Alias

```bash
echo "alias ipqm='sudo /usr/local/bin/ipquorum-instance-manager.sh'" >> ~/.bashrc
source ~/.bashrc

# Now use: ipqm create, ipqm start, ipqm status, etc.
```

---

## 🔧 How It Works

### Service Startup Flow

```
1. systemctl start ipquorum@svc_cluster01
   ↓
2. systemd loads /etc/systemd/system/ipquorum@.service
   ↓
3. Service executes /usr/local/bin/ipquorum-start-multi.sh svc_cluster01
   ↓
4. Start script:
   - Reads /etc/ipquorum/instances/svc_cluster01.conf
   - If DOWNLOADIPQ=yes: Runs download tool to get ip_quorum.jar
   - Launches: java -jar /var/lib/ipquorum/svc_cluster01/ip_quorum.jar
   ↓
5. Java process runs ip_quorum.jar
   ↓
6. JAR connects to IBM Storage Virtualize on port 1260
```

### Download Tool (Optional)

If `DOWNLOADIPQ=yes`, the service uses one of these tools to download the JAR:

**Go Binary (Recommended):**
```bash
ipquorum-download \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-file /var/lib/ipquorum/svc_cluster01/.password \
  --download --insecure \
  --output /var/lib/ipquorum/svc_cluster01/ip_quorum.jar
```

This tool:
- Connects to IBM Storage Virtualize REST API (port 7443)
- Authenticates with provided credentials
- Downloads `/dumps/ip_quorum.jar` from the system
- Saves to instance directory

---

## 📊 Distribution Comparison

| Method | Ease | Updates | Air-Gap | Enterprise | Automation | Recommended |
|--------|------|---------|---------|------------|------------|-------------|
| **GitHub Release** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ✅ **Yes** |
| RPM/DEB Package | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⚠️ Future |
| Container Image | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⚠️ Future |
| Ansible Playbook | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⚠️ Future |

---

## 🔄 Update Process

### For Customers

**Check for updates:**
```bash
# Current version
cat /opt/IBM/ip-quorum/VERSION 2>/dev/null || echo "unknown"

# Latest version
curl -s https://api.github.com/repos/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest | grep '"tag_name"'
```

**Update installation:**
```bash
# Download new version
wget https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest/download/ipquorum-service-latest.tar.gz

# Backup configurations
sudo cp -r /etc/ipquorum /etc/ipquorum.backup

# Extract and run installer
tar -xzf ipquorum-service-latest.tar.gz
cd ipquorum-service-*
sudo ./install-ipquorum-service.sh

# Restart instances (will re-download JAR if configured)
sudo systemctl restart ipquorum@*.service
```

---

## 🎯 Multi-Instance Use Cases

### Use Case 1: Multiple Storage Systems

**Scenario:** Customer has 3 IBM Storage Virtualize systems

```bash
# Create instances (one per storage system)
sudo ipqm create datacenter-a
sudo ipqm create datacenter-b
sudo ipqm create dr-site

# Configure each (different API endpoints)
sudo vi /etc/ipquorum/instances/datacenter-a.conf  # API_ENDPOINT=10.33.7.80
sudo vi /etc/ipquorum/instances/datacenter-b.conf  # API_ENDPOINT=10.33.7.81
sudo vi /etc/ipquorum/instances/dr-site.conf       # API_ENDPOINT=10.33.7.82

# Start all (each runs separate Java process with own JAR)
sudo systemctl start ipquorum@{datacenter-a,datacenter-b,dr-site}

# Enable auto-start
sudo systemctl enable ipquorum@{datacenter-a,datacenter-b,dr-site}
```

Each instance:
- Runs its own Java process
- Has its own `ip_quorum.jar` file
- Connects to different IBM Storage Virtualize system
- Has independent logs and configuration

### Use Case 2: PBHA Configuration

**Scenario:** PowerHA setup with partner systems

```bash
# Create instance for cluster01
sudo ipqm create svc_cluster01

# Edit config to set partner system
sudo vi /etc/ipquorum/instances/svc_cluster01.conf
# Set: partnersystem=svc_cluster02
# Set: mkquorumapp=yes

# Start (will create Quorum App automatically via REST API, then download JAR)
sudo ipqm start svc_cluster01
```

### Use Case 3: Production + DR

**Scenario:** Production and disaster recovery sites

```bash
# Production instance (always running)
sudo ipqm create production
sudo ipqm enable production
sudo ipqm start production

# DR instance (start only when needed)
sudo ipqm create dr-site
# Start only during DR scenarios
sudo ipqm start dr-site
```

---

## 🔒 Security Considerations

### Distribution Security

- ✅ **Sign releases** with GPG keys
- ✅ **Provide SHA256 checksums** for all files
- ✅ **Use HTTPS** for all downloads
- ✅ **Scan binaries** for vulnerabilities
- ✅ **Document security** best practices

### Customer Security

**Password Management:**
```bash
# Secure password file
echo 'password' | sudo tee /var/lib/ipquorum/<instance>/.password > /dev/null
sudo chmod 400 /var/lib/ipquorum/<instance>/.password
sudo chown root:root /var/lib/ipquorum/<instance>/.password
```

**TLS Configuration:**
```bash
# Production: use secure TLS
insecure_api=""  # or --secure

# Development: allow self-signed certs
insecure_api="--insecure"
```

**Service Hardening:**
- Runs as dedicated `ipquorum` user (not root)
- Systemd security features enabled
- SELinux contexts properly configured
- File permissions restricted

---

## 📋 Pre-Deployment Checklist

### For Developers

- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version numbers bumped
- [ ] Go binaries built for all platforms (download tool)
- [ ] Checksums generated
- [ ] Release notes written
- [ ] GitHub release created

### For Customers

- [ ] **Java 8+ installed** (`java -version`) - **REQUIRED to run ip_quorum.jar**
- [ ] Network access to Storage Virtualize (port 7443 for download, port 1260 for quorum)
- [ ] Port 1260/TCP available for IP Quorum communication
- [ ] Root/sudo access
- [ ] Firewall rules configured
- [ ] Storage Virtualize credentials ready
- [ ] Backup plan for existing installations

---

## 🚀 Quick Deployment Script

For automated deployments:

```bash
#!/bin/bash
# deploy-ipquorum.sh - Automated deployment script

set -e

VERSION="2.0.5"
INSTANCE_NAME="${1:-production}"
API_ENDPOINT="${2}"
USERNAME="${3}"
PASSWORD="${4}"

if [ -z "$API_ENDPOINT" ] || [ -z "$USERNAME" ] || [ -z "$PASSWORD" ]; then
    echo "Usage: $0 <instance_name> <api_endpoint> <username> <password>"
    exit 1
fi

# Download and install systemd service
echo "Downloading IP Quorum Service v${VERSION}..."
curl -fsSL "https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/v${VERSION}/ipquorum-service-${VERSION}.tar.gz" | tar -xz
cd "ipquorum-service-${VERSION}"

# Run installer (choose option 1 for GitHub download of Go binary)
echo "1" | sudo ./install-ipquorum-service.sh

# Create instance
echo "Creating instance: ${INSTANCE_NAME}..."
sudo /usr/local/bin/ipquorum-instance-manager.sh create "${INSTANCE_NAME}"

# Configure instance
echo "Configuring instance..."
sudo sed -i "s/<HOSTNAME_OR_IP>/${API_ENDPOINT}/g" "/etc/ipquorum/instances/${INSTANCE_NAME}.conf"
sudo sed -i "s/VIRTUALIZE_USERNAME=.*/VIRTUALIZE_USERNAME=${USERNAME}/" "/etc/ipquorum/instances/${INSTANCE_NAME}.conf"

# Set password
echo "${PASSWORD}" | sudo tee "/var/lib/ipquorum/${INSTANCE_NAME}/.password" > /dev/null
sudo chmod 400 "/var/lib/ipquorum/${INSTANCE_NAME}/.password"

# Enable and start (will download JAR and launch Java)
echo "Starting instance..."
sudo /usr/local/bin/ipquorum-instance-manager.sh enable "${INSTANCE_NAME}"
sudo /usr/local/bin/ipquorum-instance-manager.sh start "${INSTANCE_NAME}"

# Verify
echo "Verifying installation..."
sleep 5
sudo /usr/local/bin/ipquorum-instance-manager.sh status "${INSTANCE_NAME}"

echo "Deployment complete!"
echo "Instance: ${INSTANCE_NAME}"
echo "Java process running: ip_quorum.jar"
echo "Status: sudo /usr/local/bin/ipquorum-instance-manager.sh status ${INSTANCE_NAME}"
echo "Logs: sudo /usr/local/bin/ipquorum-instance-manager.sh logs ${INSTANCE_NAME}"
```

**Usage:**
```bash
chmod +x deploy-ipquorum.sh
./deploy-ipquorum.sh production 10.33.7.80 superuser 'password'
```

---

## 📞 Support & Documentation

### Customer Resources

- **Main README**: [readme.md](readme.md)
- **Multi-Instance Guide**: [ipquorum-systemd/multi-instance/README.md](ipquorum-systemd/multi-instance/README.md)
- **Architecture**: [ipquorum-systemd/multi-instance/ARCHITECTURE.md](ipquorum-systemd/multi-instance/ARCHITECTURE.md)
- **Security Guide**: [ipquorum-download/python-version/SECURITY-GUIDE.md](ipquorum-download/python-version/SECURITY-GUIDE.md)
- **Changelog**: [CHANGELOG.md](CHANGELOG.md)

### IBM Documentation

- [IP Quorum Application Requirements](https://www.ibm.com/support/pages/ibm-storage-virtualize-ip-quorum-application-requirements-1)
- [IP Quorum Application Guide](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=quorum-ip-application)
- [Storage Virtualize REST API](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=interface-storage-virtualize-restful-api)

### Getting Help

- **GitHub Issues**: Report bugs or request features
- **Documentation**: Check guides and troubleshooting sections
- **Community**: Share experiences and solutions

---

## 🎯 Success Metrics

Track these metrics for successful deployments:

- ✅ **Installation time**: < 5 minutes
- ✅ **Instance creation**: < 2 minutes
- ✅ **First connection**: < 1 minute
- ✅ **Zero manual configuration** for basic setup
- ✅ **Automatic JAR updates** via download tool
- ✅ **Multi-instance support** without conflicts

---

## 📝 Customer Communication Template

```
Subject: IP Quorum Service v2.0.5 - Automated Multi-Instance Support

Dear Customer,

We're excited to announce IP Quorum Service v2.0.5 with fully automated installation and multi-instance support!

🚀 One-Command Installation:
   curl -fsSL https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest/download/ipquorum-service-latest.tar.gz | tar -xz
   cd ipquorum-service-*
   sudo ./install-ipquorum-service.sh

✨ What It Does:
   • Installs systemd service to run ip_quorum.jar
   • Manages Java runtime
   • Supports multiple independent instances
   • Optional automatic JAR download via Go/Python/Bash tools
   • Production-ready security hardening

📖 Quick Start:
   sudo /usr/local/bin/ipquorum-instance-manager.sh create svc_cluster01
   sudo /usr/local/bin/ipquorum-instance-manager.sh start svc_cluster01

📚 Documentation:
   https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service

Need help? Check our comprehensive guides or open an issue on GitHub.

Best regards,
Your Team
```

---

## 🎯 Recommended Deployment Strategy

### Phase 1: Initial Deployment (Current)
✅ **GitHub Releases** with automated installer
✅ **Multi-instance systemd service** out of the box
✅ **Optional Go binary** for JAR downloads (no dependencies)
✅ **Instance manager** for lifecycle management
✅ **Comprehensive documentation**

### Phase 2: Enterprise Support (Future)
⚠️ **RPM packages** for RHEL/CentOS/Rocky/AlmaLinux
⚠️ **DEB packages** for Ubuntu/Debian
⚠️ **Ansible playbook** for large-scale automation
⚠️ **Terraform modules** for infrastructure as code

### Phase 3: Advanced Options (Future)
⚠️ **Container images** for Kubernetes
⚠️ **Helm charts** for cloud deployments
⚠️ **Operator** for Kubernetes automation

---

**Current Recommendation**: Use **GitHub Releases with automated installer** for maximum flexibility, ease of use, and rapid deployment. The multi-instance architecture supports all use cases from single systems to complex multi-site deployments.

---

**Made with ❤️ for IBM Storage Virtualize**

*Automate your IP Quorum deployment today!*