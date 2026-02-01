# IPQuorum Download Script

A Python utility to **download the IBM Virtualize IPQuorum JAR** via REST API and optionally **create a new Quorum App** (mkquorumapp).

> Script name: `ipquorum-restapi-download.py`

---

## Table of Contents
- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Options](#options)
  - [General](#general)
  - [Password Options](#password-options)
  - [mkquorumapp Payload](#mkquorumapp-payload)
- [Examples](#examples)
- [Environment Variables](#environment-variables)
- [Features](#features)
- [Error Handling](#error-handling)
- [Logging](#logging)
- [Troubleshooting](#troubleshooting)
- [Security & Best Practices](#security--best-practices)

---

## Overview

This Python script interacts with the IBM Storage Virtualize REST API to:
1. **Download** the IPQuorum JAR artifact to a local file
2. Optionally **create a fresh Quorum App** in the target system through the `mkquorumapp` API call

**Key Features:**
- **Type-safe**: Full type hints throughout the codebase
- **Secure**: Multiple password input methods with masking
- **Smart retry logic**: Handles rate limiting and transient errors
- **Professional logging**: Structured logging with configurable levels
- **Modular design**: Class-based architecture for maintainability
- **Cross-platform**: Works on Windows, Linux, and macOS

---

## Prerequisites

- **Python 3.6 or higher** (Python 3.8+ recommended)
- **pip** (Python package installer)
- Network access to the IBM Storage Virtualize API endpoint on port 7443
- Remote System in PBHA (MANDATORY if option mkquorumapp is enabled)
- Valid credentials:
  - **Monitor** role for IPQuorum JAR download
  - **Minimum Restricted Administrator** for creating a Quorum App

---

## Installation

1. **Clone or download the script**:
   ```bash
   cd ipquorum-download
   ```

2. **Install Python dependencies**:
   ```bash
   pip install -r requirements.txt
   ```
   
   Or install manually:
   ```bash
   pip install requests urllib3
   # For Python 3.6 only:
   pip install dataclasses
   ```

3. **Make the script executable** (Linux/macOS):
   ```bash
   chmod +x ipquorum-restapi-download.py
   ```

For detailed installation instructions, see [INSTALL.md](INSTALL.md).

---

## Quick Start

### Secure Interactive Use (Recommended):
```bash
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-prompt \
  --download --insecure
```

### Secure Automation:
```bash
# Create password file
echo 'password' > ~/.ipquorum_password
chmod 400 ~/.ipquorum_password

# Use password file
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-file ~/.ipquorum_password \
  --download --insecure
```

---

## Usage

```bash
python3 ipquorum-restapi-download.py [options]
```

Or if executable:
```bash
./ipquorum-restapi-download.py [options]
```

You can combine **general** options with **mkquorumapp payload** options. When `--mkquorumapp` is enabled, `--partnersystem` becomes **mandatory**.

---

## Options

### General

```
--mkquorumapp / --no-mkquorumapp     Enable/disable mkquorumapp call (default: enabled)
--download / --no-download           Enable/disable jar download (default: enabled)
--insecure / --secure                Use insecure TLS or strict TLS (default: insecure)
--api-endpoint <host>                Set API endpoint IP/hostname (REQUIRED)
--user <username>                    Auth username (env: VIRTUALIZE_USERNAME)
--output <file>                      Output jar filename (default: ip_quorum.jar)
--debug                              Enable debug logging
```

### Password Options

**Choose ONE of the following (in order of security):**

```
--pass-prompt                        Prompt for password interactively (MOST SECURE)
--pass-file <file>                   Read password from file (SECURE for automation)
--pass <password>                    Password in command line (INSECURE - testing only)
                                     (env: VIRTUALIZE_PASSWORD)
```

**Security Note:** See [SECURITY-GUIDE.md](SECURITY-GUIDE.md) for password best practices.

### mkquorumapp Payload

```
--ip6[=true|false] or --ip_6[=true|false]   Set IPv6 flag (default: false)
--nometadata[=true|false]                   Set nometadata flag (default: false)
--partnersystem <name>                      Set Partnersystem - Remote System name in PBHA 
                                            (MANDATORY if mkquorumapp is enabled)
--partnerip6[=true|false]                   Set partner IPv6 flag (default: false)
```

---

## Examples

### Example 1: Interactive Password (Most Secure)
```bash
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-prompt \
  --download --insecure
```

### Example 2: Password File (Secure for Automation)
```bash
# Create password file with restricted permissions
echo 'password' > ~/.ipquorum_pass
chmod 400 ~/.ipquorum_pass

# Use password file
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-file ~/.ipquorum_pass \
  --download --insecure
```

### Example 3: Create Quorum App + Download JAR
```bash
python3 ipquorum-restapi-download.py \
  --mkquorumapp --partnersystem svc_cluster02 \
  --api-endpoint 10.33.7.80 \
  --ip6=false --partnerip6=false --nometadata=false \
  --download --insecure \
  --user superuser --pass-prompt
```

### Example 4: Download JAR Only
```bash
python3 ipquorum-restapi-download.py \
  --no-mkquorumapp --download \
  --api-endpoint 10.33.7.80 \
  --output ip_quorum.jar \
  --user superuser --pass-prompt \
  --insecure
```

### Example 5: Using Environment Variables
```bash
export API_ENDPOINT=10.33.7.80
export VIRTUALIZE_USERNAME=superuser

python3 ipquorum-restapi-download.py \
  --download --insecure --pass-prompt
```

### Example 6: Debug Mode
```bash
python3 ipquorum-restapi-download.py \
  --debug \
  --api-endpoint 10.33.7.80 \
  --download --insecure \
  --user superuser --pass-prompt
```

### Example 7: Secure TLS (Production)
```bash
python3 ipquorum-restapi-download.py \
  --secure \
  --api-endpoint storage.example.com \
  --download \
  --user monitor_user --pass-prompt
```

---

## Environment Variables

The script supports the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `API_ENDPOINT` | API endpoint hostname/IP | (none) |
| `VIRTUALIZE_USERNAME` | Authentication username | (none) |
| `VIRTUALIZE_PASSWORD` | Authentication password | (none) |
| `IPQ_OUTPUT_FILE` | Output filename | `ip_quorum.jar` |

Example:
```bash
export API_ENDPOINT=10.33.7.80
export VIRTUALIZE_USERNAME=superuser
python3 ipquorum-restapi-download.py --download --insecure --pass-prompt
```

---

## Features

### 1. Secure Password Handling
- **Interactive prompt**: Hidden password input (most secure for manual use)
- **Password file**: File-based with permission control (secure for automation)
- **Environment variable**: Temporary use
- **Password masking**: Never logs passwords in plain text

### 2. Smart Authentication
- Exponential backoff for failed attempts
- **Fail fast** on wrong credentials (401/403)
- Automatic rate limit handling (HTTP 429)
- Retry-After header parsing
- Multiple token extraction methods
- Configurable max retries (default: 8)

### 3. Pre-flight Validation
- Endpoint reachability check before operations
- Network and TLS error detection
- Clear error messages for connectivity issues

### 4. Professional Logging
- Structured log messages with timestamps
- Configurable log levels (INFO, DEBUG)
- Debug mode for troubleshooting
- Password masking in all logs

### 5. Error Handling
- Custom exception types for different error scenarios
- Meaningful exit codes:
  - `0`: Success
  - `1`: General error (API, authentication)
  - `2`: Validation error (configuration)
  - `3`: Network error (connectivity)
  - `130`: User cancelled (Ctrl+C)

### 6. Type Safety
- Full type hints throughout the code
- Dataclass for configuration
- Better IDE support and code completion

---

## Error Handling

The script provides clear error messages and appropriate exit codes:

### Exit Codes
- **0**: Success
- **1**: General error (API call failed, authentication failed)
- **2**: Validation error (missing required parameters, invalid configuration)
- **3**: Network error (cannot reach endpoint, TLS error)
- **130**: User cancelled (Ctrl+C)

### Common Errors

**ValidationError**: Missing required parameters
```
Error: --partnersystem <name> is MANDATORY when --mkquorumapp is enabled
Error: --api-endpoint <host> is required
```

**NetworkError**: Cannot reach endpoint
```
Pre-flight: SSL/TLS error. Try --insecure or verify connectivity.
Pre-flight: Connection error. Verify endpoint and network connectivity.
```

**AuthenticationError**: Authentication failed
```
Authentication failed: Invalid username or password (HTTP 401)
Authentication failed: Insufficient permissions or invalid credentials (HTTP 403)
```

**APIError**: API call failed
```
mkquorumapp failed with status 400
Download failed with status 404
```

---

## Logging

### Log Levels

**INFO** (default): Shows important operations and results
```
2026-01-28 20:00:00 - INFO - Pre-flight: validating inputs & endpoint reachability...
2026-01-28 20:00:01 - INFO - Pre-flight: endpoint is reachable.
2026-01-28 20:00:02 - INFO - Get Token, please wait...
2026-01-28 20:00:03 - INFO - Authentication successful
2026-01-28 20:00:04 - INFO - Downloaded ip_quorum.jar (size: 1234567 bytes)
```

**DEBUG**: Shows detailed information for troubleshooting
```bash
python3 ipquorum-restapi-download.py --debug ...
```

Debug output includes:
- Configuration parameters (with masked passwords)
- HTTP request/response details
- Headers and payloads
- Retry attempts and delays

**Password Masking**: Passwords are always masked in logs:
```
2026-01-28 20:00:00 - DEBUG - password='********'
```

---

## Troubleshooting

### Issue: ModuleNotFoundError: No module named 'requests'
**Solution**: Install dependencies
```bash
pip install -r requirements.txt
```

### Issue: ModuleNotFoundError: No module named 'dataclasses'
**Solution**: Install dataclasses backport (Python 3.6 only)
```bash
pip install dataclasses
```

### Issue: Permission denied
**Solution**: Make script executable (Linux/macOS)
```bash
chmod +x ipquorum-restapi-download.py
```

### Issue: SSL/TLS errors
**Solution**: Use `--insecure` flag for testing (not recommended for production)
```bash
python3 ipquorum-restapi-download.py --insecure ...
```

### Issue: Authentication fails immediately
**Solution**: 
1. Verify username and password are correct
2. Check user role (Monitor for download, Restricted Admin for mkquorumapp)
3. The script fails fast on wrong credentials (401/403) - this is intentional
4. Enable debug mode to see detailed error messages:
```bash
python3 ipquorum-restapi-download.py --debug ...
```

### Issue: Connection timeout
**Solution**:
1. Verify endpoint is reachable: `ping <endpoint>`
2. Check firewall rules for port 7443
3. Verify DNS resolution
4. Try with IP address instead of hostname

### Issue: Rate limiting (429 errors)
**Solution**: The script automatically handles rate limiting with exponential backoff. If you see repeated 429 errors, wait a few minutes before retrying.

---

## Security & Best Practices

### 1. Protect Credentials

**Use secure password methods:**
- ✅ **Interactive prompt** (`--pass-prompt`) for manual use
- ✅ **Password file** (`--pass-file`) with `chmod 400` for automation
- ⚠️ **Environment variable** for temporary use only
- ❌ **Command line** (`--pass`) NEVER in production

See [SECURITY-GUIDE.md](SECURITY-GUIDE.md) for detailed password security practices.

### 2. TLS Settings
- **Production**: Always use `--secure` to enforce TLS verification
- **Testing**: Use `--insecure` only in controlled environments
- **Best practice**: Install proper CA certificates and use `--secure`

### 3. Least Privilege
- Use **Monitor** role for download-only operations
- Use **Restricted Administrator** only when creating Quorum Apps
- Don't use superuser unless absolutely necessary

### 4. Audit & Logs
- Enable debug logging for troubleshooting: `--debug`
- Store logs securely for audit purposes
- Review logs regularly for security events
- Passwords are automatically masked in all logs

### 5. Network Security
- Use VPN or secure network when accessing storage systems
- Restrict API access to authorized networks
- Use firewall rules to limit access to port 7443

---

## Additional Documentation

- **[WORKFLOW-DIAGRAM.md](WORKFLOW-DIAGRAM.md)** - Visual workflow and execution flow
- **[SECURITY-GUIDE.md](SECURITY-GUIDE.md)** - Comprehensive password security guide
- **[INSTALL.md](INSTALL.md)** - Detailed installation instructions
- **[MIGRATION-GUIDE.md](MIGRATION-GUIDE.md)** - Migration from bash version

### IBM Documentation
- [IPQuorum Info](https://www.ibm.com/support/pages/ibm-storage-virtualize-ip-quorum-application-requirements-1)
- [IP quorum application](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=quorum-ip-application)
- [Storage Virtualize RESTful API](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=interface-storage-virtualize-restful-api)

---

## Maintainers
- Ole Kristian Myklebust

---

## License
MIT License