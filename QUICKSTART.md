# Quick Start Guide

Get GitHub Reverse Proxy up and running in minutes.

## Quick Install

### Option 1: Using Make (Recommended)

```bash
# Clone the repository
git clone https://github.com/kexi/github-reverse-proxy.git
cd github-reverse-proxy

# Install dependencies and build
make install
make build

# Run the server
make run
```

The server will start on:
- HTTP: http://localhost:8080
- SSH: ssh://localhost:2222
- Metrics: http://localhost:9090/metrics

### Option 2: Using Docker

```bash
# Clone the repository
git clone https://github.com/kexi/github-reverse-proxy.git
cd github-reverse-proxy

# Start with Docker Compose
docker-compose up -d

# Check logs
docker-compose logs -f
```

### Option 3: Pre-built Binary

```bash
# Download latest release
wget https://github.com/kexi/github-reverse-proxy/releases/latest/download/github-proxy-linux-amd64.tar.gz

# Extract
tar -xzf github-proxy-linux-amd64.tar.gz

# Run
./github-proxy -config config.yaml
```

## First Steps

### 1. Test HTTP Proxy

Download a file from GitHub:

```bash
curl http://localhost:8080/owner/repo/raw/main/README.md
```

### 2. Test with Git

Clone a repository:

```bash
git clone http://localhost:8080/owner/repo.git
```

### 3. Test SSH Proxy

Clone via SSH:

```bash
git clone ssh://username:token@localhost:2222/owner/repo.git
```

### 4. Check Metrics

View Prometheus metrics:

```bash
curl http://localhost:9090/metrics
```

## Configuration

### Minimal Configuration

Create `configs/config.yaml`:

```yaml
server:
  http_port: 8080

cache:
  enabled: true
  disk_path: ./cache

logging:
  level: info
```

### Enable SOCKS5 Proxy

```yaml
proxy:
  enabled: true
  type: socks5
  address: 127.0.0.1:1080
```

### Enable Authentication

```yaml
auth:
  enabled: true
  tokens:
    - "your-secret-token"
  allow_anonymous: false
```

## Common Use Cases

### Use Case 1: Download GitHub Releases

```bash
# Download a release binary
curl http://localhost:8080/owner/repo/releases/download/v1.0.0/app.tar.gz -o app.tar.gz
```

### Use Case 2: Access Raw Files

```bash
# Get a specific file
curl http://localhost:8080/owner/repo/raw/main/config.json

# Download a script
curl http://localhost:8080/owner/repo/raw/main/install.sh | bash
```

### Use Case 3: Clone Private Repositories

```bash
# HTTP
git clone http://username:ghp_token@localhost:8080/owner/private-repo.git

# SSH
git clone ssh://username:ghp_token@localhost:2222/owner/private-repo.git
```

### Use Case 4: CI/CD Integration

```yaml
# .gitlab-ci.yml
download_dependencies:
  script:
    - curl http://github-proxy:8080/owner/repo/releases/download/v1.0/dep.tar.gz -o dep.tar.gz
    - tar -xzf dep.tar.gz
```

## Environment Variables

Override config with environment variables:

```bash
# Server port
export GITHUB_PROXY_SERVER_HTTP_PORT=8080

# Enable proxy
export GITHUB_PROXY_PROXY_ENABLED=true
export GITHUB_PROXY_PROXY_ADDRESS=127.0.0.1:1080

# Cache settings
export GITHUB_PROXY_CACHE_DISK_PATH=/var/cache/github-proxy

# Run
./github-proxy
```

## Troubleshooting

### Server won't start

Check if ports are in use:
```bash
sudo lsof -i :8080
sudo lsof -i :2222
```

### Can't connect to proxy

Test SOCKS5 proxy connection:
```bash
curl --socks5 127.0.0.1:1080 https://api.github.com
```

### Cache not working

Check cache directory permissions:
```bash
ls -la ./cache
chmod 755 ./cache
```

### High memory usage

Reduce cache size in config:
```yaml
cache:
  max_memory_size: 52428800  # 50MB
  max_disk_size: 1073741824  # 1GB
```

## Next Steps

- Read the full [README.md](README.md) for detailed documentation
- Configure [authentication](README.md#authentication)
- Set up [monitoring](README.md#monitoring)
- Deploy to [production](README.md#deployment)

## Getting Help

- Issues: https://github.com/kexi/github-reverse-proxy/issues
- Documentation: https://github.com/kexi/github-reverse-proxy/wiki
