// pkg/logger/gin.go
package logger

import (
	"time"

	"github.com/gin-gonic/gin"
)

// GinMiddleware returns a gin middleware that logs requests using the structured logger
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if query != "" {
			path = path + "?" + query
		}

		// Skip logging for healthcheck endpoints if you wish
		if path == "/health" || path == "/ping" {
			return
		}

		// Log the request details
		logger := With(
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			"latency", latency,
			"ip", c.ClientIP(),
			"user-agent", c.Request.UserAgent(),
			"bytes", c.Writer.Size(),
		)

		if len(c.Errors) > 0 {
			// Append errors if any
			logger.With("errors", c.Errors.String()).Error("Request processing failed")
		} else if c.Writer.Status() >= 500 {
			logger.Error("Server error")
		} else if c.Writer.Status() >= 400 {
			logger.Warn("Client error")
		} else {
			logger.Debug("Request processed")
		}
	}
}
