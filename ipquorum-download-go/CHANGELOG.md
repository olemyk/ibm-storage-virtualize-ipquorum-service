# Changelog - Go Version

All notable changes to the Go version of the IPQuorum download tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---
## [1.0.1] - 2026-02-01

### Fixed
- **Password file permission check**: Fixed false warning for 440 permissions
  - Previously warned about group-readable files (440 permissions)
  - Now only warns if file is world-readable (other permissions set)
  - 440 (owner+group read) is acceptable for systemd services
  - 400 (owner read only) is still recommended for maximum security
  - Updated warning message to suggest both 400 and 440 as valid options

### Technical Details
- Changed permission check from `mode&0077` (group+other) to `mode&0007` (other only)
- This allows group-readable files which are needed when the service runs as a different user
- Systemd services often use `root:serviceuser` ownership with 440 permissions


## [1.0.0] - 2026-01-29

### Added - Initial Release

#### Core Features
- **Complete Go implementation** of IPQuorum download tool
- **Single binary distribution**: No runtime dependencies
- **Cross-platform support**: Native binaries for Linux, macOS, Windows
- **Type-safe implementation**: Full compile-time type checking
- **High performance**: Fast startup (<10ms), low memory (<20MB)

#### Security Features
- **Multiple password input methods**:
  - `--pass-prompt`: Interactive hidden input (most secure)
  - `--pass-file`: File-based with permission checking
  - `--pass`: Command-line (testing only)
  - Environment variable support
- **Password masking**: Never logs passwords in plain text
- **Password file validation**: Warns if permissions are too permissive
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
- **Structured logging**: Clear, timestamped log messages
- **Configurable log levels**: INFO (default) and DEBUG modes
- **Debug mode**: `--debug` flag for detailed troubleshooting
- **Password masking in logs**: All passwords masked as `********`
- **Request/response logging**: Full HTTP details in debug mode

#### Error Handling
- **Custom error types**:
  - `ValidationError`: Configuration validation failures
  - `AuthenticationError`: Authentication failures
  - `NetworkError`: Network connectivity issues
  - `APIError`: API call failures
- **Meaningful exit codes**:
  - `0`: Success
  - `1`: General error (API, authentication)
  - `2`: Validation error (configuration)
  - `3`: Network error (connectivity)
- **Clear error messages**: Actionable error descriptions

#### CLI & User Interface
- **Cobra-based CLI**: Professional command-line interface
- **Comprehensive flags**: All options from Python/Bash versions
- **Help text**: Detailed usage information
- **Version command**: Display version, commit, and build date
- **Environment variable support**: All major settings

#### Build & Distribution
- **Makefile**: Convenient build targets
- **Cross-compilation**: Build for all platforms with one command
- **Version embedding**: Build-time version information
- **Optimized binaries**: Stripped symbols for smaller size
- **Multiple architectures**: amd64 and arm64 support

---

## Architecture

### Package Structure
```
ipquorum-go/
├── main.go                    # Entry point & CLI
├── pkg/                       # Public packages
│   ├── config/               # Configuration & validation
│   ├── client/               # HTTP client & API operations
│   ├── password/             # Password handling
│   └── logger/               # Logging system
└── internal/                  # Private packages
    ├── errors/               # Custom error types
    └── utils/                # Helper functions
```

### Key Components

#### Configuration (`pkg/config`)
- Type-safe configuration struct
- Validation with clear error messages
- Default values for all settings
- Environment variable support

#### HTTP Client (`pkg/client`)
- Custom HTTP client with retry logic
- TLS configuration (secure/insecure)
- Pre-flight endpoint validation
- Authentication with fail-fast
- mkquorumapp API operations
- Download functionality

#### Password Handling (`pkg/password`)
- Interactive password prompt (hidden input)
- Password file reading with permission checks
- Environment variable support
- Password masking for logs

#### Logging (`pkg/logger`)
- Structured logging with timestamps
- Configurable log levels (INFO, DEBUG)
- Clean, readable output

#### Error Handling (`internal/errors`)
- Custom error types for different scenarios
- Error wrapping and unwrapping
- Meaningful error messages

---

## Comparison with Other Versions

| Feature | Bash | Python | **Go** |
|---------|------|--------|--------|
| **Platform** | Linux/macOS | Windows/Linux/macOS | ✅ **All platforms** |
| **Dependencies** | bash, curl, jq | Python 3.6+, requests | ✅ **None (single binary)** |
| **Type Safety** | ❌ | ✅ | ✅ **Compile-time** |
| **Startup Time** | Fast | ~100ms | ✅ **<10ms** |
| **Memory Usage** | Low | ~50MB | ✅ **<20MB** |
| **Binary Size** | N/A | N/A | ✅ **<10MB** |
| **Cross-Compile** | ❌ | ❌ | ✅ **Yes** |
| **Password Prompt** | ✅ | ✅ | ✅ |
| **Password File** | ✅ | ✅ | ✅ |
| **Password Masking** | ✅ | ✅ | ✅ |
| **Debug Mode** | ✅ | ✅ | ✅ |
| **Fail-fast Auth** | ✅ | ✅ | ✅ |
| **Rate Limiting** | ✅ | ✅ | ✅ |
| **Performance** | Good | Good | ✅ **Excellent** |
| **Maintainability** | Good | Excellent | ✅ **Excellent** |

---

## Why Go Version?

### Advantages
1. **Single Binary**: No runtime dependencies, just distribute one executable
2. **Performance**: Compiled code, faster than interpreted languages
3. **Cross-Platform**: Build for all platforms from any platform
4. **Type Safety**: Catch errors at compile time
5. **Memory Efficient**: Low memory footprint
6. **Fast Startup**: Near-instant startup time
7. **Easy Distribution**: Single file, no installation needed
8. **Enterprise Ready**: Used by major cloud-native projects

### Use Cases
- **DevOps/CI/CD**: Single binary for pipelines
- **Containers**: Minimal Docker images (scratch/alpine)
- **Air-Gapped Systems**: No package installation
- **Multi-Platform**: One codebase for all platforms
- **Performance Critical**: Large-scale automation

---

## Building from Source

### Prerequisites
- Go 1.19 or higher
- Make (optional, for convenience)

### Quick Build
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Install to $GOPATH/bin
make install
```

### Manual Build
```bash
# Current platform
go build -o ipquorum main.go

# Linux amd64
GOOS=linux GOARCH=amd64 go build -o ipquorum-linux-amd64 main.go

# macOS arm64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o ipquorum-darwin-arm64 main.go

# Windows amd64
GOOS=windows GOARCH=amd64 go build -o ipquorum-windows-amd64.exe main.go
```

---

## Installation

### Pre-built Binaries
Download from GitHub Releases:
```bash
# Linux
curl -LO https://github.com/user/ipquorum-go/releases/latest/download/ipquorum-linux-amd64
chmod +x ipquorum-linux-amd64
sudo mv ipquorum-linux-amd64 /usr/local/bin/ipquorum

# macOS
curl -LO https://github.com/user/ipquorum-go/releases/latest/download/ipquorum-darwin-arm64
chmod +x ipquorum-darwin-arm64
sudo mv ipquorum-darwin-arm64 /usr/local/bin/ipquorum
```

### Using Go Install
```bash
go install github.com/olemyk/ipquorum-go@latest
```

---

## Usage Examples

### Interactive Password (Most Secure)
```bash
ipquorum \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-prompt \
  --download --insecure
```

### Password File (Automation)
```bash
echo 'password' > ~/.ipquorum_pass
chmod 400 ~/.ipquorum_pass

ipquorum \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-file ~/.ipquorum_pass \
  --download --insecure
```

### Create Quorum App + Download
```bash
ipquorum \
  --api-endpoint 10.33.7.80 \
  --mkquorumapp --partnersystem svc_cluster02 \
  --ip6=false --partnerip6=false --nometadata=false \
  --download --insecure \
  --user superuser --pass-prompt
```

### Debug Mode
```bash
ipquorum --debug \
  --api-endpoint 10.33.7.80 \
  --user superuser --pass-prompt \
  --download --insecure
```

---

## Performance Benchmarks

| Operation | Python | Go |
|-----------|--------|-----|
| Startup | ~100ms | <10ms |
| Authentication | ~1-2s | ~1-2s |
| Download 10MB | ~5-10s | ~5-10s |
| Memory Usage | ~50MB | <20MB |
| Binary Size | N/A | ~8MB |

---

## Future Enhancements

### Planned Features
- [ ] Progress bars for downloads
- [ ] Configuration file support (YAML/JSON)
- [ ] Shell completion (bash, zsh, fish)
- [ ] Concurrent downloads
- [ ] Checksum verification
- [ ] Homebrew formula
- [ ] Docker image

### Under Consideration
- [ ] GUI version (Fyne/Wails)
- [ ] REST API wrapper library
- [ ] Plugin system
- [ ] Kubernetes operator

---

## Contributing

Contributions are welcome! Please ensure:
- Go 1.19+ compatibility
- All tests pass
- Code follows Go conventions
- Documentation is updated

---

## Maintainer
Ole Kristian Myklebust

## License
MIT License