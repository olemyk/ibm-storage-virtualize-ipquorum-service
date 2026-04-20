# Deprecated Components Archive

This directory contains deprecated components that are no longer actively maintained but kept for historical reference.

## ⚠️ WARNING: DEPRECATED CONTENT

**All files in this directory are DEPRECATED and should NOT be used for new installations.**

Please use the current supported versions instead.

---

## Archived Components

### 1. Single-Instance IP Quorum Service (Deprecated: 2026-04-19)

**Location:** `archive/ipquorum-systemd-single-instance/`

**Reason for Deprecation:** Replaced by multi-instance installer which provides:
- Support for both single and multiple instances
- Better management tools (instance manager CLI)
- More flexible architecture
- Better security and isolation
- Easier troubleshooting

**Migration Path:**
Use the multi-instance installer instead:
```bash
# New installation location
cd ipquorum-systemd/multi-instance/

# Install (works for single instance too)
sudo ./install-multi-instance.sh

# Create instance
sudo ipquorum-instance-manager.sh create myinstance
```

**Files Archived:**
- `ibm-virtualize-ipquorum-improved.service` - Single-instance systemd service
- `install-ipquorum-service.sh` - Single-instance installer
- `ipquorum-download.sh` - Single-instance download script
- `ipquorum-start.sh` - Single-instance start script
- `ipquorum.conf` - Single-instance configuration
- `README-IMPROVED-SERVICE.md` - Single-instance documentation
- `ARCHITECTURE.md` - Single-instance architecture docs
- `CHANGELOG.md` - Single-instance changelog
- `manual-config/` - Manual configuration files

---

## Using Archived Components

**⚠️ NOT RECOMMENDED**

If you absolutely must use archived components (not recommended):

1. Understand they are no longer maintained
2. No bug fixes or updates will be provided
3. Security vulnerabilities will not be patched
4. Use at your own risk

---

## Current Supported Versions

### Multi-Instance IP Quorum Service ✅

**Location:** `ipquorum-systemd/multi-instance/`

**Features:**
- ✅ Single AND multiple instance support
- ✅ Instance manager CLI tool
- ✅ Template-based systemd units
- ✅ Per-instance configuration and credentials
- ✅ Isolated logs and data directories
- ✅ Auto-download with fallback options
- ✅ Interactive configuration
- ✅ Comprehensive documentation

**Documentation:**
- [Multi-Instance README](../ipquorum-systemd/multi-instance/README.md)
- [Quick Reference](../ipquorum-systemd/multi-instance/QUICK-REFERENCE.md)
- [Implementation Summary](../ipquorum-systemd/multi-instance/IMPLEMENTATION-SUMMARY.md)

---

## Questions?

If you have questions about:
- **Migration:** See the multi-instance README
- **Why deprecated:** See reasons above
- **Support:** Only multi-instance version is supported

---

**Last Updated:** 2026-04-19  
**Archive Created:** 2026-04-19