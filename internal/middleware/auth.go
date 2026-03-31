package middleware

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		tokenString, err := c.Cookie("auth_token")
		if err != nil {
			utils.ErrorResponse(c, 401, "UNAUTHORIZED", "Authentication required")
			c.Abort()
			return
		}

		// Check Bearer prefix
		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			utils.ErrorResponse(c, 401, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user claims in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRoles checks if the user has one of the required roles
func RequireRoles(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			utils.ErrorResponse(c, 403, "FORBIDDEN", "User role not found in context")
			c.Abort()
			return
		}

		userRoleStr := userRole.(string)

		// Check if user role matches any of the required roles
		hasPermission := false
		for _, role := range roles {
			if userRoleStr == role.String() {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			utils.ErrorResponse(c, 403, "FORBIDDEN", "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}
