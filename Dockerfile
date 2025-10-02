# Build stage
FROM golang:1.21-alpine AS builder

# Set build arguments
ARG VERSION=3.16.0
ARG BUILD_DATE
ARG VCS_REF

# Install git and ca-certificates (needed for downloading dependencies)
RUN apk add --no-cache git ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildDate=${BUILD_DATE} -X main.gitCommit=${VCS_REF}" \
    -a -installsuffix cgo \
    -o litmuschaos-mcp-server .

# Final stage - minimal runtime image
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary
COPY --from=builder /build/litmuschaos-mcp-server /usr/local/bin/litmuschaos-mcp-server

# Use non-root user
USER appuser

# Set environment variables
ENV CHAOS_CENTER_ENDPOINT=""
ENV LITMUS_PROJECT_ID=""
ENV LITMUS_ACCESS_TOKEN=""
ENV DEFAULT_INFRA_ID=""
ENV DEFAULT_ENVIRONMENT_ID="production"

# Labels for metadata
LABEL maintainer="LitmusChaos MCP Server" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.name="litmuschaos-mcp-server" \
      org.label-schema.version="${VERSION}" \
      org.label-schema.description="Model Context Protocol server for LitmusChaos 3.x" \
      org.label-schema.build-date="${BUILD_DATE}" \
      org.label-schema.vcs-ref="${VCS_REF}" \
      org.label-schema.vendor="LitmusChaos" \
      org.opencontainers.image.title="LitmusChaos MCP Server" \
      org.opencontainers.image.description="Model Context Protocol server for LitmusChaos 3.x chaos engineering platform" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.vendor="LitmusChaos" 

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/litmuschaos-mcp-server", "--health-check"] || exit 1

# Entry point
ENTRYPOINT ["/usr/local/bin/litmuschaos-mcp-server"]

# Default command
CMD []
