# Build Fix for Fresh Clone

If you encounter this error after cloning or pulling:
```
no required module provides package github.com/LZUOSS/gh-proxy/internal/cache
```

## Quick Fix (Recommended)

```bash
cd ~/gh-proxy  # or wherever you cloned

# Run the automated fix script
./fix-build.sh
```

## Manual Fix

### Step 1: Ensure you're in the project root

```bash
cd ~/gh-proxy
ls go.mod  # Should exist
```

### Step 2: Clean Go cache

```bash
go clean -modcache
go clean -cache
```

### Step 3: Restore go.sum from git

```bash
git checkout go.sum
```

### Step 4: Rebuild dependencies

```bash
go mod download
go mod verify
go mod tidy
```

### Step 5: Build

```bash
make build
# or
go build -o build/github-proxy ./cmd/server
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
