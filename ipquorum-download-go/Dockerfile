# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION:-dev} -X main.commit=${COMMIT:-unknown} -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o ipquorum-download-go \
    main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 ipquorum && \
    adduser -D -u 1000 -G ipquorum ipquorum

WORKDIR /home/ipquorum

# Copy binary from builder
COPY --from=builder /app/ipquorum-download-go /usr/local/bin/ipquorum-download-go

# Set ownership
RUN chown -R ipquorum:ipquorum /home/ipquorum

# Switch to non-root user
USER ipquorum

# Set entrypoint
ENTRYPOINT ["ipquorum-download-go"]

# Default command (show help)
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="IPQuorum Download Tool"
LABEL org.opencontainers.image.description="IBM Storage Virtualize IP Quorum download tool - Go version"
LABEL org.opencontainers.image.vendor="Ole Kristian Myklebust"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service"
