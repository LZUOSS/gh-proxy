package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logging returns a middleware that logs HTTP requests.
// It logs: method, path, status, duration, IP, user-agent, and request ID.
func Logging(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Get request ID (set by Gin's request ID middleware if available)
		requestID := c.GetString("X-Request-ID")
		if requestID == "" {
			requestID = c.GetHeader("X-Request-ID")
		}

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Get client IP (should be set by RealIP middleware)
		clientIP := c.GetString("client_ip")
		if clientIP == "" {
			clientIP = c.ClientIP()
		}

		// Log request details
		logger.Info("http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("ip", clientIP),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("request_id", requestID),
			zap.Int("response_size", c.Writer.Size()),
		)
	}
}
