// File: internal/delivery/http/middleware/auth_middleware.go
package middleware

import (
	"mikrobill/internal/port/service"
	pkg_logger "mikrobill/pkg/logger"
	"mikrobill/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware validates JWT token and sets user context
func AuthMiddleware(jwtService *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, 401, "Authorization header is required", utils.ErrUnauthorized)
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, 401, "Invalid authorization header format", utils.ErrInvalidToken)
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			pkg_logger.Error("Token validation failed", zap.Error(err))
			utils.ErrorResponse(c, 401, "Invalid or expired token", utils.ErrInvalidToken)
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		// Set single role as string
		c.Set("user_role", claims.Role)
		// Set role as array for Casbin compatibility
		c.Set("user_roles", []string{claims.Role})

		pkg_logger.Debug("User authenticated",
			zap.Int64("user_id", claims.UserID),
			zap.String("email", claims.Email),
			zap.String("role", claims.Role),
		)

		c.Next()
	}
}

// RequireRole checks if user has required role
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			utils.ErrorResponse(c, 403, "User role not found", utils.ErrForbidden)
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			utils.ErrorResponse(c, 403, "Invalid user role", utils.ErrForbidden)
			c.Abort()
			return
		}

		// Check if user has required role
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		pkg_logger.Warn("Insufficient permissions",
			zap.String("user_role", userRole),
			zap.Strings("allowed_roles", allowedRoles),
		)

		utils.ErrorResponse(c, 403, "Insufficient permissions", utils.ErrForbidden)
		c.Abort()
	}
}