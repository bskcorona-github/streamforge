package database

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// Metric represents a metric data point
type Metric struct {
	ID        string            `json:"id" db:"id"`
	Service   string            `json:"service" db:"service"`
	Metric    string            `json:"metric" db:"metric"`
	Value     float64           `json:"value" db:"value"`
	Timestamp time.Time         `json:"timestamp" db:"timestamp"`
	Labels    map[string]string `json:"labels" db:"labels"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
}

// Log represents a log entry
type Log struct {
	ID         string            `json:"id" db:"id"`
	Timestamp  time.Time         `json:"timestamp" db:"timestamp"`
	Level      string            `json:"level" db:"level"`
	Service    string            `json:"service" db:"service"`
	Message    string            `json:"message" db:"message"`
	TraceID    string            `json:"trace_id" db:"trace_id"`
	SpanID     string            `json:"span_id" db:"span_id"`
	Attributes map[string]string `json:"attributes" db:"attributes"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
}

// Trace represents a distributed trace
type Trace struct {
	ID        string    `json:"id" db:"id"`
	TraceID   string    `json:"trace_id" db:"trace_id"`
	Service   string    `json:"service" db:"service"`
	Operation string    `json:"operation" db:"operation"`
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`
	Duration  int64     `json:"duration" db:"duration"` // in nanoseconds
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Alert represents an alert rule
type Alert struct {
	ID          string            `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Description string            `json:"description" db:"description"`
	Service     string            `json:"service" db:"service"`
	Metric      string            `json:"metric" db:"metric"`
	Condition   string            `json:"condition" db:"condition"` // e.g., ">", "<", "=="
	Threshold   float64           `json:"threshold" db:"threshold"`
	Severity    string            `json:"severity" db:"severity"` // "low", "medium", "high", "critical"
	Status      string            `json:"status" db:"status"`     // "active", "inactive", "triggered"
	Labels      map[string]string `json:"labels" db:"labels"`
	CreatedBy   string            `json:"created_by" db:"created_by"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// AlertHistory represents alert trigger history
type AlertHistory struct {
	ID        string    `json:"id" db:"id"`
	AlertID   string    `json:"alert_id" db:"alert_id"`
	Service   string    `json:"service" db:"service"`
	Metric    string    `json:"metric" db:"metric"`
	Value     float64   `json:"value" db:"value"`
	Threshold float64   `json:"threshold" db:"threshold"`
	Severity  string    `json:"severity" db:"severity"`
	Message   string    `json:"message" db:"message"`
	TriggeredAt time.Time `json:"triggered_at" db:"triggered_at"`
	ResolvedAt *time.Time `json:"resolved_at" db:"resolved_at"`
	Status    string    `json:"status" db:"status"` // "triggered", "resolved", "acknowledged"
}

// Service represents a registered service
type Service struct {
	ID          string            `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Version     string            `json:"version" db:"version"`
	Environment string            `json:"environment" db:"environment"`
	Host        string            `json:"host" db:"host"`
	Port        int               `json:"port" db:"port"`
	Health      string            `json:"health" db:"health"` // "healthy", "unhealthy", "unknown"
	Metadata    map[string]string `json:"metadata" db:"metadata"`
	LastSeen    time.Time         `json:"last_seen" db:"last_seen"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// RefreshToken represents a refresh token for JWT authentication
type RefreshToken struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IsRevoked bool      `json:"is_revoked" db:"is_revoked"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// NewUser creates a new user with default values
func NewUser(username, email, password, role string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New().String(),
		Username:  username,
		Email:     email,
		Password:  password,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}
}

// NewMetric creates a new metric with default values
func NewMetric(service, metric string, value float64, labels map[string]string) *Metric {
	now := time.Now()
	return &Metric{
		ID:        uuid.New().String(),
		Service:   service,
		Metric:    metric,
		Value:     value,
		Timestamp: now,
		Labels:    labels,
		CreatedAt: now,
	}
}

// NewLog creates a new log entry with default values
func NewLog(level, service, message, traceID, spanID string, attributes map[string]string) *Log {
	now := time.Now()
	return &Log{
		ID:         uuid.New().String(),
		Timestamp:  now,
		Level:      level,
		Service:    service,
		Message:    message,
		TraceID:    traceID,
		SpanID:     spanID,
		Attributes: attributes,
		CreatedAt:  now,
	}
}

// NewAlert creates a new alert with default values
func NewAlert(name, description, service, metric, condition string, threshold float64, severity, createdBy string, labels map[string]string) *Alert {
	now := time.Now()
	return &Alert{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Service:     service,
		Metric:      metric,
		Condition:   condition,
		Threshold:   threshold,
		Severity:    severity,
		Status:      "active",
		Labels:      labels,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewService creates a new service with default values
func NewService(name, version, environment, host string, port int, metadata map[string]string) *Service {
	now := time.Now()
	return &Service{
		ID:          uuid.New().String(),
		Name:        name,
		Version:     version,
		Environment: environment,
		Host:        host,
		Port:        port,
		Health:      "unknown",
		Metadata:    metadata,
		LastSeen:    now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewRefreshToken creates a new refresh token with default values
func NewRefreshToken(userID, token string, expiresAt time.Time) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		IsRevoked: false,
		CreatedAt: now,
	}
} 