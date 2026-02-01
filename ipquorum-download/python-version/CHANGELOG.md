# Changelog - Python Version

All notable changes to the Python version of the IPQuorum download script will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2026-01-28

### Added - Initial Release

#### Core Features
- **Complete Python rewrite** of bash script with enhanced functionality
- **Cross-platform support**: Windows, Linux, and macOS
- **Type-safe implementation**: Full type hints throughout codebase
- **Class-based architecture**: IPQuorumClient for maintainability

#### Security Features
- **Multiple password input methods**:
  - `--pass-prompt`: Interactive hidden input (most secure)
  - `--pass-file`: File-based with permission checking
  - `--pass`: Command-line (testing only)
  - Environment variable support
- **Password masking**: Never logs passwords in plain text
- **Password file validation**: Checks file permissions and warns if too permissive
- **Secure defaults**: Prompts for password if none provided

#### Authentication
- **Smart retry logic**: Exponential backoff for transient errors
- **Fail-fast on auth errors**: Immediate exit on 401/403
- **Rate limiting support**: Automatic handling of HTTP 429
- **Retry-After header parsing**: Respects server rate limit guidance
- **Multiple token extraction methods**: Supports various API response formats
- **Configurable retries**: Default 8 attempts with 3-second base delay

#### API Operations
- **Pre-flight validation**: Endpoint reachability check before operations
- **mkquorumapp support**: Create new Quorum Apps with full parameter support
- **Download functionality**: Robust JAR file download with validation
- **IPv6 support**: Full support for IPv6 configurations
- **Partner system configuration**: PBHA remote system support

#### Logging & Debugging
- **Professional logging**: Structured logs with timestamps
- **Configurable log levels**: INFO (default) and DEBUG modes
- **Debug mode**: `--debug` flag for detailed troubleshooting
- **Password masking in logs**: All passwords masked as `********`
- **Request/response logging**: Full HTTP details in debug mode

#### Error Handling
- **Custom exception types**:
  - `AuthenticationError`: Authentication failures
  - `APIError`: API call failures
  - `ValidationError`: Configuration errors
  - `NetworkError`: Connectivity issues
- **Meaningful exit codes**:
  - `0`: Success
  - `1`: General error (API, authentication)
  - `2`: Validation error (configuration)
  - `3`: Network error (connectivity)
  - `130`: User cancelled (Ctrl+C)
- **Clear error messages**: Actionable error descriptions

#### Configuration
- **Environment variable support**:
  - `API_ENDPOINT`: API endpoint hostname/IP
  - `VIRTUALIZE_USERNAME`: Authentication username
  - `VIRTUALIZE_PASSWORD`: Authentication password
  - `IPQ_OUTPUT_FILE`: Output filename
- **Dataclass configuration**: Type-safe config with validation
- **Flexible CLI arguments**: Support for both `--flag` and `--flag=value` formats

#### Documentation
- **Comprehensive README**: Full usage guide with examples
- **INSTALL.md**: Detailed installation instructions
- **SECURITY-GUIDE.md**: Password security best practices
- **WORKFLOW-DIAGRAM.md**: Visual workflow documentation
- **MIGRATION-GUIDE.md**: Bash to Python migration guide
- **Inline documentation**: Docstrings for all functions and classes

#### Dependencies
- **Minimal requirements**:
  - `requests>=2.27.0`: HTTP client
  - `urllib3>=1.26.0`: HTTP library
  - `dataclasses>=0.6`: Python 3.6 backport (if needed)
- **Python 3.6+ compatible**: Works with older Python versions
- **Python 3.8+ recommended**: Best performance and features

### Technical Details

#### Architecture
```
IPQuorumClient (main class)
├── __init__(): Initialize with config
├── preflight(): Validate endpoint reachability
├── authenticate(): Get auth token with retry
├── create_quorum_app(): Call mkquorumapp API
└── download_jar(): Download IPQuorum JAR
```

#### Key Functions
- `mask_password()`: Password masking for logs
- `to_bool()`: Boolean conversion from strings
- `setup_logging()`: Configure logging system
- `read_password_from_file()`: Secure file reading
- `prompt_password()`: Interactive password input

#### Code Quality
- **Type hints**: Full type annotations
- **Docstrings**: Comprehensive documentation
- **Error handling**: Try-except blocks with specific exceptions
- **Code organization**: Logical separation of concerns
- **PEP 8 compliant**: Follows Python style guide

---

## Comparison with Bash Version

| Feature | Bash | Python |
|---------|------|--------|
| **Platform** | Linux/macOS | Windows/Linux/macOS |
| **Dependencies** | bash, curl, jq | Python 3.6+, requests |
| **Type Safety** | ❌ | ✅ Full type hints |
| **Password Prompt** | ✅ | ✅ |
| **Password File** | ✅ | ✅ |
| **Password Masking** | ✅ | ✅ Advanced |
| **Debug Mode** | ✅ | ✅ |
| **Fail-fast Auth** | ✅ | ✅ |
| **Rate Limiting** | ✅ | ✅ |
| **Logging** | Basic | Professional |
| **Error Handling** | Basic | Advanced |
| **Exit Codes** | Basic | Comprehensive |
| **Documentation** | Good | Comprehensive |
| **Code Structure** | Procedural | Object-oriented |
| **Maintainability** | Good | Excellent |

---

## Why Python Version?

### Advantages
1. **Cross-platform**: Works on Windows without WSL/Cygwin
2. **Better error handling**: Custom exceptions and clear messages
3. **Professional logging**: Structured logs with levels
4. **Type safety**: Catch errors before runtime
5. **Maintainability**: Class-based, modular design
6. **Testing**: Easier to unit test
7. **IDE support**: Better autocomplete and refactoring

### Use Cases
- **Enterprise environments**: Need cross-platform support
- **Windows users**: Native Windows support
- **Automation**: Better error handling and logging
- **Development**: Easier to extend and maintain
- **Security-conscious**: Advanced password masking

---

## Future Enhancements

### Planned Features
- [ ] Configuration file support (YAML/JSON)
- [ ] Multiple endpoint support (cluster failover)
- [ ] Async/await for concurrent operations
- [ ] Progress bars for downloads
- [ ] Checksum verification for downloads
- [ ] Automatic retry on network failures
- [ ] Webhook notifications on completion
- [ ] Metrics and telemetry collection

### Under Consideration
- [ ] GUI version (tkinter/PyQt)
- [ ] REST API wrapper library
- [ ] Plugin system for extensions
- [ ] Docker container support
- [ ] Kubernetes operator

---

## Migration from Bash

The Python version is designed to be a drop-in replacement for the bash version with the same command-line interface. See [MIGRATION-GUIDE.md](MIGRATION-GUIDE.md) for detailed migration instructions.

**Quick Migration:**
```bash
# Bash version
./ipquorum-restapi-download.sh --api-endpoint 10.33.7.80 --user superuser --pass-prompt

# Python version (same arguments)
python3 ipquorum-restapi-download.py --api-endpoint 10.33.7.80 --user superuser --pass-prompt
```

---

## Contributing

Contributions are welcome! Please ensure:
- Python 3.6+ compatibility
- Type hints for all functions
- Docstrings for public APIs
- Unit tests for new features
- Updated documentation

---

## Maintainer
Ole Kristian Myklebust

## License
MIT License