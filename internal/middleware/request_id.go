package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware generates a unique request ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate UUID for request
		requestID := uuid.New().String()

		// Set in context
		c.Set("request_id", requestID)

		// Set in response header
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}
