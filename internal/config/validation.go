package config

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Validate checks if the configuration is valid
func Validate(cfg *Config) error {
	if err := validateServer(&cfg.Server); err != nil {
		return fmt.Errorf("server config: %w", err)
	}

	if err := validateProxy(&cfg.Proxy); err != nil {
		return fmt.Errorf("proxy config: %w", err)
	}

	if err := validateCache(&cfg.Cache); err != nil {
		return fmt.Errorf("cache config: %w", err)
	}

	if err := validateRateLimit(&cfg.RateLimit); err != nil {
		return fmt.Errorf("rate limit config: %w", err)
	}

	if err := validateAuth(&cfg.Auth); err != nil {
		return fmt.Errorf("auth config: %w", err)
	}

	if err := validateSecurity(&cfg.Security); err != nil {
		return fmt.Errorf("security config: %w", err)
	}

	if err := validateMetrics(&cfg.Metrics); err != nil {
		return fmt.Errorf("metrics config: %w", err)
	}

	if err := validateLogging(&cfg.Logging); err != nil {
		return fmt.Errorf("logging config: %w", err)
	}

	return nil
}

// validateServer validates server configuration
func validateServer(cfg *ServerConfig) error {
	// Validate HTTP port
	if cfg.HTTPPort < 1 || cfg.HTTPPort > 65535 {
		return fmt.Errorf("http_port must be between 1 and 65535, got %d", cfg.HTTPPort)
	}

	// Validate HTTPS port
	if cfg.EnableHTTPS {
		if cfg.HTTPSPort < 1 || cfg.HTTPSPort > 65535 {
			return fmt.Errorf("https_port must be between 1 and 65535, got %d", cfg.HTTPSPort)
		}

		if cfg.HTTPPort == cfg.HTTPSPort {
			return fmt.Errorf("http_port and https_port cannot be the same")
		}

		// Check TLS certificate files
		if cfg.TLSCertFile == "" {
			return fmt.Errorf("tls_cert_file is required when HTTPS is enabled")
		}
		if cfg.TLSKeyFile == "" {
			return fmt.Errorf("tls_key_file is required when HTTPS is enabled")
		}

		if _, err := os.Stat(cfg.TLSCertFile); os.IsNotExist(err) {
			return fmt.Errorf("tls_cert_file does not exist: %s", cfg.TLSCertFile)
		}
		if _, err := os.Stat(cfg.TLSKeyFile); os.IsNotExist(err) {
			return fmt.Errorf("tls_key_file does not exist: %s", cfg.TLSKeyFile)
		}
	}

	// Validate timeouts
	if cfg.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be greater than 0")
	}
	if cfg.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be greater than 0")
	}
	if cfg.IdleTimeout <= 0 {
		return fmt.Errorf("idle_timeout must be greater than 0")
	}
	if cfg.ShutdownTimeout < 0 {
		return fmt.Errorf("shutdown_timeout cannot be negative")
	}

	// Validate max header bytes
	if cfg.MaxHeaderBytes <= 0 {
		return fmt.Errorf("max_header_bytes must be greater than 0")
	}

	return nil
}

// validateProxy validates proxy configuration
func validateProxy(cfg *ProxyConfig) error {
	if !cfg.Enabled {
		return nil
	}

	// Validate proxy type
	validTypes := []string{"socks5", "http", "https"}
	if !contains(validTypes, cfg.Type) {
		return fmt.Errorf("proxy type must be one of %v, got %s", validTypes, cfg.Type)
	}

	// Validate proxy address
	if cfg.Address == "" {
		return fmt.Errorf("proxy address is required when proxy is enabled")
	}

	// Validate address format (host:port)
	host, port, err := net.SplitHostPort(cfg.Address)
	if err != nil {
		return fmt.Errorf("invalid proxy address format (expected host:port): %w", err)
	}
	if host == "" {
		return fmt.Errorf("proxy host cannot be empty")
	}
	if port == "" {
		return fmt.Errorf("proxy port cannot be empty")
	}

	// Validate timeouts
	if cfg.Timeout <= 0 {
		return fmt.Errorf("proxy timeout must be greater than 0")
	}
	if cfg.DialTimeout <= 0 {
		return fmt.Errorf("proxy dial_timeout must be greater than 0")
	}
	if cfg.KeepAlive < 0 {
		return fmt.Errorf("proxy keep_alive cannot be negative")
	}
	if cfg.IdleConnTimeout <= 0 {
		return fmt.Errorf("proxy idle_conn_timeout must be greater than 0")
	}

	// Validate connection pool settings
	if cfg.MaxIdleConns < 0 {
		return fmt.Errorf("proxy max_idle_conns cannot be negative")
	}
	if cfg.MaxIdleConnsPerHost < 0 {
		return fmt.Errorf("proxy max_idle_conns_per_host cannot be negative")
	}
	if cfg.MaxIdleConnsPerHost > cfg.MaxIdleConns {
		return fmt.Errorf("proxy max_idle_conns_per_host cannot be greater than max_idle_conns")
	}

	return nil
}

// validateCache validates cache configuration
func validateCache(cfg *CacheConfig) error {
	if !cfg.Enabled {
		return nil
	}

	// Validate cache type
	validTypes := []string{"memory", "disk", "hybrid"}
	if !contains(validTypes, cfg.Type) {
		return fmt.Errorf("cache type must be one of %v, got %s", validTypes, cfg.Type)
	}

	// Validate memory size
	if cfg.MaxMemorySize <= 0 {
		return fmt.Errorf("cache max_memory_size must be greater than 0")
	}

	// Validate disk settings if disk caching is enabled
	if cfg.Type == "disk" || cfg.Type == "hybrid" {
		if cfg.MaxDiskSize <= 0 {
			return fmt.Errorf("cache max_disk_size must be greater than 0 for disk/hybrid cache")
		}
		if cfg.DiskPath == "" {
			return fmt.Errorf("cache disk_path is required for disk/hybrid cache")
		}
	}

	// Validate TTL
	if cfg.TTL <= 0 {
		return fmt.Errorf("cache ttl must be greater than 0")
	}

	// Validate cleanup interval
	if cfg.CleanupInterval <= 0 {
		return fmt.Errorf("cache cleanup_interval must be greater than 0")
	}

	return nil
}

// validateRateLimit validates rate limit configuration
func validateRateLimit(cfg *RateLimitConfig) error {
	if !cfg.Enabled {
		return nil
	}

	// Validate requests per second
	if cfg.RequestsPerSecond <= 0 {
		return fmt.Errorf("requests_per_second must be greater than 0")
	}

	// Validate burst
	if cfg.Burst < cfg.RequestsPerSecond {
		return fmt.Errorf("burst must be at least equal to requests_per_second")
	}

	// Validate strategy
	validStrategies := []string{"ip", "token", "both"}
	if !contains(validStrategies, cfg.Strategy) {
		return fmt.Errorf("rate limit strategy must be one of %v, got %s", validStrategies, cfg.Strategy)
	}

	// Validate cleanup interval
	if cfg.CleanupInterval <= 0 {
		return fmt.Errorf("rate limit cleanup_interval must be greater than 0")
	}

	// Validate ban settings
	if cfg.BanDuration < 0 {
		return fmt.Errorf("rate limit ban_duration cannot be negative")
	}
	if cfg.BanThreshold < 0 {
		return fmt.Errorf("rate limit ban_threshold cannot be negative")
	}

	return nil
}

// validateAuth validates authentication configuration
func validateAuth(cfg *AuthConfig) error {
	if !cfg.Enabled {
		return nil
	}

	// Validate auth type
	validTypes := []string{"token", "basic", "both"}
	if !contains(validTypes, cfg.Type) {
		return fmt.Errorf("auth type must be one of %v, got %s", validTypes, cfg.Type)
	}

	// Validate token configuration
	if cfg.Type == "token" || cfg.Type == "both" {
		if cfg.TokenHeader == "" {
			return fmt.Errorf("token_header is required when using token authentication")
		}
		if !cfg.AllowAnonymous && len(cfg.Tokens) == 0 {
			return fmt.Errorf("at least one token must be configured when allow_anonymous is false")
		}
	}

	return nil
}

// validateSecurity validates security configuration
func validateSecurity(cfg *SecurityConfig) error {
	// Validate max request size
	if cfg.MaxRequestSize <= 0 {
		return fmt.Errorf("max_request_size must be greater than 0")
	}

	// Validate allowed domains for SSRF protection
	if cfg.EnableSSRFProtection {
		if len(cfg.AllowedDomains) == 0 {
			return fmt.Errorf("at least one allowed domain must be specified when SSRF protection is enabled")
		}
	}

	// Validate blocked IPs format
	for _, ip := range cfg.BlockedIPs {
		if net.ParseIP(ip) == nil {
			// Try parsing as CIDR
			if _, _, err := net.ParseCIDR(ip); err != nil {
				return fmt.Errorf("invalid IP address or CIDR format in blocked_ips: %s", ip)
			}
		}
	}

	// Validate CORS settings
	if cfg.EnableCORS {
		if len(cfg.CORSAllowedOrigins) == 0 {
			return fmt.Errorf("at least one allowed origin must be specified when CORS is enabled")
		}
	}

	// Validate HSTS settings
	if cfg.EnableHSTS {
		if cfg.HSTSMaxAge <= 0 {
			return fmt.Errorf("hsts_max_age must be greater than 0 when HSTS is enabled")
		}
	}

	return nil
}

// validateMetrics validates metrics configuration
func validateMetrics(cfg *MetricsConfig) error {
	if !cfg.Enabled {
		return nil
	}

	// Validate metrics port
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("metrics port must be between 1 and 65535, got %d", cfg.Port)
	}

	// Validate metrics path
	if cfg.Path == "" {
		return fmt.Errorf("metrics path cannot be empty")
	}
	if !strings.HasPrefix(cfg.Path, "/") {
		return fmt.Errorf("metrics path must start with /")
	}

	// Validate namespace
	if cfg.Namespace == "" {
		return fmt.Errorf("metrics namespace cannot be empty")
	}

	return nil
}

// validateLogging validates logging configuration
func validateLogging(cfg *LoggingConfig) error {
	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, cfg.Level) {
		return fmt.Errorf("log level must be one of %v, got %s", validLevels, cfg.Level)
	}

	// Validate log format
	validFormats := []string{"json", "text"}
	if !contains(validFormats, cfg.Format) {
		return fmt.Errorf("log format must be one of %v, got %s", validFormats, cfg.Format)
	}

	// Validate log output
	validOutputs := []string{"stdout", "stderr", "file"}
	if !contains(validOutputs, cfg.Output) {
		return fmt.Errorf("log output must be one of %v, got %s", validOutputs, cfg.Output)
	}

	// Validate file settings if output is file
	if cfg.Output == "file" {
		if cfg.FilePath == "" {
			return fmt.Errorf("log file_path is required when output is set to file")
		}
		if cfg.MaxSize <= 0 {
			return fmt.Errorf("log max_size must be greater than 0")
		}
		if cfg.MaxBackups < 0 {
			return fmt.Errorf("log max_backups cannot be negative")
		}
		if cfg.MaxAge < 0 {
			return fmt.Errorf("log max_age cannot be negative")
		}
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
