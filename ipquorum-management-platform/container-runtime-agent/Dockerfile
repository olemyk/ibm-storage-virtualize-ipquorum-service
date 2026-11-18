# Build stage
FROM docker.io/library/golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
    -o /build/agent \
    ./cmd/agent

# Runtime stage
FROM docker.io/library/alpine:3.20

# Install runtime dependencies (wget for healthcheck, no CLI tools needed)
RUN apk add --no-cache ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1000 agent && \
    adduser -D -u 1000 -G agent agent

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/agent /app/agent

# Copy example config
COPY config.example.yaml /app/config.example.yaml

# Create directories for logs and data
RUN mkdir -p /app/logs /app/data && \
    chown -R agent:agent /app

# Switch to non-root user
USER agent

# Expose ports (8081 for API, 9090 for metrics)
EXPOSE 8081 9090

# Health check (agent runs on port 8081, use GET not HEAD)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 -O /dev/null http://localhost:8081/health || exit 1

# Set entrypoint
ENTRYPOINT ["/app/agent"]

# Default command (can be overridden)
CMD ["--config", "/app/config.yaml"]

# Made with Bob
