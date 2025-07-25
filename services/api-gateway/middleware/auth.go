package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"streamforge/services/api-gateway/auth"
)

func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := c.MustGet("logger").(*zap.Logger)

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			logger.Warn("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			logger.Warn("Invalid token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		logger.Debug("User authenticated",
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.String("role", claims.Role),
		)

		c.Next()
	}
}

// OptionalAuthMiddleware allows requests without authentication but sets user info if provided
func OptionalAuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := c.MustGet("logger").(*zap.Logger)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := tokenParts[1]

		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			logger.Debug("Optional auth failed", zap.Error(err))
			c.Next()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		logger.Debug("Optional user authenticated",
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username),
		)

		c.Next()
	}
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := c.MustGet("logger").(*zap.Logger)
		userRole, exists := c.Get("role")
		
		if !exists {
			logger.Warn("User role not found in context")
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			c.Abort()
			return
		}

		if userRole != requiredRole {
			logger.Warn("Insufficient permissions",
				zap.String("user_role", userRole.(string)),
				zap.String("required_role", requiredRole),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
} 