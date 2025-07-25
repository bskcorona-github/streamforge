package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery middleware for handling panics
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logger.Error("Panic recovered",
				zap.String("error", err),
				zap.String("stack", string(debug.Stack())),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	})
} 