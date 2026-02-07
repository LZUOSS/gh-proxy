#!/bin/bash
# Build binaries for multiple platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version from git tag or use default
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Build information
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}"

# Output directory
OUTPUT_DIR=${OUTPUT_DIR:-"dist"}
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

echo -e "${GREEN}Building gh-proxy ${VERSION}${NC}"
echo "Commit: ${COMMIT}"
echo "Build Time: ${BUILD_TIME}"
echo ""

# Platforms to build for
# Format: "OS/ARCH"
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/386"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
    "freebsd/amd64"
)

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS="/" read -r GOOS GOARCH <<< "$PLATFORM"

    OUTPUT_NAME="gh-proxy-${VERSION}-${GOOS}-${GOARCH}"

    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi

    OUTPUT_PATH="${OUTPUT_DIR}/${OUTPUT_NAME}"

    echo -e "${YELLOW}Building for ${GOOS}/${GOARCH}...${NC}"

    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
        -trimpath \
        -ldflags="${LDFLAGS}" \
        -o "$OUTPUT_PATH" \
        ./cmd/server

    if [ $? -ne 0 ]; then
        echo -e "${RED}✗ Failed to build for ${GOOS}/${GOARCH}${NC}"
        exit 1
    fi

    # Create tarball for non-Windows platforms
    if [ "$GOOS" != "windows" ]; then
        ARCHIVE_NAME="gh-proxy-${VERSION}-${GOOS}-${GOARCH}.tar.gz"
        tar -czf "${OUTPUT_DIR}/${ARCHIVE_NAME}" -C "$OUTPUT_DIR" "$(basename "$OUTPUT_PATH")"
        rm "$OUTPUT_PATH"
        echo -e "${GREEN}✓ Created ${ARCHIVE_NAME}${NC}"
    else
        # Create zip for Windows
        ARCHIVE_NAME="gh-proxy-${VERSION}-${GOOS}-${GOARCH}.zip"
        (cd "$OUTPUT_DIR" && zip -q "$ARCHIVE_NAME" "$(basename "$OUTPUT_PATH")")
        rm "$OUTPUT_PATH"
        echo -e "${GREEN}✓ Created ${ARCHIVE_NAME}${NC}"
    fi
done

echo ""
echo -e "${GREEN}✓ All builds completed successfully!${NC}"
echo ""
echo "Artifacts in ${OUTPUT_DIR}:"
ls -lh "$OUTPUT_DIR"

# Generate checksums
echo ""
echo -e "${YELLOW}Generating checksums...${NC}"
cd "$OUTPUT_DIR"
sha256sum * > SHA256SUMS
cd - > /dev/null

echo -e "${GREEN}✓ Checksums generated: ${OUTPUT_DIR}/SHA256SUMS${NC}"
