package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/LZUOSS/gh-proxy/internal/metrics"
)

// Metrics returns a middleware that collects Prometheus metrics.
// It records request count, duration, and response size.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Increment active connections
		metrics.IncrementActiveConnections()
		defer metrics.DecrementActiveConnections()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get response status
		status := strconv.Itoa(c.Writer.Status())

		// Record metrics
		metrics.RecordRequest(c.Request.Method, c.Request.URL.Path, status)
		metrics.RecordRequestDuration(c.Request.Method, c.Request.URL.Path, duration)

		// Record response size if available
		if size := c.Writer.Size(); size > 0 {
			metrics.RecordResponseSize(c.Request.URL.Path, float64(size))
		}
	}
}
