package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func TrackingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Store duration in context for logging
		c.Set("view_duration", int(duration.Seconds()))
	}
}
