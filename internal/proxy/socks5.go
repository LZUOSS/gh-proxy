package proxy

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/proxy"
)

// createSOCKS5Dialer creates a SOCKS5 proxy dialer
func createSOCKS5Dialer(cfg *ProxyConfig) (proxy.Dialer, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("SOCKS5 proxy address is required")
	}

	// Create auth if credentials are provided
	var auth *proxy.Auth
	if cfg.Username != "" || cfg.Password != "" {
		auth = &proxy.Auth{
			User:     cfg.Username,
			Password: cfg.Password,
		}
	}

	// Create base dialer with timeout
	baseDialer := &net.Dialer{
		Timeout:   cfg.Timeout,
		KeepAlive: 30 * time.Second,
	}

	// Create SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", cfg.Address, auth, baseDialer)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	return dialer, nil
}

// socks5DialContext creates a DialContext function for SOCKS5 proxy
func socks5DialContext(cfg *ProxyConfig) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer, err := createSOCKS5Dialer(cfg)
		if err != nil {
			return nil, err
		}

		// Use the SOCKS5 dialer
		conn, err := dialer.Dial(network, addr)
		if err != nil {
			return nil, fmt.Errorf("SOCKS5 dial failed: %w", err)
		}

		return conn, nil
	}
}
