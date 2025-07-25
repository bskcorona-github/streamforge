package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TraceData represents the structure of incoming trace data
type TraceData struct {
	TraceID       string            `json:"trace_id" binding:"required"`
	SpanID        string            `json:"span_id" binding:"required"`
	OperationName string            `json:"operation_name" binding:"required"`
	StartTime     string            `json:"start_time"`
	Duration      int64             `json:"duration"`
	ServiceName   string            `json:"service_name" binding:"required"`
	Tags          map[string]string `json:"tags,omitempty"`
}

// SendTrace handles trace data submission
func SendTrace(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	var traceData TraceData
	if err := c.ShouldBindJSON(&traceData); err != nil {
		logger.Error("Invalid trace data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid trace data",
			"details": err.Error(),
		})
		return
	}

	// タイムスタンプの設定
	if traceData.StartTime == "" {
		traceData.StartTime = time.Now().UTC().Format(time.RFC3339)
	}

	// TODO: トレースデータをデータベースに保存
	// if err := saveTraceToDatabase(traceData); err != nil {
	//     logger.Error("Failed to save trace", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to save trace",
	//     })
	//     return
	// }

	logger.Info("Trace received",
		zap.String("trace_id", traceData.TraceID),
		zap.String("span_id", traceData.SpanID),
		zap.String("operation", traceData.OperationName),
		zap.String("service", traceData.ServiceName),
	)

	c.JSON(http.StatusCreated, gin.H{
		"id":     generateID(),
		"status": "accepted",
		"trace":  traceData,
	})
}

// GetTrace handles trace data retrieval by trace ID
func GetTrace(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	traceID := c.Param("trace_id")
	if traceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Trace ID is required",
		})
		return
	}

	// TODO: データベースからトレースを取得
	// trace, err := getTraceFromDatabase(traceID)
	// if err != nil {
	//     logger.Error("Failed to retrieve trace", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to retrieve trace",
	//     })
	//     return
	// }

	// 仮のレスポンス
	trace := TraceData{
		TraceID:       traceID,
		SpanID:        "sample-span-id",
		OperationName: "sample_operation",
		StartTime:     time.Now().UTC().Format(time.RFC3339),
		Duration:      100,
		ServiceName:   "sample-service",
		Tags:          map[string]string{"test": "data"},
	}

	logger.Debug("Trace retrieved", zap.String("trace_id", traceID))

	c.JSON(http.StatusOK, trace)
}

// StreamTraces handles real-time trace streaming
func StreamTraces(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	// TODO: WebSocketまたはServer-Sent Eventsによるストリーミング実装
	logger.Debug("Traces streaming requested")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Traces streaming endpoint - TODO: implement",
	})
} 