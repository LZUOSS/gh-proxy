# GitHub Reverse Proxy - Usage Examples

This document provides practical examples of using the GitHub Reverse Proxy with full URL support and path-based deployment.

## Quick Start

### 1. Default Configuration (Root Path, All Interfaces)

**Configuration:**
```yaml
server:
  host: ""  # Listen on all interfaces (0.0.0.0)
  http_port: 8080
  base_path: ""
```

**Examples:**

```bash
# Download a specific file using full GitHub URL
curl http://localhost:8080/https://github.com/ZhiShengYuan/inningbo-go/raw/refs/heads/main/ARCHITECTURE_REFACTORING.md

# Same file using traditional path
curl http://localhost:8080/ZhiShengYuan/inningbo-go/raw/refs/heads/main/ARCHITECTURE_REFACTORING.md

# Download release asset
curl http://localhost:8080/https://github.com/golang/go/releases/download/go1.21.0/go1.21.0.linux-amd64.tar.gz -o go.tar.gz

# Get repository archive
curl http://localhost:8080/https://github.com/owner/repo/archive/refs/tags/v1.0.0.tar.gz -o repo.tar.gz
```

### 1b. Localhost Only (Development/Security)

For development or when running behind a reverse proxy (recommended for security):

**Configuration:**
```yaml
server:
  host: "127.0.0.1"  # Listen on localhost only
  http_port: 8080
  base_path: ""
```

**Examples:**

```bash
# Only accessible from localhost
curl http://localhost:8080/https://github.com/owner/repo/raw/main/README.md

# Not accessible from other machines (more secure)
```

**Use cases:**
- Development environment
- Behind Nginx/Apache reverse proxy
- Running on same machine as other services
- Enhanced security (prevent external access)

### 2. Subpath Deployment

**Configuration:**
```yaml
server:
  host: ""  # All interfaces
  http_port: 8080
  base_path: /ghproxy
```

**Examples:**

```bash
# All requests are under /ghproxy prefix
curl http://localhost:8080/ghproxy/https://github.com/owner/repo/raw/main/README.md

# Traditional path-based under /ghproxy
curl http://localhost:8080/ghproxy/owner/repo/raw/main/README.md

# Works with any GitHub domain
curl http://localhost:8080/ghproxy/https://raw.githubusercontent.com/owner/repo/main/file.txt
curl http://localhost:8080/ghproxy/https://api.github.com/repos/owner/repo
curl http://localhost:8080/ghproxy/https://gist.github.com/user/id/raw/file.txt
```

## URL Format Support

### GitHub.com URLs

All of these work:

```bash
# HTTPS URLs
curl http://localhost:8080/https://github.com/owner/repo/raw/main/file.md

# HTTP URLs (will be upgraded to HTTPS internally)
curl http://localhost:8080/http://github.com/owner/repo/raw/main/file.md

# URLs without scheme
curl http://localhost:8080/github.com/owner/repo/raw/main/file.md
```

### Different GitHub URL Types

#### 1. Raw Files

```bash
# Standard raw URL
http://localhost:8080/https://github.com/owner/repo/raw/main/src/main.go

# With refs/heads/ prefix
http://localhost:8080/https://github.com/owner/repo/raw/refs/heads/main/src/main.go

# With refs/tags/ prefix
http://localhost:8080/https://github.com/owner/repo/raw/refs/tags/v1.0.0/README.md

# blob URLs (auto-converted to raw)
http://localhost:8080/https://github.com/owner/repo/blob/main/README.md

# raw.githubusercontent.com URLs
http://localhost:8080/https://raw.githubusercontent.com/owner/repo/main/README.md
```

#### 2. Releases

```bash
# Release asset download
http://localhost:8080/https://github.com/owner/repo/releases/download/v1.0.0/app-linux-amd64.tar.gz
```

#### 3. Archives

```bash
# Tarball
http://localhost:8080/https://github.com/owner/repo/archive/refs/tags/v1.0.0.tar.gz

# Zipball
http://localhost:8080/https://github.com/owner/repo/archive/refs/heads/main.zip
```

#### 4. Gists

```bash
# Gist raw file
http://localhost:8080/https://gist.github.com/username/abc123/raw/example.js
```

#### 5. API

```bash
# GitHub API
http://localhost:8080/https://api.github.com/repos/owner/repo
http://localhost:8080/https://api.github.com/user/repos
```

## Production Deployment Examples

### Behind Nginx Reverse Proxy

**Nginx config:**
```nginx
server {
    listen 80;
    server_name example.com;

    location /ghproxy/ {
        proxy_pass http://localhost:8080/ghproxy/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Increase timeouts for large files
        proxy_connect_timeout 60s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
    }
}
```

**GitHub Proxy config:**
```yaml
server:
  http_port: 8080
  base_path: /ghproxy
```

**Usage:**
```bash
curl https://example.com/ghproxy/https://github.com/owner/repo/raw/main/file.md
```

### With Authentication and HTTP Proxy

**Config:**
```yaml
server:
  http_port: 8080
  base_path: /api/github

proxy:
  enabled: true
  type: http
  address: corporate-proxy.example.com:8080
  username: proxy_user
  password: proxy_pass

auth:
  enabled: true
  type: token
  tokens:
    - "production-token-123"
    - "ci-token-456"
```

**Usage:**
```bash
# With token in header
curl -H "X-Auth-Token: production-token-123" \
  http://localhost:8080/api/github/https://github.com/owner/repo/raw/main/config.yaml

# With basic auth
curl -u user:production-token-123 \
  http://localhost:8080/api/github/https://github.com/owner/repo/releases/download/v1.0.0/app.tar.gz
```

### Docker Compose with Reverse Proxy

**docker-compose.yml:**
```yaml
version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - github-proxy

  github-proxy:
    build: .
    environment:
      - GITHUB_PROXY_SERVER_BASE_PATH=/ghproxy
      - GITHUB_PROXY_PROXY_ENABLED=true
      - GITHUB_PROXY_PROXY_TYPE=http
      - GITHUB_PROXY_PROXY_ADDRESS=corporate-proxy:8080
    volumes:
      - cache-data:/var/cache/github-proxy

volumes:
  cache-data:
```

## Use Cases

### 1. China/Restricted Network Access

Access GitHub resources when direct access is blocked:

```bash
# Set up with SOCKS5 proxy
export GITHUB_PROXY_PROXY_TYPE=socks5
export GITHUB_PROXY_PROXY_ADDRESS=127.0.0.1:1080

# Download Go binary
curl http://localhost:8080/https://github.com/golang/go/releases/download/go1.21.0/go1.21.0.linux-amd64.tar.gz -o go.tar.gz

# Clone via HTTP
git clone http://localhost:8080/owner/repo.git
```

### 2. Corporate Environment

Access GitHub through corporate proxy:

```bash
# Configure HTTP proxy
export GITHUB_PROXY_PROXY_TYPE=http
export GITHUB_PROXY_PROXY_ADDRESS=proxy.corp.com:8080
export GITHUB_PROXY_PROXY_USERNAME=corp_user
export GITHUB_PROXY_PROXY_PASSWORD=corp_pass

# Access GitHub resources
curl http://localhost:8080/https://github.com/owner/repo/raw/main/file.md
```

### 3. CI/CD Pipeline

Speed up builds with caching:

```yaml
# GitHub Actions example
steps:
  - name: Download dependencies via cache proxy
    run: |
      curl http://github-proxy:8080/https://github.com/owner/tool/releases/download/v1.0.0/tool.tar.gz -o tool.tar.gz
      tar -xzf tool.tar.gz
```

### 4. Download Manager Integration

Use with wget, curl, or download managers:

```bash
# wget
wget http://localhost:8080/https://github.com/owner/repo/releases/download/v1.0.0/large-file.zip

# aria2c (parallel downloads)
aria2c -x 16 http://localhost:8080/https://github.com/owner/repo/releases/download/v1.0.0/large-file.zip

# curl with resume
curl -C - http://localhost:8080/https://github.com/owner/repo/releases/download/v1.0.0/large-file.zip -o file.zip
```

## Environment Variables

All configuration options can be set via environment variables:

```bash
# Server
export GITHUB_PROXY_SERVER_HTTP_PORT=8080
export GITHUB_PROXY_SERVER_BASE_PATH=/ghproxy

# Proxy
export GITHUB_PROXY_PROXY_ENABLED=true
export GITHUB_PROXY_PROXY_TYPE=http
export GITHUB_PROXY_PROXY_ADDRESS=proxy.example.com:8080

# Cache
export GITHUB_PROXY_CACHE_ENABLED=true
export GITHUB_PROXY_CACHE_TYPE=hybrid
export GITHUB_PROXY_CACHE_MAX_MEMORY_ENTRIES=1000

# Auth
export GITHUB_PROXY_AUTH_ENABLED=true
```

## Testing

Test the proxy is working:

```bash
# Health check
curl http://localhost:8080/health

# Test with a small file
curl http://localhost:8080/https://github.com/octocat/Hello-World/raw/master/README

# Test with full URL
curl -v http://localhost:8080/https://raw.githubusercontent.com/octocat/Hello-World/master/README

# Check metrics
curl http://localhost:9090/metrics
```

## Troubleshooting

### URLs not being recognized as GitHub URLs

Make sure the URL includes the domain:
- ✅ `http://localhost:8080/https://github.com/owner/repo/raw/main/file.md`
- ✅ `http://localhost:8080/github.com/owner/repo/raw/main/file.md`
- ❌ `http://localhost:8080/owner/repo/raw/main/file.md` (traditional path-based, still works but different route)

### Base path not working

Make sure base_path is configured:
```yaml
server:
  base_path: /ghproxy  # Don't forget the leading slash
```

And use it in all requests:
- ✅ `http://localhost:8080/ghproxy/https://github.com/...`
- ❌ `http://localhost:8080/https://github.com/...` (will 404 if base_path is set)
