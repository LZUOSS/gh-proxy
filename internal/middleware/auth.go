package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kexi/github-reverse-proxy/internal/auth"
	"github.com/kexi/github-reverse-proxy/internal/config"
	"go.uber.org/zap"
)

// Auth returns a middleware that validates authentication.
// It supports both Basic authentication and Bearer token authentication.
// The middleware is optional and can be disabled via configuration.
func Auth(cfg *config.AuthConfig, cache *auth.Cache, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication if not enabled
		if !cfg.Enabled {
			c.Next()
			return
		}

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Check if anonymous access is allowed
			if cfg.AllowAnonymous {
				c.Next()
				return
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
				"message": "Authorization header required",
			})
			return
		}

		var token *auth.Token
		var err error

		// Parse authorization header
		if strings.HasPrefix(authHeader, "Basic ") {
			// Handle Basic authentication
			token, err = handleBasicAuth(authHeader, cache, logger)
		} else if strings.HasPrefix(authHeader, "Bearer ") {
			// Handle Bearer token authentication
			token, err = handleBearerAuth(authHeader, cache, logger)
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
				"message": "Invalid Authorization header format",
			})
			return
		}

		if err != nil {
			logger.Warn("authentication failed",
				zap.Error(err),
				zap.String("ip", c.GetString("client_ip")),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
				"message": "Invalid credentials",
			})
			return
		}

		// Set token in context for use by handlers
		c.Set("auth_token", token)

		c.Next()
	}
}

// handleBasicAuth handles Basic authentication by treating the password as a GitHub PAT.
func handleBasicAuth(authHeader string, cache *auth.Cache, logger *zap.Logger) (*auth.Token, error) {
	// Remove "Basic " prefix
	encodedCreds := strings.TrimPrefix(authHeader, "Basic ")

	// Decode base64
	decodedCreds, err := base64.StdEncoding.DecodeString(encodedCreds)
	if err != nil {
		return nil, err
	}

	// Split username:password
	credentials := string(decodedCreds)
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return nil, err
	}

	username := parts[0]
	password := parts[1]

	// Check cache first
	if token := cache.Get(username, password); token != nil {
		logger.Debug("auth cache hit", zap.String("username", username))
		return token, nil
	}

	// Validate with GitHub API
	logger.Debug("auth cache miss, validating with GitHub", zap.String("username", username))
	token, err := auth.ValidateBasicAuth(username, password)
	if err != nil {
		return nil, err
	}

	// Cache the validated token
	cache.Set(username, password, token)

	return token, nil
}

// handleBearerAuth handles Bearer token authentication.
func handleBearerAuth(authHeader string, cache *auth.Cache, logger *zap.Logger) (*auth.Token, error) {
	// Remove "Bearer " prefix
	tokenValue := strings.TrimPrefix(authHeader, "Bearer ")

	// For Bearer tokens, use empty username
	username := ""

	// Check cache first
	if token := cache.Get(username, tokenValue); token != nil {
		logger.Debug("auth cache hit (bearer)", zap.String("token_prefix", tokenValue[:min(8, len(tokenValue))]))
		return token, nil
	}

	// Validate with GitHub API
	logger.Debug("auth cache miss (bearer), validating with GitHub")
	token, err := auth.ValidateBasicAuth(username, tokenValue)
	if err != nil {
		return nil, err
	}

	// Cache the validated token
	cache.Set(username, tokenValue, token)

	return token, nil
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
