# Release Process

This document describes how to create and publish releases for the GitHub Reverse Proxy.

## Prerequisites

1. **GitHub CLI (gh)** - Required for creating releases
   ```bash
   # Install on macOS
   brew install gh

   # Install on Linux
   sudo apt install gh  # Debian/Ubuntu
   sudo yum install gh  # RHEL/CentOS

   # Or download from: https://cli.github.com/
   ```

2. **Authentication** - Authenticate with GitHub
   ```bash
   gh auth login
   ```

3. **Permissions** - You need write access to the repository

## Quick Release

### Option 1: Using the release script (Recommended)

```bash
# Create and publish release v1.0.0
./scripts/release.sh v1.0.0
```

This will:
1. Create a git tag (if it doesn't exist)
2. Build binaries for all platforms
3. Create archives (tar.gz for Unix, zip for Windows)
4. Generate SHA256 checksums
5. Create GitHub release with release notes
6. Upload all artifacts

### Option 2: Using Makefile

```bash
# Create GitHub release
make github-release VERSION=v1.0.0
```

## Manual Release Process

If you prefer to do it step by step:

### Step 1: Update Version

Decide on the version number following [Semantic Versioning](https://semver.org/):
- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Step 2: Create Git Tag

```bash
# Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Push tag to remote
git push origin v1.0.0
```

### Step 3: Build Binaries

```bash
# Build for all platforms
./scripts/build-all.sh

# Or use Makefile
make build-all-platforms
```

This creates binaries in the `dist/` directory for:
- Linux (amd64, arm64, 386)
- macOS (amd64, arm64)
- Windows (amd64, arm64)
- FreeBSD (amd64)

### Step 4: Create GitHub Release

#### Option A: Using the script
```bash
# The script will build and upload everything
./scripts/release.sh v1.0.0
```

#### Option B: Using gh CLI manually
```bash
# Create release and upload files
gh release create v1.0.0 \
    --title "Release v1.0.0" \
    --notes "Release notes here" \
    dist/*
```

#### Option C: Using GitHub Web UI
1. Go to https://github.com/LZUOSS/gh-proxy/releases/new
2. Choose tag: v1.0.0
3. Fill in release title and description
4. Upload files from `dist/` directory
5. Click "Publish release"

## Release Checklist

Before creating a release:

- [ ] All tests pass (`make test`)
- [ ] Code is properly formatted (`make fmt`)
- [ ] Linters pass (`make lint`)
- [ ] Documentation is updated
- [ ] CHANGELOG is updated (if you have one)
- [ ] Version number follows semver
- [ ] All changes are committed and pushed

## Versioning Strategy

### Version Format
```
vMAJOR.MINOR.PATCH[-PRERELEASE][+BUILDMETADATA]
```

### Examples
- `v1.0.0` - First stable release
- `v1.1.0` - Added new features
- `v1.1.1` - Bug fixes
- `v2.0.0` - Breaking changes
- `v1.0.0-beta.1` - Pre-release
- `v1.0.0+20240101` - Build metadata

### When to Bump

**MAJOR version** when you make incompatible changes:
- Change configuration file format
- Remove or rename API endpoints
- Change command-line flags
- Change default behavior significantly

**MINOR version** when you add functionality in a backward-compatible manner:
- Add new features
- Add new configuration options
- Add new endpoints
- Deprecate features (but don't remove)

**PATCH version** when you make backward-compatible bug fixes:
- Fix bugs
- Update dependencies
- Improve performance
- Update documentation

## Build Artifacts

Each release includes:

### Binaries
- `gh-proxy-{version}-linux-amd64.tar.gz`
- `gh-proxy-{version}-linux-arm64.tar.gz`
- `gh-proxy-{version}-linux-386.tar.gz`
- `gh-proxy-{version}-darwin-amd64.tar.gz`
- `gh-proxy-{version}-darwin-arm64.tar.gz`
- `gh-proxy-{version}-windows-amd64.zip`
- `gh-proxy-{version}-windows-arm64.zip`
- `gh-proxy-{version}-freebsd-amd64.tar.gz`

### Checksums
- `SHA256SUMS` - SHA256 checksums for all binaries

## Automated Releases (CI/CD)

You can automate releases using GitHub Actions. Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Build binaries
        run: ./scripts/build-all.sh

      - name: Create release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          gh release create ${{ github.ref_name }} \
            --title "Release ${{ github.ref_name }}" \
            --generate-notes \
            dist/*
```

## Release Notes Template

```markdown
## GitHub Reverse Proxy v1.0.0

### 🚀 New Features
- Feature 1
- Feature 2

### 🐛 Bug Fixes
- Fix 1
- Fix 2

### ⚡ Improvements
- Improvement 1
- Improvement 2

### 📝 Documentation
- Doc update 1
- Doc update 2

### 🔧 Maintenance
- Dependency updates
- Code cleanup

### Installation

See the [README](https://github.com/LZUOSS/gh-proxy#installation) for installation instructions.

### Checksums

See `SHA256SUMS` for file verification.
```

## Troubleshooting

### "gh: command not found"
Install GitHub CLI from https://cli.github.com/

### "permission denied"
Make sure scripts are executable:
```bash
chmod +x scripts/*.sh
```

### "failed to create release: tag already exists"
Delete the tag and recreate:
```bash
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0
```

### "build failed for platform X"
Check Go cross-compilation support:
```bash
go tool dist list | grep platform
```

## Post-Release

After creating a release:

1. **Announce** - Post announcement (Twitter, blog, etc.)
2. **Update documentation** - Update installation instructions
3. **Monitor** - Watch for issues reported by users
4. **Plan next release** - Create milestone for next version

## Emergency Hotfix Release

For critical bugs in production:

1. Create hotfix branch from the release tag
   ```bash
   git checkout -b hotfix/v1.0.1 v1.0.0
   ```

2. Fix the bug and commit

3. Create new patch release
   ```bash
   git tag -a v1.0.1 -m "Hotfix: critical bug fix"
   git push origin v1.0.1
   ./scripts/release.sh v1.0.1
   ```

4. Merge hotfix back to main
   ```bash
   git checkout main
   git merge hotfix/v1.0.1
   git push origin main
   ```
