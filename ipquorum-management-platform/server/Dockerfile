# Build stage
FROM docker.io/library/golang:1.26-alpine AS builder

WORKDIR /build

# Install build dependencies (including gcc for CGO/SQLite)
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

RUN CGO_ENABLED=1 go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
    -o /build/server \
    ./cmd/server

# Runtime stage
FROM docker.io/library/alpine:3.20

# Update packages and install runtime dependencies
RUN apk upgrade --no-cache && \
    apk add --no-cache ca-certificates tzdata wget sqlite-libs && \
    rm -rf /var/cache/apk/*

# Create non-root user with no login shell for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser -s /sbin/nologin appuser

WORKDIR /app

# Copy binary with restricted permissions
COPY --from=builder --chown=appuser:appuser /build/server /app/server

# Copy migrations with restricted permissions
COPY --chown=appuser:appuser migrations /app/migrations

# Set strict permissions
RUN chmod 500 /app/server && \
    chmod -R 400 /app/migrations

USER appuser

EXPOSE 8080

# Health check (server runs HTTP on port 8080, use GET not HEAD)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 -O /dev/null http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/server"]

# Made with Bob
