package repository

import (
	"encoding/json"
	"time"

	"github.com/bskcorona-github/streamforge/packages/proto/streamforge/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// BaseEntity は全てのエンティティの基底構造です
type BaseEntity struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate はエンティティ作成前にIDを生成します
func (b *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// MetricEntity はメトリクスエンティティです
type MetricEntity struct {
	BaseEntity
	Name        string          `gorm:"not null;index" json:"name"`
	Type        int32           `gorm:"not null" json:"type"` // MetricType enum
	Unit        string          `gorm:"not null" json:"unit"`
	Description string          `gorm:"type:text" json:"description"`
	Value       float64         `gorm:"not null" json:"value"`
	Timestamp   time.Time       `gorm:"not null;index" json:"timestamp"`
	Tags        json.RawMessage `gorm:"type:jsonb" json:"tags"`
	Resource    json.RawMessage `gorm:"type:jsonb" json:"resource"`
}

// ToProto はエンティティをプロトコルバッファに変換します
func (m *MetricEntity) ToProto() *v1.Metric {
	metric := &v1.Metric{
		Name:        m.Name,
		Type:        v1.MetricType(m.Type),
		Unit:        m.Unit,
		Description: m.Description,
		Resource:    &v1.Resource{},
	}

	// データポイントの作成
	dataPoint := &v1.MetricDataPoint{
		Timestamp: timestamppb.New(m.Timestamp),
		Value:     m.Value,
	}

	// タグの解析
	if m.Tags != nil {
		var tags []map[string]string
		if err := json.Unmarshal(m.Tags, &tags); err == nil {
			for _, tagMap := range tags {
				for k, v := range tagMap {
					dataPoint.Tags = append(dataPoint.Tags, &v1.Tag{
						Key:   k,
						Value: v,
					})
				}
			}
		}
	}

	metric.DataPoints = append(metric.DataPoints, dataPoint)

	// リソース情報の解析
	if m.Resource != nil {
		var resourceMap map[string]interface{}
		if err := json.Unmarshal(m.Resource, &resourceMap); err == nil {
			if serviceName, ok := resourceMap["service_name"].(string); ok {
				metric.Resource.ServiceName = serviceName
			}
			if hostName, ok := resourceMap["host_name"].(string); ok {
				metric.Resource.HostName = hostName
			}
			if instanceID, ok := resourceMap["instance_id"].(string); ok {
				metric.Resource.InstanceId = instanceID
			}
		}
	}

	return metric
}

// FromProto はプロトコルバッファからエンティティを作成します
func (m *MetricEntity) FromProto(metric *v1.Metric) {
	m.Name = metric.Name
	m.Type = int32(metric.Type)
	m.Unit = metric.Unit
	m.Description = metric.Description

	if len(metric.DataPoints) > 0 {
		dataPoint := metric.DataPoints[0]
		m.Value = dataPoint.Value
		m.Timestamp = dataPoint.Timestamp.AsTime()

		// タグの保存
		if len(dataPoint.Tags) > 0 {
			tags := make(map[string]string)
			for _, tag := range dataPoint.Tags {
				tags[tag.Key] = tag.Value
			}
			if tagsJSON, err := json.Marshal(tags); err == nil {
				m.Tags = tagsJSON
			}
		}
	}

	// リソース情報の保存
	if metric.Resource != nil {
		resource := map[string]interface{}{
			"service_name": metric.Resource.ServiceName,
			"host_name":    metric.Resource.HostName,
			"instance_id":  metric.Resource.InstanceId,
		}
		if resourceJSON, err := json.Marshal(resource); err == nil {
			m.Resource = resourceJSON
		}
	}
}

// LogEntity はログエンティティです
type LogEntity struct {
	BaseEntity
	Message     string          `gorm:"type:text;not null" json:"message"`
	Level       int32           `gorm:"not null;index" json:"level"` // LogLevel enum
	Timestamp   time.Time       `gorm:"not null;index" json:"timestamp"`
	ServiceName string          `gorm:"index" json:"service_name"`
	HostName    string          `gorm:"index" json:"host_name"`
	TraceID     string          `gorm:"index" json:"trace_id"`
	SpanID      string          `gorm:"index" json:"span_id"`
	Attributes  json.RawMessage `gorm:"type:jsonb" json:"attributes"`
}

// ToProto はエンティティをプロトコルバッファに変換します
func (l *LogEntity) ToProto() *v1.Log {
	log := &v1.Log{
		Id:          l.ID,
		Timestamp:   timestamppb.New(l.Timestamp),
		Level:       v1.LogLevel(l.Level),
		Message:     l.Message,
		ServiceName: l.ServiceName,
		HostName:    l.HostName,
		TraceId:     l.TraceID,
		SpanId:      l.SpanID,
		Attributes:  make(map[string]string),
	}

	// 属性の解析
	if l.Attributes != nil {
		var attrs map[string]string
		if err := json.Unmarshal(l.Attributes, &attrs); err == nil {
			log.Attributes = attrs
		}
	}

	return log
}

// FromProto はプロトコルバッファからエンティティを作成します
func (l *LogEntity) FromProto(log *v1.Log) {
	l.ID = log.Id
	l.Message = log.Message
	l.Level = int32(log.Level)
	l.Timestamp = log.Timestamp.AsTime()
	l.ServiceName = log.ServiceName
	l.HostName = log.HostName
	l.TraceID = log.TraceId
	l.SpanID = log.SpanId

	// 属性の保存
	if len(log.Attributes) > 0 {
		if attrsJSON, err := json.Marshal(log.Attributes); err == nil {
			l.Attributes = attrsJSON
		}
	}
}

// AlertEntity はアラートエンティティです
type AlertEntity struct {
	BaseEntity
	Name        string          `gorm:"not null" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	Query       string          `gorm:"type:text;not null" json:"query"`
	Interval    int64           `gorm:"not null" json:"interval"` // seconds
	Duration    int64           `gorm:"not null" json:"duration"` // seconds
	Severity    string          `gorm:"not null;index" json:"severity"`
	Status      string          `gorm:"not null;index" json:"status"`
	LastFiredAt *time.Time      `gorm:"index" json:"last_fired_at"`
	Recipients  json.RawMessage `gorm:"type:jsonb" json:"recipients"`
	Labels      json.RawMessage `gorm:"type:jsonb" json:"labels"`
}

// ToProto はエンティティをプロトコルバッファに変換します
func (a *AlertEntity) ToProto() *v1.Alert {
	alert := &v1.Alert{
		Id:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Query:       a.Query,
		Severity:    a.Severity,
		Status:      a.Status,
		CreatedAt:   timestamppb.New(a.CreatedAt),
		UpdatedAt:   timestamppb.New(a.UpdatedAt),
	}

	// 間隔と期間の設定
	alert.Interval = &timestamppb.Duration{Seconds: a.Interval}
	alert.Duration = &timestamppb.Duration{Seconds: a.Duration}

	// 最終発火時刻の設定
	if a.LastFiredAt != nil {
		alert.LastFiredAt = timestamppb.New(*a.LastFiredAt)
	}

	// 受信者の解析
	if a.Recipients != nil {
		var recipients []string
		if err := json.Unmarshal(a.Recipients, &recipients); err == nil {
			alert.Recipients = recipients
		}
	}

	// ラベルの解析
	if a.Labels != nil {
		var labels []map[string]string
		if err := json.Unmarshal(a.Labels, &labels); err == nil {
			for _, labelMap := range labels {
				for k, v := range labelMap {
					alert.Labels = append(alert.Labels, &v1.Tag{
						Key:   k,
						Value: v,
					})
				}
			}
		}
	}

	return alert
}

// FromProto はプロトコルバッファからエンティティを作成します
func (a *AlertEntity) FromProto(alert *v1.Alert) {
	a.ID = alert.Id
	a.Name = alert.Name
	a.Description = alert.Description
	a.Query = alert.Query
	a.Severity = alert.Severity
	a.Status = alert.Status

	// 間隔と期間の設定
	if alert.Interval != nil {
		a.Interval = alert.Interval.Seconds
	}
	if alert.Duration != nil {
		a.Duration = alert.Duration.Seconds
	}

	// 最終発火時刻の設定
	if alert.LastFiredAt != nil {
		lastFiredAt := alert.LastFiredAt.AsTime()
		a.LastFiredAt = &lastFiredAt
	}

	// 受信者の保存
	if len(alert.Recipients) > 0 {
		if recipientsJSON, err := json.Marshal(alert.Recipients); err == nil {
			a.Recipients = recipientsJSON
		}
	}

	// ラベルの保存
	if len(alert.Labels) > 0 {
		labels := make(map[string]string)
		for _, label := range alert.Labels {
			labels[label.Key] = label.Value
		}
		if labelsJSON, err := json.Marshal(labels); err == nil {
			a.Labels = labelsJSON
		}
	}
}

// DashboardEntity はダッシュボードエンティティです
type DashboardEntity struct {
	BaseEntity
	Name        string          `gorm:"not null" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	LayoutJSON  string          `gorm:"type:text;not null" json:"layout_json"`
	Tags        json.RawMessage `gorm:"type:jsonb" json:"tags"`
}

// ToProto はエンティティをプロトコルバッファに変換します
func (d *DashboardEntity) ToProto() *v1.Dashboard {
	dashboard := &v1.Dashboard{
		Id:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		LayoutJson:  d.LayoutJSON,
		CreatedAt:   timestamppb.New(d.CreatedAt),
		UpdatedAt:   timestamppb.New(d.UpdatedAt),
	}

	// タグの解析
	if d.Tags != nil {
		var tags []map[string]string
		if err := json.Unmarshal(d.Tags, &tags); err == nil {
			for _, tagMap := range tags {
				for k, v := range tagMap {
					dashboard.Tags = append(dashboard.Tags, &v1.Tag{
						Key:   k,
						Value: v,
					})
				}
			}
		}
	}

	return dashboard
}

// FromProto はプロトコルバッファからエンティティを作成します
func (d *DashboardEntity) FromProto(dashboard *v1.Dashboard) {
	d.ID = dashboard.Id
	d.Name = dashboard.Name
	d.Description = dashboard.Description
	d.LayoutJSON = dashboard.LayoutJson

	// タグの保存
	if len(dashboard.Tags) > 0 {
		tags := make(map[string]string)
		for _, tag := range dashboard.Tags {
			tags[tag.Key] = tag.Value
		}
		if tagsJSON, err := json.Marshal(tags); err == nil {
			d.Tags = tagsJSON
		}
	}
}

// SpanEntity はトレーススパンエンティティです
type SpanEntity struct {
	BaseEntity
	TraceID       string          `gorm:"not null;index" json:"trace_id"`
	SpanID        string          `gorm:"not null;index" json:"span_id"`
	ParentSpanID  string          `gorm:"index" json:"parent_span_id"`
	Name          string          `gorm:"not null" json:"name"`
	StartTime     time.Time       `gorm:"not null;index" json:"start_time"`
	EndTime       time.Time       `gorm:"not null;index" json:"end_time"`
	Status        int32           `gorm:"not null" json:"status"` // Status enum
	Resource      json.RawMessage `gorm:"type:jsonb" json:"resource"`
	Attributes    json.RawMessage `gorm:"type:jsonb" json:"attributes"`
	Events        json.RawMessage `gorm:"type:jsonb" json:"events"`
}

// ToProto はエンティティをプロトコルバッファに変換します
func (s *SpanEntity) ToProto() *v1.Span {
	span := &v1.Span{
		TraceId:      s.TraceID,
		SpanId:       s.SpanID,
		ParentSpanId: s.ParentSpanID,
		Name:         s.Name,
		StartTime:    timestamppb.New(s.StartTime),
		EndTime:      timestamppb.New(s.EndTime),
		Status:       v1.Status(s.Status),
		Duration:     &timestamppb.Duration{Seconds: int64(s.EndTime.Sub(s.StartTime).Seconds())},
		Resource:     &v1.Resource{},
	}

	// リソース情報の解析
	if s.Resource != nil {
		var resourceMap map[string]interface{}
		if err := json.Unmarshal(s.Resource, &resourceMap); err == nil {
			if serviceName, ok := resourceMap["service_name"].(string); ok {
				span.Resource.ServiceName = serviceName
			}
			if hostName, ok := resourceMap["host_name"].(string); ok {
				span.Resource.HostName = hostName
			}
			if instanceID, ok := resourceMap["instance_id"].(string); ok {
				span.Resource.InstanceId = instanceID
			}
		}
	}

	// 属性の解析
	if s.Attributes != nil {
		var attrs []map[string]interface{}
		if err := json.Unmarshal(s.Attributes, &attrs); err == nil {
			for _, attrMap := range attrs {
				for k, v := range attrMap {
					attr := &v1.Attribute{Key: k}
					switch val := v.(type) {
					case string:
						attr.Value = &v1.Attribute_StringValue{StringValue: val}
					case bool:
						attr.Value = &v1.Attribute_BoolValue{BoolValue: val}
					case int64:
						attr.Value = &v1.Attribute_IntValue{IntValue: val}
					case float64:
						attr.Value = &v1.Attribute_DoubleValue{DoubleValue: val}
					}
					span.Attributes = append(span.Attributes, attr)
				}
			}
		}
	}

	// イベントの解析
	if s.Events != nil {
		var events []map[string]interface{}
		if err := json.Unmarshal(s.Events, &events); err == nil {
			for _, eventMap := range events {
				event := &v1.Event{}
				if id, ok := eventMap["id"].(string); ok {
					event.Id = id
				}
				if eventType, ok := eventMap["type"].(string); ok {
					event.Type = eventType
				}
				if message, ok := eventMap["message"].(string); ok {
					event.Message = message
				}
				span.Events = append(span.Events, event)
			}
		}
	}

	return span
}

// FromProto はプロトコルバッファからエンティティを作成します
func (s *SpanEntity) FromProto(span *v1.Span) {
	s.TraceID = span.TraceId
	s.SpanID = span.SpanId
	s.ParentSpanID = span.ParentSpanId
	s.Name = span.Name
	s.StartTime = span.StartTime.AsTime()
	s.EndTime = span.EndTime.AsTime()
	s.Status = int32(span.Status)

	// リソース情報の保存
	if span.Resource != nil {
		resource := map[string]interface{}{
			"service_name": span.Resource.ServiceName,
			"host_name":    span.Resource.HostName,
			"instance_id":  span.Resource.InstanceId,
		}
		if resourceJSON, err := json.Marshal(resource); err == nil {
			s.Resource = resourceJSON
		}
	}

	// 属性の保存
	if len(span.Attributes) > 0 {
		attrs := make([]map[string]interface{}, 0, len(span.Attributes))
		for _, attr := range span.Attributes {
			attrMap := make(map[string]interface{})
			switch v := attr.Value.(type) {
			case *v1.Attribute_StringValue:
				attrMap[attr.Key] = v.StringValue
			case *v1.Attribute_BoolValue:
				attrMap[attr.Key] = v.BoolValue
			case *v1.Attribute_IntValue:
				attrMap[attr.Key] = v.IntValue
			case *v1.Attribute_DoubleValue:
				attrMap[attr.Key] = v.DoubleValue
			}
			attrs = append(attrs, attrMap)
		}
		if attrsJSON, err := json.Marshal(attrs); err == nil {
			s.Attributes = attrsJSON
		}
	}

	// イベントの保存
	if len(span.Events) > 0 {
		events := make([]map[string]interface{}, 0, len(span.Events))
		for _, event := range span.Events {
			eventMap := map[string]interface{}{
				"id":      event.Id,
				"type":    event.Type,
				"message": event.Message,
			}
			events = append(events, eventMap)
		}
		if eventsJSON, err := json.Marshal(events); err == nil {
			s.Events = eventsJSON
		}
	}
} 