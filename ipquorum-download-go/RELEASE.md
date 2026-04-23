# Release Guide for IPQuorum Download Tool - Go Version

This document describes how to create and publish releases for the Go version of the IPQuorum download tool.

## Overview

The project uses GitHub Actions for automated builds and releases. When you push a version tag, the workflow automatically:

1. Builds binaries for all supported platforms
2. Creates checksums for verification
3. Creates a GitHub Release with all artifacts
4. Builds and pushes a Docker image (optional)

## Supported Platforms

The automated build creates binaries for:

- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

## Release Process

### 1. Update Version Information

Before creating a release, update the version in relevant files:

```bash
# Update CHANGELOG.md
vi CHANGELOG.md
# Add new version section with changes

# Update version in main.go (if not using build-time injection)
# The version is typically injected during build, so this is optional
```

### 2. Commit Changes

```bash
git add CHANGELOG.md
git commit -m "Prepare release v1.0.1"
git push origin main
```

### 3. Create and Push Tag

```bash
# Create annotated tag
git tag -a v1.0.1 -m "Release v1.0.1"

# Push tag to trigger release workflow
git push origin v1.0.1
```

### 4. Monitor Workflow

1. Go to GitHub Actions tab
2. Watch the "Build and Release Go Binary" workflow
3. Verify all jobs complete successfully

### 5. Verify Release

1. Go to GitHub Releases page
2. Verify the new release is created
3. Check that all binaries are attached
4. Verify checksums are present
5. Test download and execution of binaries

## Manual Release (if needed)

If you need to create a release manually:

### Build All Binaries

```bash
cd ipquorum-download-go

# Create dist directory
mkdir -p dist

# Build for all platforms
make build-all

# Or manually:
VERSION=1.0.1
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE"

# Linux amd64
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o dist/ipquorum-download-go-linux-amd64 main.go

# Linux arm64
GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o dist/ipquorum-download-go-linux-arm64 main.go

# macOS amd64
GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o dist/ipquorum-download-go-darwin-amd64 main.go

# macOS arm64
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o dist/ipquorum-download-go-darwin-arm64 main.go

# Windows amd64
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o dist/ipquorum-download-go-windows-amd64.exe main.go
```

### Create Checksums

```bash
cd dist
for file in ipquorum-download-go-*; do
    sha256sum "$file" > "$file.sha256"
done
```

### Create GitHub Release

1. Go to GitHub repository
2. Click "Releases" → "Draft a new release"
3. Choose tag: v1.0.1
4. Release title: v1.0.1
5. Add release notes from CHANGELOG.md
6. Upload all binaries and checksums from dist/
7. Publish release

## Docker Image

### Automated Build

The workflow automatically builds and pushes Docker images when a tag is pushed.

Images are pushed to: `ghcr.io/olemyk/ibm-storage-virtualize-ipquorum-service/ipquorum-download-go`

Tags created:
- `latest` (for main branch releases)
- `v1.0.1` (specific version)
- `v1.0` (major.minor)
- `v1` (major)

### Manual Docker Build

```bash
cd ipquorum-download-go

# Build image
docker build -t ipquorum-download-go:v1.0.1 .

# Test image
docker run --rm ipquorum-download-go:v1.0.1 --help

# Tag for registry
docker tag ipquorum-download-go:v1.0.1 ghcr.io/olemyk/ibm-storage-virtualize-ipquorum-service/ipquorum-download-go:v1.0.1
docker tag ipquorum-download-go:v1.0.1 ghcr.io/olemyk/ibm-storage-virtualize-ipquorum-service/ipquorum-download-go:latest

# Push to registry
docker push ghcr.io/olemyk/ibm-storage-virtualize-ipquorum-service/ipquorum-download-go:v1.0.1
docker push ghcr.io/olemyk/ibm-storage-virtualize-ipquorum-service/ipquorum-download-go:latest
```

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v2.0.0): Incompatible API changes
- **MINOR** version (v1.1.0): New functionality, backwards compatible
- **PATCH** version (v1.0.1): Bug fixes, backwards compatible

## Pre-release Versions

For pre-release versions, use suffixes:

- `v1.0.0-alpha.1` - Alpha release
- `v1.0.0-beta.1` - Beta release
- `v1.0.0-rc.1` - Release candidate

Mark as "pre-release" in GitHub when creating the release.

## Hotfix Process

For urgent fixes:

1. Create hotfix branch from tag
2. Make fix and test
3. Update CHANGELOG.md
4. Create new patch version tag
5. Push tag to trigger release

```bash
# Create hotfix branch
git checkout -b hotfix/v1.0.2 v1.0.1

# Make fixes
git add .
git commit -m "Fix critical bug"

# Create tag
git tag -a v1.0.2 -m "Hotfix v1.0.2"

# Push
git push origin hotfix/v1.0.2
git push origin v1.0.2

# Merge back to main
git checkout main
git merge hotfix/v1.0.2
git push origin main
```

## Rollback

If a release has issues:

1. Delete the tag locally and remotely:
   ```bash
   git tag -d v1.0.1
   git push origin :refs/tags/v1.0.1
   ```

2. Delete the GitHub Release

3. Fix issues and create new release

## Testing Releases

Before announcing a release:

1. **Download and test each binary**:
   ```bash
   # Linux
   curl -LO https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/v1.0.1/ipquorum-download-go-linux-amd64
   chmod +x ipquorum-download-go-linux-amd64
   ./ipquorum-download-go-linux-amd64 --help
   ```

2. **Verify checksums**:
   ```bash
   sha256sum -c ipquorum-download-go-linux-amd64.sha256
   ```

3. **Test functionality**:
   ```bash
   ./ipquorum-download-go-linux-amd64 \
     --api-endpoint test.example.com \
     --user test \
     --pass test \
     --download
   ```

4. **Test Docker image**:
   ```bash
   docker pull ghcr.io/olemyk/ibm-storage-virtualize-ipquorum-service/ipquorum-download-go:v1.0.1
   docker run --rm ghcr.io/olemyk/ibm-storage-virtualize-ipquorum-service/ipquorum-download-go:v1.0.1 --help
   ```

## Announcement

After successful release:

1. Update main README.md with new version
2. Announce in relevant channels
3. Update documentation if needed

## Troubleshooting

### Workflow Fails

1. Check GitHub Actions logs
2. Verify Go version compatibility
3. Check for build errors
4. Ensure all dependencies are available

### Binary Doesn't Work

1. Verify correct GOOS/GOARCH combination
2. Check ldflags are correct
3. Test on target platform
4. Verify no CGO dependencies

### Docker Build Fails

1. Check Dockerfile syntax
2. Verify base image availability
3. Test build locally first
4. Check registry permissions

## Checklist

Before creating a release:

- [ ] All tests pass
- [ ] CHANGELOG.md updated
- [ ] Version number follows semver
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
- [ ] Tested on target platforms
- [ ] Tag created and pushed
- [ ] Workflow completed successfully
- [ ] Binaries tested
- [ ] Docker image tested
- [ ] Release notes clear and complete

---

## Maintainer

Ole Kristian Myklebust

## License

MIT License