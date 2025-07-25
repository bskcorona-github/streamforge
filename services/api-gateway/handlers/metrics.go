package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MetricData represents the structure of incoming metric data
type MetricData struct {
	Service   string            `json:"service" binding:"required"`
	Metric    string            `json:"metric" binding:"required"`
	Value     float64           `json:"value" binding:"required"`
	Timestamp string            `json:"timestamp"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// SendMetrics handles metric data submission
func SendMetrics(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	var metricData MetricData
	if err := c.ShouldBindJSON(&metricData); err != nil {
		logger.Error("Invalid metric data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid metric data",
			"details": err.Error(),
		})
		return
	}

	// タイムスタンプの設定
	if metricData.Timestamp == "" {
		metricData.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	// TODO: メトリクスデータをデータベースに保存
	// if err := saveMetricToDatabase(metricData); err != nil {
	//     logger.Error("Failed to save metric", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to save metric",
	//     })
	//     return
	// }

	// TODO: ストリームプロセッサーにメトリクスを送信
	// if err := sendMetricToStreamProcessor(metricData); err != nil {
	//     logger.Error("Failed to send metric to stream processor", zap.Error(err))
	//     // エラーをログに記録するが、レスポンスは成功とする
	// }

	logger.Info("Metric received",
		zap.String("service", metricData.Service),
		zap.String("metric", metricData.Metric),
		zap.Float64("value", metricData.Value),
	)

	c.JSON(http.StatusCreated, gin.H{
		"id":     generateID(),
		"status": "accepted",
		"metric": metricData,
	})
}

// GetMetrics handles metric data retrieval
func GetMetrics(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	// クエリパラメータの取得
	service := c.Query("service")
	metric := c.Query("metric")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	limit := c.DefaultQuery("limit", "100")

	// TODO: データベースからメトリクスを取得
	// metrics, err := getMetricsFromDatabase(service, metric, startTime, endTime, limit)
	// if err != nil {
	//     logger.Error("Failed to retrieve metrics", zap.Error(err))
	//     c.JSON(http.StatusInternalServerError, gin.H{
	//         "error": "Failed to retrieve metrics",
	//     })
	//     return
	// }

	// 仮のレスポンス
	metrics := []MetricData{
		{
			Service:   service,
			Metric:    metric,
			Value:     100.0,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Labels:    map[string]string{"test": "data"},
		},
	}

	logger.Debug("Metrics retrieved",
		zap.String("service", service),
		zap.String("metric", metric),
		zap.Int("count", len(metrics)),
	)

	c.JSON(http.StatusOK, metrics)
}

// GetTimeSeries handles time series data retrieval
func GetTimeSeries(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	// TODO: 時系列データの取得実装
	logger.Debug("Time series requested")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Time series endpoint - TODO: implement",
	})
}

// StreamMetrics handles real-time metric streaming
func StreamMetrics(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	// TODO: WebSocketまたはServer-Sent Eventsによるストリーミング実装
	logger.Debug("Metrics streaming requested")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Metrics streaming endpoint - TODO: implement",
	})
}

// generateID generates a unique ID for the metric
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
} 