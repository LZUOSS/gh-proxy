# Build Fix for Fresh Clone

If you encounter this error after cloning:
```
no required module provides package github.com/LZUOSS/gh-proxy/internal/cache
```

## Solution

This happens because Go is trying to fetch internal packages from GitHub. Here's how to fix it:

### Option 1: Clean and Rebuild (Recommended)

```bash
cd ~/gh-proxy  # or wherever you cloned

# Clean Go cache
go clean -modcache

# Remove go.sum if it exists
rm -f go.sum

# Initialize fresh
go mod download
go mod tidy

# Build
make build
```

### Option 2: Verify go.mod

Make sure your `go.mod` starts with:
```
module github.com/LZUOSS/gh-proxy

go 1.24.0
```

### Option 3: Use replace directive (temporary workaround)

Add this to the bottom of your `go.mod`:
```
replace github.com/LZUOSS/gh-proxy => ./
```

Then run:
```bash
go mod tidy
make build
```

### Option 4: Direct build without make

```bash
go build -o build/github-proxy ./cmd/server
```

## Root Cause

The issue occurs when:
1. Go tries to resolve internal packages as external dependencies
2. Module cache has stale entries
3. Fresh clone doesn't have a go.sum file yet

## Verification

After fixing, verify with:
```bash
# Should show the module path
go list -m

# Should build successfully
go build -o build/github-proxy ./cmd/server

# Check the binary
./build/github-proxy -h
```
