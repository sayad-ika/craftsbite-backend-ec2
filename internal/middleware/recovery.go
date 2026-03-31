package middleware

import (
	"craftsbite-backend/internal/utils"
	"craftsbite-backend/pkg/logger"
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
				)

				// Return 500 error
				utils.ErrorResponse(c, 500, "INTERNAL_SERVER_ERROR", fmt.Sprintf("Internal server error: %v", err))
				c.Abort()
			}
		}()

		c.Next()
	}
}
