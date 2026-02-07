# Multi-stage Dockerfile for GitHub Reverse Proxy

# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o github-proxy \
    ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Create directories
RUN mkdir -p /var/cache/github-proxy && \
    chown -R appuser:appuser /var/cache/github-proxy

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/github-proxy .

# Copy configuration
COPY configs/config.yaml ./configs/

# Change ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080 2222 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the application
ENTRYPOINT ["./github-proxy"]
CMD ["-config", "./configs/config.yaml"]
