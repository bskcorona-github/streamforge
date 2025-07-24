package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/service"
	"github.com/bskcorona-github/streamforge/packages/proto/streamforge/v1"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler はHTTPハンドラーを実装します
type Handler struct {
	service *service.Service
	logger  *zap.Logger
}

// NewHandler は新しいHandlerインスタンスを作成します
func NewHandler(svc *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

// GetMetrics はメトリクスを取得するハンドラーです
func (h *Handler) GetMetrics(c *gin.Context) {
	query := c.Query("query")
	timeRange := h.parseTimeRange(c)
	pagination := h.parsePagination(c)

	req := &v1.GetMetricsRequest{
		Query:      query,
		TimeRange:  timeRange,
		Pagination: pagination,
	}

	resp, err := h.service.GetMetrics(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMetric は特定のメトリクスを取得するハンドラーです
func (h *Handler) GetMetric(c *gin.Context) {
	name := c.Param("name")
	timeRange := h.parseTimeRange(c)

	req := &v1.GetMetricRequest{
		Name:      name,
		TimeRange: timeRange,
	}

	resp, err := h.service.GetMetric(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get metric", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTimeSeries は時系列データを取得するハンドラーです
func (h *Handler) GetTimeSeries(c *gin.Context) {
	metricName := c.Param("name")
	timeRange := h.parseTimeRange(c)
	tags := h.parseTags(c)
	aggregation := c.DefaultQuery("aggregation", "avg")
	intervalStr := c.DefaultQuery("interval", "5m")

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interval format"})
		return
	}

	req := &v1.GetTimeSeriesRequest{
		MetricName: metricName,
		TimeRange:  timeRange,
		Tags:       tags,
		Aggregation: aggregation,
		Interval:    durationpb.New(interval),
	}

	resp, err := h.service.GetTimeSeries(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get time series", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetLogs はログを取得するハンドラーです
func (h *Handler) GetLogs(c *gin.Context) {
	query := c.Query("query")
	timeRange := h.parseTimeRange(c)
	minLevel := h.parseLogLevel(c)
	pagination := h.parsePagination(c)

	req := &v1.GetLogsRequest{
		Query:      query,
		TimeRange:  timeRange,
		MinLevel:   minLevel,
		Pagination: pagination,
	}

	resp, err := h.service.GetLogs(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SearchLogs はログを検索するハンドラーです
func (h *Handler) SearchLogs(c *gin.Context) {
	query := c.Query("query")
	timeRange := h.parseTimeRange(c)
	minLevel := h.parseLogLevel(c)
	pagination := h.parsePagination(c)

	req := &v1.SearchLogsRequest{
		Query:      query,
		TimeRange:  timeRange,
		MinLevel:   minLevel,
		Pagination: pagination,
	}

	resp, err := h.service.SearchLogs(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to search logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTraces はトレースを取得するハンドラーです
func (h *Handler) GetTraces(c *gin.Context) {
	serviceName := c.Query("service_name")
	operationName := c.Query("operation_name")
	timeRange := h.parseTimeRange(c)
	status := h.parseStatus(c)
	pagination := h.parsePagination(c)

	req := &v1.GetTracesRequest{
		ServiceName:   serviceName,
		OperationName: operationName,
		TimeRange:     timeRange,
		Status:        status,
		Pagination:    pagination,
	}

	resp, err := h.service.GetTraces(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get traces", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTrace は特定のトレースを取得するハンドラーです
func (h *Handler) GetTrace(c *gin.Context) {
	traceID := c.Param("traceId")

	req := &v1.GetTraceRequest{
		TraceId: traceID,
	}

	resp, err := h.service.GetTrace(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get trace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTraceGraph はトレースグラフを取得するハンドラーです
func (h *Handler) GetTraceGraph(c *gin.Context) {
	traceID := c.Param("traceId")

	req := &v1.GetTraceGraphRequest{
		TraceId: traceID,
	}

	resp, err := h.service.GetTraceGraph(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get trace graph", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetAlerts はアラートを取得するハンドラーです
func (h *Handler) GetAlerts(c *gin.Context) {
	status := c.Query("status")
	severity := c.Query("severity")
	pagination := h.parsePagination(c)

	req := &v1.GetAlertsRequest{
		Status:     status,
		Severity:   severity,
		Pagination: pagination,
	}

	resp, err := h.service.GetAlerts(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetAlert は特定のアラートを取得するハンドラーです
func (h *Handler) GetAlert(c *gin.Context) {
	id := c.Param("id")

	req := &v1.GetAlertRequest{
		Id: id,
	}

	resp, err := h.service.GetAlert(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateAlert はアラートを作成するハンドラーです
func (h *Handler) CreateAlert(c *gin.Context) {
	var req v1.CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateAlert(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// UpdateAlert はアラートを更新するハンドラーです
func (h *Handler) UpdateAlert(c *gin.Context) {
	id := c.Param("id")
	var req v1.UpdateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Id = id

	resp, err := h.service.UpdateAlert(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteAlert はアラートを削除するハンドラーです
func (h *Handler) DeleteAlert(c *gin.Context) {
	id := c.Param("id")

	req := &v1.DeleteAlertRequest{
		Id: id,
	}

	resp, err := h.service.DeleteAlert(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to delete alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetDashboards はダッシュボードを取得するハンドラーです
func (h *Handler) GetDashboards(c *gin.Context) {
	query := c.Query("query")
	pagination := h.parsePagination(c)

	req := &v1.GetDashboardsRequest{
		Query:      query,
		Pagination: pagination,
	}

	resp, err := h.service.GetDashboards(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get dashboards", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetDashboard は特定のダッシュボードを取得するハンドラーです
func (h *Handler) GetDashboard(c *gin.Context) {
	id := c.Param("id")

	req := &v1.GetDashboardRequest{
		Id: id,
	}

	resp, err := h.service.GetDashboard(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateDashboard はダッシュボードを作成するハンドラーです
func (h *Handler) CreateDashboard(c *gin.Context) {
	var req v1.CreateDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateDashboard(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// UpdateDashboard はダッシュボードを更新するハンドラーです
func (h *Handler) UpdateDashboard(c *gin.Context) {
	id := c.Param("id")
	var req v1.UpdateDashboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Id = id

	resp, err := h.service.UpdateDashboard(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteDashboard はダッシュボードを削除するハンドラーです
func (h *Handler) DeleteDashboard(c *gin.Context) {
	id := c.Param("id")

	req := &v1.DeleteDashboardRequest{
		Id: id,
	}

	resp, err := h.service.DeleteDashboard(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to delete dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ヘルパー関数

func (h *Handler) parseTimeRange(c *gin.Context) *v1.TimeRange {
	startStr := c.Query("start_time")
	endStr := c.Query("end_time")

	if startStr == "" && endStr == "" {
		// デフォルトで過去1時間
		end := time.Now()
		start := end.Add(-1 * time.Hour)
		return &v1.TimeRange{
			StartTime: timestamppb.New(start),
			EndTime:   timestamppb.New(end),
		}
	}

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			h.logger.Warn("Invalid start_time format", zap.String("start_time", startStr))
		}
	}

	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			h.logger.Warn("Invalid end_time format", zap.String("end_time", endStr))
		}
	}

	return &v1.TimeRange{
		StartTime: timestamppb.New(start),
		EndTime:   timestamppb.New(end),
	}
}

func (h *Handler) parsePagination(c *gin.Context) *v1.Pagination {
	pageSizeStr := c.DefaultQuery("page_size", "50")
	pageTokenStr := c.DefaultQuery("page_token", "0")

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 50
	}

	pageToken, err := strconv.Atoi(pageTokenStr)
	if err != nil {
		pageToken = 0
	}

	return &v1.Pagination{
		PageSize:  int32(pageSize),
		PageToken: int32(pageToken),
	}
}

func (h *Handler) parseTags(c *gin.Context) []*v1.Tag {
	tags := make([]*v1.Tag, 0)
	
	// クエリパラメータからタグを解析
	for key, values := range c.Request.URL.Query() {
		if key == "tag" {
			for _, value := range values {
				// "key=value" 形式を想定
				tag := &v1.Tag{
					Key:   key,
					Value: value,
				}
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

func (h *Handler) parseLogLevel(c *gin.Context) v1.LogLevel {
	levelStr := c.DefaultQuery("min_level", "INFO")
	
	switch levelStr {
	case "TRACE":
		return v1.LogLevel_LOG_LEVEL_TRACE
	case "DEBUG":
		return v1.LogLevel_LOG_LEVEL_DEBUG
	case "INFO":
		return v1.LogLevel_LOG_LEVEL_INFO
	case "WARN":
		return v1.LogLevel_LOG_LEVEL_WARN
	case "ERROR":
		return v1.LogLevel_LOG_LEVEL_ERROR
	case "FATAL":
		return v1.LogLevel_LOG_LEVEL_FATAL
	default:
		return v1.LogLevel_LOG_LEVEL_INFO
	}
}

func (h *Handler) parseStatus(c *gin.Context) v1.Status {
	statusStr := c.Query("status")
	
	switch statusStr {
	case "OK":
		return v1.Status_STATUS_OK
	case "ERROR":
		return v1.Status_STATUS_ERROR
	case "INVALID_ARGUMENT":
		return v1.Status_STATUS_INVALID_ARGUMENT
	case "NOT_FOUND":
		return v1.Status_STATUS_NOT_FOUND
	case "UNAUTHENTICATED":
		return v1.Status_STATUS_UNAUTHENTICATED
	case "PERMISSION_DENIED":
		return v1.Status_STATUS_PERMISSION_DENIED
	case "UNAVAILABLE":
		return v1.Status_STATUS_UNAVAILABLE
	default:
		return v1.Status_STATUS_UNSPECIFIED
	}
} 