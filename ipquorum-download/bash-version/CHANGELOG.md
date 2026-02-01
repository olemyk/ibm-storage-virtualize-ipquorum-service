# Changelog - Bash Version

All notable changes to the bash version of the IPQuorum download script will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.0.0] - 2026-01-29

### Added
- **Debug Mode**: New `--debug` flag for verbose troubleshooting output
- **Password Security Enhancements**:
  - `--pass-prompt`: Interactive password prompt with hidden input (most secure)
  - `--pass-file`: Read password from file with permission checking
  - Password file permission warnings (recommends chmod 400)
  - Password masking in all debug output
- **Enhanced Validation**:
  - Required field validation for `--api-endpoint`
  - Required field validation for `--user`
  - Clear error messages for missing parameters
- **Fail-Fast Authentication**:
  - Immediate exit on 401 (Invalid credentials)
  - Immediate exit on 403 (Insufficient permissions)
  - No retry on authentication errors
- **Rate Limiting Support**:
  - Automatic handling of HTTP 429 responses
  - Exponential backoff with Retry-After header parsing
  - Configurable max retries (default: 8)

### Changed
- **Debug Output**: Now only shown when `--debug` flag is used
- **Token Display**: Token no longer printed to stdout by default
- **Help Text**: Updated to mark `--api-endpoint` as REQUIRED
- **Curl Flags**: Fixed inconsistent use of `-k` flag, now uses `${insecure_api}` consistently

### Fixed
- Removed duplicate `INSECURE` variable (consolidated to `insecure_api`)
- Fixed hardcoded `-k` flag in mkquorumapp curl call
- All debug messages now properly go to stderr

### Security
- Password masking in all output (shows `********`)
- Password file permission checking (warns if not 400/600)
- No sensitive data in normal output mode
- Debug mode shows only first 20 characters of token

---

## [1.0.0] - 2026-01-28

### Added
- Initial release with core functionality
- Download IPQuorum JAR via REST API
- Create new Quorum App via mkquorumapp API
- Basic authentication with retry logic
- Pre-flight endpoint validation
- Support for IPv6 configurations
- Environment variable support
- Configurable TLS verification (--insecure/--secure)

### Features
- Command-line argument parsing
- JSON payload construction for mkquorumapp
- HTTP status code validation
- File size reporting for downloads
- Multiple token extraction methods
- Support for partnersystem configuration

---

## Version Comparison

| Feature | v1.0.0 | v2.0.0 |
|---------|--------|--------|
| Basic download | ✅ | ✅ |
| mkquorumapp | ✅ | ✅ |
| Password prompt | ❌ | ✅ |
| Password file | ❌ | ✅ |
| Password masking | ❌ | ✅ |
| Debug mode | ❌ | ✅ |
| Fail-fast auth | ❌ | ✅ |
| Rate limiting | ❌ | ✅ |
| Validation | Basic | Enhanced |
| Security | Basic | Advanced |

---

## Migration Guide

### From v1.0.0 to v2.0.0

**Breaking Changes:**
- None - v2.0.0 is fully backward compatible

**Recommended Changes:**

1. **Use secure password methods:**
   ```bash
   # Old (v1.0.0) - password in command line
   ./ipquorum-restapi-download.sh --pass password
   
   # New (v2.0.0) - interactive prompt (recommended)
   ./ipquorum-restapi-download.sh --pass-prompt
   
   # New (v2.0.0) - password file (for automation)
   echo 'password' > ~/.ipquorum_pass && chmod 400 ~/.ipquorum_pass
   ./ipquorum-restapi-download.sh --pass-file ~/.ipquorum_pass
   ```

2. **Enable debug mode for troubleshooting:**
   ```bash
   # Old (v1.0.0) - debug always on
   ./ipquorum-restapi-download.sh --download
   
   # New (v2.0.0) - clean output by default, debug on demand
   ./ipquorum-restapi-download.sh --download --debug
   ```

3. **Explicit required parameters:**
   ```bash
   # Now required (was optional via env vars)
   ./ipquorum-restapi-download.sh \
     --api-endpoint 10.33.7.80 \
     --user superuser \
     --pass-prompt
   ```

---

## Maintainer
Ole Kristian Myklebust

## License
MIT License