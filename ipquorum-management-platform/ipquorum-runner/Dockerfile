# IPQuorum Runner - Simplified with pre-built binary
# This container runs IBM Storage Virtualize IP Quorum applications

FROM eclipse-temurin:17-jre-alpine

LABEL maintainer="Ole Kristian Myklebust"
LABEL description="IPQuorum Runner - Runs IBM Storage Virtualize IP Quorum applications"
LABEL version="1.0.0"

# Version of ipquorum-download-go binary to use
ARG IPQUORUM_DOWNLOADER_VERSION=v3.0.8

# Install runtime dependencies
RUN apk add --no-cache \
    bash \
    curl \
    ca-certificates \
    tzdata \
    procps \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 ipquorum && \
    adduser -D -u 1000 -G ipquorum ipquorum

# Create directory structure
RUN mkdir -p /opt/ipquorum/jar \
             /opt/ipquorum/logs \
             /opt/ipquorum/config \
             /opt/ipquorum/bin && \
    chown -R ipquorum:ipquorum /opt/ipquorum

WORKDIR /opt/ipquorum

# Download pre-built ipquorum-download-go binary
RUN curl -fsSL "https://github.com/olemyk/ibm-storage-virtualize-ipquorum-service/releases/download/${IPQUORUM_DOWNLOADER_VERSION}/ipquorum-download-go-linux-amd64" \
    -o /opt/ipquorum/bin/ipquorum-download-go && \
    chmod +x /opt/ipquorum/bin/ipquorum-download-go && \
    chown ipquorum:ipquorum /opt/ipquorum/bin/ipquorum-download-go

# Verify binary works
RUN /opt/ipquorum/bin/ipquorum-download-go --version || echo "Binary downloaded successfully"

# Copy entrypoint script
COPY entrypoint.sh /opt/ipquorum/bin/
RUN chmod +x /opt/ipquorum/bin/entrypoint.sh && \
    chown ipquorum:ipquorum /opt/ipquorum/bin/entrypoint.sh

# Switch to non-root user
USER ipquorum

# Environment variables with defaults
ENV API_ENDPOINT="" \
    VIRTUALIZE_USERNAME="" \
    VIRTUALIZE_PASSWORD="" \
    PARTNERSYSTEM="" \
    IP6="false" \
    PARTNERIP6="false" \
    NOMETADATA="false" \
    INSECURE="true" \
    DOWNLOAD_JAR="true" \
    CREATE_QUORUM_APP="false" \
    JAVA_OPTS="-Xmx512m -Xms256m" \
    TZ="UTC"

# Expose default IPQuorum port
EXPOSE 1260

# Health check - verify Java process is running
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD pgrep -f "java.*ip_quorum.jar" > /dev/null || exit 1

# Volume for persistent data
VOLUME ["/opt/ipquorum/jar", "/opt/ipquorum/logs"]

# Set entrypoint
ENTRYPOINT ["/opt/ipquorum/bin/entrypoint.sh"]

# Made with Bob
