# GitHub Reverse Proxy

A high-performance reverse proxy server for GitHub that provides HTTP and SSH access to GitHub resources with built-in caching, rate limiting, authentication, and SOCKS5/HTTP proxy support.

## Features

- **HTTP Proxy**: Access GitHub releases, raw files, archives, gists, and API endpoints
- **SSH Proxy**: Git operations via SSH with authentication support
- **Caching**: Hybrid memory + disk caching with ARC eviction policy
- **Rate Limiting**: IP-based or token-based rate limiting with configurable limits
- **Authentication**: Token-based or basic authentication support
- **Proxy Support**: Route traffic through SOCKS5 or HTTP proxies
- **Security**: SSRF protection, request validation, and security headers
- **Metrics**: Prometheus metrics for monitoring
- **Graceful Shutdown**: Clean shutdown with configurable timeout

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [HTTP Proxy](#http-proxy)
  - [SSH Proxy](#ssh-proxy)
  - [Authentication](#authentication)
- [API Endpoints](#api-endpoints)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Monitoring](#monitoring)
- [License](#license)

## Installation

### From Source

**Prerequisites:**
- Go 1.24 or higher
- Make (optional, for using Makefile)

**Build:**

```bash
# Clone the repository
git clone https://github.com/kexi/github-reverse-proxy.git
cd github-reverse-proxy

# Install dependencies
make install

# Build the binary
make build

# The binary will be available at ./build/github-proxy
```

### Using Docker

```bash
# Build Docker image
make docker-build

# Run with Docker
make docker-run
```

### Pre-built Binaries

Download pre-built binaries from the [releases page](https://github.com/kexi/github-reverse-proxy/releases).

```bash
# Linux AMD64
wget https://github.com/kexi/github-reverse-proxy/releases/download/v1.0.0/github-proxy-linux-amd64.tar.gz
tar -xzf github-proxy-linux-amd64.tar.gz
chmod +x github-proxy
sudo mv github-proxy /usr/local/bin/
```

## Configuration

The application uses a YAML configuration file. By default, it looks for `./configs/config.yaml`.

### Configuration File

Create a configuration file at `./configs/config.yaml`:

```yaml
server:
  http_port: 8080
  read_timeout: 30s
  write_timeout: 300s
  idle_timeout: 120s

proxy:
  enabled: true
  type: socks5
  address: 127.0.0.1:1080

cache:
  enabled: true
  type: hybrid
  max_memory_size: 104857600  # 100MB
  max_disk_size: 10737418240  # 10GB
  disk_path: /var/cache/github-proxy

ratelimit:
  enabled: true
  requests_per_second: 100
  burst: 200

auth:
  enabled: false
  allow_anonymous: true

metrics:
  enabled: true
  port: 9090
```

See [configs/config.yaml](configs/config.yaml) for a complete configuration example with all available options.

### Environment Variables

Configuration values can be overridden using environment variables with the `GITHUB_PROXY_` prefix:

```bash
export GITHUB_PROXY_SERVER_HTTP_PORT=8080
export GITHUB_PROXY_PROXY_ENABLED=true
export GITHUB_PROXY_PROXY_ADDRESS=127.0.0.1:1080
export GITHUB_PROXY_AUTH_ENABLED=true
```

### Configuration Options

| Section | Option | Description | Default |
|---------|--------|-------------|---------|
| `server.http_port` | HTTP port | Port for HTTP server | `8080` |
| `server.read_timeout` | Read timeout | Maximum duration for reading requests | `30s` |
| `server.write_timeout` | Write timeout | Maximum duration for writing responses | `300s` |
| `proxy.enabled` | Enable proxy | Route GitHub requests through proxy | `false` |
| `proxy.type` | Proxy type | Type of proxy: `socks5` or `http` | `socks5` |
| `proxy.address` | Proxy address | Proxy server address | `127.0.0.1:1080` |
| `cache.enabled` | Enable caching | Enable response caching | `true` |
| `cache.type` | Cache type | Cache strategy: `memory`, `disk`, or `hybrid` | `hybrid` |
| `ratelimit.enabled` | Enable rate limiting | Enable request rate limiting | `true` |
| `ratelimit.requests_per_second` | Requests per second | Maximum requests per second per IP | `100` |
| `auth.enabled` | Enable authentication | Require authentication for requests | `false` |
| `metrics.enabled` | Enable metrics | Enable Prometheus metrics | `true` |
| `metrics.port` | Metrics port | Port for metrics endpoint | `9090` |

## Usage

### Running the Server

```bash
# Run with default configuration
./build/github-proxy

# Run with custom configuration file
./build/github-proxy -config /path/to/config.yaml

# Using Make
make run
```

### HTTP Proxy

The HTTP proxy provides access to various GitHub resources:

#### Download Release Assets

```bash
# Download a release binary
curl http://localhost:8080/owner/repo/releases/download/v1.0.0/binary.tar.gz -o binary.tar.gz

# With authentication
curl -u username:ghp_token http://localhost:8080/owner/repo/releases/download/v1.0.0/binary.tar.gz -o binary.tar.gz
```

#### Get Raw Files

```bash
# Get a raw file from a repository
curl http://localhost:8080/owner/repo/raw/main/README.md

# Get file from a specific commit
curl http://localhost:8080/owner/repo/raw/abc123/src/main.go
```

#### Download Archives

```bash
# Download tarball archive
curl http://localhost:8080/owner/repo/archive/refs/tags/v1.0.0.tar.gz -o repo-v1.0.0.tar.gz

# Download zipball archive
curl http://localhost:8080/owner/repo/archive/refs/heads/main.zip -o repo-main.zip
```

#### Git Operations

```bash
# Clone via HTTP
git clone http://localhost:8080/owner/repo.git

# Clone with authentication
git clone http://username:ghp_token@localhost:8080/owner/repo.git

# Push to repository
cd repo
git remote set-url origin http://username:ghp_token@localhost:8080/owner/repo.git
git push origin main
```

#### Access Gists

```bash
# Get gist content
curl http://localhost:8080/gist/username/gist-id/raw/file.txt

# Download gist archive
curl http://localhost:8080/gist/username/gist-id/archive.tar.gz -o gist.tar.gz
```

#### GitHub API

```bash
# Access GitHub API endpoints
curl http://localhost:8080/api/user

# With authentication
curl -u username:ghp_token http://localhost:8080/api/user

# List repositories
curl -u username:ghp_token http://localhost:8080/api/user/repos

# Get repository information
curl http://localhost:8080/api/repos/owner/repo
```

### SSH Proxy

The SSH proxy allows Git operations over SSH protocol:

```bash
# Clone repository via SSH
git clone ssh://username:ghp_token@localhost:2222/owner/repo.git

# Configure SSH (add to ~/.ssh/config)
Host github-proxy
    HostName localhost
    Port 2222
    User username

# Clone using SSH config
git clone github-proxy:owner/repo.git

# Push via SSH
cd repo
git remote set-url origin ssh://username:ghp_token@localhost:2222/owner/repo.git
git push origin main
```

#### SSH Authentication

The SSH proxy supports two authentication methods:

1. **Password authentication**: Use your GitHub personal access token as the password
2. **Public key authentication**: Use your SSH key (any key is accepted in the current implementation)

```bash
# Password authentication (token as password)
git clone ssh://username@localhost:2222/owner/repo.git
# Enter your GitHub personal access token when prompted

# Public key authentication
ssh-add ~/.ssh/id_rsa
git clone ssh://username@localhost:2222/owner/repo.git
```

### Authentication

When authentication is enabled, you can authenticate using:

#### Token Header

```bash
curl -H "X-Auth-Token: your-secret-token" http://localhost:8080/api/user
```

#### HTTP Basic Auth

```bash
curl -u username:token http://localhost:8080/api/user
```

#### Git Credentials

```bash
# For HTTP Git operations
git clone http://username:token@localhost:8080/owner/repo.git

# For SSH Git operations
git clone ssh://username:token@localhost:2222/owner/repo.git
```

## API Endpoints

### HTTP Endpoints

| Endpoint | Description | Example |
|----------|-------------|---------|
| `/:owner/:repo/releases/download/:tag/:file` | Download release asset | `/owner/repo/releases/download/v1.0.0/app.tar.gz` |
| `/:owner/:repo/raw/:ref/*path` | Get raw file content | `/owner/repo/raw/main/README.md` |
| `/:owner/:repo/archive/*path` | Download repository archive | `/owner/repo/archive/refs/tags/v1.0.0.tar.gz` |
| `/:owner/:repo.git/*path` | Git HTTP operations | `/owner/repo.git/info/refs` |
| `/gist/:username/:gistid/*path` | Access gist content | `/gist/user/abc123/raw/file.txt` |
| `/api/*path` | GitHub API proxy | `/api/user` |

### Metrics Endpoint

Prometheus metrics are available at:

```
http://localhost:9090/metrics
```

**Available metrics:**
- `github_proxy_requests_total` - Total number of requests
- `github_proxy_request_duration_seconds` - Request duration histogram
- `github_proxy_cache_hits_total` - Cache hit counter
- `github_proxy_cache_misses_total` - Cache miss counter
- `github_proxy_active_connections` - Current active connections
- `github_proxy_bytes_transferred_total` - Total bytes transferred

## Development

### Project Structure

```
github-reverse-proxy/
├── cmd/
│   └── server/          # Main application entry point
│       └── main.go
├── internal/
│   ├── auth/            # Authentication system
│   ├── cache/           # Caching system (ARC + disk)
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP request handlers
│   ├── metrics/         # Prometheus metrics
│   ├── middleware/      # HTTP middleware
│   ├── proxy/           # Proxy client (SOCKS5/HTTP)
│   ├── ratelimit/       # Rate limiting
│   ├── security/        # Security features (SSRF protection)
│   ├── server/          # HTTP server
│   ├── ssh/             # SSH proxy server
│   └── util/            # Utility functions
├── configs/
│   └── config.yaml      # Configuration file
├── Makefile             # Build automation
└── README.md            # This file
```

### Building from Source

```bash
# Install dependencies
make install

# Format code
make fmt

# Run linters
make lint

# Run tests
make test

# Build
make build
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/cache/...
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Run go vet
make vet
```

## Testing

### Unit Tests

```bash
# Run all unit tests
make test

# Run with race detection
go test -race ./...

# Run with coverage
make test-coverage
```

### Integration Tests

```bash
# Start the server
make run

# In another terminal, test endpoints
curl http://localhost:8080/owner/repo/raw/main/README.md
curl http://localhost:9090/metrics
```

### Performance Testing

```bash
# Using Apache Bench
ab -n 1000 -c 10 http://localhost:8080/owner/repo/raw/main/README.md

# Using wrk
wrk -t4 -c100 -d30s http://localhost:8080/owner/repo/raw/main/README.md
```

## Deployment

### Docker

**Dockerfile:**

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o github-proxy ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/github-proxy .
COPY --from=builder /app/configs/config.yaml ./configs/
EXPOSE 8080 2222 9090
CMD ["./github-proxy"]
```

**Build and run:**

```bash
# Build image
docker build -t github-proxy:latest .

# Run container
docker run -d \
  -p 8080:8080 \
  -p 2222:2222 \
  -p 9090:9090 \
  -v /var/cache/github-proxy:/var/cache/github-proxy \
  -e GITHUB_PROXY_PROXY_ENABLED=true \
  -e GITHUB_PROXY_PROXY_ADDRESS=host.docker.internal:1080 \
  --name github-proxy \
  github-proxy:latest
```

### Docker Compose

**docker-compose.yml:**

```yaml
version: '3.8'

services:
  github-proxy:
    build: .
    ports:
      - "8080:8080"
      - "2222:2222"
      - "9090:9090"
    volumes:
      - cache-data:/var/cache/github-proxy
      - ./configs/config.yaml:/root/configs/config.yaml
    environment:
      - GITHUB_PROXY_PROXY_ENABLED=true
      - GITHUB_PROXY_PROXY_ADDRESS=socks5-proxy:1080
    restart: unless-stopped

  socks5-proxy:
    image: serjs/go-socks5-proxy
    ports:
      - "1080:1080"
    restart: unless-stopped

volumes:
  cache-data:
```

### Systemd Service

Create `/etc/systemd/system/github-proxy.service`:

```ini
[Unit]
Description=GitHub Reverse Proxy
After=network.target

[Service]
Type=simple
User=github-proxy
Group=github-proxy
WorkingDirectory=/opt/github-proxy
ExecStart=/opt/github-proxy/github-proxy -config /etc/github-proxy/config.yaml
Restart=on-failure
RestartSec=5s

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/cache/github-proxy

[Install]
WantedBy=multi-user.target
```

**Install and start:**

```bash
# Create user
sudo useradd -r -s /bin/false github-proxy

# Create directories
sudo mkdir -p /opt/github-proxy
sudo mkdir -p /etc/github-proxy
sudo mkdir -p /var/cache/github-proxy

# Copy files
sudo cp build/github-proxy /opt/github-proxy/
sudo cp configs/config.yaml /etc/github-proxy/
sudo chown -R github-proxy:github-proxy /opt/github-proxy
sudo chown -R github-proxy:github-proxy /var/cache/github-proxy

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable github-proxy
sudo systemctl start github-proxy
sudo systemctl status github-proxy
```

### Kubernetes

**deployment.yaml:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: github-proxy
spec:
  replicas: 3
  selector:
    matchLabels:
      app: github-proxy
  template:
    metadata:
      labels:
        app: github-proxy
    spec:
      containers:
      - name: github-proxy
        image: github-proxy:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 2222
          name: ssh
        - containerPort: 9090
          name: metrics
        env:
        - name: GITHUB_PROXY_CACHE_DISK_PATH
          value: /cache
        volumeMounts:
        - name: cache
          mountPath: /cache
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: cache
        persistentVolumeClaim:
          claimName: github-proxy-cache

---
apiVersion: v1
kind: Service
metadata:
  name: github-proxy
spec:
  selector:
    app: github-proxy
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: ssh
    port: 2222
    targetPort: 2222
  - name: metrics
    port: 9090
    targetPort: 9090
  type: LoadBalancer
```

## Monitoring

### Prometheus

**prometheus.yml:**

```yaml
scrape_configs:
  - job_name: 'github-proxy'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 15s
```

### Grafana Dashboard

Import the provided Grafana dashboard or create custom panels using metrics:

- Request rate: `rate(github_proxy_requests_total[5m])`
- Error rate: `rate(github_proxy_requests_total{status=~"5.."}[5m])`
- Cache hit rate: `rate(github_proxy_cache_hits_total[5m]) / (rate(github_proxy_cache_hits_total[5m]) + rate(github_proxy_cache_misses_total[5m]))`
- Request duration (p95): `histogram_quantile(0.95, rate(github_proxy_request_duration_seconds_bucket[5m]))`

## Troubleshooting

### Common Issues

**1. Connection refused on port 8080**
- Check if the server is running: `systemctl status github-proxy`
- Verify the port in configuration matches
- Check firewall rules: `sudo ufw status`

**2. Proxy connection failures**
- Verify SOCKS5/HTTP proxy is running
- Check proxy address in configuration
- Test proxy connectivity: `curl --socks5 127.0.0.1:1080 https://api.github.com`

**3. Cache not working**
- Check disk space: `df -h`
- Verify cache directory permissions
- Check cache configuration in config.yaml

**4. Rate limiting issues**
- Adjust `requests_per_second` in configuration
- Check if rate limiting is enabled
- Review logs for rate limit events

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

- GitHub Issues: https://github.com/kexi/github-reverse-proxy/issues
- Documentation: https://github.com/kexi/github-reverse-proxy/wiki

## Acknowledgments

Built with:
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Viper Configuration](https://github.com/spf13/viper)
- [Prometheus Client](https://github.com/prometheus/client_golang)
- [Zap Logger](https://github.com/uber-go/zap)
- [Go SSH](https://golang.org/x/crypto/ssh)
