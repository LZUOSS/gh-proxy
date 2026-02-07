package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write minimal valid config
	configContent := `
server:
  http_port: 8080
`
	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Load config
	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify defaults are set
	if cfg.Server.HTTPPort != 8080 {
		t.Errorf("Expected HTTPPort 8080, got %d", cfg.Server.HTTPPort)
	}
	if cfg.Cache.Enabled != true {
		t.Errorf("Expected Cache.Enabled true, got %v", cfg.Cache.Enabled)
	}
	if cfg.Metrics.Port != 9090 {
		t.Errorf("Expected Metrics.Port 9090, got %d", cfg.Metrics.Port)
	}
}

func TestValidateServerConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ServerConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: ServerConfig{
				HTTPPort:        8080,
				HTTPSPort:       8443,
				EnableHTTPS:     false,
				ReadTimeout:     30 * time.Second,
				WriteTimeout:    30 * time.Second,
				IdleTimeout:     120 * time.Second,
				ShutdownTimeout: 30 * time.Second,
				MaxHeaderBytes:  1 << 20,
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			cfg: ServerConfig{
				HTTPPort:     0,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  120 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			cfg: ServerConfig{
				HTTPPort:     70000,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  120 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "same http and https ports",
			cfg: ServerConfig{
				HTTPPort:     8080,
				HTTPSPort:    8080,
				EnableHTTPS:  true,
				TLSCertFile:  "cert.pem",
				TLSKeyFile:   "key.pem",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  120 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			cfg: ServerConfig{
				HTTPPort:     8080,
				ReadTimeout:  -1 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  120 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServer(&tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateProxyConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ProxyConfig
		wantErr bool
	}{
		{
			name: "disabled proxy",
			cfg: ProxyConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "valid socks5 proxy",
			cfg: ProxyConfig{
				Enabled:             true,
				Type:                "socks5",
				Address:             "127.0.0.1:1080",
				Timeout:             30 * time.Second,
				DialTimeout:         10 * time.Second,
				KeepAlive:           30 * time.Second,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid proxy type",
			cfg: ProxyConfig{
				Enabled: true,
				Type:    "invalid",
				Address: "127.0.0.1:1080",
			},
			wantErr: true,
		},
		{
			name: "missing address",
			cfg: ProxyConfig{
				Enabled: true,
				Type:    "socks5",
			},
			wantErr: true,
		},
		{
			name: "invalid address format",
			cfg: ProxyConfig{
				Enabled: true,
				Type:    "socks5",
				Address: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProxy(&tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProxy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCacheConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     CacheConfig
		wantErr bool
	}{
		{
			name: "valid memory cache",
			cfg: CacheConfig{
				Enabled:       true,
				Type:          "memory",
				MaxMemorySize: 100 * 1024 * 1024,
				TTL:           1 * time.Hour,
				CleanupInterval: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "valid disk cache",
			cfg: CacheConfig{
				Enabled:       true,
				Type:          "disk",
				MaxMemorySize: 100 * 1024 * 1024,
				MaxDiskSize:   1024 * 1024 * 1024,
				DiskPath:      "/tmp/cache",
				TTL:           1 * time.Hour,
				CleanupInterval: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "invalid cache type",
			cfg: CacheConfig{
				Enabled: true,
				Type:    "invalid",
			},
			wantErr: true,
		},
		{
			name: "disk cache without path",
			cfg: CacheConfig{
				Enabled:         true,
				Type:            "disk",
				MaxMemorySize:   100 * 1024 * 1024,
				MaxDiskSize:     1024 * 1024 * 1024,
				TTL:             1 * time.Hour,
				CleanupInterval: 5 * time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCache(&tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRateLimitConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     RateLimitConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: RateLimitConfig{
				Enabled:           true,
				RequestsPerSecond: 100,
				Burst:             200,
				Strategy:          "ip",
				CleanupInterval:   1 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "burst less than rps",
			cfg: RateLimitConfig{
				Enabled:           true,
				RequestsPerSecond: 100,
				Burst:             50,
				Strategy:          "ip",
			},
			wantErr: true,
		},
		{
			name: "invalid strategy",
			cfg: RateLimitConfig{
				Enabled:           true,
				RequestsPerSecond: 100,
				Burst:             200,
				Strategy:          "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRateLimit(&tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRateLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAuthConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     AuthConfig
		wantErr bool
	}{
		{
			name: "valid token auth",
			cfg: AuthConfig{
				Enabled:        true,
				Type:           "token",
				TokenHeader:    "X-Auth-Token",
				AllowAnonymous: true,
			},
			wantErr: false,
		},
		{
			name: "missing token header",
			cfg: AuthConfig{
				Enabled: true,
				Type:    "token",
			},
			wantErr: true,
		},
		{
			name: "no tokens with anonymous disabled",
			cfg: AuthConfig{
				Enabled:        true,
				Type:           "token",
				TokenHeader:    "X-Auth-Token",
				AllowAnonymous: false,
				Tokens:         []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAuth(&tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
