package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Proxy     ProxyConfig     `mapstructure:"proxy"`
	Cache     CacheConfig     `mapstructure:"cache"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Security  SecurityConfig  `mapstructure:"security"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	Logging   LoggingConfig   `mapstructure:"logging"`
}

// ServerConfig contains HTTP/HTTPS server settings
type ServerConfig struct {
	HTTPPort         int           `mapstructure:"http_port"`
	HTTPSPort        int           `mapstructure:"https_port"`
	EnableHTTPS      bool          `mapstructure:"enable_https"`
	TLSCertFile      string        `mapstructure:"tls_cert_file"`
	TLSKeyFile       string        `mapstructure:"tls_key_file"`
	BasePath         string        `mapstructure:"base_path"` // Base path for all routes (e.g., "/ghproxy")
	ReadTimeout      time.Duration `mapstructure:"read_timeout"`
	WriteTimeout     time.Duration `mapstructure:"write_timeout"`
	IdleTimeout      time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes   int           `mapstructure:"max_header_bytes"`
	ShutdownTimeout  time.Duration `mapstructure:"shutdown_timeout"`
	EnableGracefulShutdown bool     `mapstructure:"enable_graceful_shutdown"`
}

// ProxyConfig contains proxy client settings
type ProxyConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	Type             string        `mapstructure:"type"` // "socks5" or "http"
	Address          string        `mapstructure:"address"`
	Username         string        `mapstructure:"username"`
	Password         string        `mapstructure:"password"`
	Timeout          time.Duration `mapstructure:"timeout"`
	DialTimeout      time.Duration `mapstructure:"dial_timeout"`
	KeepAlive        time.Duration `mapstructure:"keep_alive"`
	MaxIdleConns     int           `mapstructure:"max_idle_conns"`
	MaxIdleConnsPerHost int        `mapstructure:"max_idle_conns_per_host"`
	IdleConnTimeout  time.Duration `mapstructure:"idle_conn_timeout"`
}

// CacheConfig contains caching settings
type CacheConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	Type              string        `mapstructure:"type"` // "memory", "disk", "hybrid"
	MaxMemorySize     int64         `mapstructure:"max_memory_size"`     // Maximum memory cache size in bytes
	MaxMemoryEntries  int           `mapstructure:"max_memory_entries"`  // Maximum number of entries in memory cache
	MaxDiskSize       int64         `mapstructure:"max_disk_size"`
	DiskPath          string        `mapstructure:"disk_path"`
	TTL               time.Duration `mapstructure:"ttl"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
	EnableCompression bool          `mapstructure:"enable_compression"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	RequestsPerSecond int           `mapstructure:"requests_per_second"`
	Burst             int           `mapstructure:"burst"`
	Strategy          string        `mapstructure:"strategy"` // "ip", "token", "both"
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
	BanDuration       time.Duration `mapstructure:"ban_duration"`
	BanThreshold      int           `mapstructure:"ban_threshold"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Enabled         bool     `mapstructure:"enabled"`
	Type            string   `mapstructure:"type"` // "token", "basic", "both"
	Tokens          []string `mapstructure:"tokens"`
	TokenHeader     string   `mapstructure:"token_header"`
	AllowAnonymous  bool     `mapstructure:"allow_anonymous"`
	RequireAuth     []string `mapstructure:"require_auth"` // Paths that require authentication
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	EnableSSRFProtection bool     `mapstructure:"enable_ssrf_protection"`
	AllowedDomains       []string `mapstructure:"allowed_domains"`
	BlockedIPs           []string `mapstructure:"blocked_ips"`
	BlockPrivateIPs      bool     `mapstructure:"block_private_ips"`
	MaxRequestSize       int64    `mapstructure:"max_request_size"`
	EnableCORS           bool     `mapstructure:"enable_cors"`
	CORSAllowedOrigins   []string `mapstructure:"cors_allowed_origins"`
	EnableHSTS           bool     `mapstructure:"enable_hsts"`
	HSTSMaxAge           int      `mapstructure:"hsts_max_age"`
	EnableCSP            bool     `mapstructure:"enable_csp"`
	CSPDirectives        string   `mapstructure:"csp_directives"`
}

// MetricsConfig contains metrics/monitoring settings
type MetricsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Port       int    `mapstructure:"port"`
	Path       string `mapstructure:"path"`
	Namespace  string `mapstructure:"namespace"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level       string `mapstructure:"level"` // "debug", "info", "warn", "error"
	Format      string `mapstructure:"format"` // "json", "text"
	Output      string `mapstructure:"output"` // "stdout", "file"
	FilePath    string `mapstructure:"file_path"`
	MaxSize     int    `mapstructure:"max_size"` // megabytes
	MaxBackups  int    `mapstructure:"max_backups"`
	MaxAge      int    `mapstructure:"max_age"` // days
	Compress    bool   `mapstructure:"compress"`
}

// Load reads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// Enable environment variable overrides
	v.SetEnvPrefix("GITHUB_PROXY")
	v.AutomaticEnv()

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, use defaults and env vars
	}

	// Unmarshal configuration
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := Validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.http_port", 8080)
	v.SetDefault("server.https_port", 8443)
	v.SetDefault("server.enable_https", false)
	v.SetDefault("server.base_path", "") // Default to root path
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.idle_timeout", 120*time.Second)
	v.SetDefault("server.max_header_bytes", 1<<20) // 1MB
	v.SetDefault("server.shutdown_timeout", 30*time.Second)
	v.SetDefault("server.enable_graceful_shutdown", true)

	// Proxy defaults
	v.SetDefault("proxy.enabled", false)
	v.SetDefault("proxy.type", "socks5")
	v.SetDefault("proxy.timeout", 30*time.Second)
	v.SetDefault("proxy.dial_timeout", 10*time.Second)
	v.SetDefault("proxy.keep_alive", 30*time.Second)
	v.SetDefault("proxy.max_idle_conns", 100)
	v.SetDefault("proxy.max_idle_conns_per_host", 10)
	v.SetDefault("proxy.idle_conn_timeout", 90*time.Second)

	// Cache defaults
	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.type", "hybrid")
	v.SetDefault("cache.max_memory_size", 100*1024*1024) // 100MB
	v.SetDefault("cache.max_disk_size", 1024*1024*1024)  // 1GB
	v.SetDefault("cache.disk_path", "./cache")
	v.SetDefault("cache.ttl", 1*time.Hour)
	v.SetDefault("cache.cleanup_interval", 5*time.Minute)
	v.SetDefault("cache.enable_compression", true)

	// Rate limit defaults
	v.SetDefault("ratelimit.enabled", true)
	v.SetDefault("ratelimit.requests_per_second", 100)
	v.SetDefault("ratelimit.burst", 200)
	v.SetDefault("ratelimit.strategy", "ip")
	v.SetDefault("ratelimit.cleanup_interval", 1*time.Minute)
	v.SetDefault("ratelimit.ban_duration", 1*time.Hour)
	v.SetDefault("ratelimit.ban_threshold", 1000)

	// Auth defaults
	v.SetDefault("auth.enabled", false)
	v.SetDefault("auth.type", "token")
	v.SetDefault("auth.token_header", "X-Auth-Token")
	v.SetDefault("auth.allow_anonymous", true)

	// Security defaults
	v.SetDefault("security.enable_ssrf_protection", true)
	v.SetDefault("security.allowed_domains", []string{"github.com", "raw.githubusercontent.com"})
	v.SetDefault("security.block_private_ips", true)
	v.SetDefault("security.max_request_size", 100*1024*1024) // 100MB
	v.SetDefault("security.enable_cors", true)
	v.SetDefault("security.cors_allowed_origins", []string{"*"})
	v.SetDefault("security.enable_hsts", false)
	v.SetDefault("security.hsts_max_age", 31536000) // 1 year
	v.SetDefault("security.enable_csp", false)

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.port", 9090)
	v.SetDefault("metrics.path", "/metrics")
	v.SetDefault("metrics.namespace", "github_proxy")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.max_size", 100) // megabytes
	v.SetDefault("logging.max_backups", 3)
	v.SetDefault("logging.max_age", 28) // days
	v.SetDefault("logging.compress", true)
}

// Get returns a copy of the configuration value
func (c *Config) Get() Config {
	return *c
}
