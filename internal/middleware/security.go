package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking attacks
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Prevent information leakage
		c.Header("X-Powered-By", "")

		// Content Security Policy (basic)
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Force HTTPS in browsers (when served over HTTPS)
		// Commented out as it should only be enabled when running with TLS
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
