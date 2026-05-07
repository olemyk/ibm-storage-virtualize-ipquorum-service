# Changelog

All notable changes to the IP Quorum Management Platform will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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