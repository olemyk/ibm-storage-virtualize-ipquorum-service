# Installation Guide - Go Version

Complete installation instructions for the IPQuorum Go tool.

---

## Prerequisites

- **Go 1.19 or higher** (Go 1.21+ recommended)
- **Make** (optional, for convenience)
- **Git** (for cloning the repository)

---

## Quick Install

### Option 1: Pre-built Binaries (Recommended)

Download the latest release for your platform:

#### Linux (amd64)
```bash
curl -LO https://github.com/olemyk/ipquorum-go/releases/latest/download/ipquorum-linux-amd64
chmod +x ipquorum-linux-amd64
sudo mv ipquorum-linux-amd64 /usr/local/bin/ipquorum
```

#### Linux (arm64)
```bash
curl -LO https://github.com/olemyk/ipquorum-go/releases/latest/download/ipquorum-linux-arm64
chmod +x ipquorum-linux-arm64
sudo mv ipquorum-linux-arm64 /usr/local/bin/ipquorum
```

#### macOS (Intel)
```bash
curl -LO https://github.com/olemyk/ipquorum-go/releases/latest/download/ipquorum-darwin-amd64
chmod +x ipquorum-darwin-amd64
sudo mv ipquorum-darwin-amd64 /usr/local/bin/ipquorum
```

#### macOS (Apple Silicon)
```bash
curl -LO https://github.com/olemyk/ipquorum-go/releases/latest/download/ipquorum-darwin-arm64
chmod +x ipquorum-darwin-arm64
sudo mv ipquorum-darwin-arm64 /usr/local/bin/ipquorum
```

#### Windows (PowerShell)
```powershell
Invoke-WebRequest -Uri "https://github.com/olemyk/ipquorum-go/releases/latest/download/ipquorum-windows-amd64.exe" -OutFile "ipquorum.exe"
# Add to PATH or move to desired location
```

### Option 2: Using Go Install

```bash
go install github.com/olemyk/ipquorum-go@latest
```

The binary will be installed to `$GOPATH/bin/ipquorum` (usually `~/go/bin/ipquorum`).

Make sure `$GOPATH/bin` is in your PATH:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

---

## Building from Source

### Step 1: Clone the Repository

```bash
git clone https://github.com/olemyk/ipquorum-go.git
cd ipquorum-go
```

### Step 2: Download Dependencies

```bash
go mod download
```

Or let Go download them automatically during build:
```bash
go mod tidy
```

### Step 3: Build

#### Using Make (Recommended)

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Install to $GOPATH/bin
make install
```

#### Manual Build

```bash
# Build for current platform
go build -o ipquorum main.go

# Build with version information
go build -ldflags "-s -w -X main.version=1.0.0" -o ipquorum main.go
```

### Step 4: Verify Installation

```bash
./ipquorum version
./ipquorum --help
```

---

## Cross-Compilation

Build for multiple platforms from any platform:

### All Platforms at Once

```bash
make build-all
```

This creates binaries in the `dist/` directory:
- `ipquorum-linux-amd64`
- `ipquorum-linux-arm64`
- `ipquorum-darwin-amd64`
- `ipquorum-darwin-arm64`
- `ipquorum-windows-amd64.exe`

### Individual Platforms

#### Linux (amd64)
```bash
GOOS=linux GOARCH=amd64 go build -o ipquorum-linux-amd64 main.go
```

#### Linux (arm64)
```bash
GOOS=linux GOARCH=arm64 go build -o ipquorum-linux-arm64 main.go
```

#### macOS (Intel)
```bash
GOOS=darwin GOARCH=amd64 go build -o ipquorum-darwin-amd64 main.go
```

#### macOS (Apple Silicon)
```bash
GOOS=darwin GOARCH=arm64 go build -o ipquorum-darwin-arm64 main.go
```

#### Windows (amd64)
```bash
GOOS=windows GOARCH=amd64 go build -o ipquorum-windows-amd64.exe main.go
```

---

## Optimized Builds

### Minimal Binary Size

```bash
# Strip debug symbols and reduce binary size
go build -ldflags="-s -w" -o ipquorum main.go
```

### With Version Information

```bash
VERSION=1.0.0
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -ldflags="-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" -o ipquorum main.go
```

### Using UPX Compression (Optional)

```bash
# Build
go build -ldflags="-s -w" -o ipquorum main.go

# Compress with UPX (requires UPX to be installed)
upx --best --lzma ipquorum
```

---

## Docker Build

### Using Dockerfile

Create a `Dockerfile`:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ipquorum main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/ipquorum .

ENTRYPOINT ["./ipquorum"]
```

Build and run:

```bash
# Build image
docker build -t ipquorum:latest .

# Run
docker run --rm ipquorum:latest --help

# Download JAR
docker run --rm -v $(pwd):/output ipquorum:latest \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-prompt \
  --download --insecure \
  --output /output/ip_quorum.jar
```

---

## Platform-Specific Instructions

### Linux

#### Ubuntu/Debian

```bash
# Install Go if not already installed
sudo apt-get update
sudo apt-get install golang-go

# Clone and build
git clone https://github.com/olemyk/ipquorum-go.git
cd ipquorum-go
make build

# Install
sudo mv ipquorum /usr/local/bin/
```

#### RHEL/CentOS/Fedora

```bash
# Install Go if not already installed
sudo dnf install golang

# Clone and build
git clone https://github.com/olemyk/ipquorum-go.git
cd ipquorum-go
make build

# Install
sudo mv ipquorum /usr/local/bin/
```

### macOS

#### Using Homebrew

```bash
# Install Go if not already installed
brew install go

# Clone and build
git clone https://github.com/olemyk/ipquorum-go.git
cd ipquorum-go
make build

# Install
sudo mv ipquorum /usr/local/bin/
```

### Windows

#### Using PowerShell

```powershell
# Install Go from https://golang.org/dl/

# Clone and build
git clone https://github.com/olemyk/ipquorum-go.git
cd ipquorum-go
go build -o ipquorum.exe main.go

# Add to PATH or move to desired location
Move-Item ipquorum.exe C:\Windows\System32\
```

---

## Verification

After installation, verify the tool works:

```bash
# Check version
ipquorum version

# Show help
ipquorum --help

# Test connection (will fail without credentials, but verifies binary works)
ipquorum --api-endpoint test.example.com --user test --pass test --download
```

---

## Troubleshooting

### Issue: "command not found: ipquorum"

**Solution**: Add the binary location to your PATH

```bash
# For current session
export PATH=$PATH:/usr/local/bin

# Permanently (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc
```

### Issue: "permission denied"

**Solution**: Make the binary executable

```bash
chmod +x ipquorum
```

### Issue: Go version too old

**Solution**: Update Go

```bash
# Linux
sudo rm -rf /usr/local/go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# macOS
brew upgrade go

# Windows
# Download and install from https://golang.org/dl/
```

### Issue: Build fails with "cannot find package"

**Solution**: Download dependencies

```bash
go mod download
go mod tidy
```

### Issue: Cross-compilation fails

**Solution**: Ensure you have the correct Go version and try again

```bash
go version  # Should be 1.19 or higher
go clean -cache
make build-all
```

---

## Uninstallation

### Remove Binary

```bash
# If installed to /usr/local/bin
sudo rm /usr/local/bin/ipquorum

# If installed via go install
rm $(go env GOPATH)/bin/ipquorum
```

### Remove Source

```bash
rm -rf ipquorum-go/
```

---

## Next Steps

- Read [README.md](README.md) for usage instructions
- Check [SECURITY-GUIDE.md](SECURITY-GUIDE.md) for password best practices
- Review [CHANGELOG.md](CHANGELOG.md) for version history

---

## Support

For issues or questions:
- Check the troubleshooting section above
- Review the documentation
- Open an issue on GitHub

---

## Maintainer

Ole Kristian Myklebust

## License

MIT License