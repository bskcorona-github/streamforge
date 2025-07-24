package handler

import (
	"context"
	"time"

	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/service"
	"github.com/bskcorona-github/streamforge/packages/proto/streamforge/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

// GrpcHandler implements the gRPC API service
type GrpcHandler struct {
	v1.UnimplementedApiServiceServer
	service service.Service
	logger  *zap.Logger
}

// NewGrpcHandler creates a new gRPC handler
func NewGrpcHandler(service service.Service, logger *zap.Logger) *GrpcHandler {
	return &GrpcHandler{
		service: service,
		logger:  logger,
	}
}

// GetMetrics retrieves metrics based on query and time range
func (h *GrpcHandler) GetMetrics(ctx context.Context, req *v1.GetMetricsRequest) (*v1.GetMetricsResponse, error) {
	h.logger.Info("GetMetrics called", zap.String("query", req.Query))
	
	metrics, pagination, err := h.service.GetMetrics(ctx, req.Query, req.TimeRange, req.Pagination)
	if err != nil {
		h.logger.Error("Failed to get metrics", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get metrics: %v", err)
	}
	
	return &v1.GetMetricsResponse{
		Metrics:    metrics,
		Pagination: pagination,
	}, nil
}

// GetMetric retrieves a specific metric by name
func (h *GrpcHandler) GetMetric(ctx context.Context, req *v1.GetMetricRequest) (*v1.GetMetricResponse, error) {
	h.logger.Info("GetMetric called", zap.String("name", req.Name))
	
	metric, err := h.service.GetMetric(ctx, req.Name, req.TimeRange)
	if err != nil {
		h.logger.Error("Failed to get metric", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get metric: %v", err)
	}
	
	return &v1.GetMetricResponse{
		Metric: metric,
	}, nil
}

// GetTimeSeries retrieves time series data for a metric
func (h *GrpcHandler) GetTimeSeries(ctx context.Context, req *v1.GetTimeSeriesRequest) (*v1.GetTimeSeriesResponse, error) {
	h.logger.Info("GetTimeSeries called", zap.String("metric_name", req.MetricName))
	
	interval := time.Duration(0)
	if req.Interval != nil {
		interval = req.Interval.AsDuration()
	}
	
	series, err := h.service.GetTimeSeries(ctx, req.MetricName, req.TimeRange, req.Tags, req.Aggregation, interval)
	if err != nil {
		h.logger.Error("Failed to get time series", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get time series: %v", err)
	}
	
	return &v1.GetTimeSeriesResponse{
		Series: series,
	}, nil
}

// GetLogs retrieves logs based on query and time range
func (h *GrpcHandler) GetLogs(ctx context.Context, req *v1.GetLogsRequest) (*v1.GetLogsResponse, error) {
	h.logger.Info("GetLogs called", zap.String("query", req.Query))
	
	logs, pagination, err := h.service.GetLogs(ctx, req.Query, req.TimeRange, req.MinLevel, req.Pagination)
	if err != nil {
		h.logger.Error("Failed to get logs", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get logs: %v", err)
	}
	
	return &v1.GetLogsResponse{
		Logs:       logs,
		Pagination: pagination,
	}, nil
}

// SearchLogs performs a search query on logs
func (h *GrpcHandler) SearchLogs(ctx context.Context, req *v1.SearchLogsRequest) (*v1.SearchLogsResponse, error) {
	h.logger.Info("SearchLogs called", zap.String("query", req.Query))
	
	logs, pagination, err := h.service.SearchLogs(ctx, req.Query, req.TimeRange, req.MinLevel, req.Pagination)
	if err != nil {
		h.logger.Error("Failed to search logs", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to search logs: %v", err)
	}
	
	return &v1.SearchLogsResponse{
		Logs:       logs,
		Pagination: pagination,
	}, nil
}

// GetTraces retrieves traces based on filters
func (h *GrpcHandler) GetTraces(ctx context.Context, req *v1.GetTracesRequest) (*v1.GetTracesResponse, error) {
	h.logger.Info("GetTraces called", zap.String("service_name", req.ServiceName))
	
	traces, pagination, err := h.service.GetTraces(ctx, req.ServiceName, req.OperationName, req.TimeRange, req.Status, req.Pagination)
	if err != nil {
		h.logger.Error("Failed to get traces", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get traces: %v", err)
	}
	
	return &v1.GetTracesResponse{
		Traces:     traces,
		Pagination: pagination,
	}, nil
}

// GetTrace retrieves a specific trace by ID
func (h *GrpcHandler) GetTrace(ctx context.Context, req *v1.GetTraceRequest) (*v1.GetTraceResponse, error) {
	h.logger.Info("GetTrace called", zap.String("trace_id", req.TraceId))
	
	spans, err := h.service.GetTrace(ctx, req.TraceId)
	if err != nil {
		h.logger.Error("Failed to get trace", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get trace: %v", err)
	}
	
	return &v1.GetTraceResponse{
		Spans: spans,
	}, nil
}

// GetTraceGraph retrieves a trace graph visualization
func (h *GrpcHandler) GetTraceGraph(ctx context.Context, req *v1.GetTraceGraphRequest) (*v1.GetTraceGraphResponse, error) {
	h.logger.Info("GetTraceGraph called", zap.String("trace_id", req.TraceId))
	
	graph, err := h.service.GetTraceGraph(ctx, req.TraceId)
	if err != nil {
		h.logger.Error("Failed to get trace graph", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get trace graph: %v", err)
	}
	
	return &v1.GetTraceGraphResponse{
		Graph: graph,
	}, nil
}

// GetAlerts retrieves alerts based on filters
func (h *GrpcHandler) GetAlerts(ctx context.Context, req *v1.GetAlertsRequest) (*v1.GetAlertsResponse, error) {
	h.logger.Info("GetAlerts called", zap.String("status", req.Status))
	
	alerts, pagination, err := h.service.GetAlerts(ctx, req.Status, req.Severity, req.Pagination)
	if err != nil {
		h.logger.Error("Failed to get alerts", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get alerts: %v", err)
	}
	
	return &v1.GetAlertsResponse{
		Alerts:     alerts,
		Pagination: pagination,
	}, nil
}

// GetAlert retrieves a specific alert by ID
func (h *GrpcHandler) GetAlert(ctx context.Context, req *v1.GetAlertRequest) (*v1.GetAlertResponse, error) {
	h.logger.Info("GetAlert called", zap.String("id", req.Id))
	
	alert, err := h.service.GetAlert(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get alert", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get alert: %v", err)
	}
	
	return &v1.GetAlertResponse{
		Alert: alert,
	}, nil
}

// CreateAlert creates a new alert
func (h *GrpcHandler) CreateAlert(ctx context.Context, req *v1.CreateAlertRequest) (*v1.CreateAlertResponse, error) {
	h.logger.Info("CreateAlert called", zap.String("name", req.Name))
	
	alert := &v1.Alert{
		Name:        req.Name,
		Description: req.Description,
		Query:       req.Query,
		Interval:    req.Interval,
		Duration:    req.Duration,
		Severity:    req.Severity,
		Recipients:  req.Recipients,
		Labels:      req.Labels,
	}
	
	createdAlert, err := h.service.CreateAlert(ctx, alert)
	if err != nil {
		h.logger.Error("Failed to create alert", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create alert: %v", err)
	}
	
	return &v1.CreateAlertResponse{
		Alert:   createdAlert,
		Status:  v1.Status_STATUS_OK,
		Message: "Alert created successfully",
	}, nil
}

// UpdateAlert updates an existing alert
func (h *GrpcHandler) UpdateAlert(ctx context.Context, req *v1.UpdateAlertRequest) (*v1.UpdateAlertResponse, error) {
	h.logger.Info("UpdateAlert called", zap.String("id", req.Id))
	
	alert := &v1.Alert{
		Id:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Query:       req.Query,
		Interval:    req.Interval,
		Duration:    req.Duration,
		Severity:    req.Severity,
		Recipients:  req.Recipients,
		Labels:      req.Labels,
	}
	
	updatedAlert, err := h.service.UpdateAlert(ctx, alert)
	if err != nil {
		h.logger.Error("Failed to update alert", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update alert: %v", err)
	}
	
	return &v1.UpdateAlertResponse{
		Alert:   updatedAlert,
		Status:  v1.Status_STATUS_OK,
		Message: "Alert updated successfully",
	}, nil
}

// DeleteAlert deletes an alert
func (h *GrpcHandler) DeleteAlert(ctx context.Context, req *v1.DeleteAlertRequest) (*v1.DeleteAlertResponse, error) {
	h.logger.Info("DeleteAlert called", zap.String("id", req.Id))
	
	err := h.service.DeleteAlert(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to delete alert", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete alert: %v", err)
	}
	
	return &v1.DeleteAlertResponse{
		Status:  v1.Status_STATUS_OK,
		Message: "Alert deleted successfully",
	}, nil
}

// GetDashboards retrieves dashboards
func (h *GrpcHandler) GetDashboards(ctx context.Context, req *v1.GetDashboardsRequest) (*v1.GetDashboardsResponse, error) {
	h.logger.Info("GetDashboards called", zap.String("query", req.Query))
	
	dashboards, pagination, err := h.service.GetDashboards(ctx, req.Query, req.Pagination)
	if err != nil {
		h.logger.Error("Failed to get dashboards", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get dashboards: %v", err)
	}
	
	return &v1.GetDashboardsResponse{
		Dashboards: dashboards,
		Pagination: pagination,
	}, nil
}

// GetDashboard retrieves a specific dashboard by ID
func (h *GrpcHandler) GetDashboard(ctx context.Context, req *v1.GetDashboardRequest) (*v1.GetDashboardResponse, error) {
	h.logger.Info("GetDashboard called", zap.String("id", req.Id))
	
	dashboard, err := h.service.GetDashboard(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get dashboard", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get dashboard: %v", err)
	}
	
	return &v1.GetDashboardResponse{
		Dashboard: dashboard,
	}, nil
}

// CreateDashboard creates a new dashboard
func (h *GrpcHandler) CreateDashboard(ctx context.Context, req *v1.CreateDashboardRequest) (*v1.CreateDashboardResponse, error) {
	h.logger.Info("CreateDashboard called", zap.String("name", req.Name))
	
	dashboard := &v1.Dashboard{
		Name:        req.Name,
		Description: req.Description,
		LayoutJson:  req.LayoutJson,
		Tags:        req.Tags,
	}
	
	createdDashboard, err := h.service.CreateDashboard(ctx, dashboard)
	if err != nil {
		h.logger.Error("Failed to create dashboard", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create dashboard: %v", err)
	}
	
	return &v1.CreateDashboardResponse{
		Dashboard: createdDashboard,
		Status:    v1.Status_STATUS_OK,
		Message:   "Dashboard created successfully",
	}, nil
}

// UpdateDashboard updates an existing dashboard
func (h *GrpcHandler) UpdateDashboard(ctx context.Context, req *v1.UpdateDashboardRequest) (*v1.UpdateDashboardResponse, error) {
	h.logger.Info("UpdateDashboard called", zap.String("id", req.Id))
	
	dashboard := &v1.Dashboard{
		Id:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		LayoutJson:  req.LayoutJson,
		Tags:        req.Tags,
	}
	
	updatedDashboard, err := h.service.UpdateDashboard(ctx, dashboard)
	if err != nil {
		h.logger.Error("Failed to update dashboard", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update dashboard: %v", err)
	}
	
	return &v1.UpdateDashboardResponse{
		Dashboard: updatedDashboard,
		Status:    v1.Status_STATUS_OK,
		Message:   "Dashboard updated successfully",
	}, nil
}

// DeleteDashboard deletes a dashboard
func (h *GrpcHandler) DeleteDashboard(ctx context.Context, req *v1.DeleteDashboardRequest) (*v1.DeleteDashboardResponse, error) {
	h.logger.Info("DeleteDashboard called", zap.String("id", req.Id))
	
	err := h.service.DeleteDashboard(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to delete dashboard", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete dashboard: %v", err)
	}
	
	return &v1.DeleteDashboardResponse{
		Status:  v1.Status_STATUS_OK,
		Message: "Dashboard deleted successfully",
	}, nil
} 