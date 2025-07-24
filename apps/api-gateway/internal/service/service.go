package service

import (
	"context"
	"time"

	"github.com/bskcorona-github/streamforge/packages/proto/streamforge/v1"
	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service はビジネスロジック層のインターフェースです
type Service interface {
	// メトリクス関連
	GetMetrics(ctx context.Context, req *v1.GetMetricsRequest) (*v1.GetMetricsResponse, error)
	GetMetric(ctx context.Context, req *v1.GetMetricRequest) (*v1.GetMetricResponse, error)
	GetTimeSeries(ctx context.Context, req *v1.GetTimeSeriesRequest) (*v1.GetTimeSeriesResponse, error)

	// ログ関連
	GetLogs(ctx context.Context, req *v1.GetLogsRequest) (*v1.GetLogsResponse, error)
	SearchLogs(ctx context.Context, req *v1.SearchLogsRequest) (*v1.SearchLogsResponse, error)

	// トレース関連
	GetTraces(ctx context.Context, req *v1.GetTracesRequest) (*v1.GetTracesResponse, error)
	GetTrace(ctx context.Context, req *v1.GetTraceRequest) (*v1.GetTraceResponse, error)
	GetTraceGraph(ctx context.Context, req *v1.GetTraceGraphRequest) (*v1.GetTraceGraphResponse, error)

	// アラート関連
	GetAlerts(ctx context.Context, req *v1.GetAlertsRequest) (*v1.GetAlertsResponse, error)
	GetAlert(ctx context.Context, req *v1.GetAlertRequest) (*v1.GetAlertResponse, error)
	CreateAlert(ctx context.Context, req *v1.CreateAlertRequest) (*v1.CreateAlertResponse, error)
	UpdateAlert(ctx context.Context, req *v1.UpdateAlertRequest) (*v1.UpdateAlertResponse, error)
	DeleteAlert(ctx context.Context, req *v1.DeleteAlertRequest) (*v1.DeleteAlertResponse, error)

	// ダッシュボード関連
	GetDashboards(ctx context.Context, req *v1.GetDashboardsRequest) (*v1.GetDashboardsResponse, error)
	GetDashboard(ctx context.Context, req *v1.GetDashboardRequest) (*v1.GetDashboardResponse, error)
	CreateDashboard(ctx context.Context, req *v1.CreateDashboardRequest) (*v1.CreateDashboardResponse, error)
	UpdateDashboard(ctx context.Context, req *v1.UpdateDashboardRequest) (*v1.UpdateDashboardResponse, error)
	DeleteDashboard(ctx context.Context, req *v1.DeleteDashboardRequest) (*v1.DeleteDashboardResponse, error)
}

// ApiService はAPIサービスの実装です
type ApiService struct {
	repo   repository.Repository
	logger *zap.Logger
}

// NewApiService は新しいApiServiceインスタンスを作成します
func NewApiService(repo repository.Repository, logger *zap.Logger) *ApiService {
	return &ApiService{
		repo:   repo,
		logger: logger,
	}
}

// GetMetrics はメトリクスを取得します
func (s *ApiService) GetMetrics(ctx context.Context, req *v1.GetMetricsRequest) (*v1.GetMetricsResponse, error) {
	s.logger.Info("GetMetrics service called", zap.String("query", req.Query))

	metrics, pagination, err := s.repo.GetMetrics(ctx, req.Query, req.TimeRange, req.Pagination)
	if err != nil {
		s.logger.Error("Failed to get metrics", zap.Error(err))
		return &v1.GetMetricsResponse{
			Metrics:    []*v1.Metric{},
			Pagination: &v1.Pagination{},
		}, err
	}

	return &v1.GetMetricsResponse{
		Metrics:    metrics,
		Pagination: pagination,
	}, nil
}

// GetMetric は特定のメトリクスを取得します
func (s *ApiService) GetMetric(ctx context.Context, req *v1.GetMetricRequest) (*v1.GetMetricResponse, error) {
	s.logger.Info("GetMetric service called", zap.String("name", req.Name))

	metric, err := s.repo.GetMetric(ctx, req.Name, req.TimeRange)
	if err != nil {
		s.logger.Error("Failed to get metric", zap.Error(err))
		return &v1.GetMetricResponse{
			Metric: &v1.Metric{},
		}, err
	}

	return &v1.GetMetricResponse{
		Metric: metric,
	}, nil
}

// GetTimeSeries は時系列データを取得します
func (s *ApiService) GetTimeSeries(ctx context.Context, req *v1.GetTimeSeriesRequest) (*v1.GetTimeSeriesResponse, error) {
	s.logger.Info("GetTimeSeries service called", zap.String("metricName", req.MetricName))

	interval, err := time.ParseDuration(req.Interval.AsDuration().String())
	if err != nil {
		s.logger.Error("Failed to parse interval", zap.Error(err))
		return &v1.GetTimeSeriesResponse{
			Series: []*v1.TimeSeries{},
		}, err
	}

	series, err := s.repo.GetTimeSeries(ctx, req.MetricName, req.TimeRange, req.Tags, req.Aggregation, interval)
	if err != nil {
		s.logger.Error("Failed to get time series", zap.Error(err))
		return &v1.GetTimeSeriesResponse{
			Series: []*v1.TimeSeries{},
		}, err
	}

	return &v1.GetTimeSeriesResponse{
		Series: series,
	}, nil
}

// GetLogs はログを取得します
func (s *ApiService) GetLogs(ctx context.Context, req *v1.GetLogsRequest) (*v1.GetLogsResponse, error) {
	s.logger.Info("GetLogs service called", zap.String("query", req.Query))

	logs, pagination, err := s.repo.GetLogs(ctx, req.Query, req.TimeRange, req.MinLevel, req.Pagination)
	if err != nil {
		s.logger.Error("Failed to get logs", zap.Error(err))
		return &v1.GetLogsResponse{
			Logs:       []*v1.Log{},
			Pagination: &v1.Pagination{},
		}, err
	}

	return &v1.GetLogsResponse{
		Logs:       logs,
		Pagination: pagination,
	}, nil
}

// SearchLogs はログを検索します
func (s *ApiService) SearchLogs(ctx context.Context, req *v1.SearchLogsRequest) (*v1.SearchLogsResponse, error) {
	s.logger.Info("SearchLogs service called", zap.String("query", req.Query))

	logs, pagination, err := s.repo.SearchLogs(ctx, req.Query, req.TimeRange, req.MinLevel, req.Pagination)
	if err != nil {
		s.logger.Error("Failed to search logs", zap.Error(err))
		return &v1.SearchLogsResponse{
			Logs:       []*v1.Log{},
			Pagination: &v1.Pagination{},
		}, err
	}

	return &v1.SearchLogsResponse{
		Logs:       logs,
		Pagination: pagination,
	}, nil
}

// GetTraces はトレースを取得します
func (s *ApiService) GetTraces(ctx context.Context, req *v1.GetTracesRequest) (*v1.GetTracesResponse, error) {
	s.logger.Info("GetTraces service called", zap.String("serviceName", req.ServiceName))

	traces, pagination, err := s.repo.GetTraces(ctx, req.ServiceName, req.OperationName, req.TimeRange, req.Status, req.Pagination)
	if err != nil {
		s.logger.Error("Failed to get traces", zap.Error(err))
		return &v1.GetTracesResponse{
			Traces:     []*v1.Span{},
			Pagination: &v1.Pagination{},
		}, err
	}

	return &v1.GetTracesResponse{
		Traces:     traces,
		Pagination: pagination,
	}, nil
}

// GetTrace は特定のトレースを取得します
func (s *ApiService) GetTrace(ctx context.Context, req *v1.GetTraceRequest) (*v1.GetTraceResponse, error) {
	s.logger.Info("GetTrace service called", zap.String("traceID", req.TraceId))

	spans, err := s.repo.GetTrace(ctx, req.TraceId)
	if err != nil {
		s.logger.Error("Failed to get trace", zap.Error(err))
		return &v1.GetTraceResponse{
			Spans: []*v1.Span{},
		}, err
	}

	return &v1.GetTraceResponse{
		Spans: spans,
	}, nil
}

// GetTraceGraph はトレースグラフを取得します
func (s *ApiService) GetTraceGraph(ctx context.Context, req *v1.GetTraceGraphRequest) (*v1.GetTraceGraphResponse, error) {
	s.logger.Info("GetTraceGraph service called", zap.String("traceID", req.TraceId))

	graph, err := s.repo.GetTraceGraph(ctx, req.TraceId)
	if err != nil {
		s.logger.Error("Failed to get trace graph", zap.Error(err))
		return &v1.GetTraceGraphResponse{
			Graph: &v1.TraceGraph{},
		}, err
	}

	return &v1.GetTraceGraphResponse{
		Graph: graph,
	}, nil
}

// GetAlerts はアラートを取得します
func (s *ApiService) GetAlerts(ctx context.Context, req *v1.GetAlertsRequest) (*v1.GetAlertsResponse, error) {
	s.logger.Info("GetAlerts service called", zap.String("status", req.Status))

	alerts, pagination, err := s.repo.GetAlerts(ctx, req.Status, req.Severity, req.Pagination)
	if err != nil {
		s.logger.Error("Failed to get alerts", zap.Error(err))
		return &v1.GetAlertsResponse{
			Alerts:     []*v1.Alert{},
			Pagination: &v1.Pagination{},
		}, err
	}

	return &v1.GetAlertsResponse{
		Alerts:     alerts,
		Pagination: pagination,
	}, nil
}

// GetAlert は特定のアラートを取得します
func (s *ApiService) GetAlert(ctx context.Context, req *v1.GetAlertRequest) (*v1.GetAlertResponse, error) {
	s.logger.Info("GetAlert service called", zap.String("id", req.Id))

	alert, err := s.repo.GetAlert(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get alert", zap.Error(err))
		return &v1.GetAlertResponse{
			Alert: &v1.Alert{},
		}, err
	}

	return &v1.GetAlertResponse{
		Alert: alert,
	}, nil
}

// CreateAlert はアラートを作成します
func (s *ApiService) CreateAlert(ctx context.Context, req *v1.CreateAlertRequest) (*v1.CreateAlertResponse, error) {
	s.logger.Info("CreateAlert service called", zap.String("name", req.Name))

	alert := &v1.Alert{
		Name:        req.Name,
		Description: req.Description,
		Query:       req.Query,
		Interval:    req.Interval,
		Duration:    req.Duration,
		Severity:    req.Severity,
		Recipients:  req.Recipients,
		Labels:      req.Labels,
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
	}

	createdAlert, err := s.repo.CreateAlert(ctx, alert)
	if err != nil {
		s.logger.Error("Failed to create alert", zap.Error(err))
		return &v1.CreateAlertResponse{
			Alert:   &v1.Alert{},
			Status:  v1.Status_STATUS_ERROR,
			Message: err.Error(),
		}, err
	}

	return &v1.CreateAlertResponse{
		Alert:   createdAlert,
		Status:  v1.Status_STATUS_OK,
		Message: "Alert created successfully",
	}, nil
}

// UpdateAlert はアラートを更新します
func (s *ApiService) UpdateAlert(ctx context.Context, req *v1.UpdateAlertRequest) (*v1.UpdateAlertResponse, error) {
	s.logger.Info("UpdateAlert service called", zap.String("id", req.Id))

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
		UpdatedAt:   timestamppb.Now(),
	}

	updatedAlert, err := s.repo.UpdateAlert(ctx, alert)
	if err != nil {
		s.logger.Error("Failed to update alert", zap.Error(err))
		return &v1.UpdateAlertResponse{
			Alert:   &v1.Alert{},
			Status:  v1.Status_STATUS_ERROR,
			Message: err.Error(),
		}, err
	}

	return &v1.UpdateAlertResponse{
		Alert:   updatedAlert,
		Status:  v1.Status_STATUS_OK,
		Message: "Alert updated successfully",
	}, nil
}

// DeleteAlert はアラートを削除します
func (s *ApiService) DeleteAlert(ctx context.Context, req *v1.DeleteAlertRequest) (*v1.DeleteAlertResponse, error) {
	s.logger.Info("DeleteAlert service called", zap.String("id", req.Id))

	err := s.repo.DeleteAlert(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to delete alert", zap.Error(err))
		return &v1.DeleteAlertResponse{
			Status:  v1.Status_STATUS_ERROR,
			Message: err.Error(),
		}, err
	}

	return &v1.DeleteAlertResponse{
		Status:  v1.Status_STATUS_OK,
		Message: "Alert deleted successfully",
	}, nil
}

// GetDashboards はダッシュボードを取得します
func (s *ApiService) GetDashboards(ctx context.Context, req *v1.GetDashboardsRequest) (*v1.GetDashboardsResponse, error) {
	s.logger.Info("GetDashboards service called", zap.String("query", req.Query))

	dashboards, pagination, err := s.repo.GetDashboards(ctx, req.Query, req.Pagination)
	if err != nil {
		s.logger.Error("Failed to get dashboards", zap.Error(err))
		return &v1.GetDashboardsResponse{
			Dashboards: []*v1.Dashboard{},
			Pagination: &v1.Pagination{},
		}, err
	}

	return &v1.GetDashboardsResponse{
		Dashboards: dashboards,
		Pagination: pagination,
	}, nil
}

// GetDashboard は特定のダッシュボードを取得します
func (s *ApiService) GetDashboard(ctx context.Context, req *v1.GetDashboardRequest) (*v1.GetDashboardResponse, error) {
	s.logger.Info("GetDashboard service called", zap.String("id", req.Id))

	dashboard, err := s.repo.GetDashboard(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get dashboard", zap.Error(err))
		return &v1.GetDashboardResponse{
			Dashboard: &v1.Dashboard{},
		}, err
	}

	return &v1.GetDashboardResponse{
		Dashboard: dashboard,
	}, nil
}

// CreateDashboard はダッシュボードを作成します
func (s *ApiService) CreateDashboard(ctx context.Context, req *v1.CreateDashboardRequest) (*v1.CreateDashboardResponse, error) {
	s.logger.Info("CreateDashboard service called", zap.String("name", req.Name))

	dashboard := &v1.Dashboard{
		Name:        req.Name,
		Description: req.Description,
		LayoutJson:  req.LayoutJson,
		Tags:        req.Tags,
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
	}

	createdDashboard, err := s.repo.CreateDashboard(ctx, dashboard)
	if err != nil {
		s.logger.Error("Failed to create dashboard", zap.Error(err))
		return &v1.CreateDashboardResponse{
			Dashboard: &v1.Dashboard{},
			Status:    v1.Status_STATUS_ERROR,
			Message:   err.Error(),
		}, err
	}

	return &v1.CreateDashboardResponse{
		Dashboard: createdDashboard,
		Status:    v1.Status_STATUS_OK,
		Message:   "Dashboard created successfully",
	}, nil
}

// UpdateDashboard はダッシュボードを更新します
func (s *ApiService) UpdateDashboard(ctx context.Context, req *v1.UpdateDashboardRequest) (*v1.UpdateDashboardResponse, error) {
	s.logger.Info("UpdateDashboard service called", zap.String("id", req.Id))

	dashboard := &v1.Dashboard{
		Id:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		LayoutJson:  req.LayoutJson,
		Tags:        req.Tags,
		UpdatedAt:   timestamppb.Now(),
	}

	updatedDashboard, err := s.repo.UpdateDashboard(ctx, dashboard)
	if err != nil {
		s.logger.Error("Failed to update dashboard", zap.Error(err))
		return &v1.UpdateDashboardResponse{
			Dashboard: &v1.Dashboard{},
			Status:    v1.Status_STATUS_ERROR,
			Message:   err.Error(),
		}, err
	}

	return &v1.UpdateDashboardResponse{
		Dashboard: updatedDashboard,
		Status:    v1.Status_STATUS_OK,
		Message:   "Dashboard updated successfully",
	}, nil
}

// DeleteDashboard はダッシュボードを削除します
func (s *ApiService) DeleteDashboard(ctx context.Context, req *v1.DeleteDashboardRequest) (*v1.DeleteDashboardResponse, error) {
	s.logger.Info("DeleteDashboard service called", zap.String("id", req.Id))

	err := s.repo.DeleteDashboard(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to delete dashboard", zap.Error(err))
		return &v1.DeleteDashboardResponse{
			Status:  v1.Status_STATUS_ERROR,
			Message: err.Error(),
		}, err
	}

	return &v1.DeleteDashboardResponse{
		Status:  v1.Status_STATUS_OK,
		Message: "Dashboard deleted successfully",
	}, nil
} 