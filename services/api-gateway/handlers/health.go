package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthCheck handles health check requests
func HealthCheck(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	// 基本的なヘルスチェック情報
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "api-gateway",
		"version":   "1.0.0",
	}

	// データベース接続チェック（TODO: 実装予定）
	// if err := checkDatabaseConnection(); err != nil {
	//     health["status"] = "unhealthy"
	//     health["database"] = "disconnected"
	//     logger.Error("Database health check failed", zap.Error(err))
	//     c.JSON(http.StatusServiceUnavailable, health)
	//     return
	// }

	// Redis接続チェック（TODO: 実装予定）
	// if err := checkRedisConnection(); err != nil {
	//     health["status"] = "unhealthy"
	//     health["redis"] = "disconnected"
	//     logger.Error("Redis health check failed", zap.Error(err))
	//     c.JSON(http.StatusServiceUnavailable, health)
	//     return
	// }

	logger.Debug("Health check requested", zap.String("status", health["status"].(string)))
	c.JSON(http.StatusOK, health)
} 