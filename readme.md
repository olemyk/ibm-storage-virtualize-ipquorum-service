# Repo for the IBM Storage Virtualize — IP Quorum application Service

This repository provides resources and guides for running the IP Quorum application for IBM Storage Virtualize. It includes multiple deployment options and download tools for managing the service effectively.

## 📘 About the IP Quorum Application for IBM Storage Virtualize

## 🏗 Architecture
Here's a simplified view of how IP Quorum interacts with IBM Storage Virtualize clusters:

<img src="ipquorum-systemd/images/ipquorum-smal.png" alt="drawing" style="width:600px;"/>

-----

A quorum device is used to break a tie when a SAN fault occurs, when exactly half of the nodes that were previously a member of the cluster are present.

The IP quorum application is a Java application that runs on a separate server or host. (This can be physical or Virtual Machine.)
An IP quorum application is used in IP networks to resolve failure scenarios where half the control canisters/nodes on the cluster become unavailable.

The application determines which nodes or enclosures can continue processing host operations and avoids a split cluster, where both halves of the system continue to process I/O independently.

IBM Storage Virtualize powers the IBM Storage FlashSystem and IBM SVC

---

## 📋 Prerequisites

Before deploying the IP Quorum service, ensure you have:

### Required
- **Linux host or VM** (physical or virtual machine, or container)
- **Java Runtime Environment** (OpenJDK 8, 11, 14, or later)
- **Network connectivity** to IBM Storage Virtualize cluster
  - Port 1260/TCP (for IP Quorum communication)
  - Port 7443/HTTPS (for REST API - if using automatic download)
- **Valid credentials** for IBM Storage Virtualize:
  - **Monitor** role (for downloading IP Quorum JAR)
  - **Restricted Administrator** role (for creating new Quorum Apps)

### Optional (for automatic download)
- **One of the download tools**: Go (recommended), Python, or Bash
- **IBM Storage Virtualize Code 8.6.1 or later** (for REST API support)

---

## 🚀 Deployment Options for IP Quorum Service

Choose the deployment method that best fits your environment:

### Comparison of Deployment Options

| Feature | Automated Systemd | Manual Systemd | Container |
|---------|------------------|----------------|-----------|
| **Ease of Setup** | ✅ Easiest (automated script) | ⚠️ Manual steps | ⚠️ Manual steps |
| **Auto-Download JAR** | ✅ Yes (optional) | ❌ No | ❌ No |
| **Auto-Start on Boot** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Isolation** | ⚠️ System service | ⚠️ System service | ✅ Container isolated |
| **Resource Management** | ✅ Systemd limits | ✅ Systemd limits | ✅ Container limits |
| **Best For** | Production, automation | Custom setups | Container environments |

### 1. 🎯 Automated Systemd Service (Recommended)

**Features:**
- ✅ Automated installation script
- ✅ Automatic JAR download on service start/restart
- ✅ Centralized configuration file
- ✅ Support for Go, Python, or Bash download tools
- ✅ Automatic backup before updates
- ✅ Enhanced security with systemd hardening

**Quick Start:**
```bash
# Run the installation script
sudo ./install-ipquorum-service.sh

# Configure the service
sudo vi /etc/ipquorum/ipquorum.conf

# Start the service
sudo systemctl enable --now ipquorum
```

📖 **[Complete Guide: Automated Systemd Service with Auto-Download](ipquorum-systemd/README-IMPROVED-SERVICE.md)**

---

### 2. 🔧 Manual Systemd Service

**Features:**
- ⚙️ Full control over configuration
- 📝 Manual JAR placement
- 🔒 Custom security settings
- 🎛️ Flexible for specific requirements



📖 **[Complete Guide: Manual Systemd Configuration](ipquorum-systemd/manual-config/readme-ipquorum-systemd.md)**

---

### 3. 🐳 Container Deployment (Podman/Docker)

**Features:**
- 🚢 Runs in isolated container
- 📦 Portable across environments
- 🔄 Easy to replicate
- 🛡️ Container-level security


📖 **[Complete Guide: Container Deployment](ipquorum-container/ibm-virtualize-ipquorum-container.md)**

---

## 📥 Downloading the IP Quorum JAR File through RestAPI calls

Starting with **IBM Storage Virtualize Code 8.6.1**, you can download the IP Quorum application using the REST API. Choose the method that best fits your needs:

### Comparison of Download Tools

| Feature | **Go (Recommended)** | Python | Bash | Manual |
|---------|---------------------|--------|------|--------|
| **Single Binary** | ✅ Yes | ❌ No | ❌ No | ❌ No |
| **Dependencies** | ✅ None | ⚠️ Python 3.8+ | ⚠️ curl, jq | ⚠️ curl, jq |
| **Cross-Platform** | ✅ Linux/macOS/Windows | ✅ Linux/macOS/Windows | ⚠️ Linux/macOS | ⚠️ Linux/macOS |
| **Performance** | ✅ Excellent (<10ms startup) | ✅ Good (~100ms startup) | ✅ Good | ⚠️ Manual |
| **Memory Usage** | ✅ <20MB | ⚠️ ~50MB | ✅ Low | N/A |
| **Type Safety** | ✅ Compile-time | ✅ Runtime | ❌ No | N/A |
| **Ease of Use** | ✅ Very Easy | ✅ Easy | ✅ Easy | ⚠️ Complex |
| **Best For** | Production, CI/CD, automation | Development, scripting | Quick scripts | Learning API |

---

### 1. 🏆 Go Version (Recommended)

**Why Go?**
- ✅ **Single binary** - No dependencies, no runtime required
- ✅ **Fast** - <10ms startup time, <20MB memory
- ✅ **Cross-platform** - Native binaries for Linux, macOS, Windows
- ✅ **Production-ready** - Type-safe, concurrent, reliable
- ✅ **Perfect for CI/CD** - Single file deployment

**Quick Start:**
```bash
# Download and install
curl -LO https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest/download/ipquorum-download-go-linux-amd64
chmod +x ipquorum-download-go-linux-amd64
# Rename binary to 'ipquorum' for easier use
sudo mv ipquorum-download-go-linux-amd64 /usr/local/bin/ipquorum

# Interactive use (most secure)
# Note: --insecure is the default behavior (skips TLS verification)
# Use --secure flag to enable strict TLS verification instead
ipquorum \
  --api-endpoint 10.33.7.80 \
  --mkquorumapp --partnersystem svc_cluster02 \
  --user superuser \
  --pass-prompt \
  --download

# Automated use with password file
echo 'password' > ~/.ipquorum_pass
chmod 400 ~/.ipquorum_pass
ipquorum \
  --api-endpoint 10.33.7.80 \
  --mkquorumapp --partnersystem svc_cluster02 \
  --user superuser \
  --pass-file ~/.ipquorum_pass \
  --download
```

📖 **[Complete Guide: Go Download Tool](ipquorum-download-go/README.md)**

---

### 2. 🐍 Python Version

**Features:**
- ✅ Cross-platform (Windows, Linux, macOS)
- ✅ Type-safe with full type hints
- ✅ Professional logging and error handling
- ✅ Smart retry logic with exponential backoff
- ✅ Multiple password input methods

**Quick Start:**
```bash
# Install dependencies
pip install -r requirements.txt

# Interactive use (most secure)
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --mkquorumapp --partnersystem svc_cluster02 \
  --user superuser \
  --pass-prompt \
  --download --insecure

# Automated use with password file
echo 'password' > ~/.ipquorum_password
chmod 400 ~/.ipquorum_password
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --mkquorumapp --partnersystem svc_cluster02 \
  --user superuser \
  --pass-file ~/.ipquorum_password \
  --download --insecure
```

📖 **[Complete Guide: Python Download Tool](ipquorum-download/python-version/README.md)**

---

### 3. 🔧 Bash Version

**Features:**
- ✅ Simple and lightweight
- ✅ No Python required
- ✅ Works on any Linux/Unix system
- ✅ Easy to understand and modify

**Quick Start:**
```bash
# Make executable
chmod +x ipquorum-restapi-download.sh

# Download JAR only
./ipquorum-restapi-download.sh \
  --api-endpoint 10.33.7.80 \
  --no-mkquorumapp --download \
  --user superuser --pass password \
  --insecure

# Create Quorum App + Download
./ipquorum-restapi-download.sh \
  --api-endpoint 10.33.7.80 \
  --mkquorumapp --partnersystem svc_cluster02 \
  --download --insecure \
  --user superuser --pass password
```

📖 **[Complete Guide: Bash Download Script](ipquorum-download/bash-version/README.md)**

---

### 4. 📖 Manual Download (curl)

**For learning the API or custom automation:**

```bash
# 1. Authenticate
export VIP="10.33.7.80"
export VUSERNAME="superuser"
export VPASSWORD="password"

AUTH_RESPONSE=$(curl -ks -X POST "https://${VIP}:7443/rest/v1/auth" \
  -H "accept: application/json" \
  -H "X-Auth-Username: ${VUSERNAME}" \
  -H "X-Auth-Password: ${VPASSWORD}" \
  -d "")

TOKEN=$(echo "${AUTH_RESPONSE}" | jq -r '.token // .X_Auth_Token // .authToken // empty')

# 2. Create new IP Quorum app (optional)
curl -ks -X POST "https://${VIP}:7443/rest/v1/mkquorumapp" \
  -H "accept: application/json" \
  -H "X-Auth-Token: ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"ip_6": false, "nometadata": false, "partnersystem": "svc_cluster02", "partnerip6": false}'

# 3. Download JAR file
curl -ks -X POST "https://${VIP}:7443/rest/v1/download" \
  -H "accept: application/json" \
  -H "X-Auth-Token: ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"prefix":"/dumps","filename":"ip_quorum.jar"}' \
  --output ip_quorum.jar
```

📖 **[Complete Guide: Manual Download with curl](ipquorum-download/ipquorum-download-readme.md)**

---

## 💡 Tips and Tricks

### Security Best Practices
- ✅ Use `--pass-prompt` for interactive password entry (most secure)
- ✅ Use password files with `chmod 400` for automation
- ✅ Never use `--pass` with password in command line for production
- ✅ Use `--secure` flag in production if you have valid TLS certificates
- ✅ Run services with dedicated user accounts (not root)

### Monitoring
```bash
# Check service status
sudo systemctl status ipquorum

# View real-time logs
sudo journalctl -u ipquorum -f

# Check from Storage Virtualize
ssh superuser@YOUR_SV_IP
lsquorum
```

### Troubleshooting
```bash
# Test connectivity
curl -k https://YOUR_API_ENDPOINT:7443/rest/v1/auth

# Check firewall
sudo firewall-cmd --list-ports
sudo firewall-cmd --add-port=1260/tcp --permanent

# Verify Java
java -version

# Check JAR file
ls -l /opt/IBM/ip-quorum/ip_quorum.jar
```

---

## 📚 Additional Resources

### IBM Documentation
- [IPQuorum Info](https://www.ibm.com/support/pages/ibm-storage-virtualize-ip-quorum-application-requirements-1)
- [IP quorum application](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=quorum-ip-application)
- [Storage Virtualize RESTful API](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=interface-storage-virtualize-restful-api)

### Repository Documentation
- [Architecture Overview](ipquorum-systemd/ARCHITECTURE.md)
- [Changelog](ipquorum-systemd/CHANGELOG.md)
- [Security Guide](ipquorum-download/python-version/SECURITY-GUIDE.md)

---

## 🤝 Contributing

Contributions are welcome! Please submit issues or pull requests to improve this repository.

---

## 👤 Maintainer

Ole Kristian Myklebust

---

**Made with ❤️ for IBM Storage Virtualize**
