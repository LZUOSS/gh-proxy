#!/bin/bash
# Fix build issues after fresh clone

set -e

echo "🔧 Fixing build environment..."

# Clean cache
echo "1. Cleaning Go cache..."
go clean -modcache 2>/dev/null || true

# Remove old go.sum
if [ -f go.sum ]; then
    echo "2. Removing old go.sum..."
    rm -f go.sum
fi

# Download dependencies
echo "3. Downloading dependencies..."
go mod download

# Tidy modules
echo "4. Tidying modules..."
go mod tidy

# Verify
echo "5. Verifying module..."
go list -m

# Build
echo "6. Building..."
go build -o build/github-proxy ./cmd/server

echo "✅ Build successful! Binary at: build/github-proxy"
