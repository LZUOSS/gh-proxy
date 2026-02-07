#!/bin/bash
# Create GitHub release and upload binaries

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed${NC}"
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${RED}Error: Not authenticated with GitHub CLI${NC}"
    echo "Run: gh auth login"
    exit 1
fi

# Get version from argument or git tag
VERSION=${1:-$(git describe --tags --abbrev=0 2>/dev/null)}

if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: No version specified and no git tags found${NC}"
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

# Ensure version starts with 'v'
if [[ ! "$VERSION" =~ ^v ]]; then
    VERSION="v${VERSION}"
fi

echo -e "${BLUE}Creating release ${VERSION}${NC}"
echo ""

# Check if tag exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo -e "${YELLOW}Tag ${VERSION} already exists${NC}"
else
    echo -e "${YELLOW}Creating tag ${VERSION}...${NC}"
    git tag -a "$VERSION" -m "Release ${VERSION}"
    git push origin "$VERSION"
    echo -e "${GREEN}✓ Tag created and pushed${NC}"
fi

# Build binaries
echo ""
echo -e "${YELLOW}Building binaries...${NC}"
VERSION=$VERSION ./scripts/build-all.sh

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# Check if dist directory exists and has files
DIST_DIR="dist"
if [ ! -d "$DIST_DIR" ] || [ -z "$(ls -A $DIST_DIR)" ]; then
    echo -e "${RED}Error: No build artifacts found in ${DIST_DIR}${NC}"
    exit 1
fi

# Generate release notes
RELEASE_NOTES_FILE=$(mktemp)
cat > "$RELEASE_NOTES_FILE" << EOF
## GitHub Reverse Proxy ${VERSION}

### Features

- Full GitHub URL support (use complete GitHub URLs directly)
- Path-based deployment (deploy on subpaths like \`/ghproxy\`)
- HTTP/HTTPS/SOCKS5 proxy support
- Hybrid memory + disk caching with ARC eviction
- IP-based and token-based rate limiting
- Token-based and basic authentication
- SSRF protection and security headers
- Prometheus metrics
- SSH proxy for Git operations

### Installation

Download the appropriate binary for your platform below.

#### Linux (AMD64)
\`\`\`bash
wget https://github.com/LZUOSS/gh-proxy/releases/download/${VERSION}/gh-proxy-${VERSION}-linux-amd64.tar.gz
tar -xzf gh-proxy-${VERSION}-linux-amd64.tar.gz
chmod +x gh-proxy-${VERSION}-linux-amd64
sudo mv gh-proxy-${VERSION}-linux-amd64 /usr/local/bin/gh-proxy
\`\`\`

#### macOS (ARM64 - Apple Silicon)
\`\`\`bash
wget https://github.com/LZUOSS/gh-proxy/releases/download/${VERSION}/gh-proxy-${VERSION}-darwin-arm64.tar.gz
tar -xzf gh-proxy-${VERSION}-darwin-arm64.tar.gz
chmod +x gh-proxy-${VERSION}-darwin-arm64
sudo mv gh-proxy-${VERSION}-darwin-arm64 /usr/local/bin/gh-proxy
\`\`\`

#### Windows (AMD64)
Download \`gh-proxy-${VERSION}-windows-amd64.zip\` and extract.

### Quick Start

\`\`\`bash
# Run with default config
gh-proxy

# Run with custom config
gh-proxy -config /path/to/config.yaml
\`\`\`

### Configuration

See [README.md](https://github.com/LZUOSS/gh-proxy#configuration) for full configuration options.

### Checksums

See \`SHA256SUMS\` for file checksums.
EOF

# Check if release already exists
if gh release view "$VERSION" &> /dev/null; then
    echo -e "${YELLOW}Release ${VERSION} already exists. Deleting and recreating...${NC}"
    gh release delete "$VERSION" -y
fi

# Create release
echo ""
echo -e "${YELLOW}Creating GitHub release...${NC}"
gh release create "$VERSION" \
    --title "Release ${VERSION}" \
    --notes-file "$RELEASE_NOTES_FILE" \
    ${DIST_DIR}/*

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Failed to create release${NC}"
    rm "$RELEASE_NOTES_FILE"
    exit 1
fi

# Clean up
rm "$RELEASE_NOTES_FILE"

echo ""
echo -e "${GREEN}✓ Release ${VERSION} created successfully!${NC}"
echo ""
echo -e "${BLUE}View release: https://github.com/LZUOSS/gh-proxy/releases/tag/${VERSION}${NC}"
echo ""
echo "Uploaded files:"
ls -lh "$DIST_DIR"
