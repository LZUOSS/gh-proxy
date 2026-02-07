package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery returns a middleware that recovers from panics and logs them.
// It returns a 500 Internal Server Error to the client.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := debug.Stack()

				// Log the panic with stack trace
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("stack", string(stack)),
				)

				// Return 500 Internal Server Error
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
					"message": fmt.Sprintf("%v", err),
				})
			}
		}()

		c.Next()
	}
}
