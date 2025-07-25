package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LogData represents the structure of incoming log data
type LogData struct {
	Timestamp  string            `json:"timestamp"`
	Level      string            `json:"level" binding:"required"`
	Service    string            `json:"service" binding:"required"`
	Message    string            `json:"message" binding:"required"`
	TraceID    string            `json:"trace_id,omitempty"`
	SpanID     string            `json:"span_id,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// SendLogs handles log data submission
func SendLogs(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	var logData LogData
	if err := c.ShouldBindJSON(&logData); err != nil {
		logger.Error("Invalid log data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid log data",
			"details": err.Error(),
		})
		return
	}

	// タイムスタンプの設定
	if logData.Timestamp == "" {
		logData.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	// ログレベルの検証
	validLevels := map[string]bool{
		"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true, "FATAL": true,
		"debug": true, "info": true, "warn": true, "error": true, "fatal": true,
	}
	if !validLevels[logData.Level] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid log level",
			"valid_levels": []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
		})
		return
	}

	// TODO: ログデータをデータベースに保存
	// if err := saveLogToDatabase(logData); err != nil {
	//     logger.Error("Failed to save log", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to save log",
	//     })
	//     return
	// }

	logger.Info("Log received",
		zap.String("service", logData.Service),
		zap.String("level", logData.Level),
		zap.String("message", logData.Message),
		zap.String("trace_id", logData.TraceID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"id":     generateID(),
		"status": "accepted",
		"log":    logData,
	})
}

// GetLogs handles log data retrieval
func GetLogs(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	// クエリパラメータの取得
	service := c.Query("service")
	level := c.Query("level")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	limit := c.DefaultQuery("limit", "100")

	// TODO: データベースからログを取得
	// logs, err := getLogsFromDatabase(service, level, startTime, endTime, limit)
	// if err != nil {
	//     logger.Error("Failed to retrieve logs", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to retrieve logs",
	//     })
	//     return
	// }

	// 仮のレスポンス
	logs := []LogData{
		{
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			Level:      level,
			Service:    service,
			Message:    "Sample log message",
			TraceID:    "sample-trace-id",
			SpanID:     "sample-span-id",
			Attributes: map[string]string{"test": "data"},
		},
	}

	logger.Debug("Logs retrieved",
		zap.String("service", service),
		zap.String("level", level),
		zap.Int("count", len(logs)),
	)

	c.JSON(http.StatusOK, logs)
}

// StreamLogs handles real-time log streaming
func StreamLogs(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	// TODO: WebSocketまたはServer-Sent Eventsによるストリーミング実装
	logger.Debug("Logs streaming requested")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Logs streaming endpoint - TODO: implement",
	})
} 