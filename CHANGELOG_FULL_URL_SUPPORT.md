# Changelog: Full URL Support & Path-Based Deployment

## Summary

Added support for:
1. **Full GitHub URL proxying** - Use complete GitHub URLs instead of path-based routing
2. **Path-based deployment** - Deploy on subpaths (e.g., `/ghproxy`)
3. **HTTP/HTTPS proxy support** - Already implemented in previous commit

## Features Added

### 1. Full GitHub URL Support

You can now use complete GitHub URLs in your requests:

**Before (only path-based):**
```bash
curl http://localhost:8080/owner/repo/raw/main/file.md
```

**Now (both work):**
```bash
# Traditional path-based (still works)
curl http://localhost:8080/owner/repo/raw/main/file.md

# Full GitHub URL
curl http://localhost:8080/https://github.com/owner/repo/raw/refs/heads/main/file.md
curl http://localhost:8080/github.com/owner/repo/raw/main/file.md
curl http://localhost:8080/https://raw.githubusercontent.com/owner/repo/main/file.md
```

**Supported URL formats:**
- `https://github.com/...` - Any GitHub.com URL
- `https://raw.githubusercontent.com/...` - Raw content URLs
- `https://api.github.com/...` - API endpoints
- `https://gist.github.com/...` - Gist URLs
- URLs without scheme (e.g., `github.com/...`)

### 2. Path-Based Deployment

Deploy the proxy on a subpath instead of root:

**Configuration:**
```yaml
server:
  base_path: /ghproxy
```

**Usage:**
```bash
# All routes are now under /ghproxy
curl http://example.com/ghproxy/https://github.com/owner/repo/raw/main/file.md
curl http://example.com/ghproxy/owner/repo/raw/main/file.md
```

**Use cases:**
- Deploy behind Nginx/Apache on a subpath
- Run multiple proxies on the same domain
- Better URL organization for corporate environments

### 3. Intelligent URL Routing

The proxy automatically detects and routes URLs:
- Recognizes GitHub URLs vs. traditional paths
- Converts `blob` URLs to `raw` URLs
- Handles `refs/heads/` and `refs/tags/` prefixes correctly
- Supports all GitHub URL patterns

## Files Changed

### New Files
1. **`internal/handler/url.go`** - Full URL parsing and routing handler
2. **`internal/handler/url_test.go`** - Comprehensive tests for URL parsing
3. **`EXAMPLES.md`** - Detailed usage examples
4. **`CHANGELOG_FULL_URL_SUPPORT.md`** - This file

### Modified Files
1. **`internal/config/config.go`**
   - Added `BasePath` field to `ServerConfig`
   - Added default value for `base_path`

2. **`configs/config.yaml`**
   - Added `base_path` configuration option
   - Updated comments for `proxy.type`

3. **`internal/server/http.go`**
   - Added URL handler initialization
   - Implemented base path routing logic
   - Added `isGitHubURL()` helper function
   - Modified route setup to support both traditional and full URL formats

4. **`README.md`**
   - Updated feature list
   - Added full URL support documentation
   - Added configuration table entries
   - Added usage examples

## Configuration Changes

### New Configuration Option

```yaml
server:
  base_path: ""  # Optional: base path for all routes (e.g., "/ghproxy")
```

**Environment variable:**
```bash
export GITHUB_PROXY_SERVER_BASE_PATH=/ghproxy
```

## Examples

### Example 1: Default Setup (Root Path)

```yaml
server:
  base_path: ""
```

```bash
# Works with full URLs
curl http://localhost:8080/https://github.com/ZhiShengYuan/inningbo-go/raw/refs/heads/main/ARCHITECTURE_REFACTORING.md

# Still works with traditional paths
curl http://localhost:8080/ZhiShengYuan/inningbo-go/raw/refs/heads/main/ARCHITECTURE_REFACTORING.md
```

### Example 2: Subpath Deployment

```yaml
server:
  base_path: /ghproxy
```

```bash
# All routes under /ghproxy
curl http://localhost:8080/ghproxy/https://github.com/owner/repo/raw/main/README.md
```

### Example 3: With HTTP Proxy

```yaml
server:
  base_path: /ghproxy

proxy:
  enabled: true
  type: http  # HTTP proxy support
  address: proxy.example.com:8080
```

```bash
# Access GitHub through HTTP proxy and custom path
curl http://localhost:8080/ghproxy/https://github.com/owner/repo/releases/download/v1.0.0/app.tar.gz
```

## Testing

All tests pass successfully:

```bash
# Run all tests
go test ./...

# Run URL handler tests specifically
go test -v ./internal/handler/... -run TestParseGitHub
go test -v ./internal/handler/... -run TestIsGitHub
```

**Test coverage:**
- ✅ Full GitHub URLs with `refs/heads/` prefix
- ✅ GitHub URLs without scheme
- ✅ Release download URLs
- ✅ raw.githubusercontent.com URLs
- ✅ Archive URLs
- ✅ API URLs
- ✅ Gist URLs
- ✅ Blob URLs (converted to raw)
- ✅ Base path routing
- ✅ URL detection logic

## Backward Compatibility

✅ **Fully backward compatible** - All existing traditional path-based routes continue to work:

```bash
# These all still work exactly as before
curl http://localhost:8080/owner/repo/raw/main/file.md
curl http://localhost:8080/owner/repo/releases/download/v1.0.0/app.tar.gz
curl http://localhost:8080/owner/repo/archive/refs/tags/v1.0.0.tar.gz
curl http://localhost:8080/api/repos/owner/repo
```

## Migration Guide

### No Changes Required

If you're using traditional path-based URLs, **no changes are needed**. Everything continues to work as before.

### To Use Full URL Support

Simply start using GitHub URLs directly:

```bash
# Before
curl http://proxy:8080/owner/repo/raw/main/README.md

# After (both work)
curl http://proxy:8080/owner/repo/raw/main/README.md  # Still works
curl http://proxy:8080/https://github.com/owner/repo/raw/main/README.md  # New feature
```

### To Use Base Path

1. Update configuration:
```yaml
server:
  base_path: /ghproxy
```

2. Update your URLs to include the base path:
```bash
# Before
curl http://proxy:8080/owner/repo/raw/main/file.md

# After
curl http://proxy:8080/ghproxy/owner/repo/raw/main/file.md
curl http://proxy:8080/ghproxy/https://github.com/owner/repo/raw/main/file.md
```

## Performance

- **No performance impact** on traditional path-based routes
- **Minimal overhead** for URL parsing (simple string operations + regex)
- **Same caching behavior** for both URL formats
- **Same proxy behavior** for both URL formats

## Security

- All existing security features apply to both URL formats:
  - SSRF protection
  - Rate limiting
  - Authentication
  - Request validation

## Future Enhancements

Potential future improvements:
1. Support for GitHub Enterprise URLs
2. URL shortening/aliasing
3. Custom domain mapping
4. URL rewriting rules

## Acknowledgments

This feature was implemented to support:
- Easier integration with existing GitHub workflows
- Better compatibility with tools expecting full URLs
- Flexible deployment options (subpath support)
- Corporate environments with specific URL requirements
