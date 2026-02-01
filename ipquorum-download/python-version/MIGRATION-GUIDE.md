# Migration Guide: Bash to Python

This guide helps you migrate from the bash version to the Python version of the IPQuorum download script.

## Quick Migration

The Python version maintains **full CLI compatibility** with the bash version. In most cases, you can simply replace:

```bash
# Bash version
./ipquorum-restapi-download.sh [options]

# Python version
python3 ipquorum-restapi-download.py [options]
# or
./ipquorum-restapi-download.py [options]
```

## Installation Steps

1. **Install Python dependencies**:
   ```bash
   pip install -r requirements.txt
   ```

2. **Make script executable** (Linux/macOS):
   ```bash
   chmod +x ipquorum-restapi-download.py
   ```

3. **Test the script**:
   ```bash
   python3 ipquorum-restapi-download.py --help
   ```

## Command Equivalence

All bash commands work identically in Python:

### Example 1: Download only
**Bash:**
```bash
./ipquorum-restapi-download.sh \
  --no-mkquorumapp --download \
  --api-endpoint 10.33.7.80 \
  --output ip_quorum.jar \
  --user superuser --pass password \
  --insecure
```

**Python:**
```bash
python3 ipquorum-restapi-download.py \
  --no-mkquorumapp --download \
  --api-endpoint 10.33.7.80 \
  --output ip_quorum.jar \
  --user superuser --pass password \
  --insecure
```

### Example 2: Create Quorum App + Download
**Bash:**
```bash
./ipquorum-restapi-download.sh \
  --api-endpoint 10.33.7.80 \
  --mkquorumapp --partnersystem svc_cluster02 \
  --ip6=false --partnerip6=false --nometadata=false \
  --download --insecure \
  --user superuser --pass password
```

**Python:**
```bash
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --mkquorumapp --partnersystem svc_cluster02 \
  --ip6=false --partnerip6=false --nometadata=false \
  --download --insecure \
  --user superuser --pass password
```

## Environment Variables

Both versions support the same environment variables:

```bash
export API_ENDPOINT=10.33.7.80
export VIRTUALIZE_USERNAME=superuser
export VIRTUALIZE_PASSWORD=password
export IPQ_OUTPUT_FILE=ip_quorum.jar

# Works with both versions
./ipquorum-restapi-download.sh --download --insecure
python3 ipquorum-restapi-download.py --download --insecure
```

## New Features in Python Version

### 1. Debug Mode
The Python version adds a `--debug` flag for detailed troubleshooting:

```bash
python3 ipquorum-restapi-download.py --debug \
  --api-endpoint 10.33.7.80 \
  --download --insecure \
  --user superuser --pass password
```

### 2. Better Error Messages
Python version provides more detailed error messages:

**Bash:**
```
Error: --partnersystem <name> is MANDATORY when --mkquorumapp is enabled.
```

**Python:**
```
2026-01-28 20:00:00 - ERROR - Validation error: --partnersystem <name> is MANDATORY when --mkquorumapp is enabled
```

### 3. Structured Logging
Python version uses professional logging with timestamps:

```
2026-01-28 20:00:00 - INFO - Pre-flight: validating inputs & endpoint reachability...
2026-01-28 20:00:01 - INFO - Pre-flight: endpoint is reachable.
2026-01-28 20:00:02 - INFO - Get Token, please wait...
2026-01-28 20:00:03 - INFO - Authentication successful
2026-01-28 20:00:04 - INFO - Downloaded ip_quorum.jar (size: 1234567 bytes)
2026-01-28 20:00:05 - INFO - Operation completed successfully
```

## Exit Codes

Both versions use the same exit codes:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (API, authentication) |
| 2 | Validation error (configuration) |
| 3 | Network error (connectivity) |
| 130 | User cancelled (Ctrl+C) |

## Troubleshooting

### Issue: Python not found
**Solution**: Install Python 3.8 or higher
```bash
# macOS
brew install python3

# Ubuntu/Debian
sudo apt-get install python3 python3-pip

# RHEL/CentOS
sudo yum install python3 python3-pip
```

### Issue: Module not found
**Solution**: Install dependencies
```bash
pip install -r requirements.txt
```

### Issue: Permission denied
**Solution**: Make script executable
```bash
chmod +x ipquorum-restapi-download.py
```

## Performance Comparison

Both versions have similar performance:

| Operation | Bash | Python |
|-----------|------|--------|
| Authentication | ~1-2s | ~1-2s |
| mkquorumapp | ~2-3s | ~2-3s |
| Download (10MB) | ~5-10s | ~5-10s |
| Total overhead | Minimal | Minimal |

## When to Use Which Version

### Use Bash Version When:
- You're on a system without Python
- You prefer minimal dependencies (only curl and jq)
- You're already using bash scripts in your workflow

### Use Python Version When:
- You need better error handling and debugging
- You want professional logging
- You're on Windows (without WSL)
- You need to extend or customize the script
- You want better IDE support and code completion
- You need unit testing capabilities

## Gradual Migration Strategy

You can use both versions side-by-side during migration:

1. **Phase 1**: Install Python version alongside bash version
2. **Phase 2**: Test Python version in non-production environments
3. **Phase 3**: Gradually migrate production scripts
4. **Phase 4**: Keep bash version as backup until fully confident

## Support

Both versions are maintained and supported. Choose the one that best fits your environment and requirements.

For issues or questions:
- Check the respective README files
- Review the troubleshooting sections
- Enable debug mode for detailed diagnostics

---

**Maintainer**: Ole Kristian Myklebust