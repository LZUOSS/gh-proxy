package proxy

import "time"

// ProxyType represents the type of proxy
type ProxyType string

const (
	ProxyTypeSOCKS5 ProxyType = "socks5"
	ProxyTypeHTTP   ProxyType = "http"
	ProxyTypeHTTPS  ProxyType = "https"
	ProxyTypeNone   ProxyType = "none"
)

// ProxyConfig holds the proxy client configuration
type ProxyConfig struct {
	// Type is the proxy type (socks5, http, none)
	Type ProxyType

	// Address is the proxy server address (host:port)
	Address string

	// Username for proxy authentication (optional)
	Username string

	// Password for proxy authentication (optional)
	Password string

	// Timeout for proxy connections
	Timeout time.Duration

	// MaxIdleConns controls the maximum number of idle connections
	MaxIdleConns int

	// MaxIdleConnsPerHost controls the maximum idle connections per host
	MaxIdleConnsPerHost int
}

// DefaultProxyConfig returns a ProxyConfig with sensible defaults
func DefaultProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		Type:                ProxyTypeNone,
		Timeout:             30 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
	}
}
