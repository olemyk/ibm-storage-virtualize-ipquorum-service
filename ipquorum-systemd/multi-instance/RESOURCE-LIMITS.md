# Customizing Resource Limits Per Instance

The systemd service file has default resource limits (512M memory, 50% CPU). To customize these per instance, use systemd drop-in files.

## Default Limits

The service file `ipquorum@.service` has these defaults:
- **Memory**: 512M
- **CPU**: 50%
- **File Descriptors**: 65536

## Customizing Per Instance

### Method 1: Using systemctl (Recommended)

```bash
# Create drop-in directory and edit
sudo systemctl edit ipquorum@myinstance.service
```

This opens an editor. Add your custom limits:

```ini
[Service]
MemoryMax=1G
CPUQuota=75%
```

Save and exit. Systemd will automatically:
- Create `/etc/systemd/system/ipquorum@myinstance.service.d/override.conf`
- Reload the daemon
- Apply the new limits

### Method 2: Manual Drop-in File

```bash
# Create drop-in directory
sudo mkdir -p /etc/systemd/system/ipquorum@myinstance.service.d/

# Create override file
sudo tee /etc/systemd/system/ipquorum@myinstance.service.d/resources.conf <<EOF
[Service]
MemoryMax=1G
CPUQuota=75%
LimitNOFILE=131072
EOF

# Reload systemd
sudo systemctl daemon-reload

# Restart instance
sudo systemctl restart ipquorum@myinstance.service
```

## Examples

### High-Performance Instance

```ini
[Service]
MemoryMax=2G
CPUQuota=100%
LimitNOFILE=131072
```

### Low-Resource Instance

```ini
[Service]
MemoryMax=256M
CPUQuota=25%
```

### Production PBHA Instance

```ini
[Service]
MemoryMax=1G
CPUQuota=75%
LimitNOFILE=131072
Nice=-10
```

## Verifying Limits

```bash
# Check current limits
sudo systemctl show ipquorum@myinstance.service | grep -E 'Memory|CPU|LimitNOFILE'

# View effective configuration
sudo systemctl cat ipquorum@myinstance.service
```

## Removing Custom Limits

```bash
# Remove drop-in file
sudo rm -rf /etc/systemd/system/ipquorum@myinstance.service.d/

# Or use systemctl
sudo systemctl revert ipquorum@myinstance.service

# Reload and restart
sudo systemctl daemon-reload
sudo systemctl restart ipquorum@myinstance.service
```

## Available Resource Limits

Common systemd resource directives:

| Directive | Description | Example |
|-----------|-------------|---------|
| `MemoryMax` | Maximum memory | `512M`, `1G`, `2G` |
| `CPUQuota` | CPU percentage | `25%`, `50%`, `100%` |
| `LimitNOFILE` | Max open files | `65536`, `131072` |
| `TasksMax` | Max tasks/threads | `100`, `200` |
| `Nice` | Process priority | `-20` to `19` (lower = higher priority) |
| `IOWeight` | I/O priority | `10` to `1000` |

## Notes

- Drop-in files override the main service file
- Changes require `systemctl daemon-reload`
- Restart the service to apply changes
- Each instance can have different limits
- Drop-in files are preserved during updates

---

**Made with ❤️ and systemd**