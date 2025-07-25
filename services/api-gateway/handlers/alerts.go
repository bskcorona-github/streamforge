package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AlertData represents the structure of alert data
type AlertData struct {
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Severity    string            `json:"severity" binding:"required"`
	Service     string            `json:"service" binding:"required"`
	Condition   string            `json:"condition" binding:"required"`
	Status      string            `json:"status,omitempty"`
	CreatedAt   string            `json:"created_at,omitempty"`
	UpdatedAt   string            `json:"updated_at,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// CreateAlert handles alert creation
func CreateAlert(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	var alertData AlertData
	if err := c.ShouldBindJSON(&alertData); err != nil {
		logger.Error("Invalid alert data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid alert data",
			"details": err.Error(),
		})
		return
	}

	// アラートの検証
	if err := validateAlert(alertData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// タイムスタンプの設定
	now := time.Now().UTC().Format(time.RFC3339)
	alertData.ID = generateID()
	alertData.CreatedAt = now
	alertData.UpdatedAt = now
	alertData.Status = "active"

	// TODO: アラートをデータベースに保存
	// if err := saveAlertToDatabase(alertData); err != nil {
	//     logger.Error("Failed to save alert", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to save alert",
	//     })
	//     return
	// }

	logger.Info("Alert created",
		zap.String("alert_id", alertData.ID),
		zap.String("name", alertData.Name),
		zap.String("service", alertData.Service),
		zap.String("severity", alertData.Severity),
	)

	c.JSON(http.StatusCreated, alertData)
}

// ListAlerts handles alert listing
func ListAlerts(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	// クエリパラメータの取得
	service := c.Query("service")
	severity := c.Query("severity")
	status := c.Query("status")
	limit := c.DefaultQuery("limit", "100")

	// TODO: データベースからアラートを取得
	// alerts, err := getAlertsFromDatabase(service, severity, status, limit)
	// if err != nil {
	//     logger.Error("Failed to retrieve alerts", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to retrieve alerts",
	//     })
	//     return
	// }

	// 仮のレスポンス
	alerts := []AlertData{
		{
			ID:          generateID(),
			Name:        "High CPU Usage",
			Description: "CPU usage is above 90%",
			Severity:    "high",
			Service:     service,
			Condition:   "cpu_usage > 90",
			Status:      "active",
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
			UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
		},
	}

	logger.Debug("Alerts retrieved",
		zap.String("service", service),
		zap.String("severity", severity),
		zap.Int("count", len(alerts)),
	)

	c.JSON(http.StatusOK, alerts)
}

// GetAlert handles alert retrieval by ID
func GetAlert(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	alertID := c.Param("alert_id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alert ID is required",
		})
		return
	}

	// TODO: データベースからアラートを取得
	// alert, err := getAlertFromDatabase(alertID)
	// if err != nil {
	//     logger.Error("Failed to retrieve alert", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to retrieve alert",
	//     })
	//     return
	// }

	// 仮のレスポンス
	alert := AlertData{
		ID:          alertID,
		Name:        "Sample Alert",
		Description: "This is a sample alert",
		Severity:    "medium",
		Service:     "sample-service",
		Condition:   "sample_condition > 50",
		Status:      "active",
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	logger.Debug("Alert retrieved", zap.String("alert_id", alertID))

	c.JSON(http.StatusOK, alert)
}

// UpdateAlert handles alert updates
func UpdateAlert(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	alertID := c.Param("alert_id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alert ID is required",
		})
		return
	}

	var alertData AlertData
	if err := c.ShouldBindJSON(&alertData); err != nil {
		logger.Error("Invalid alert data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid alert data",
			"details": err.Error(),
		})
		return
	}

	alertData.ID = alertID
	alertData.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	// TODO: アラートをデータベースで更新
	// if err := updateAlertInDatabase(alertData); err != nil {
	//     logger.Error("Failed to update alert", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to update alert",
	//     })
	//     return
	// }

	logger.Info("Alert updated", zap.String("alert_id", alertID))

	c.JSON(http.StatusOK, alertData)
}

// DeleteAlert handles alert deletion
func DeleteAlert(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	alertID := c.Param("alert_id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alert ID is required",
		})
		return
	}

	// TODO: アラートをデータベースから削除
	// if err := deleteAlertFromDatabase(alertID); err != nil {
	//     logger.Error("Failed to delete alert", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to delete alert",
	//     })
	//     return
	// }

	logger.Info("Alert deleted", zap.String("alert_id", alertID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert deleted successfully",
		"id":      alertID,
	})
}

// StreamAlerts handles real-time alert streaming
func StreamAlerts(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	// TODO: WebSocketまたはServer-Sent Eventsによるストリーミング実装
	logger.Debug("Alerts streaming requested")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Alerts streaming endpoint - TODO: implement",
	})
}

// validateAlert validates alert data
func validateAlert(alert AlertData) error {
	// 重要度の検証
	validSeverities := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	if !validSeverities[alert.Severity] {
		return fmt.Errorf("invalid severity: %s", alert.Severity)
	}

	return nil
} 