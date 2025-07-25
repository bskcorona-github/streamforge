package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ServiceData represents the structure of service data
type ServiceData struct {
	Name        string            `json:"name" binding:"required"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Endpoint    string            `json:"endpoint"`
	Status      string            `json:"status,omitempty"`
	LastSeen    string            `json:"last_seen,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// RegisterService handles service registration
func RegisterService(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	var serviceData ServiceData
	if err := c.ShouldBindJSON(&serviceData); err != nil {
		logger.Error("Invalid service data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid service data",
			"details": err.Error(),
		})
		return
	}

	// サービス名の検証
	if serviceData.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	// タイムスタンプの設定
	now := time.Now().UTC().Format(time.RFC3339)
	serviceData.LastSeen = now
	serviceData.Status = "active"

	// TODO: サービスをデータベースに登録
	// if err := registerServiceInDatabase(serviceData); err != nil {
	//     logger.Error("Failed to register service", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to register service",
	//     })
	//     return
	// }

	logger.Info("Service registered",
		zap.String("service_name", serviceData.Name),
		zap.String("version", serviceData.Version),
		zap.String("endpoint", serviceData.Endpoint),
	)

	c.JSON(http.StatusCreated, serviceData)
}

// ListServices handles service listing
func ListServices(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	// クエリパラメータの取得
	status := c.Query("status")
	limit := c.DefaultQuery("limit", "100")

	// TODO: データベースからサービスを取得
	// services, err := getServicesFromDatabase(status, limit)
	// if err != nil {
	//     logger.Error("Failed to retrieve services", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to retrieve services",
	//     })
	//     return
	// }

	// 仮のレスポンス
	services := []ServiceData{
		{
			Name:        "api-gateway",
			Version:     "1.0.0",
			Description: "API Gateway Service",
			Endpoint:    "http://localhost:8080",
			Status:      "active",
			LastSeen:    time.Now().UTC().Format(time.RFC3339),
			Labels:      map[string]string{"type": "gateway"},
		},
		{
			Name:        "collector",
			Version:     "1.0.0",
			Description: "Data Collector Service",
			Endpoint:    "http://localhost:8081",
			Status:      "active",
			LastSeen:    time.Now().UTC().Format(time.RFC3339),
			Labels:      map[string]string{"type": "collector"},
		},
	}

	logger.Debug("Services retrieved",
		zap.String("status", status),
		zap.Int("count", len(services)),
	)

	c.JSON(http.StatusOK, services)
}

// UpdateService handles service updates
func UpdateService(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	serviceName := c.Param("service_name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	var serviceData ServiceData
	if err := c.ShouldBindJSON(&serviceData); err != nil {
		logger.Error("Invalid service data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid service data",
			"details": err.Error(),
		})
		return
	}

	serviceData.Name = serviceName
	serviceData.LastSeen = time.Now().UTC().Format(time.RFC3339)

	// TODO: サービスをデータベースで更新
	// if err := updateServiceInDatabase(serviceData); err != nil {
	//     logger.Error("Failed to update service", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to update service",
	//     })
	//     return
	// }

	logger.Info("Service updated", zap.String("service_name", serviceName))

	c.JSON(http.StatusOK, serviceData)
}

// UnregisterService handles service unregistration
func UnregisterService(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	serviceName := c.Param("service_name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	// TODO: サービスをデータベースから削除
	// if err := unregisterServiceFromDatabase(serviceName); err != nil {
	//     logger.Error("Failed to unregister service", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to unregister service",
	//     })
	//     return
	// }

	logger.Info("Service unregistered", zap.String("service_name", serviceName))

	c.JSON(http.StatusOK, gin.H{
		"message": "Service unregistered successfully",
		"name":    serviceName,
	})
}

// Heartbeat handles service heartbeat
func Heartbeat(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	serviceName := c.Param("service_name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	// ハートビートデータの取得
	var heartbeatData struct {
		Status   string            `json:"status,omitempty"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}
	if err := c.ShouldBindJSON(&heartbeatData); err != nil {
		logger.Error("Invalid heartbeat data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid heartbeat data",
			"details": err.Error(),
		})
		return
	}

	// TODO: ハートビートをデータベースに記録
	// if err := recordHeartbeat(serviceName, heartbeatData); err != nil {
	//     logger.Error("Failed to record heartbeat", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to record heartbeat",
	//     })
	//     return
	// }

	logger.Debug("Heartbeat received",
		zap.String("service_name", serviceName),
		zap.String("status", heartbeatData.Status),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Heartbeat recorded",
		"service": serviceName,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
} 