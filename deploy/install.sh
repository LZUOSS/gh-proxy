#!/bin/bash
# GitHub Reverse Proxy Installation Script
# This script installs the GitHub Reverse Proxy service on Linux systems

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/opt/github-proxy"
CONFIG_DIR="/etc/github-proxy"
CACHE_DIR="/var/cache/github-proxy"
LOG_DIR="/var/log/github-proxy"
SERVICE_USER="github-proxy"
SERVICE_FILE="/etc/systemd/system/github-proxy.service"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    exit 1
fi

echo -e "${GREEN}GitHub Reverse Proxy Installation${NC}"
echo "=================================="
echo ""

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        BINARY_ARCH="amd64"
        ;;
    aarch64)
        BINARY_ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "${YELLOW}Detected architecture: $ARCH ($BINARY_ARCH)${NC}"

# Check if binary exists in build directory
BINARY_PATH="./build/github-proxy"
if [ ! -f "$BINARY_PATH" ]; then
    echo -e "${RED}Error: Binary not found at $BINARY_PATH${NC}"
    echo "Please run 'make build' first"
    exit 1
fi

# Create service user
echo -e "${YELLOW}Creating service user...${NC}"
if ! id "$SERVICE_USER" &>/dev/null; then
    useradd -r -s /bin/false "$SERVICE_USER"
    echo -e "${GREEN}Created user: $SERVICE_USER${NC}"
else
    echo -e "${YELLOW}User $SERVICE_USER already exists${NC}"
fi

# Create directories
echo -e "${YELLOW}Creating directories...${NC}"
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$CACHE_DIR"
mkdir -p "$LOG_DIR"

# Copy binary
echo -e "${YELLOW}Installing binary...${NC}"
cp "$BINARY_PATH" "$INSTALL_DIR/github-proxy"
chmod +x "$INSTALL_DIR/github-proxy"
echo -e "${GREEN}Binary installed to: $INSTALL_DIR/github-proxy${NC}"

# Copy configuration
echo -e "${YELLOW}Installing configuration...${NC}"
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    cp ./configs/config.yaml "$CONFIG_DIR/config.yaml"
    # Update cache path in config
    sed -i "s|disk_path: /var/cache/github-proxy|disk_path: $CACHE_DIR|g" "$CONFIG_DIR/config.yaml"
    echo -e "${GREEN}Configuration installed to: $CONFIG_DIR/config.yaml${NC}"
else
    echo -e "${YELLOW}Configuration already exists at $CONFIG_DIR/config.yaml${NC}"
    echo -e "${YELLOW}Creating backup: $CONFIG_DIR/config.yaml.backup${NC}"
    cp "$CONFIG_DIR/config.yaml" "$CONFIG_DIR/config.yaml.backup"
    cp ./configs/config.yaml "$CONFIG_DIR/config.yaml.new"
    echo -e "${YELLOW}New configuration saved as: $CONFIG_DIR/config.yaml.new${NC}"
fi

# Set permissions
echo -e "${YELLOW}Setting permissions...${NC}"
chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"
chown -R "$SERVICE_USER:$SERVICE_USER" "$CACHE_DIR"
chown -R "$SERVICE_USER:$SERVICE_USER" "$LOG_DIR"
chmod 755 "$INSTALL_DIR"
chmod 755 "$CACHE_DIR"
chmod 755 "$LOG_DIR"
chmod 644 "$CONFIG_DIR/config.yaml"

# Install systemd service
echo -e "${YELLOW}Installing systemd service...${NC}"
cp ./deploy/github-proxy.service "$SERVICE_FILE"
systemctl daemon-reload
echo -e "${GREEN}Service installed${NC}"

# Ask to enable and start service
echo ""
read -p "Do you want to enable and start the service now? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    systemctl enable github-proxy
    systemctl start github-proxy
    echo -e "${GREEN}Service enabled and started${NC}"
    echo ""
    echo -e "${YELLOW}Checking service status...${NC}"
    sleep 2
    systemctl status github-proxy --no-pager || true
fi

# Print summary
echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo "=================================="
echo ""
echo "Binary location:    $INSTALL_DIR/github-proxy"
echo "Configuration:      $CONFIG_DIR/config.yaml"
echo "Cache directory:    $CACHE_DIR"
echo "Log directory:      $LOG_DIR"
echo "Service user:       $SERVICE_USER"
echo ""
echo "Service commands:"
echo "  Start:   systemctl start github-proxy"
echo "  Stop:    systemctl stop github-proxy"
echo "  Restart: systemctl restart github-proxy"
echo "  Status:  systemctl status github-proxy"
echo "  Logs:    journalctl -u github-proxy -f"
echo ""
echo "Default ports:"
echo "  HTTP:    8080"
echo "  SSH:     2222"
echo "  Metrics: 9090"
echo ""
echo "Test the installation:"
echo "  curl http://localhost:8080/"
echo "  curl http://localhost:9090/metrics"
echo ""

# Check if firewall is running
if command -v ufw &> /dev/null && ufw status | grep -q "Status: active"; then
    echo -e "${YELLOW}Note: UFW firewall is active. You may need to allow ports:${NC}"
    echo "  sudo ufw allow 8080/tcp"
    echo "  sudo ufw allow 2222/tcp"
    echo "  sudo ufw allow 9090/tcp"
    echo ""
fi

echo -e "${GREEN}Installation successful!${NC}"
