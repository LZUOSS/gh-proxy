# Proxy Client Package

This package provides a flexible HTTP client with support for multiple proxy types.

## Features

- **Multiple Proxy Types**: Supports SOCKS5, HTTP proxy, and direct connections
- **Connection Pooling**: Configurable connection pool settings (MaxIdleConns: 100, MaxIdleConnsPerHost: 10)
- **Timeouts**: Configurable connection and request timeouts
- **Authentication**: Support for proxy authentication (username/password)
- **Streaming**: Utilities for efficient data streaming

## Components

### client.go
Main ProxyClient implementation with:
- `ProxyClient` struct wrapping `http.Client`
- `NewProxyClient(cfg *ProxyConfig)` constructor
- Connection pooling configuration
- Timeout support
- Methods: `Do()`, `DoWithContext()`, `Get()`, `Post()`, `Client()`, `Config()`, `Close()`

### config.go
Configuration structures:
- `ProxyConfig` struct with all proxy settings
- `ProxyType` constants: `socks5`, `http`, `none`
- `DefaultProxyConfig()` function

### socks5.go
SOCKS5 proxy implementation using `golang.org/x/net/proxy`:
- `createSOCKS5Dialer()` function
- `socks5DialContext()` function
- Support for SOCKS5 authentication

### http.go
HTTP proxy implementation:
- `createHTTPTransport()` function
- `parseProxyURL()` function
- `httpProxyDialContext()` function
- HTTP proxy authentication support

### stream.go
Streaming utilities:
- `StreamResponse()` - stream HTTP response body
- `StreamRequest()` - stream HTTP request body
- `CopyHeaders()` - copy HTTP headers
- `BufferedCopy()` - buffered data copy (32KB buffer)

## Usage

```go
// Direct connection (no proxy)
cfg := &proxy.ProxyConfig{
    Type:                proxy.ProxyTypeNone,
    Timeout:             30 * time.Second,
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
}
client, err := proxy.NewProxyClient(cfg)

// SOCKS5 proxy
cfg := &proxy.ProxyConfig{
    Type:                proxy.ProxyTypeSOCKS5,
    Address:             "localhost:1080",
    Username:            "user",
    Password:            "pass",
    Timeout:             30 * time.Second,
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
}
client, err := proxy.NewProxyClient(cfg)

// HTTP proxy
cfg := &proxy.ProxyConfig{
    Type:                proxy.ProxyTypeHTTP,
    Address:             "proxy.example.com:8080",
    Username:            "user",
    Password:            "pass",
    Timeout:             30 * time.Second,
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
}
client, err := proxy.NewProxyClient(cfg)

// Use default configuration
client, err := proxy.NewProxyClient(nil)
```

## Dependencies

- `golang.org/x/net/proxy` - for SOCKS5 support

## Testing

Run tests with:
```bash
go test ./internal/proxy/... -v
```
