package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// createHTTPTransport creates an HTTP transport with proxy configuration
func createHTTPTransport(cfg *ProxyConfig) (*http.Transport, error) {
	transport := &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   cfg.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	// Configure proxy if HTTP or HTTPS proxy is specified
	if (cfg.Type == ProxyTypeHTTP || cfg.Type == ProxyTypeHTTPS) && cfg.Address != "" {
		proxyURL, err := parseProxyURL(cfg)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return transport, nil
}

// parseProxyURL parses the proxy configuration into a URL
func parseProxyURL(cfg *ProxyConfig) (*url.URL, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("HTTP/HTTPS proxy address is required")
	}

	// Determine scheme based on proxy type
	scheme := "http"
	if cfg.Type == ProxyTypeHTTPS {
		scheme = "https"
	}

	// Build proxy URL
	proxyURL := &url.URL{
		Scheme: scheme,
		Host:   cfg.Address,
	}

	// Add authentication if provided
	if cfg.Username != "" {
		if cfg.Password != "" {
			proxyURL.User = url.UserPassword(cfg.Username, cfg.Password)
		} else {
			proxyURL.User = url.User(cfg.Username)
		}
	}

	return proxyURL, nil
}

// httpProxyDialContext creates a DialContext function for HTTP proxy
func httpProxyDialContext(cfg *ProxyConfig) func(ctx context.Context, network, addr string) (net.Conn, error) {
	transport, _ := createHTTPTransport(cfg)
	if transport.Proxy == nil {
		// Fallback to direct connection if no proxy configured
		return (&net.Dialer{
			Timeout:   cfg.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext
	}

	return transport.DialContext
}
