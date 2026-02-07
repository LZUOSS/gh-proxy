// Package middleware provides HTTP middleware handlers for the GitHub reverse proxy.
//
// Middleware Stack Order (CRITICAL):
//
// The middleware must be applied in this specific order:
//
//  1. Recovery      - Panic recovery with stack trace logging
//  2. Logging       - Request/response logging with zap
//  3. Metrics       - Prometheus metrics collection
//  4. RealIP        - Client IP extraction from headers
//  5. Security      - Security headers (SSRF, headers, validation)
//  6. RateLimit     - Per-IP rate limiting
//  7. Auth          - Optional authentication (Basic/Bearer)
//
// Example usage:
//
//	router := gin.New()
//
//	// Apply middleware in order
//	router.Use(middleware.Recovery(logger))
//	router.Use(middleware.Logging(logger))
//	router.Use(middleware.Metrics())
//	router.Use(middleware.RealIP())
//	router.Use(middleware.SecurityHeaders())
//	router.Use(middleware.RateLimit(rateLimiter))
//	router.Use(middleware.Auth(&cfg.Auth, authCache, logger))
//
// Middleware Details:
//
// Recovery: Recovers from panics, logs stack traces, returns 500 error
// Logging: Logs method, path, status, duration, IP, user-agent, request ID
// Metrics: Records request count, duration, response size to Prometheus
// RealIP: Extracts client IP from X-Real-IP, X-Forwarded-For, etc.
// Security: Adds security headers (X-Content-Type-Options, CSP, etc.)
// RateLimit: Enforces per-IP rate limiting using token bucket algorithm
// Auth: Optional authentication via Basic or Bearer tokens, with caching
//
// Context Values:
//
// The middleware sets the following values in the Gin context:
//
//  - "client_ip"   (string)      - Real client IP address (set by RealIP)
//  - "auth_token"  (*auth.Token) - Validated authentication token (set by Auth)
//
// Handlers can retrieve these values using c.GetString("client_ip") or c.Get("auth_token").
package middleware
