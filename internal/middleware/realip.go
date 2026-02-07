package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kexi/github-reverse-proxy/internal/util"
)

// RealIP returns a middleware that extracts the real client IP address
// and sets it in the Gin context as "client_ip".
func RealIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract real IP using util.ExtractRealIP
		ip := util.ExtractRealIP(c.Request)

		// Set client IP in context for use by other middleware
		c.Set("client_ip", ip)

		c.Next()
	}
}
