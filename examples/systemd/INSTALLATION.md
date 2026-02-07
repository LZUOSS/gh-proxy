# Systemd Service Installation Guide

This guide shows how to install and run gh-proxy as a systemd service.

## Prerequisites

- Linux system with systemd
- Root or sudo access
- gh-proxy binary downloaded

## Quick Installation

### Step 1: Create User and Directories

```bash
# Create dedicated user
sudo useradd -r -s /bin/false gh-proxy

# Create directories
sudo mkdir -p /opt/gh-proxy
sudo mkdir -p /etc/gh-proxy
sudo mkdir -p /var/cache/gh-proxy
sudo mkdir -p /var/log/gh-proxy

# Set ownership
sudo chown -R gh-proxy:gh-proxy /opt/gh-proxy
sudo chown -R gh-proxy:gh-proxy /var/cache/gh-proxy
sudo chown -R gh-proxy:gh-proxy /var/log/gh-proxy
sudo chown -R root:gh-proxy /etc/gh-proxy
sudo chmod 750 /etc/gh-proxy
```

### Step 2: Install Binary

```bash
# Download latest release
wget https://github.com/LZUOSS/gh-proxy/releases/download/v0.2.0/gh-proxy-v0.2.0-linux-amd64.tar.gz

# Extract
tar -xzf gh-proxy-v0.2.0-linux-amd64.tar.gz

# Install binary
sudo mv gh-proxy-v0.2.0-linux-amd64 /usr/local/bin/gh-proxy
sudo chmod 755 /usr/local/bin/gh-proxy

# Verify
gh-proxy --version
```

### Step 3: Create Configuration

```bash
# Copy example config
sudo cp configs/config.yaml /etc/gh-proxy/config.yaml
sudo chown root:gh-proxy /etc/gh-proxy/config.yaml
sudo chmod 640 /etc/gh-proxy/config.yaml

# Edit configuration
sudo nano /etc/gh-proxy/config.yaml
```

### Step 4: Choose and Install Service Unit

Choose one of the service files based on your needs:

#### Option A: Basic Service (Recommended)

```bash
sudo cp examples/systemd/gh-proxy.service /etc/systemd/system/
```

#### Option B: With SOCKS5 Proxy

```bash
# Copy service file
sudo cp examples/systemd/gh-proxy-socks5.service /etc/systemd/system/gh-proxy.service

# Create credentials file
sudo cp examples/systemd/socks5.env.example /etc/gh-proxy/socks5.env
sudo nano /etc/gh-proxy/socks5.env

# Set permissions (IMPORTANT!)
sudo chmod 600 /etc/gh-proxy/socks5.env
sudo chown root:gh-proxy /etc/gh-proxy/socks5.env
```

Edit `/etc/gh-proxy/socks5.env`:
```bash
GITHUB_PROXY_PROXY_USERNAME=your_username
GITHUB_PROXY_PROXY_PASSWORD=your_password
```

#### Option C: Localhost Only (Behind Nginx)

```bash
sudo cp examples/systemd/gh-proxy-localhost.service /etc/systemd/system/gh-proxy.service
```

### Step 5: Enable and Start Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service (start on boot)
sudo systemctl enable gh-proxy

# Start service
sudo systemctl start gh-proxy

# Check status
sudo systemctl status gh-proxy
```

## Service Management

### Basic Commands

```bash
# Start service
sudo systemctl start gh-proxy

# Stop service
sudo systemctl stop gh-proxy

# Restart service
sudo systemctl restart gh-proxy

# Reload configuration (graceful)
sudo systemctl reload gh-proxy

# Check status
sudo systemctl status gh-proxy

# Enable auto-start
sudo systemctl enable gh-proxy

# Disable auto-start
sudo systemctl disable gh-proxy
```

### View Logs

```bash
# View recent logs
sudo journalctl -u gh-proxy -n 100

# Follow logs (live)
sudo journalctl -u gh-proxy -f

# View logs since boot
sudo journalctl -u gh-proxy -b

# View logs for specific time
sudo journalctl -u gh-proxy --since "1 hour ago"

# View application logs (if file logging enabled)
sudo tail -f /var/log/gh-proxy/app.log
```

## Configuration Examples

### Example 1: Basic Setup

**/etc/gh-proxy/config.yaml:**
```yaml
server:
  host: ""
  http_port: 8080

cache:
  enabled: true
  type: hybrid
  disk_path: /var/cache/gh-proxy
```

### Example 2: Behind Nginx (Localhost)

**/etc/gh-proxy/config.yaml:**
```yaml
server:
  host: "127.0.0.1"  # Localhost only
  http_port: 8080
  base_path: /ghproxy

cache:
  enabled: true
  type: hybrid
  disk_path: /var/cache/gh-proxy
```

**/etc/nginx/sites-available/gh-proxy:**
```nginx
server {
    listen 80;
    server_name proxy.example.com;

    location /ghproxy/ {
        proxy_pass http://127.0.0.1:8080/ghproxy/;
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

### Example 3: With SOCKS5 Proxy

**/etc/gh-proxy/config.yaml:**
```yaml
server:
  host: "127.0.0.1"
  http_port: 8080

proxy:
  enabled: true
  type: socks5
  address: 127.0.0.1:1080
  # username/password set via environment file

cache:
  enabled: true
  type: hybrid
  disk_path: /var/cache/gh-proxy
```

**/etc/gh-proxy/socks5.env:**
```bash
GITHUB_PROXY_PROXY_USERNAME=myuser
GITHUB_PROXY_PROXY_PASSWORD=mypassword
```

## Security Hardening

### File Permissions

```bash
# Configuration file
sudo chmod 640 /etc/gh-proxy/config.yaml
sudo chown root:gh-proxy /etc/gh-proxy/config.yaml

# Credentials file (if used)
sudo chmod 600 /etc/gh-proxy/socks5.env
sudo chown root:gh-proxy /etc/gh-proxy/socks5.env

# Binary
sudo chmod 755 /usr/local/bin/gh-proxy
sudo chown root:root /usr/local/bin/gh-proxy

# Data directories
sudo chmod 750 /var/cache/gh-proxy
sudo chown gh-proxy:gh-proxy /var/cache/gh-proxy
```

### Firewall Rules

```bash
# If listening on all interfaces
sudo ufw allow 8080/tcp

# If behind reverse proxy (block direct access)
sudo ufw deny 8080/tcp
sudo ufw allow from 127.0.0.1 to any port 8080
```

### SELinux (RHEL/CentOS)

```bash
# Allow network connections
sudo setsebool -P httpd_can_network_connect 1

# Set context for binary
sudo semanage fcontext -a -t bin_t /usr/local/bin/gh-proxy
sudo restorecon -v /usr/local/bin/gh-proxy

# Set context for cache
sudo semanage fcontext -a -t var_cache_t "/var/cache/gh-proxy(/.*)?"
sudo restorecon -Rv /var/cache/gh-proxy
```

## Monitoring

### Systemd Status

```bash
# Detailed status
systemctl status gh-proxy

# Check if service is active
systemctl is-active gh-proxy

# Check if service is enabled
systemctl is-enabled gh-proxy
```

### Prometheus Metrics

```bash
# If metrics enabled on port 9090
curl http://localhost:9090/metrics

# Add to Prometheus scrape config
# /etc/prometheus/prometheus.yml
scrape_configs:
  - job_name: 'gh-proxy'
    static_configs:
      - targets: ['localhost:9090']
```

### Health Check

```bash
# Simple health check
curl http://localhost:8080/health

# Add to monitoring system
watch -n 10 'curl -s http://localhost:8080/health | jq'
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
sudo journalctl -u gh-proxy -n 50 --no-pager

# Check configuration
sudo /usr/local/bin/gh-proxy -config /etc/gh-proxy/config.yaml

# Check permissions
ls -la /usr/local/bin/gh-proxy
ls -la /etc/gh-proxy/
ls -la /var/cache/gh-proxy/
```

### Permission Denied Errors

```bash
# Fix ownership
sudo chown -R gh-proxy:gh-proxy /var/cache/gh-proxy
sudo chown -R gh-proxy:gh-proxy /var/log/gh-proxy

# Check SELinux (if applicable)
sudo ausearch -m avc -ts recent
```

### High Memory Usage

Edit `/etc/systemd/system/gh-proxy.service`:
```ini
[Service]
MemoryLimit=1G  # Adjust as needed
Environment="GITHUB_PROXY_CACHE_MAX_MEMORY_ENTRIES=500"
```

Then:
```bash
sudo systemctl daemon-reload
sudo systemctl restart gh-proxy
```

### Port Already in Use

```bash
# Check what's using the port
sudo lsof -i :8080
sudo netstat -tulpn | grep 8080

# Change port in config
sudo nano /etc/gh-proxy/config.yaml
# Or use environment variable
echo "Environment=\"GITHUB_PROXY_SERVER_HTTP_PORT=8081\"" | sudo tee -a /etc/systemd/system/gh-proxy.service
```

## Upgrading

```bash
# Download new version
wget https://github.com/LZUOSS/gh-proxy/releases/download/v0.3.0/gh-proxy-v0.3.0-linux-amd64.tar.gz
tar -xzf gh-proxy-v0.3.0-linux-amd64.tar.gz

# Stop service
sudo systemctl stop gh-proxy

# Backup old binary
sudo cp /usr/local/bin/gh-proxy /usr/local/bin/gh-proxy.old

# Install new binary
sudo mv gh-proxy-v0.3.0-linux-amd64 /usr/local/bin/gh-proxy
sudo chmod 755 /usr/local/bin/gh-proxy

# Start service
sudo systemctl start gh-proxy

# Check status
sudo systemctl status gh-proxy
```

## Uninstallation

```bash
# Stop and disable service
sudo systemctl stop gh-proxy
sudo systemctl disable gh-proxy

# Remove service file
sudo rm /etc/systemd/system/gh-proxy.service
sudo systemctl daemon-reload

# Remove binary
sudo rm /usr/local/bin/gh-proxy

# Remove configuration
sudo rm -rf /etc/gh-proxy

# Remove cache (optional)
sudo rm -rf /var/cache/gh-proxy

# Remove logs (optional)
sudo rm -rf /var/log/gh-proxy

# Remove user (optional)
sudo userdel gh-proxy
```

## Additional Resources

- **Project Repository:** https://github.com/LZUOSS/gh-proxy
- **Issue Tracker:** https://github.com/LZUOSS/gh-proxy/issues
- **Documentation:** https://github.com/LZUOSS/gh-proxy#readme
