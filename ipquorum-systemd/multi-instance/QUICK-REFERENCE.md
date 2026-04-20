# IP Quorum Multi-Instance - Quick Reference Guide

## 🚀 Quick Commands

### Instance Management

```bash
# List all instances
sudo systemctl list-units 'ipquorum@*'

# Start instance
sudo systemctl start ipquorum@<instance>.service

# Stop instance
sudo systemctl stop ipquorum@<instance>.service

# Restart instance
sudo systemctl restart ipquorum@<instance>.service

# Status
sudo systemctl status ipquorum@<instance>.service

# Enable (auto-start on boot)
sudo systemctl enable ipquorum@<instance>.service

# Disable
sudo systemctl disable ipquorum@<instance>.service
```

### Using Instance Manager

```bash
# List instances
ipqm list

# Create instance
ipqm create <name>

# Start/stop/restart
ipqm start <name>
ipqm stop <name>
ipqm restart <name>

# View logs
ipqm logs <name>

# Show info
ipqm info <name>

# Validate config
ipqm validate <name>

# Delete instance
ipqm delete <name>
```

### Logs

```bash
# Systemd journal (follow)
sudo journalctl -u ipquorum@<instance>.service -f

# Systemd journal (last 100 lines)
sudo journalctl -u ipquorum@<instance>.service -n 100

# Download log
sudo tail -f /var/log/ipquorum/<instance>/download.log

# Application log
sudo tail -f /var/log/ipquorum/<instance>/ip_quorum.log.1

# All logs for instance
ipqm logs <instance> 100
```

### Multi-Instance Operations

```bash
# Start multiple
sudo systemctl start ipquorum@{inst1,inst2,inst3}.service

# Stop all
sudo systemctl stop 'ipquorum@*'

# Status all
sudo systemctl status 'ipquorum@*'

# Enable multiple
sudo systemctl enable ipquorum@{inst1,inst2,inst3}.service
```

## 📁 File Locations

```
Configuration:     /etc/ipquorum/instances/<instance>.conf
Password:          /etc/ipquorum/instances/.passwords/<instance>.password
Data Directory:    /var/lib/ipquorum/<instance>/
Log Directory:     /var/log/ipquorum/<instance>/
JAR File:          /var/lib/ipquorum/<instance>/ip_quorum.jar
Template Service:  /etc/systemd/system/ipquorum@.service
```

## ⚙️ Configuration Essentials

### Minimum Required Settings

```bash
INSTANCE_NAME=<instance-name>
IBM_STORAGE_SYSTEM="<description>"
API_ENDPOINT=<ip-or-hostname>
VIRTUALIZE_USERNAME=<username>
VIRTUALIZE_PASSWORD_FILE=/etc/ipquorum/instances/.passwords/<instance>.password
```

### Common Settings

```bash
# Download
IPQUORUM_DOWNLOAD_ENABLED=true
IPQUORUM_DOWNLOAD_TOOL=go              # go, python, or bash

# TLS
IPQUORUM_TLS_VERIFY=false              # true for valid certs

# Resources
IPQUORUM_MEMORY_MAX=512M
IPQUORUM_CPU_QUOTA=50%

# Logging
IPQUORUM_DEBUG=false
IPQUORUM_LOG_ROTATION=5
IPQUORUM_LOG_SIZE=5120
```

### PBHA Configuration

```bash
IPQUORUM_MKQUORUMAPP_ENABLED=true
IPQUORUM_PARTNERSYSTEM=<remote-system-name>
IPQUORUM_IP6=false
IPQUORUM_NOMETADATA=false
IPQUORUM_PARTNERIP6=false
```

## 🔧 Common Tasks

### Create New Instance

```bash
# 1. Create config
sudo cp /etc/ipquorum/instance.conf.template /etc/ipquorum/instances/myinst.conf
sudo vi /etc/ipquorum/instances/myinst.conf

# 2. Create password
echo 'password' | sudo tee /etc/ipquorum/instances/.passwords/myinst.password
sudo chmod 400 /etc/ipquorum/instances/.passwords/myinst.password
sudo chown root:ipquorum /etc/ipquorum/instances/.passwords/myinst.password

# 3. Create directories
sudo mkdir -p /var/lib/ipquorum/myinst /var/log/ipquorum/myinst
sudo chown -R ipquorum:ipquorum /var/lib/ipquorum/myinst /var/log/ipquorum/myinst

# 4. Enable and start
sudo systemctl enable --now ipquorum@myinst.service
```

### Update Configuration

```bash
# 1. Edit config
sudo vi /etc/ipquorum/instances/<instance>.conf

# 2. Reload systemd (if service file changed)
sudo systemctl daemon-reload

# 3. Restart instance
sudo systemctl restart ipquorum@<instance>.service
```

### Change Password

```bash
# 1. Update password file
echo 'new_password' | sudo tee /etc/ipquorum/instances/.passwords/<instance>.password

# 2. Verify permissions
sudo chmod 400 /etc/ipquorum/instances/.passwords/<instance>.password

# 3. Restart instance
sudo systemctl restart ipquorum@<instance>.service
```

### Backup Instance

```bash
# Backup configuration
sudo cp /etc/ipquorum/instances/<instance>.conf /backup/

# Backup password (encrypted)
sudo cp /etc/ipquorum/instances/.passwords/<instance>.password /backup/

# Backup JAR file
sudo cp /var/lib/ipquorum/<instance>/ip_quorum.jar /backup/
```

## 🔍 Troubleshooting

### Instance Won't Start

```bash
# Check status
sudo systemctl status ipquorum@<instance>.service

# Check logs
sudo journalctl -u ipquorum@<instance>.service -n 50

# Validate config
ipqm validate <instance>

# Check password file
sudo ls -la /etc/ipquorum/instances/.passwords/<instance>.password
sudo -u ipquorum cat /etc/ipquorum/instances/.passwords/<instance>.password
```

### Download Fails

```bash
# Check download log
sudo tail -f /var/log/ipquorum/<instance>/download.log

# Test API connectivity
curl -k https://<api-endpoint>:7443/rest/v1/auth

# Check credentials
ipqm validate <instance>
```

### Permission Errors

```bash
# Fix directory permissions
sudo chown -R ipquorum:ipquorum /var/lib/ipquorum/<instance>
sudo chown -R ipquorum:ipquorum /var/log/ipquorum/<instance>

# Fix password file
sudo chmod 440 /etc/ipquorum/instances/.passwords/<instance>.password
sudo chown root:ipquorum /etc/ipquorum/instances/.passwords/<instance>.password
```

## 📊 Monitoring

### Check All Instances

```bash
# List with status
ipqm list

# Status of all
sudo systemctl status 'ipquorum@*' --no-pager

# Running instances
ps aux | grep ip_quorum.jar
```

### Resource Usage

```bash
# Memory usage
sudo systemctl status ipquorum@<instance>.service | grep Memory

# CPU usage
top -p $(pgrep -f "ip_quorum.jar.*<instance>")

# Disk usage
du -sh /var/lib/ipquorum/<instance>
du -sh /var/log/ipquorum/<instance>
```

### Network Connections

```bash
# Check port 1260
sudo netstat -tlnp | grep 1260
sudo ss -tlnp | grep 1260

# Check connections
sudo lsof -i :1260
```

## 🎯 Best Practices

### Security

- ✅ Use `chmod 400` or `440` for password files
- ✅ Never commit passwords to version control
- ✅ Use `IPQUORUM_TLS_VERIFY=true` in production
- ✅ Rotate passwords regularly
- ✅ Use different passwords per instance

### Naming

- ✅ Use descriptive names: `prod-dc-a`, `test-partition-1`
- ✅ Avoid special characters
- ✅ Keep names short and meaningful
- ✅ Document in `IBM_STORAGE_SYSTEM` field

### Monitoring

- ✅ Enable auto-start: `systemctl enable`
- ✅ Monitor logs regularly
- ✅ Set up alerts for failures
- ✅ Check disk space for logs

### Maintenance

- ✅ Backup configurations regularly
- ✅ Test recovery procedures
- ✅ Keep documentation updated
- ✅ Review logs periodically

## 📞 Getting Help

```bash
# View full documentation
less /etc/ipquorum/README-ADVANCED.md

# Instance manager help
ipqm help

# Systemd service details
systemctl cat ipquorum@.service

# Check systemd logs
sudo journalctl -u ipquorum@<instance>.service --since today
```

## 🔗 Useful Links

- IBM FlashSystem Docs: https://www.ibm.com/docs/en/flashsystem-7x00/8.6.x
- IP Quorum Info: https://www.ibm.com/support/pages/node/7013877
- Systemd Documentation: https://www.freedesktop.org/software/systemd/man/

---

**Quick Reference v1.0** | Last updated: 2026-04-16