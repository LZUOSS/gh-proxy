package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/LZUOSS/gh-proxy/internal/ratelimit"
)

// RateLimit returns a middleware that enforces rate limiting per IP address.
// It uses the client IP from the context (set by RealIP middleware).
func RateLimit(limiter *ratelimit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP from context
		clientIP := c.GetString("client_ip")
		if clientIP == "" {
			// Fallback to Gin's ClientIP if not set
			clientIP = c.ClientIP()
		}

		// Check if request is allowed
		if !limiter.Allow(clientIP) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
