package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kexi/github-reverse-proxy/internal/auth"
	"github.com/kexi/github-reverse-proxy/internal/cache"
	"github.com/kexi/github-reverse-proxy/internal/config"
	"github.com/kexi/github-reverse-proxy/internal/handler"
	"github.com/kexi/github-reverse-proxy/internal/metrics"
	"github.com/kexi/github-reverse-proxy/internal/middleware"
	"github.com/kexi/github-reverse-proxy/internal/proxy"
	"github.com/kexi/github-reverse-proxy/internal/ratelimit"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// HTTPServer represents the HTTP server with all dependencies.
type HTTPServer struct {
	router       *gin.Engine
	server       *http.Server
	config       *config.Config
	proxyClient  *proxy.ProxyClient
	cache        *cache.Cache
	rateLimiter  *ratelimit.RateLimiter
	authCache    *auth.Cache
	logger       *zap.Logger
}

// NewHTTPServer creates a new HTTP server with all dependencies initialized.
func NewHTTPServer(cfg *config.Config) (*HTTPServer, error) {
	// Initialize logger
	logger, err := initLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize proxy client
	proxyConfig := &proxy.ProxyConfig{
		Type:                proxy.ProxyType(cfg.Proxy.Type),
		Address:             cfg.Proxy.Address,
		Username:            cfg.Proxy.Username,
		Password:            cfg.Proxy.Password,
		Timeout:             cfg.Proxy.Timeout,
		MaxIdleConns:        cfg.Proxy.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.Proxy.MaxIdleConnsPerHost,
	}

	// If proxy is not enabled, use direct connection
	if !cfg.Proxy.Enabled {
		proxyConfig.Type = proxy.ProxyTypeNone
	}

	proxyClient, err := proxy.NewProxyClient(proxyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy client: %w", err)
	}

	// Initialize cache
	memorySize := cfg.Cache.MaxMemoryEntries
	if memorySize <= 0 {
		// If MaxMemoryEntries not set, estimate from MaxMemorySize
		// Assume average entry size of 1KB
		memorySize = int(cfg.Cache.MaxMemorySize / 1024)
	}
	if memorySize <= 0 {
		memorySize = 1000 // Default to 1000 entries
	}

	cacheConfig := cache.Config{
		MemorySize: memorySize,
		DiskPath:   cfg.Cache.DiskPath,
		EnableDisk: cfg.Cache.Enabled && (cfg.Cache.Type == "disk" || cfg.Cache.Type == "hybrid"),
	}

	cacheSystem, err := cache.NewCache(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	// Initialize rate limiter
	var rateLimiter *ratelimit.RateLimiter
	if cfg.RateLimit.Enabled {
		rateLimiter = ratelimit.NewRateLimiter(
			rate.Limit(cfg.RateLimit.RequestsPerSecond),
			cfg.RateLimit.Burst,
		)
	}

	// Initialize auth cache
	var authCache *auth.Cache
	if cfg.Auth.Enabled {
		authCache = auth.NewCache(1 * time.Hour)
		// Start cleanup task
		authCache.StartCleanupTask(10 * time.Minute)
	}

	// Initialize Prometheus metrics if enabled
	if cfg.Metrics.Enabled {
		metrics.InitPrometheus()
	}

	// Create HTTP server instance
	httpServer := &HTTPServer{
		config:      cfg,
		proxyClient: proxyClient,
		cache:       cacheSystem,
		rateLimiter: rateLimiter,
		authCache:   authCache,
		logger:      logger,
	}

	// Setup router
	httpServer.setupRouter()

	return httpServer, nil
}

// setupRouter configures the Gin router with all middleware and routes.
func (s *HTTPServer) setupRouter() {
	// Set Gin mode based on logging level
	if s.config.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Setup middleware in order
	router.Use(middleware.Recovery(s.logger))
	router.Use(middleware.Logging(s.logger))

	if s.config.Metrics.Enabled {
		router.Use(middleware.Metrics())
	}

	router.Use(middleware.RealIP())
	router.Use(middleware.SecurityHeaders())

	if s.config.RateLimit.Enabled && s.rateLimiter != nil {
		router.Use(middleware.RateLimit(s.rateLimiter))
	}

	if s.config.Auth.Enabled && s.authCache != nil {
		router.Use(middleware.Auth(&s.config.Auth, s.authCache, s.logger))
	}

	// Setup routes
	s.setupRoutes(router)

	s.router = router
}

// setupRoutes defines all HTTP routes.
func (s *HTTPServer) setupRoutes(router *gin.Engine) {
	// Initialize handlers
	releasesHandler := handler.NewReleasesHandler(s.cache, s.proxyClient)
	rawHandler := handler.NewRawHandler(s.cache, s.proxyClient)
	archiveHandler := handler.NewArchiveHandler(s.cache, s.proxyClient)
	gitHandler := handler.NewGitHandler(s.proxyClient, "")
	gistHandler := handler.NewGistHandler(s.cache, s.proxyClient)
	apiHandler := handler.NewAPIHandler(s.cache, s.proxyClient, "")
	urlHandler := handler.NewURLHandler(s.cache, s.proxyClient)

	// Determine the base path
	basePath := s.config.Server.BasePath
	if basePath != "" {
		// Ensure base path starts with / but doesn't end with /
		if !strings.HasPrefix(basePath, "/") {
			basePath = "/" + basePath
		}
		basePath = strings.TrimSuffix(basePath, "/")
	}

	// Create a route group with the base path
	var routeGroup *gin.RouterGroup
	if basePath != "" {
		routeGroup = router.Group(basePath)
	} else {
		// Use empty group for root path
		routeGroup = router.Group("")
	}

	// Full URL handler - catches GitHub URLs like /https://github.com/owner/repo/...
	// This should be registered first to catch full URLs before path-based routes
	routeGroup.GET("/*url", func(c *gin.Context) {
		path := c.Param("url")
		// Check if this looks like a GitHub URL
		if isGitHubURL(path) {
			urlHandler.Handle(c)
			return
		}
		// If not a GitHub URL, continue to next handler
		c.Next()
	})

	// Traditional path-based routes
	// Release downloads
	routeGroup.GET("/:owner/:repo/releases/download/:tag/:filename", releasesHandler.Handle)

	// Raw content
	routeGroup.GET("/:owner/:repo/raw/:ref/*filepath", rawHandler.Handle)

	// Archive downloads
	routeGroup.GET("/:owner/:repo/archive/:ref", archiveHandler.Handle)

	// Git protocol routes
	routeGroup.GET("/:owner/:repo.git/info/refs", gitHandler.HandleInfoRefs)
	routeGroup.POST("/:owner/:repo.git/git-upload-pack", gitHandler.HandleUploadPack)
	routeGroup.POST("/:owner/:repo.git/git-receive-pack", gitHandler.HandleReceivePack)

	// Gist routes
	routeGroup.GET("/gist/:user/:gist_id/raw/:file", gistHandler.Handle)

	// API proxy
	routeGroup.Any("/api/*path", apiHandler.Handle)

	// Health check endpoint (always at root + base path)
	if basePath != "" {
		routeGroup.GET("/health", s.handleHealth)
	}
	router.GET("/health", s.handleHealth)

	// Metrics endpoint (if enabled, always at root)
	if s.config.Metrics.Enabled {
		router.GET(s.config.Metrics.Path, gin.WrapH(metrics.Handler()))
	}
}

// isGitHubURL checks if a path looks like a GitHub URL
func isGitHubURL(path string) bool {
	path = strings.TrimPrefix(path, "/")
	return strings.HasPrefix(path, "http://github.com/") ||
		strings.HasPrefix(path, "https://github.com/") ||
		strings.HasPrefix(path, "github.com/") ||
		strings.HasPrefix(path, "http://raw.githubusercontent.com/") ||
		strings.HasPrefix(path, "https://raw.githubusercontent.com/") ||
		strings.HasPrefix(path, "raw.githubusercontent.com/") ||
		strings.HasPrefix(path, "http://api.github.com/") ||
		strings.HasPrefix(path, "https://api.github.com/") ||
		strings.HasPrefix(path, "api.github.com/") ||
		strings.HasPrefix(path, "http://gist.github.com/") ||
		strings.HasPrefix(path, "https://gist.github.com/") ||
		strings.HasPrefix(path, "gist.github.com/")
}

// handleHealth handles health check requests.
func (s *HTTPServer) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Unix(),
	})
}

// Start starts the HTTP server.
func (s *HTTPServer) Start() error {
	// Create HTTP server
	addr := fmt.Sprintf(":%d", s.config.Server.HTTPPort)

	s.server = &http.Server{
		Addr:           addr,
		Handler:        s.router,
		ReadTimeout:    s.config.Server.ReadTimeout,
		WriteTimeout:   s.config.Server.WriteTimeout,
		IdleTimeout:    s.config.Server.IdleTimeout,
		MaxHeaderBytes: s.config.Server.MaxHeaderBytes,
	}

	s.logger.Info("starting HTTP server",
		zap.String("addr", addr),
		zap.Duration("read_timeout", s.config.Server.ReadTimeout),
		zap.Duration("write_timeout", s.config.Server.WriteTimeout),
	)

	// Start server
	if s.config.Server.EnableHTTPS {
		return s.server.ListenAndServeTLS(
			s.config.Server.TLSCertFile,
			s.config.Server.TLSKeyFile,
		)
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")

	if s.server == nil {
		return nil
	}

	// Shutdown the server
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("error shutting down server", zap.Error(err))
		return err
	}

	// Close proxy client connections
	if s.proxyClient != nil {
		s.proxyClient.Close()
	}

	s.logger.Info("HTTP server shutdown complete")
	return nil
}

// initLogger initializes the zap logger based on configuration.
func initLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	var zapConfig zap.Config

	// Set base config based on format
	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set log level
	switch cfg.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Set output paths
	if cfg.Output == "file" && cfg.FilePath != "" {
		zapConfig.OutputPaths = []string{cfg.FilePath}
		zapConfig.ErrorOutputPaths = []string{cfg.FilePath}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	}

	return zapConfig.Build()
}
