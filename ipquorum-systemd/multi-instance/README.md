# IBM Storage Virtualize IP Quorum - Multi-Instance Service

Run multiple independent IP Quorum instances on a single host, each connecting to a different IBM Storage Virtualize System.

## 🚀 Quick Start

### Method 1: Instance Manager (Recommended - Interactive)

```bash
# 1. Install the multi-instance system
sudo ./install-multi-instance.sh

# 2. Create instance with interactive configuration
sudo ipquorum-instance-manager.sh create svc_cluster01

# The manager will prompt you for:
#   - Description
#   - API endpoint (IP or hostname)
#   - Username and password
#   - Download settings
#   - mkquorumapp configuration (if needed)

# 3. Enable and start the instance
sudo ipquorum-instance-manager.sh enable svc_cluster01
sudo ipquorum-instance-manager.sh start svc_cluster01

# 4. Check status
sudo ipquorum-instance-manager.sh status svc_cluster01
sudo ipquorum-instance-manager.sh logs svc_cluster01
```

### Method 2: Manual Configuration

```bash
# 1. Make scripts executable (if not already)
chmod +x install-multi-instance.sh
chmod +x ipquorum-instance-manager.sh

# 2. Install the multi-instance system
sudo ./install-multi-instance.sh

# 2. Create instance configuration from example
sudo cp /etc/ipquorum/examples/system1.conf /etc/ipquorum/instances/myinstance.conf
sudo vi /etc/ipquorum/instances/myinstance.conf

# 3. Set password (in /var/lib for SELinux compatibility)
sudo mkdir -p /var/lib/ipquorum/.passwords
echo 'your_password' | sudo tee /var/lib/ipquorum/.passwords/myinstance.password
sudo chmod 440 /var/lib/ipquorum/.passwords/myinstance.password
sudo chown ipquorum:ipquorum /var/lib/ipquorum/.passwords/myinstance.password

# 4. Create directories
sudo mkdir -p /var/lib/ipquorum/myinstance /var/log/ipquorum/myinstance
sudo chown -R ipquorum:ipquorum /var/lib/ipquorum/myinstance /var/log/ipquorum/myinstance

# 5. Start instance
sudo systemctl enable --now ipquorum@myinstance.service
```

## 🎯 Instance Manager Commands

```bash
# Create instance (interactive)
sudo ipquorum-instance-manager.sh create <instance-name>

# Create instance (non-interactive - requires manual config)
sudo ipquorum-instance-manager.sh create <instance-name> --non-interactive

# List all instances
sudo ipquorum-instance-manager.sh list

# Manage instances
sudo ipquorum-instance-manager.sh start <instance-name>
sudo ipquorum-instance-manager.sh stop <instance-name>
sudo ipquorum-instance-manager.sh restart <instance-name>
sudo ipquorum-instance-manager.sh status <instance-name>

# Enable/disable auto-start
sudo ipquorum-instance-manager.sh enable <instance-name>
sudo ipquorum-instance-manager.sh disable <instance-name>

# View logs
sudo ipquorum-instance-manager.sh logs <instance-name> [lines]

# Validate configuration
sudo ipquorum-instance-manager.sh validate <instance-name>

# Show detailed info
sudo ipquorum-instance-manager.sh info <instance-name>

# Delete instance
sudo ipquorum-instance-manager.sh delete <instance-name>
```

## 📁 Files in This Directory

| File | Description |
|------|-------------|
| **ipquorum@.service** | Systemd template service unit |
| **ipquorum-download-multi.sh** | Multi-instance download script |
| **ipquorum-start-multi.sh** | Multi-instance start script |
| **install-multi-instance.sh** | Automated installation script |
| **ipquorum-instance-manager.sh** | Instance management CLI tool |
| **instance.conf.template** | Configuration template |
| **README-ADVANCED.md** | Complete user guide (781 lines) |
| **QUICK-REFERENCE.md** | Quick command reference |
| **IMPLEMENTATION-SUMMARY.md** | Implementation details |
| **examples/** | Example configurations |

## 📚 Documentation

### Start Here

1. **[README-ADVANCED.md](README-ADVANCED.md)** - Complete user guide (advanced)
   - Installation instructions
   - Configuration reference
   - Management commands
   - Examples and troubleshooting

2. **[QUICK-REFERENCE.md](QUICK-REFERENCE.md)** - Quick command reference
   - Common commands
   - Configuration essentials
   - Troubleshooting tips

3. **[IMPLEMENTATION-SUMMARY.md](IMPLEMENTATION-SUMMARY.md)** - Technical details
   - Architecture overview
   - Testing checklist
   - Deployment guide

## 🎯 Key Features

✅ **Multiple Independent Instances** - Each connects to a different IBM Storage system  
✅ **Systemd Template Units** - Easy management with `ipquorum@<name>.service`  
✅ **Instance-Specific Configs** - Dedicated configuration per instance  
✅ **Isolated Resources** - Separate directories, logs, and processes  
✅ **Systemd Aliases** - Descriptive names for identification  
✅ **Security** - Per-instance credentials with proper permissions  
✅ **Management Tools** - Comprehensive CLI for instance lifecycle  
✅ **Complete Documentation** - Guides, examples, and troubleshooting  

## 🔧 Common Commands

```bash
# List all instances
sudo systemctl list-units 'ipquorum@*'

# Start/stop/restart instance
sudo systemctl start ipquorum@myinstance.service
sudo systemctl stop ipquorum@myinstance.service
sudo systemctl restart ipquorum@myinstance.service

# Check status
sudo systemctl status ipquorum@myinstance.service

# View logs
sudo journalctl -u ipquorum@myinstance.service -f

# Using instance manager
sudo ./ipquorum-instance-manager.sh list
sudo ./ipquorum-instance-manager.sh info myinstance
sudo ./ipquorum-instance-manager.sh logs myinstance
```

## 📋 Example Configurations

Three example configurations are provided in the `examples/` directory:

1. **system1.conf** - Simple production system (download-only)
2. **datacenter-a.conf** - PBHA configuration with mkquorumapp
3. **partition-prod.conf** - Secure TLS configuration

Copy and customize these for your environment:

```bash
sudo cp examples/system1.conf /etc/ipquorum/instances/myinstance.conf
sudo vi /etc/ipquorum/instances/myinstance.conf

# Important: Update IBM_STORAGE_SYSTEM with the hostname/IP of your Storage Virtualize system
# This should match the API_ENDPOINT value and is used for connectivity checks
```

## 🏗️ Architecture

```
/etc/ipquorum/instances/          # Instance configurations
/var/lib/ipquorum/<instance>/     # Instance data (JAR files)
/var/log/ipquorum/<instance>/     # Instance logs
/etc/systemd/system/ipquorum@.service  # Template service
```

Each instance runs independently with:
- Unique configuration file
- Separate password file
- Isolated data directory
- Isolated log directory
- Independent systemd service

## 🔒 Security

**Password files must have restrictive permissions:**

```bash
# Recommended location: /var/lib/ipquorum/.passwords/ (for SELinux compatibility)
sudo mkdir -p /var/lib/ipquorum/.passwords
echo 'your_password' | sudo tee /var/lib/ipquorum/.passwords/<instance>.password > /dev/null
sudo chmod 440 /var/lib/ipquorum/.passwords/<instance>.password
sudo chown ipquorum:ipquorum /var/lib/ipquorum/.passwords/<instance>.password

# On SELinux systems (RHEL/CentOS), restore context
sudo restorecon -v /var/lib/ipquorum/.passwords/<instance>.password
```

**Never use world-readable permissions (444, 644)!**

**Note:** Files in `/var/lib` automatically get correct SELinux contexts, avoiding permission issues.

## 🆘 Troubleshooting

### Alias Error on Enable?

If you see: `Failed to enable unit: Cannot alias ipquorum@instance.service as ipquorum-instance.service`

**This is a known issue with older systemd versions and can be safely ignored.** The service will still work correctly. The alias feature was removed for compatibility.

```bash
# The service is still enabled, verify with:
sudo systemctl is-enabled ipquorum@myinstance.service

# Start the service normally:
sudo systemctl start ipquorum@myinstance.service
```

### Go Downloader Not Found?

If you see: `Go download tool not found or not executable`

The script looks for these binary names in order:
1. `/usr/local/bin/ipquorum-download-go` (configured path)
2. `/usr/local/bin/ipquorum-download-go-linux-amd64` (release binary name)
3. `./ipquorum-download-go-linux-amd64` (current directory)
4. `./ipquorum-download-go` (current directory)

**Solutions:**
```bash
# Option 1: Rename the binary to match expected name
sudo mv ipquorum-download-go-linux-amd64 /usr/local/bin/ipquorum-download-go
sudo chmod +x /usr/local/bin/ipquorum-download-go

# Option 2: Create a symlink
sudo ln -s /path/to/ipquorum-download-go-linux-amd64 /usr/local/bin/ipquorum-download-go

# Option 3: Update config to specify exact path
# Edit /etc/ipquorum/instances/myinstance.conf
DOWNLOAD_TOOL_GO=/usr/local/bin/ipquorum-download-go-linux-amd64

# Option 4: Use a different download tool
IPQUORUM_DOWNLOAD_TOOL=bash  # or python
```

### Instance won't start?

```bash
# Check logs
sudo journalctl -u ipquorum@myinstance.service -n 50

# Validate configuration
sudo ./ipquorum-instance-manager.sh validate myinstance

# Check password file
sudo ls -la /var/lib/ipquorum/.passwords/myinstance.password

# Verify IBM_STORAGE_SYSTEM is set correctly (not <HOSTNAME_OR_IP>)
grep IBM_STORAGE_SYSTEM /etc/ipquorum/instances/myinstance.conf
```

### Download fails?

```bash
# Check download log
sudo tail -f /var/log/ipquorum/myinstance/download.log

# Test API connectivity (replace <api-endpoint> with your IBM_STORAGE_SYSTEM value)
curl -k https://<api-endpoint>:7443/rest/v1/auth

# Manual download (if validation doesn't auto-download)
sudo /usr/local/bin/ipquorum-download-multi.sh myinstance
```

### Permission errors?

```bash
# Fix directory permissions
sudo chown -R ipquorum:ipquorum /var/lib/ipquorum/myinstance
sudo chown -R ipquorum:ipquorum /var/log/ipquorum/myinstance

# Fix password file (use /var/lib for SELinux compatibility)
sudo chmod 440 /var/lib/ipquorum/.passwords/myinstance.password
sudo chown ipquorum:ipquorum /var/lib/ipquorum/.passwords/myinstance.password
sudo restorecon -v /var/lib/ipquorum/.passwords/myinstance.password
```

### Running scripts manually?

```bash
# Make scripts executable first
chmod +x script-name.sh

# Then run with ./
./script-name.sh [arguments]

# Or use bash directly
bash script-name.sh [arguments]
```

### Configuration has placeholder values?

If you see `<HOSTNAME_OR_IP>` or `<USERNAME>` in logs:

```bash
# Edit the instance configuration
sudo vi /etc/ipquorum/instances/myinstance.conf

# Replace ALL placeholder values:
# IBM_STORAGE_SYSTEM="<HOSTNAME_OR_IP>"  →  IBM_STORAGE_SYSTEM="10.33.7.80"
# API_ENDPOINT=<API_ENDPOINT>            →  API_ENDPOINT=10.33.7.80
# VIRTUALIZE_USERNAME=<USERNAME>         →  VIRTUALIZE_USERNAME=superuser

# Reload and restart
sudo systemctl daemon-reload
sudo systemctl restart ipquorum@myinstance.service
```

## 📖 Full Documentation

For complete documentation, see:
- **[README-ADVANCED.md](README-ADVANCED.md)** - Complete guide (advanced)
- **[QUICK-REFERENCE.md](QUICK-REFERENCE.md)** - Quick reference
- **[IMPLEMENTATION-SUMMARY.md](IMPLEMENTATION-SUMMARY.md)** - Technical details

## 🔗 Related Links

- **IBM FlashSystem Documentation**: https://www.ibm.com/docs/en/flashsystem-7x00/8.6.x
- **IP Quorum Information**: https://www.ibm.com/support/pages/node/7013877
- **Systemd Documentation**: https://www.freedesktop.org/software/systemd/man/

## 📞 Support

1. Review documentation in this directory
2. Check IBM support documentation
3. Contact IBM support for storage-specific issues

---

**Version**: 1.0.0  
**Status**: ✅ Production Ready  
**Last Updated**: 2026-04-16

*Made with ❤️ and systemd*