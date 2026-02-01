# IPQuorum Download Tool - Go Version

A high-performance Go implementation of the IBM Storage Virtualize IP Quorum download tool. Download IPQuorum JAR files and create Quorum Apps via REST API with a single, self-contained binary.

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](https://github.com)

---

## 🚀 Why Go Version?

### Key Advantages

| Feature | Bash | Python | **Go** |
|---------|------|--------|--------|
| **Single Binary** | ❌ | ❌ | ✅ **No dependencies** |
| **Cross-Platform** | Limited | ✅ | ✅ **Native binaries** |
| **Performance** | Good | Good | ✅ **Excellent** |
| **Startup Time** | Fast | ~100ms | ✅ **<10ms** |
| **Memory Usage** | Low | ~50MB | ✅ **<20MB** |
| **Type Safety** | ❌ | ✅ | ✅ **Compile-time** |
| **Concurrency** | Limited | Good | ✅ **Built-in** |

### Perfect For

- ✅ **DevOps/CI/CD**: Single binary, no runtime dependencies
- ✅ **Containers**: Minimal Docker images (scratch/alpine)
- ✅ **Air-Gapped Systems**: No package installation needed
- ✅ **Multi-Platform**: One codebase for all platforms
- ✅ **Enterprise**: Production-ready, type-safe, performant

---

## 📦 Installation

### Pre-built Binaries (Recommended)

Download the latest release for your platform:

```bash
# Linux (amd64)
curl -LO https://github.com/user/ipquorum-go/releases/latest/download/ipquorum-linux-amd64
chmod +x ipquorum-linux-amd64
sudo mv ipquorum-linux-amd64 /usr/local/bin/ipquorum

# macOS (Apple Silicon)
curl -LO https://github.com/user/ipquorum-go/releases/latest/download/ipquorum-darwin-arm64
chmod +x ipquorum-darwin-arm64
sudo mv ipquorum-darwin-arm64 /usr/local/bin/ipquorum

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/user/ipquorum-go/releases/latest/download/ipquorum-windows-amd64.exe" -OutFile "ipquorum.exe"
```

### Using Go Install

```bash
go install github.com/user/ipquorum-go@latest
```

### Build from Source

```bash
git clone https://github.com/user/ipquorum-go.git
cd ipquorum-go
make build

# Or manually
go build -o ipquorum main.go
```

See [INSTALL.md](INSTALL.md) for detailed installation instructions.

---

## 🎯 Quick Start

### Interactive Use (Most Secure)

```bash
ipquorum \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-prompt \
  --download --insecure
```

### Automated Use (Password File)

```bash
# Create secure password file
echo 'password' > ~/.ipquorum_pass
chmod 400 ~/.ipquorum_pass

# Run with password file
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

---

## 📖 Usage

```
Usage: ipquorum [options]

General Options:
  --api-endpoint <host>        API endpoint IP/hostname (REQUIRED)
  --user <username>            Auth username (env: VIRTUALIZE_USERNAME)
  --output <file>              Output jar filename (default: ip_quorum.jar)
  --debug                      Enable debug logging
  --mkquorumapp                Enable mkquorumapp call (default: true)
  --no-mkquorumapp             Disable mkquorumapp call
  --download                   Enable jar download (default: true)
  --no-download                Disable jar download
  --insecure                   Use insecure TLS (default)
  --secure                     Use strict TLS verification

Password Options (choose ONE):
  --pass-prompt                Prompt for password interactively (MOST SECURE)
  --pass-file <file>           Read password from file (SECURE for automation)
  --pass <password>            Password in command line (INSECURE - testing only)
                               (env: VIRTUALIZE_PASSWORD)

mkquorumapp Options:
  --ip6[=true|false]           Set IPv6 flag (default: false)
  --nometadata[=true|false]    Set nometadata flag (default: false)
  --partnersystem <name>       Set Partnersystem - Remote System in PBHA (MANDATORY)
  --partnerip6[=true|false]    Set partner IPv6 flag (default: false)

Environment Variables:
  API_ENDPOINT                 API endpoint hostname/IP
  VIRTUALIZE_USERNAME          Authentication username
  VIRTUALIZE_PASSWORD          Authentication password
  IPQ_OUTPUT_FILE              Output filename

Examples:
  # Interactive password (most secure)
  ipquorum --api-endpoint 10.33.7.80 --user superuser --pass-prompt --download

  # Password file (automation)
  ipquorum --api-endpoint 10.33.7.80 --user superuser --pass-file ~/.pass --download

  # Environment variables
  export API_ENDPOINT=10.33.7.80
  export VIRTUALIZE_USERNAME=superuser
  ipquorum --download --pass-prompt

  # Debug mode
  ipquorum --debug --api-endpoint 10.33.7.80 --user superuser --pass-prompt --download
```

---

## ✨ Features

### 🔒 Security
- **Multiple password input methods**: Interactive prompt, file, environment variable
- **Password masking**: Never logs passwords in plain text
- **File permission checking**: Warns if password files are too permissive
- **Fail-fast authentication**: Immediate exit on wrong credentials (401/403)

### 🔄 Reliability
- **Smart retry logic**: Exponential backoff for transient errors
- **Rate limiting support**: Automatic handling of HTTP 429
- **Pre-flight validation**: Endpoint reachability check before operations
- **Configurable retries**: Default 8 attempts with 3-second base delay

### 📊 Observability
- **Structured logging**: Clear, timestamped log messages
- **Debug mode**: Detailed troubleshooting information
- **Progress tracking**: Real-time operation status
- **Error reporting**: Meaningful error messages with exit codes

### 🎯 Performance
- **Fast startup**: <10ms cold start
- **Low memory**: <20MB RAM usage
- **Efficient**: Compiled binary, no interpreter overhead
- **Concurrent**: Built-in support for parallel operations (future)

### 🔧 Developer Experience
- **Type-safe**: Compile-time type checking
- **Well-documented**: Comprehensive godoc comments
- **Testable**: Unit and integration tests
- **Maintainable**: Clean, idiomatic Go code

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    IPQuorum CLI                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Command-Line Interface (Cobra)                   │  │
│  │  • Argument parsing                               │  │
│  │  │  • Flag validation                              │  │
│  │  • Help text                                      │  │
│  └───────────────────┬───────────────────────────────┘  │
│                      │                                   │
│                      ▼                                   │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Configuration & Validation                       │  │
│  │  • Config struct                                  │  │
│  │  • Input validation                               │  │
│  │  • Password handling                              │  │
│  └───────────────────┬───────────────────────────────┘  │
│                      │                                   │
│                      ▼                                   │
│  ┌───────────────────────────────────────────────────┐  │
│  │  IPQuorum Client                                  │  │
│  │  ┌─────────────────────────────────────────────┐  │  │
│  │  │ HTTP Client with Retry Logic                │  │  │
│  │  │ • Pre-flight check                          │  │  │
│  │  │ • Authentication (with fail-fast)           │  │  │
│  │  │ • mkquorumapp API call                      │  │  │
│  │  │ • Download JAR file                         │  │  │
│  │  └─────────────────────────────────────────────┘  │  │
│  └───────────────────┬───────────────────────────────┘  │
│                      │                                   │
│                      ▼                                   │
│  ┌───────────────────────────────────────────────────┐  │
│  │  IBM Storage Virtualize REST API                 │  │
│  │  • /rest/v1/auth                                  │  │
│  │  • /rest/v1/mkquorumapp                           │  │
│  │  • /rest/v1/download                              │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## 🔍 Error Handling

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (API, authentication) |
| 2 | Validation error (configuration) |
| 3 | Network error (connectivity) |
| 130 | User cancelled (Ctrl+C) |

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
Authentication failed: Insufficient permissions (HTTP 403)
```

---

## 🐳 Docker Usage

### Using Pre-built Image

```bash
docker run --rm \
  -v $(pwd):/output \
  ghcr.io/user/ipquorum:latest \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-prompt \
  --download --insecure \
  --output /output/ip_quorum.jar
```

### Build Your Own

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -ldflags="-s -w" -o ipquorum main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/ipquorum /usr/local/bin/
ENTRYPOINT ["ipquorum"]
```

---

## 🧪 Development

### Prerequisites

- Go 1.19 or higher
- Make (optional, for convenience)

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run linter
make lint
```

### Testing

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Coverage
go test -cover ./...
```

---

## 📚 Documentation

- **[INSTALL.md](INSTALL.md)** - Detailed installation instructions
- **[SECURITY-GUIDE.md](SECURITY-GUIDE.md)** - Password security best practices
- **[CHANGELOG.md](CHANGELOG.md)** - Version history and changes
- **[PLAN.md](PLAN.md)** - Implementation plan and architecture

### IBM Documentation
- [IPQuorum Info](https://www.ibm.com/support/pages/ibm-storage-virtualize-ip-quorum-application-requirements-1)
- [IP quorum application](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=quorum-ip-application)
- [Storage Virtualize RESTful API](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=interface-storage-virtualize-restful-api)

---

## 🤝 Contributing

Contributions are welcome! Please ensure:
- Go 1.19+ compatibility
- All tests pass
- Code follows Go conventions
- Documentation is updated

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

---

## 👤 Maintainer

Ole Kristian Myklebust

---

## 🔗 Related Projects

- **[Bash Version](../ipquorum-download/bash-version/)** - Original bash implementation
- **[Python Version](../ipquorum-download/python-version/)** - Python implementation with enhanced features

---

## ⭐ Star History

If you find this tool useful, please consider giving it a star on GitHub!

---

**Made with ❤️ and Go**