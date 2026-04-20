# Release Instructions for ipquorum-download-go

Quick guide for creating releases of the Go download tool binaries.

---

## 🚀 Quick Release Process

### 1. Update Version Information

```bash
cd ipquorum-download-go

# Update CHANGELOG.md
vi CHANGELOG.md
# Add new version section with changes
```

### 2. Commit Changes

```bash
git add CHANGELOG.md
git commit -m "Prepare ipquorum-download-go release v1.0.2"
git push origin main
```

### 3. Create and Push Tag

```bash
# Create annotated tag for Go binary
git tag -a go-v1.0.2 -m "Release ipquorum-download-go v1.0.2"

# Push tag to trigger automated build
git push origin go-v1.0.2
```

### 4. Monitor Workflow

1. Go to: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/actions
2. Watch the "Build and Release Go Binary" workflow
3. Verify all jobs complete successfully (usually 5-10 minutes)

### 5. Verify Release

1. Go to: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases
2. Find release tagged `go-v1.0.2`
3. Verify all binaries are present:
   - `ipquorum-download-go-linux-amd64`
   - `ipquorum-download-go-linux-arm64`
   - `ipquorum-download-go-darwin-amd64`
   - `ipquorum-download-go-darwin-arm64`
   - `ipquorum-download-go-windows-amd64.exe`
   - SHA256 checksums for each

---

## 📦 What Gets Built

The automated workflow builds:

### Binaries
- **Linux amd64** - Most common Linux servers
- **Linux arm64** - ARM-based Linux (Raspberry Pi, AWS Graviton)
- **macOS amd64** - Intel Macs
- **macOS arm64** - Apple Silicon Macs (M1/M2/M3)
- **Windows amd64** - Windows 64-bit

### Checksums
- SHA256 checksum for each binary (`.sha256` files)

### Docker Image (Optional)
- `ghcr.io/olemyk/ibm-storage-virtualize-ipquorum-service/ipquorum-download-go:latest`
- `ghcr.io/olemyk/ibm-storage-virtualize-ipquorum-service/ipquorum-download-go:v1.0.2`

---

## 📥 Download URLs

After release, binaries are available at:

```
https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/go-v1.0.2/ipquorum-download-go-linux-amd64
https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/go-v1.0.2/ipquorum-download-go-linux-arm64
https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/go-v1.0.2/ipquorum-download-go-darwin-amd64
https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/go-v1.0.2/ipquorum-download-go-darwin-arm64
https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/go-v1.0.2/ipquorum-download-go-windows-amd64.exe
```

### Latest Release (Always Current)

```bash
# Linux amd64
curl -LO https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest/download/ipquorum-download-go-linux-amd64

# macOS arm64 (Apple Silicon)
curl -LO https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/latest/download/ipquorum-download-go-darwin-arm64
```

---

## 🔍 Version Numbering

Use [Semantic Versioning](https://semver.org/) with `go-` prefix:

- `go-v1.0.0` - Major release (breaking changes)
- `go-v1.1.0` - Minor release (new features)
- `go-v1.0.1` - Patch release (bug fixes)

### Why `go-` Prefix?

The prefix distinguishes Go binary releases from the main IP Quorum Service releases:
- `v2.0.5` - Full IP Quorum Service package (systemd + scripts + Go binary)
- `go-v1.0.2` - Just the Go download tool binary

---

## 🧪 Testing Before Release

### Local Build Test

```bash
cd ipquorum-download-go

# Build for current platform
make build

# Test binary
./ipquorum-download-go --help
./ipquorum-download-go version
```

### Cross-Platform Build Test

```bash
# Build for all platforms
make build-all

# Verify all binaries created
ls -lh dist/
```

### Functional Test

```bash
# Test download functionality (requires valid credentials)
./ipquorum-download-go \
  --api-endpoint YOUR_TEST_ENDPOINT \
  --user YOUR_TEST_USER \
  --pass-prompt \
  --download --insecure
```

---

## 📝 Release Checklist

Before creating a release:

- [ ] All tests pass (`make test`)
- [ ] CHANGELOG.md updated with new version
- [ ] Version number follows semver
- [ ] Local build successful
- [ ] Cross-platform builds successful
- [ ] Functional test passed
- [ ] Tag created with `go-` prefix
- [ ] Tag pushed to GitHub
- [ ] Workflow completed successfully
- [ ] All binaries present in release
- [ ] Checksums verified
- [ ] Docker image built (if applicable)
- [ ] Release notes clear and complete

---

## 🔄 Update Main Service Release

After releasing the Go binary, update the main service package:

```bash
# The main service release script will automatically include
# the latest Go binaries from GitHub releases

cd ..  # Back to repository root
./scripts/create-release.sh 2.0.6

# This will:
# 1. Download latest Go binaries from GitHub
# 2. Package them in ipquorum-downloader/ directory
# 3. Create complete service package
```

---

## 🐛 Troubleshooting

### Workflow Fails

**Check logs:**
1. Go to Actions tab
2. Click failed workflow
3. Review error messages

**Common issues:**
- Go version mismatch - Update `go-version` in workflow
- Build errors - Fix code and create new tag
- Permission errors - Check repository settings

### Binaries Don't Work

**Test locally first:**
```bash
cd ipquorum-download-go
make build-all
./dist/ipquorum-download-go-linux-amd64 --help
```

**Check:**
- Correct GOOS/GOARCH combination
- No CGO dependencies
- Ldflags are correct

### Release Not Created

**Verify:**
- Tag was pushed: `git ls-remote --tags origin`
- Tag matches pattern `go-v*.*.*`
- Workflow completed successfully
- You have write permissions

---

## 📞 Getting Help

If you encounter issues:

1. Check this guide
2. Review [GITHUB-ACTIONS-GUIDE.md](GITHUB-ACTIONS-GUIDE.md)
3. Check GitHub Actions logs
4. Review workflow file: `.github/workflows/release-go-binary.yml`
5. Open an issue on GitHub

---

## 🎯 Example: Complete Release

```bash
# 1. Update changelog
cd ipquorum-download-go
vi CHANGELOG.md
# Add: ## [1.0.2] - 2026-04-20
#      ### Fixed
#      - Fixed password file permission check

# 2. Commit
git add CHANGELOG.md
git commit -m "Prepare ipquorum-download-go release v1.0.2"
git push origin main

# 3. Create and push tag
git tag -a go-v1.0.2 -m "Release ipquorum-download-go v1.0.2"
git push origin go-v1.0.2

# 4. Wait for workflow (5-10 minutes)
# Monitor at: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/actions

# 5. Verify release
# Check: https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/tag/go-v1.0.2

# 6. Test download
curl -LO https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/go-v1.0.2/ipquorum-download-go-linux-amd64
chmod +x ipquorum-download-go-linux-amd64
./ipquorum-download-go-linux-amd64 --help

# 7. Update main service package (optional)
cd ..
./scripts/create-release.sh 2.0.6
```

---

## 👤 Maintainer

Ole Kristian Myklebust

---

## 📄 License

MIT License