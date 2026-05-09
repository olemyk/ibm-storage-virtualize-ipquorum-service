# Changelog

All notable changes to the IP Quorum Management Platform will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [3.0.8] - 2026-05-09

### Added
- **HTTPS Support for Web Dashboard**
  - Added HTTPS listener on port 443 in nginx configuration
  - HTTP (port 80) now redirects to HTTPS automatically
  - Exposed port 3443 for HTTPS access (maps to container port 443)
  - Added TLS certificate volume mount to web container
  - Added HSTS (Strict-Transport-Security) header for enhanced security
  - Updated healthcheck to use HTTPS

### Fixed
- **Deployment Script Color Code Display**
  - Fixed raw ANSI escape sequences showing in output
  - Changed from `cat << EOF` to `echo -e` for proper color rendering
  - Success messages now display with proper green color
  - All colored output now renders correctly

### Changed
- **Updated Access Instructions**
  - Primary access now via HTTPS: https://localhost:3443
  - HTTP access (port 3000) redirects to HTTPS
  - Updated deployment instructions with HTTPS URLs
  - Added note about accepting self-signed certificate warnings

### Security
- Enabled TLS 1.2 and 1.3 protocols
- Added strong cipher configuration
- Enabled SSL session caching
- Added HSTS header with 1-year max-age

## [3.0.7] - 2026-05-09

### Fixed
- **Deployment Script Health Check Protocol Mismatch**
  - Fixed health check in deploy.sh to use HTTPS instead of HTTP
  - Changed `curl -f http://localhost:8443/health` to `curl -fk https://localhost:8443/health`
  - Added `-k` flag to skip certificate verification for self-signed certificates
  - Fixed both `wait_for_services()` and `validate_deployment()` functions
  - Resolves "API server health check failed" error during deployment validation

### Changed
- Updated deployment instructions to clarify API server uses HTTPS with self-signed certificate

## [3.0.6] - 2026-05-09

### Fixed
- **Critical: Database Schema Migration Bug**
  - Added missing `started_at` and `last_health_check` columns to migrations array
  - Migration SQL files existed but were never added to hardcoded migrations in database.go
  - Fixed "no such column: started_at" error on instances page
  - Added indexes for new columns for better query performance

### Technical Details
- Migrations are hardcoded in `server/internal/storage/database.go`, not loaded from SQL files
- Added ALTER TABLE statements for both columns with NULL support
- Added CREATE INDEX statements for performance optimization

## [3.0.5] - 2026-05-09

### Fixed
- **Critical: Container Networking Configuration**
  - Fixed nginx proxy to use container name instead of localhost
  - Changed `proxy_pass https://localhost:8080` to `proxy_pass https://ipquorum-server:8080`
  - Containers in podman-compose pod use bridge networking with DNS resolution
  - Web container can now successfully communicate with API server
  - Login functionality now works correctly

### Technical Details
- Podman-compose creates pods with bridge networking, not shared network namespace
- Containers resolve each other via DNS using container names
- Internal port is 8080 (not the host-mapped 8443)

## [3.0.4] - 2026-05-09

### Fixed
- **Port Configuration in Nginx**
  - Changed nginx proxy from port 8443 to 8080 (internal container port)
  - Server listens on internal port 8080, mapped to host port 8443
  - Still had networking issues due to localhost vs container name

## [3.0.3] - 2026-05-09

### Fixed
- **Critical: JWT_EXPIRATION Configuration Error**
  - Changed JWT_EXPIRATION from "24h" (duration string) to 86400 (seconds as integer)
  - Fixed "cannot parse 'auth.token_expiry' as int" error preventing server startup
  - Updated both docker-compose.prod.yml and .env.example

### Changed
- Fixed GitHub Actions tag format (requires `v*.*.*` pattern, not `*.*.*`)

## [3.0.2] - 2026-05-08

### Fixed
- **Critical: Web Dashboard Login Issue (HTTP 400)**
  - Fixed nginx proxy configuration to use HTTPS port 8443 instead of HTTP port 8080
  - Added `proxy_ssl_verify off` for self-signed certificates
  - Updated all proxy endpoints (/api/, /health, /metrics) to use correct backend URL
  - Web dashboard login now works correctly with admin/admin123 credentials

- **Configuration File Requirement**
  - Made config.yaml file optional for container deployments
  - Added environment variable support as fallback configuration method
  - Updated config.Load() to use viper.AutomaticEnv() for environment variable binding
  - Containers can now run with environment variables only (no config file needed)

- **SELinux Volume Mount Issues (RHEL/CentOS/Fedora)**
  - Added automatic SELinux context setting in deployment script
  - Applied `svirt_sandbox_file_t` context to data, logs, and tls directories
  - Fixed "Permission denied" errors on RHEL 9.4 and similar systems

- **Container Permissions**
  - Fixed database write permissions for container user (UID 1000)
  - Added proper ownership setting in deployment script
  - Resolved "attempt to write a readonly database" errors

- **TLS Certificate Generation**
  - Added automatic self-signed certificate generation in deployment script
  - Certificates include proper SANs (localhost, ipquorum-server, 127.0.0.1)
  - Fixed "TLS certificate not found" startup errors

### Changed
- **Deployment Script Enhancements (deploy.sh)**
  - Updated default version to 3.0.2
  - Added `generate_tls_certificates()` function for automatic cert generation
  - Enhanced `create_directories()` with SELinux context handling
  - Added tls directory to deployment structure
  - Updated default credentials documentation (admin/admin123)
  - Improved error messages and warnings for common issues

- **Container Configuration**
  - Changed default server port from 8443 to 8080 in config defaults
  - Updated default paths to match container volume mounts
  - Removed hardcoded config file requirement

### Infrastructure
- All fixes tested and validated on RHEL 9.4 with Podman
- Production-ready deployment with automatic issue resolution
- Enhanced compatibility with rootless containers and SELinux

## [3.0.1] - 2026-05-07

### Added
- **Container Deployment Automation**: Complete CI/CD and deployment infrastructure
  - GitHub Actions workflow for automated builds and publishing to GitHub Container Registry
  - Multi-platform container builds (linux/amd64, linux/arm64)
  - Automated testing and security scanning with Trivy
  - Semantic versioning from git tags
- **Comprehensive Deployment Documentation** (2,300+ lines)
  - DEPLOYMENT-GUIDE.md: General deployment guide with HA, Kubernetes, monitoring
  - DOCKER-DEPLOYMENT.md: Docker-specific deployment instructions
  - PODMAN-DEPLOYMENT.md: Podman-specific deployment with rootless containers
  - AGENT-DEPLOYMENT-GUIDE.md: Agent deployment and configuration
  - MANAGER-AGENT-CONNECTION-GUIDE.md: Connection setup between manager and agents
- **Automated Deployment Script** (deploy.sh)
  - Auto-detection of Docker or Podman
  - Interactive deployment with colored output
  - Automatic image pulling from GitHub Packages
  - Environment setup with secure secret generation
  - Health validation and backup support
  - Multiple actions: deploy, update, stop, status, logs

### Changed
- Updated docker-compose.prod.yml to use GitHub Container Registry images
  - `ghcr.io/olemyk/ipquorum-manager:v3.0.1`
  - `ghcr.io/olemyk/ipquorum-web:v3.0.1`
- Added proper OCI labels and metadata to container images
- Enhanced version tracking in container configurations

### Infrastructure
- Production-ready deployment with one-command installation
- Automated CI/CD pipeline for continuous delivery
- Security-first approach with rootless containers and secrets management
- Platform-agnostic deployment (Docker and Podman support)

## [3.0.0] - 2026-05-07

### Added
- **Instance Edit Functionality**: Full implementation of instance update endpoint
  - `PUT /api/v1/instances/:id` endpoint now fully functional
  - Supports partial updates (only updates provided fields)
  - Handles all instance fields: API endpoint, credentials, storage system, description, location, partnersystem, IPQuorum name
  - Proper handling of boolean pointer fields (EnableDownload, EnableMkquorumapp, IP6, PartnerIP6, NoMetadata)
  - Secure password handling (only updates if new password provided)
  - Audit logging for all instance modifications
- **Comprehensive .gitignore**: Added project-wide gitignore file
  - Excludes build artifacts, logs, databases, temporary files
  - Excludes internal documentation and development scripts
  - Preserves essential files: source code, migrations, configuration, public documentation
  - Properly configured to track database migration files

### Fixed
- Instance edit returning HTTP 501 "Not Implemented" error
- Container deployment issues after code updates
- Network connectivity issues between frontend and backend containers after restart

### Changed
- Backend container now properly rebuilds with updated code
- Improved container restart process to ensure new binaries are loaded

### Technical Details
- Backend: `server/internal/api/server.go` - Implemented `handleUpdateInstance` function (lines 471-547)
- Frontend: Instance edit form already had full support, now backend matches functionality
- Database: `UpdateInstance` method was already implemented, now properly utilized
- Container: Full rebuild and recreation process ensures code updates are deployed

## [2.0.8] - 2026-05-06

### Added
- Manager uptime display in Settings page
- Last health check timestamp for instances
- Help page with comprehensive documentation
- Timezone handling documentation

### Fixed
- Settings page showing "Uptime: N/A" for manager
- Instance uptime displaying as raw numbers instead of human-readable format
- Last health check always showing "N/A"

### Changed
- Health endpoint now returns manager uptime in seconds
- Frontend formats uptime values for better readability
- Database schema updated with `last_health_check` column

## [2.0.0] - 2026-05-03

### Added
- Initial release of IP Quorum Management Platform
- Manager-Agent architecture
- Web-based dashboard for instance management
- RESTful API for instance operations
- Agent deployment on remote servers
- Health monitoring system
- Metrics and monitoring integration
- User authentication and authorization
- Server registry and management

### Features
- Create, read, update, delete IP Quorum instances
- Remote instance management via agents
- Real-time health checks
- Prometheus metrics integration
- Grafana dashboard support
- Multi-server support
- API key authentication for agents
- JWT-based user authentication

---

## Version History

- **3.0.0** (2026-05-07): Instance edit functionality + comprehensive gitignore
- **2.0.8** (2026-05-06): Uptime display and health check improvements
- **2.0.0** (2026-05-03): Initial platform release

---

*Made with Bob 🤖*
