package middleware

import (
	"go-template/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set the request ID in the context
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// RequestLog logs requests using the structured logger
func RequestLog() gin.HandlerFunc {
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
		logger := logger.With(
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
