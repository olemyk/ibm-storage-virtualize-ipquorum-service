# ⚠️ DEPRECATION NOTICE

## Single-Instance Files in This Directory Are DEPRECATED

**Effective Date:** 2026-04-19

---

## Deprecated Files

The following files in this directory (`ipquorum-systemd/`) are **DEPRECATED** and will be moved to the archive:

- ❌ `ibm-virtualize-ipquorum-improved.service`
- ❌ `install-ipquorum-service.sh`
- ❌ `ipquorum-download.sh`
- ❌ `ipquorum-start.sh`
- ❌ `ipquorum.conf`
- ❌ `README-IMPROVED-SERVICE.md`
- ❌ `ARCHITECTURE.md`
- ❌ `CHANGELOG.md`
- ❌ `manual-config/`

**These files have been archived in:** `archive/ipquorum-systemd-single-instance/`

---

## ✅ Use Multi-Instance Installer Instead

**Location:** `ipquorum-systemd/multi-instance/`

### Why Multi-Instance is Better

✅ **Works for Single Instance Too** - No need for separate installer  
✅ **Instance Manager CLI** - Easy management with `ipquorum-instance-manager.sh`  
✅ **Better Architecture** - Template-based systemd units  
✅ **More Flexible** - Add instances anytime  
✅ **Better Security** - Per-instance credentials in `/var/lib`  
✅ **Easier Troubleshooting** - Isolated logs per instance  
✅ **Production Ready** - Handles all edge cases  

### Quick Start with Multi-Instance

```bash
# Navigate to multi-instance directory
cd ipquorum-systemd/multi-instance/

# Install (works for single instance too!)
sudo ./install-multi-instance.sh

# Create your first instance
sudo ipquorum-instance-manager.sh create production

# Manage it
sudo ipquorum-instance-manager.sh start production
sudo ipquorum-instance-manager.sh status production
sudo ipquorum-instance-manager.sh logs production
```

### For Single Instance Users

The multi-instance installer works perfectly for single instance deployments:

```bash
# Just create ONE instance - that's it!
sudo ipquorum-instance-manager.sh create myinstance

# Use it exactly like the old single-instance service
sudo systemctl start ipquorum@myinstance.service
sudo systemctl status ipquorum@myinstance.service
```

---

## Migration from Single to Multi-Instance

If you're currently using the deprecated single-instance installer:

### Step 1: Install Multi-Instance System

```bash
cd ipquorum-systemd/multi-instance/
sudo ./install-multi-instance.sh
```

### Step 2: Stop Old Service

```bash
sudo systemctl stop ibm-virtualize-ipquorum.service
sudo systemctl disable ibm-virtualize-ipquorum.service
```

### Step 3: Create New Instance

```bash
# Use interactive configuration
sudo ipquorum-instance-manager.sh create production

# Or copy your old config values manually
sudo vi /etc/ipquorum/instances/production.conf
```

### Step 4: Start New Service

```bash
sudo systemctl enable ipquorum@production.service
sudo systemctl start ipquorum@production.service
sudo systemctl status ipquorum@production.service
```

---

## Documentation

- **Multi-Instance README:** [multi-instance/README.md](multi-instance/README.md)
- **Quick Reference:** [multi-instance/QUICK-REFERENCE.md](multi-instance/QUICK-REFERENCE.md)
- **Implementation Details:** [multi-instance/IMPLEMENTATION-SUMMARY.md](multi-instance/IMPLEMENTATION-SUMMARY.md)
- **Archived Single-Instance Docs:** [../archive/README.md](../archive/README.md)

---

## Questions?

**Q: Can I still use the single-instance installer?**  
A: Not recommended. It's deprecated and unmaintained. Use multi-instance instead.

**Q: Will multi-instance work for my single instance use case?**  
A: Yes! Just create one instance. It's actually easier to use.

**Q: What if I need help migrating?**  
A: Follow the migration steps above or see the multi-instance README.

**Q: Where did the old files go?**  
A: They're archived in `archive/ipquorum-systemd-single-instance/` for reference.

---

**Last Updated:** 2026-04-19  
**Deprecation Date:** 2026-04-19