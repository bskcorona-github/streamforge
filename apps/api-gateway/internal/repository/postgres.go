package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bskcorona-github/streamforge/packages/proto/streamforge/v1"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// PostgresRepository はPostgreSQLリポジトリの実装です
type PostgresRepository struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

// NewPostgresRepository は新しいPostgresRepositoryインスタンスを作成します
func NewPostgresRepository(db *gorm.DB, redis *redis.Client, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// GetMetrics はメトリクスを取得します
func (r *PostgresRepository) GetMetrics(ctx context.Context, query string, timeRange *v1.TimeRange, pagination *v1.Pagination) ([]*v1.Metric, *v1.Pagination, error) {
	var entities []MetricEntity
	db := r.db.WithContext(ctx)

	// クエリ条件を構築
	if query != "" {
		db = db.Where("name ILIKE ?", "%"+query+"%")
	}

	if timeRange != nil && timeRange.StartTime != nil && timeRange.EndTime != nil {
		start := timeRange.StartTime.AsTime()
		end := timeRange.EndTime.AsTime()
		db = db.Where("created_at BETWEEN ? AND ?", start, end)
	}

	// ページネーション
	offset := int(pagination.PageToken) * int(pagination.PageSize)
	limit := int(pagination.PageSize)

	var total int64
	if err := db.Model(&MetricEntity{}).Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count metrics: %w", err)
	}

	if err := db.Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// エンティティをプロトコルバッファに変換
	metrics := make([]*v1.Metric, len(entities))
	for i, entity := range entities {
		metric, err := entity.ToProto()
		if err != nil {
			r.logger.Error("Failed to convert metric entity to proto", zap.Error(err))
			continue
		}
		metrics[i] = metric
	}

	// ページネーション情報を更新
	pagination.TotalCount = int32(total)

	return metrics, pagination, nil
}

// GetMetric は特定のメトリクスを取得します
func (r *PostgresRepository) GetMetric(ctx context.Context, name string, timeRange *v1.TimeRange) (*v1.Metric, error) {
	var entity MetricEntity
	db := r.db.WithContext(ctx).Where("name = ?", name)

	if timeRange != nil && timeRange.StartTime != nil && timeRange.EndTime != nil {
		start := timeRange.StartTime.AsTime()
		end := timeRange.EndTime.AsTime()
		db = db.Where("created_at BETWEEN ? AND ?", start, end)
	}

	if err := db.First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("metric not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	return entity.ToProto()
}

// GetTimeSeries は時系列データを取得します
func (r *PostgresRepository) GetTimeSeries(ctx context.Context, metricName string, timeRange *v1.TimeRange, tags []*v1.Tag, aggregation string, interval time.Duration) ([]*v1.TimeSeries, error) {
	// ClickHouseを使用した時系列データ取得の実装
	// ここでは簡略化のため、PostgreSQLで実装
	var entities []MetricDataPointEntity
	db := r.db.WithContext(ctx).Where("metric_name = ?", metricName)

	if timeRange != nil && timeRange.StartTime != nil && timeRange.EndTime != nil {
		start := timeRange.StartTime.AsTime()
		end := timeRange.EndTime.AsTime()
		db = db.Where("timestamp BETWEEN ? AND ?", start, end)
	}

	// タグフィルタリング
	for _, tag := range tags {
		db = db.Where("tags @> ?", fmt.Sprintf(`{"%s": "%s"}`, tag.Key, tag.Value))
	}

	// 集約クエリ
	query := fmt.Sprintf(`
		SELECT 
			date_trunc('%s', timestamp) as time_bucket,
			%s(value) as aggregated_value
		FROM metric_data_points 
		WHERE metric_name = ? 
		GROUP BY time_bucket 
		ORDER BY time_bucket
	`, interval.String(), aggregation)

	if err := db.Raw(query, metricName).Scan(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to get time series: %w", err)
	}

	// 時系列データを構築
	timeSeries := &v1.TimeSeries{
		MetricName: metricName,
		Tags:       tags,
		DataPoints: make([]*v1.TimeSeriesDataPoint, len(entities)),
	}

	for i, entity := range entities {
		timeSeries.DataPoints[i] = &v1.TimeSeriesDataPoint{
			Timestamp: timestamppb.New(entity.Timestamp),
			Value:     entity.Value,
		}
	}

	return []*v1.TimeSeries{timeSeries}, nil
}

// GetLogs はログを取得します
func (r *PostgresRepository) GetLogs(ctx context.Context, query string, timeRange *v1.TimeRange, minLevel v1.LogLevel, pagination *v1.Pagination) ([]*v1.Log, *v1.Pagination, error) {
	var entities []LogEntity
	db := r.db.WithContext(ctx)

	// クエリ条件を構築
	if query != "" {
		db = db.Where("message ILIKE ? OR service_name ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	if timeRange != nil && timeRange.StartTime != nil && timeRange.EndTime != nil {
		start := timeRange.StartTime.AsTime()
		end := timeRange.EndTime.AsTime()
		db = db.Where("timestamp BETWEEN ? AND ?", start, end)
	}

	// ログレベルフィルタリング
	db = db.Where("level >= ?", int32(minLevel))

	// ページネーション
	offset := int(pagination.PageToken) * int(pagination.PageSize)
	limit := int(pagination.PageSize)

	var total int64
	if err := db.Model(&LogEntity{}).Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count logs: %w", err)
	}

	if err := db.Offset(offset).Limit(limit).Order("timestamp DESC").Find(&entities).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get logs: %w", err)
	}

	// エンティティをプロトコルバッファに変換
	logs := make([]*v1.Log, len(entities))
	for i, entity := range entities {
		log, err := entity.ToProto()
		if err != nil {
			r.logger.Error("Failed to convert log entity to proto", zap.Error(err))
			continue
		}
		logs[i] = log
	}

	// ページネーション情報を更新
	pagination.TotalCount = int32(total)

	return logs, pagination, nil
}

// SearchLogs はログを検索します
func (r *PostgresRepository) SearchLogs(ctx context.Context, query string, timeRange *v1.TimeRange, minLevel v1.LogLevel, pagination *v1.Pagination) ([]*v1.Log, *v1.Pagination, error) {
	// フルテキスト検索の実装
	// ここでは簡略化のため、GetLogsと同じ実装を使用
	return r.GetLogs(ctx, query, timeRange, minLevel, pagination)
}

// GetTraces はトレースを取得します
func (r *PostgresRepository) GetTraces(ctx context.Context, serviceName string, operationName string, timeRange *v1.TimeRange, status v1.Status, pagination *v1.Pagination) ([]*v1.Span, *v1.Pagination, error) {
	var entities []SpanEntity
	db := r.db.WithContext(ctx).Where("parent_span_id IS NULL") // ルートスパンのみ

	// クエリ条件を構築
	if serviceName != "" {
		db = db.Where("service_name = ?", serviceName)
	}

	if operationName != "" {
		db = db.Where("name = ?", operationName)
	}

	if timeRange != nil && timeRange.StartTime != nil && timeRange.EndTime != nil {
		start := timeRange.StartTime.AsTime()
		end := timeRange.EndTime.AsTime()
		db = db.Where("start_time BETWEEN ? AND ?", start, end)
	}

	if status != v1.Status_STATUS_UNSPECIFIED {
		db = db.Where("status = ?", int32(status))
	}

	// ページネーション
	offset := int(pagination.PageToken) * int(pagination.PageSize)
	limit := int(pagination.PageSize)

	var total int64
	if err := db.Model(&SpanEntity{}).Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count traces: %w", err)
	}

	if err := db.Offset(offset).Limit(limit).Order("start_time DESC").Find(&entities).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get traces: %w", err)
	}

	// エンティティをプロトコルバッファに変換
	spans := make([]*v1.Span, len(entities))
	for i, entity := range entities {
		span, err := entity.ToProto()
		if err != nil {
			r.logger.Error("Failed to convert span entity to proto", zap.Error(err))
			continue
		}
		spans[i] = span
	}

	// ページネーション情報を更新
	pagination.TotalCount = int32(total)

	return spans, pagination, nil
}

// GetTrace は特定のトレースを取得します
func (r *PostgresRepository) GetTrace(ctx context.Context, traceID string) ([]*v1.Span, error) {
	var entities []SpanEntity
	db := r.db.WithContext(ctx).Where("trace_id = ?", traceID)

	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to get trace: %w", err)
	}

	// エンティティをプロトコルバッファに変換
	spans := make([]*v1.Span, len(entities))
	for i, entity := range entities {
		span, err := entity.ToProto()
		if err != nil {
			r.logger.Error("Failed to convert span entity to proto", zap.Error(err))
			continue
		}
		spans[i] = span
	}

	return spans, nil
}

// GetTraceGraph はトレースグラフを取得します
func (r *PostgresRepository) GetTraceGraph(ctx context.Context, traceID string) (*v1.TraceGraph, error) {
	// トレースグラフの構築
	spans, err := r.GetTrace(ctx, traceID)
	if err != nil {
		return nil, err
	}

	// ノードとエッジを構築
	nodes := make([]*v1.TraceGraphNode, len(spans))
	edges := make([]*v1.TraceGraphEdge, 0)

	for i, span := range spans {
		nodes[i] = &v1.TraceGraphNode{
			Id:          span.SpanId,
			Name:        span.Name,
			ServiceName: span.Resource.ServiceName,
			Duration:    span.Duration,
			Status:      span.Status,
			Type:        "span",
		}

		// 親子関係のエッジを追加
		if span.ParentSpanId != "" {
			edges = append(edges, &v1.TraceGraphEdge{
				SourceId: span.ParentSpanId,
				TargetId: span.SpanId,
				Type:     "child_of",
			})
		}
	}

	return &v1.TraceGraph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

// GetAlerts はアラートを取得します
func (r *PostgresRepository) GetAlerts(ctx context.Context, status string, severity string, pagination *v1.Pagination) ([]*v1.Alert, *v1.Pagination, error) {
	var entities []AlertEntity
	db := r.db.WithContext(ctx)

	// クエリ条件を構築
	if status != "" {
		db = db.Where("status = ?", status)
	}

	if severity != "" {
		db = db.Where("severity = ?", severity)
	}

	// ページネーション
	offset := int(pagination.PageToken) * int(pagination.PageSize)
	limit := int(pagination.PageSize)

	var total int64
	if err := db.Model(&AlertEntity{}).Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count alerts: %w", err)
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&entities).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	// エンティティをプロトコルバッファに変換
	alerts := make([]*v1.Alert, len(entities))
	for i, entity := range entities {
		alert, err := entity.ToProto()
		if err != nil {
			r.logger.Error("Failed to convert alert entity to proto", zap.Error(err))
			continue
		}
		alerts[i] = alert
	}

	// ページネーション情報を更新
	pagination.TotalCount = int32(total)

	return alerts, pagination, nil
}

// GetAlert は特定のアラートを取得します
func (r *PostgresRepository) GetAlert(ctx context.Context, id string) (*v1.Alert, error) {
	var entity AlertEntity
	db := r.db.WithContext(ctx).Where("id = ?", id)

	if err := db.First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("alert not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return entity.ToProto()
}

// CreateAlert はアラートを作成します
func (r *PostgresRepository) CreateAlert(ctx context.Context, alert *v1.Alert) (*v1.Alert, error) {
	entity := &AlertEntity{}
	if err := entity.FromProto(alert); err != nil {
		return nil, fmt.Errorf("failed to convert proto to entity: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	return entity.ToProto()
}

// UpdateAlert はアラートを更新します
func (r *PostgresRepository) UpdateAlert(ctx context.Context, alert *v1.Alert) (*v1.Alert, error) {
	entity := &AlertEntity{}
	if err := entity.FromProto(alert); err != nil {
		return nil, fmt.Errorf("failed to convert proto to entity: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update alert: %w", err)
	}

	return entity.ToProto()
}

// DeleteAlert はアラートを削除します
func (r *PostgresRepository) DeleteAlert(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&AlertEntity{}).Error; err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	return nil
}

// GetDashboards はダッシュボードを取得します
func (r *PostgresRepository) GetDashboards(ctx context.Context, query string, pagination *v1.Pagination) ([]*v1.Dashboard, *v1.Pagination, error) {
	var entities []DashboardEntity
	db := r.db.WithContext(ctx)

	// クエリ条件を構築
	if query != "" {
		db = db.Where("name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	// ページネーション
	offset := int(pagination.PageToken) * int(pagination.PageSize)
	limit := int(pagination.PageSize)

	var total int64
	if err := db.Model(&DashboardEntity{}).Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count dashboards: %w", err)
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&entities).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get dashboards: %w", err)
	}

	// エンティティをプロトコルバッファに変換
	dashboards := make([]*v1.Dashboard, len(entities))
	for i, entity := range entities {
		dashboard, err := entity.ToProto()
		if err != nil {
			r.logger.Error("Failed to convert dashboard entity to proto", zap.Error(err))
			continue
		}
		dashboards[i] = dashboard
	}

	// ページネーション情報を更新
	pagination.TotalCount = int32(total)

	return dashboards, pagination, nil
}

// GetDashboard は特定のダッシュボードを取得します
func (r *PostgresRepository) GetDashboard(ctx context.Context, id string) (*v1.Dashboard, error) {
	var entity DashboardEntity
	db := r.db.WithContext(ctx).Where("id = ?", id)

	if err := db.First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("dashboard not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	return entity.ToProto()
}

// CreateDashboard はダッシュボードを作成します
func (r *PostgresRepository) CreateDashboard(ctx context.Context, dashboard *v1.Dashboard) (*v1.Dashboard, error) {
	entity := &DashboardEntity{}
	if err := entity.FromProto(dashboard); err != nil {
		return nil, fmt.Errorf("failed to convert proto to entity: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, fmt.Errorf("failed to create dashboard: %w", err)
	}

	return entity.ToProto()
}

// UpdateDashboard はダッシュボードを更新します
func (r *PostgresRepository) UpdateDashboard(ctx context.Context, dashboard *v1.Dashboard) (*v1.Dashboard, error) {
	entity := &DashboardEntity{}
	if err := entity.FromProto(dashboard); err != nil {
		return nil, fmt.Errorf("failed to convert proto to entity: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, fmt.Errorf("failed to update dashboard: %w", err)
	}

	return entity.ToProto()
}

// DeleteDashboard はダッシュボードを削除します
func (r *PostgresRepository) DeleteDashboard(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&DashboardEntity{}).Error; err != nil {
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	return nil
}

// Close はリソースをクリーンアップします
func (r *PostgresRepository) Close() error {
	if r.redis != nil {
		return r.redis.Close()
	}
	return nil
} 