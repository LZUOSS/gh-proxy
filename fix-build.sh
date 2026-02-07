#!/bin/bash
# Fix build issues after fresh clone

set -e

echo "🔧 Fixing build environment..."

# Check we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found. Are you in the project root?"
    exit 1
fi

# Verify module name
MODULE_NAME=$(head -1 go.mod | awk '{print $2}')
if [ "$MODULE_NAME" != "github.com/LZUOSS/gh-proxy" ]; then
    echo "❌ Error: Module name mismatch. Expected github.com/LZUOSS/gh-proxy, got $MODULE_NAME"
    exit 1
fi

# Clean cache
echo "1. Cleaning Go cache..."
go clean -modcache 2>/dev/null || true
go clean -cache 2>/dev/null || true

# Restore go.sum from git
echo "2. Restoring go.sum from git..."
git checkout go.sum 2>/dev/null || true

# Download dependencies with vendor
echo "3. Downloading dependencies..."
go mod download

# Verify dependencies
echo "4. Verifying dependencies..."
go mod verify

# Tidy modules
echo "5. Tidying modules..."
go mod tidy

# Verify module
echo "6. Verifying module..."
go list -m

# Create build directory
mkdir -p build

# Build
echo "7. Building..."
go build -o build/github-proxy ./cmd/server

# Verify binary
if [ -f "build/github-proxy" ]; then
    echo "✅ Build successful!"
    echo "📦 Binary: build/github-proxy"
    ls -lh build/github-proxy
else
    echo "❌ Build failed!"
    exit 1
fi
