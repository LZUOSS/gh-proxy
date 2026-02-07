package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// ProxyClient wraps an HTTP client with proxy support
type ProxyClient struct {
	client *http.Client
	config *ProxyConfig
}

// NewProxyClient creates a new proxy client with the given configuration
func NewProxyClient(cfg *ProxyConfig) (*ProxyClient, error) {
	if cfg == nil {
		cfg = DefaultProxyConfig()
	}

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid proxy configuration: %w", err)
	}

	// Create transport based on proxy type
	transport, err := createTransport(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Create HTTP client
	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	return &ProxyClient{
		client: client,
		config: cfg,
	}, nil
}

// validateConfig validates the proxy configuration
func validateConfig(cfg *ProxyConfig) error {
	if cfg.Type != ProxyTypeSOCKS5 && cfg.Type != ProxyTypeHTTP && cfg.Type != ProxyTypeHTTPS && cfg.Type != ProxyTypeNone {
		return fmt.Errorf("unsupported proxy type: %s", cfg.Type)
	}

	if (cfg.Type == ProxyTypeSOCKS5 || cfg.Type == ProxyTypeHTTP || cfg.Type == ProxyTypeHTTPS) && cfg.Address == "" {
		return fmt.Errorf("proxy address is required for type: %s", cfg.Type)
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 100
	}

	if cfg.MaxIdleConnsPerHost <= 0 {
		cfg.MaxIdleConnsPerHost = 10
	}

	return nil
}

// createTransport creates an HTTP transport based on proxy type
func createTransport(cfg *ProxyConfig) (*http.Transport, error) {
	baseTransport := &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	switch cfg.Type {
	case ProxyTypeSOCKS5:
		// Use SOCKS5 proxy
		baseTransport.DialContext = socks5DialContext(cfg)

	case ProxyTypeHTTP, ProxyTypeHTTPS:
		// Use HTTP or HTTPS proxy
		proxyURL, err := parseProxyURL(cfg)
		if err != nil {
			return nil, err
		}
		baseTransport.Proxy = http.ProxyURL(proxyURL)
		baseTransport.DialContext = (&net.Dialer{
			Timeout:   cfg.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext

	case ProxyTypeNone:
		// Direct connection
		baseTransport.DialContext = (&net.Dialer{
			Timeout:   cfg.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext

	default:
		return nil, fmt.Errorf("unsupported proxy type: %s", cfg.Type)
	}

	return baseTransport, nil
}

// Do executes an HTTP request using the proxy client
func (pc *ProxyClient) Do(req *http.Request) (*http.Response, error) {
	return pc.client.Do(req)
}

// DoWithContext executes an HTTP request with context using the proxy client
func (pc *ProxyClient) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return pc.client.Do(req)
}

// Get performs an HTTP GET request
func (pc *ProxyClient) Get(url string) (*http.Response, error) {
	return pc.client.Get(url)
}

// Post performs an HTTP POST request
func (pc *ProxyClient) Post(url, contentType string, body interface{}) (*http.Response, error) {
	// This is a simplified version; in production, you'd handle body properly
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return pc.client.Do(req)
}

// Client returns the underlying HTTP client
func (pc *ProxyClient) Client() *http.Client {
	return pc.client
}

// Config returns the proxy configuration
func (pc *ProxyClient) Config() *ProxyConfig {
	return pc.config
}

// Close closes idle connections
func (pc *ProxyClient) Close() {
	pc.client.CloseIdleConnections()
}
