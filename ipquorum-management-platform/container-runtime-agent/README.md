# IPQuorum Container Runtime Agent

A lightweight, production-ready container runtime agent that provides a unified REST API for managing Docker and Podman containers. Built for the IPQuorum Management Platform.

## Features

- **Unified Runtime Interface**: Single API for both Docker and Podman
- **Auto-Detection**: Automatically detects and selects available container runtime
- **REST API**: Complete HTTP API with 13 endpoints
- **Configuration Management**: YAML files + environment variables
- **Production Ready**: Proper logging, error handling, graceful shutdown
- **Secure**: JWT authentication, API keys, rate limiting, CORS
- **Containerized**: Docker image with health checks
- **Lightweight**: Alpine-based image, ~20MB

## Quick Start

### Prerequisites

- Go 1.25+ (for building from source)
- Docker or Podman installed and running
- (Optional) Docker for containerized deployment

### Installation

#### From Source

```bash
# Clone repository
git clone <repository-url>
cd container-runtime-agent

# Build
make build

# Run
./bin/ipquorum-agent --config config.yaml
```

#### Using Docker

```bash
# Build image
make docker-build

# Run with Docker socket
make docker-run

# Or run with Podman socket
make docker-run-podman
```

### Configuration

Create a `config.yaml` file (see `config.example.yaml` for all options):

```yaml
server:
  host: "0.0.0.0"
  port: 8080

runtime:
  preferred_runtime: "auto"  # auto, docker, or podman
  docker_socket_path: "/var/run/docker.sock"
  podman_socket_path: "/run/podman/podman.sock"

auth:
  enabled: true
  jwt_secret: "your-secret-key"

logging:
  level: "info"
  format: "json"
```

Or use environment variables:

```bash
export IPQUORUM_AGENT_SERVER_PORT=8080
export IPQUORUM_AGENT_RUNTIME_PREFERRED_RUNTIME=docker
export IPQUORUM_AGENT_AUTH_JWT_SECRET=your-secret-key
./bin/ipquorum-agent
```

## API Endpoints

### Health & Status

- `GET /health` - Health check
- `GET /ready` - Readiness check

### Runtime Information

- `GET /api/v1/runtime/info` - Get runtime information
- `GET /api/v1/runtime/version` - Get runtime version
- `GET /api/v1/runtime/ping` - Ping runtime

### Container Management

- `POST /api/v1/containers` - Create container
- `GET /api/v1/containers` - List containers
- `GET /api/v1/containers/:id` - Inspect container
- `POST /api/v1/containers/:id/start` - Start container
- `POST /api/v1/containers/:id/stop` - Stop container
- `POST /api/v1/containers/:id/restart` - Restart container
- `DELETE /api/v1/containers/:id` - Remove container
- `GET /api/v1/containers/:id/logs` - Get container logs
- `GET /api/v1/containers/:id/stats` - Get container stats

## API Examples

### Create Container

```bash
curl -X POST http://localhost:8080/api/v1/containers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-container",
    "image": "nginx:latest",
    "ports": [
      {"host_port": 8080, "container_port": 80}
    ]
  }'
```

### List Containers

```bash
curl http://localhost:8080/api/v1/containers?all=true
```

### Start Container

```bash
curl -X POST http://localhost:8080/api/v1/containers/my-container/start
```

### Get Logs

```bash
curl http://localhost:8080/api/v1/containers/my-container/logs?tail=100
```

## Development

### Build

```bash
make build          # Build binary
make build-linux    # Build for Linux
make docker-build   # Build Docker image
```

### Test

```bash
make test           # Run tests
make test-coverage  # Run tests with coverage
```

### Code Quality

```bash
make fmt            # Format code
make vet            # Run go vet
make lint           # Run linter (requires golangci-lint)
```

### Run

```bash
make run            # Build and run
make run-config     # Run with config file
```

## Architecture

```
container-runtime-agent/
├── cmd/agent/              # Main entry point
│   └── main.go
├── internal/api/           # HTTP API layer
│   ├── server.go          # API server
│   └── middleware.go      # Middleware (logging, CORS, auth)
├── pkg/
│   ├── config/            # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   └── runtime/           # Container runtime abstraction
│       ├── types.go       # Interface & types
│       ├── docker.go      # Docker client
│       ├── podman.go      # Podman client
│       └── factory.go     # Runtime factory
├── Dockerfile             # Container image
├── Makefile              # Build automation
├── config.example.yaml   # Example configuration
└── go.mod                # Go dependencies
```

## Configuration Reference

### Server Configuration

- `server.host`: Listen address (default: "0.0.0.0")
- `server.port`: Listen port (default: 8080)
- `server.read_timeout`: Read timeout (default: 15s)
- `server.write_timeout`: Write timeout (default: 15s)
- `server.tls.enabled`: Enable TLS (default: false)

### Runtime Configuration

- `runtime.preferred_runtime`: Preferred runtime (auto/docker/podman)
- `runtime.docker_socket_path`: Docker socket path
- `runtime.podman_socket_path`: Podman socket path
- `runtime.detection_timeout`: Runtime detection timeout
- `runtime.max_containers`: Maximum containers to manage

### Authentication Configuration

- `auth.enabled`: Enable authentication (default: true)
- `auth.jwt_secret`: JWT secret key
- `auth.jwt_expiration`: JWT token expiration (default: 24h)
- `auth.api_keys`: List of API keys
- `auth.basic_auth.enabled`: Enable basic auth

### Logging Configuration

- `logging.level`: Log level (debug/info/warn/error)
- `logging.format`: Log format (json/text)
- `logging.output`: Output destination (stdout/file)
- `logging.file`: Log file path (when output=file)

### Security Configuration

- `security.allowed_origins`: CORS allowed origins
- `security.rate_limiting.enabled`: Enable rate limiting
- `security.rate_limiting.requests_per_minute`: Rate limit
- `security.ip_whitelist`: IP whitelist
- `security.ip_blacklist`: IP blacklist

## Deployment

### Docker Compose

```yaml
version: '3.8'
services:
  ipquorum-agent:
    image: ipquorum-agent:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./config.yaml:/app/config.yaml
    environment:
      - IPQUORUM_AGENT_AUTH_JWT_SECRET=your-secret-key
    restart: unless-stopped
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ipquorum-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ipquorum-agent
  template:
    metadata:
      labels:
        app: ipquorum-agent
    spec:
      containers:
      - name: agent
        image: ipquorum-agent:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        volumeMounts:
        - name: docker-sock
          mountPath: /var/run/docker.sock
        env:
        - name: IPQUORUM_AGENT_AUTH_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: ipquorum-secrets
              key: jwt-secret
      volumes:
      - name: docker-sock
        hostPath:
          path: /var/run/docker.sock
```

## Security Considerations

1. **Never expose the agent directly to the internet** - Use a reverse proxy
2. **Always enable authentication** in production
3. **Use TLS** for production deployments
4. **Restrict socket access** - Mount Docker/Podman socket read-only if possible
5. **Use API keys or JWT** for service-to-service communication
6. **Enable rate limiting** to prevent abuse
7. **Monitor logs** for suspicious activity

## Testing

The agent includes comprehensive test coverage:

### Unit Tests (50+ tests)
```bash
# Run all tests
make test

# Run specific package tests
go test -v ./pkg/runtime/
go test -v ./pkg/metrics/
go test -v ./pkg/auth/

# With coverage
make test-coverage
```

### Integration Tests
```bash
# Run integration tests (requires Docker/Podman)
go test -v ./tests/integration/

# Run specific integration test
go test -v ./tests/integration/ -run TestContainerLifecycle
```

### Performance Tests
```bash
# Run benchmarks
make benchmark

# Run load tests (requires k6)
make load-test
```

### Test Coverage

- **JWT Authentication**: 10/10 tests passing
- **Runtime Factory**: 10/10 tests passing
- **Docker Runtime**: 11 tests passing
- **Podman Runtime**: 20+ tests passing
- **Metrics Package**: 30+ tests passing
- **Integration Tests**: 7 tests passing
- **Performance Benchmarks**: 8 benchmarks

**Total**: 50+ unit tests + 7 integration tests + 8 benchmarks

See detailed test documentation:
- [PHASE24-COMPLETE.md](PHASE24-COMPLETE.md) - JWT, Runtime, Integration, Metrics tests
- [PHASE27-28-COMPLETE.md](PHASE27-28-COMPLETE.md) - Performance & Load tests
- [PHASE29-COMPLETE.md](PHASE29-COMPLETE.md) - Podman & Metrics tests

## Troubleshooting

### Agent won't start

- Check if Docker/Podman is running: `docker ps` or `podman ps`
- Verify socket permissions: `ls -l /var/run/docker.sock`
- Check configuration: `./bin/ipquorum-agent --config config.yaml`

### Cannot connect to runtime

- Verify socket path in configuration
- Check if socket is accessible: `curl --unix-socket /var/run/docker.sock http://localhost/version`
- Try running with `--runtime docker` or `--runtime podman`

### Authentication errors

- Verify JWT secret is set
- Check token expiration
- Ensure API key is in configuration

### Test failures

- Install missing dependencies: `go get github.com/prometheus/client_golang/prometheus/testutil@v1.23.2`
- Ensure Docker/Podman is running for integration tests
- Check test logs: `go test -v ./... 2>&1 | tee test.log`

## Development Status

### Completed Phases

- ✅ **Phase 24**: Testing, Metrics & Monitoring (3,288 lines)
  - JWT Authentication tests (1,015 lines)
  - Runtime tests (1,253 lines)
  - Integration tests (420 lines)
  - Metrics implementation (600 lines)

- ✅ **Phase 25**: API Documentation (2,450 lines)
  - Swagger/OpenAPI specification (850 lines)
  - API documentation (850 lines)
  - Postman collection (750 lines)

- ✅ **Phase 26**: Performance Optimization (1,140 lines)
  - Connection pooling (420 lines)
  - Caching layer (320 lines)
  - Rate limiting (400 lines)

- ✅ **Phase 27-28**: Testing & Production Readiness (1,650 lines)
  - Performance benchmarks (580 lines)
  - Load testing with k6 (470 lines)
  - Documentation (600 lines)

- ✅ **Phase 29**: Additional Testing (1,300 lines)
  - Podman runtime tests (700 lines)
  - Metrics package tests (600 lines)

**Total**: 9,828 lines of production-ready code

### Next Phases (from ROADMAP.md)

- **Phase 30**: Security Enhancements (TLS/HTTPS, RBAC, Audit Logging)
- **Phase 31**: High Availability (Leader Election, State Sync)
- **Phase 34**: Kubernetes Integration (Operator, Helm Charts, CRDs)

## License

[Your License Here]

## Contributing

Contributions are welcome! Please read CONTRIBUTING.md for details.

## Support

For issues and questions, please open an issue on GitHub.