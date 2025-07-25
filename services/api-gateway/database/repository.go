package database

import (
	"context"
	"time"
)

// Repository defines the interface for database operations
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, limit, offset int) ([]*User, error)

	// Metric operations
	CreateMetric(ctx context.Context, metric *Metric) error
	GetMetricByID(ctx context.Context, id string) (*Metric, error)
	GetMetricsByService(ctx context.Context, service string, limit, offset int) ([]*Metric, error)
	GetMetricsByTimeRange(ctx context.Context, service, metric string, start, end time.Time) ([]*Metric, error)
	DeleteOldMetrics(ctx context.Context, olderThan time.Time) error

	// Log operations
	CreateLog(ctx context.Context, log *Log) error
	GetLogByID(ctx context.Context, id string) (*Log, error)
	GetLogsByService(ctx context.Context, service string, limit, offset int) ([]*Log, error)
	GetLogsByTimeRange(ctx context.Context, service string, start, end time.Time) ([]*Log, error)
	GetLogsByLevel(ctx context.Context, service, level string, limit, offset int) ([]*Log, error)
	DeleteOldLogs(ctx context.Context, olderThan time.Time) error

	// Trace operations
	CreateTrace(ctx context.Context, trace *Trace) error
	GetTraceByID(ctx context.Context, id string) (*Trace, error)
	GetTraceByTraceID(ctx context.Context, traceID string) (*Trace, error)
	GetTracesByService(ctx context.Context, service string, limit, offset int) ([]*Trace, error)
	GetTracesByTimeRange(ctx context.Context, service string, start, end time.Time) ([]*Trace, error)
	DeleteOldTraces(ctx context.Context, olderThan time.Time) error

	// Alert operations
	CreateAlert(ctx context.Context, alert *Alert) error
	GetAlertByID(ctx context.Context, id string) (*Alert, error)
	GetAlertsByService(ctx context.Context, service string) ([]*Alert, error)
	GetAlertsByStatus(ctx context.Context, status string) ([]*Alert, error)
	UpdateAlert(ctx context.Context, alert *Alert) error
	DeleteAlert(ctx context.Context, id string) error
	ListAlerts(ctx context.Context, limit, offset int) ([]*Alert, error)

	// AlertHistory operations
	CreateAlertHistory(ctx context.Context, history *AlertHistory) error
	GetAlertHistoryByID(ctx context.Context, id string) (*AlertHistory, error)
	GetAlertHistoryByAlertID(ctx context.Context, alertID string, limit, offset int) ([]*AlertHistory, error)
	UpdateAlertHistory(ctx context.Context, history *AlertHistory) error
	DeleteOldAlertHistory(ctx context.Context, olderThan time.Time) error

	// Service operations
	CreateService(ctx context.Context, service *Service) error
	GetServiceByID(ctx context.Context, id string) (*Service, error)
	GetServiceByName(ctx context.Context, name string) (*Service, error)
	UpdateService(ctx context.Context, service *Service) error
	DeleteService(ctx context.Context, id string) error
	ListServices(ctx context.Context, limit, offset int) ([]*Service, error)
	UpdateServiceHealth(ctx context.Context, id, health string) error
	UpdateServiceLastSeen(ctx context.Context, id string) error

	// RefreshToken operations
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshTokenByID(ctx context.Context, id string) (*RefreshToken, error)
	GetRefreshTokenByToken(ctx context.Context, token string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id string) error
	RevokeRefreshTokensByUser(ctx context.Context, userID string) error
	DeleteExpiredRefreshTokens(ctx context.Context) error

	// Health check
	HealthCheck(ctx context.Context) error
	Close() error
}

// MockRepository provides a mock implementation for testing
type MockRepository struct {
	users           map[string]*User
	metrics         map[string]*Metric
	logs            map[string]*Log
	traces          map[string]*Trace
	alerts          map[string]*Alert
	alertHistory    map[string]*AlertHistory
	services        map[string]*Service
	refreshTokens   map[string]*RefreshToken
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:         make(map[string]*User),
		metrics:       make(map[string]*Metric),
		logs:          make(map[string]*Log),
		traces:        make(map[string]*Trace),
		alerts:        make(map[string]*Alert),
		alertHistory:  make(map[string]*AlertHistory),
		services:      make(map[string]*Service),
		refreshTokens: make(map[string]*RefreshToken),
	}
}

// User operations
func (m *MockRepository) CreateUser(ctx context.Context, user *User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, ErrNotFound
	}
	return user, nil
}

func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *User) error {
	if _, exists := m.users[user.ID]; !exists {
		return ErrNotFound
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockRepository) DeleteUser(ctx context.Context, id string) error {
	if _, exists := m.users[id]; !exists {
		return ErrNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *MockRepository) ListUsers(ctx context.Context, limit, offset int) ([]*User, error) {
	users := make([]*User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

// Metric operations
func (m *MockRepository) CreateMetric(ctx context.Context, metric *Metric) error {
	m.metrics[metric.ID] = metric
	return nil
}

func (m *MockRepository) GetMetricByID(ctx context.Context, id string) (*Metric, error) {
	metric, exists := m.metrics[id]
	if !exists {
		return nil, ErrNotFound
	}
	return metric, nil
}

func (m *MockRepository) GetMetricsByService(ctx context.Context, service string, limit, offset int) ([]*Metric, error) {
	var metrics []*Metric
	for _, metric := range m.metrics {
		if metric.Service == service {
			metrics = append(metrics, metric)
		}
	}
	return metrics, nil
}

func (m *MockRepository) GetMetricsByTimeRange(ctx context.Context, service, metric string, start, end time.Time) ([]*Metric, error) {
	var metrics []*Metric
	for _, m := range m.metrics {
		if m.Service == service && m.Metric == metric && m.Timestamp.After(start) && m.Timestamp.Before(end) {
			metrics = append(metrics, m)
		}
	}
	return metrics, nil
}

func (m *MockRepository) DeleteOldMetrics(ctx context.Context, olderThan time.Time) error {
	for id, metric := range m.metrics {
		if metric.Timestamp.Before(olderThan) {
			delete(m.metrics, id)
		}
	}
	return nil
}

// Log operations
func (m *MockRepository) CreateLog(ctx context.Context, log *Log) error {
	m.logs[log.ID] = log
	return nil
}

func (m *MockRepository) GetLogByID(ctx context.Context, id string) (*Log, error) {
	log, exists := m.logs[id]
	if !exists {
		return nil, ErrNotFound
	}
	return log, nil
}

func (m *MockRepository) GetLogsByService(ctx context.Context, service string, limit, offset int) ([]*Log, error) {
	var logs []*Log
	for _, log := range m.logs {
		if log.Service == service {
			logs = append(logs, log)
		}
	}
	return logs, nil
}

func (m *MockRepository) GetLogsByTimeRange(ctx context.Context, service string, start, end time.Time) ([]*Log, error) {
	var logs []*Log
	for _, log := range m.logs {
		if log.Service == service && log.Timestamp.After(start) && log.Timestamp.Before(end) {
			logs = append(logs, log)
		}
	}
	return logs, nil
}

func (m *MockRepository) GetLogsByLevel(ctx context.Context, service, level string, limit, offset int) ([]*Log, error) {
	var logs []*Log
	for _, log := range m.logs {
		if log.Service == service && log.Level == level {
			logs = append(logs, log)
		}
	}
	return logs, nil
}

func (m *MockRepository) DeleteOldLogs(ctx context.Context, olderThan time.Time) error {
	for id, log := range m.logs {
		if log.Timestamp.Before(olderThan) {
			delete(m.logs, id)
		}
	}
	return nil
}

// Trace operations
func (m *MockRepository) CreateTrace(ctx context.Context, trace *Trace) error {
	m.traces[trace.ID] = trace
	return nil
}

func (m *MockRepository) GetTraceByID(ctx context.Context, id string) (*Trace, error) {
	trace, exists := m.traces[id]
	if !exists {
		return nil, ErrNotFound
	}
	return trace, nil
}

func (m *MockRepository) GetTraceByTraceID(ctx context.Context, traceID string) (*Trace, error) {
	for _, trace := range m.traces {
		if trace.TraceID == traceID {
			return trace, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) GetTracesByService(ctx context.Context, service string, limit, offset int) ([]*Trace, error) {
	var traces []*Trace
	for _, trace := range m.traces {
		if trace.Service == service {
			traces = append(traces, trace)
		}
	}
	return traces, nil
}

func (m *MockRepository) GetTracesByTimeRange(ctx context.Context, service string, start, end time.Time) ([]*Trace, error) {
	var traces []*Trace
	for _, trace := range m.traces {
		if trace.Service == service && trace.StartTime.After(start) && trace.StartTime.Before(end) {
			traces = append(traces, trace)
		}
	}
	return traces, nil
}

func (m *MockRepository) DeleteOldTraces(ctx context.Context, olderThan time.Time) error {
	for id, trace := range m.traces {
		if trace.StartTime.Before(olderThan) {
			delete(m.traces, id)
		}
	}
	return nil
}

// Alert operations
func (m *MockRepository) CreateAlert(ctx context.Context, alert *Alert) error {
	m.alerts[alert.ID] = alert
	return nil
}

func (m *MockRepository) GetAlertByID(ctx context.Context, id string) (*Alert, error) {
	alert, exists := m.alerts[id]
	if !exists {
		return nil, ErrNotFound
	}
	return alert, nil
}

func (m *MockRepository) GetAlertsByService(ctx context.Context, service string) ([]*Alert, error) {
	var alerts []*Alert
	for _, alert := range m.alerts {
		if alert.Service == service {
			alerts = append(alerts, alert)
		}
	}
	return alerts, nil
}

func (m *MockRepository) GetAlertsByStatus(ctx context.Context, status string) ([]*Alert, error) {
	var alerts []*Alert
	for _, alert := range m.alerts {
		if alert.Status == status {
			alerts = append(alerts, alert)
		}
	}
	return alerts, nil
}

func (m *MockRepository) UpdateAlert(ctx context.Context, alert *Alert) error {
	if _, exists := m.alerts[alert.ID]; !exists {
		return ErrNotFound
	}
	m.alerts[alert.ID] = alert
	return nil
}

func (m *MockRepository) DeleteAlert(ctx context.Context, id string) error {
	if _, exists := m.alerts[id]; !exists {
		return ErrNotFound
	}
	delete(m.alerts, id)
	return nil
}

func (m *MockRepository) ListAlerts(ctx context.Context, limit, offset int) ([]*Alert, error) {
	alerts := make([]*Alert, 0, len(m.alerts))
	for _, alert := range m.alerts {
		alerts = append(alerts, alert)
	}
	return alerts, nil
}

// AlertHistory operations
func (m *MockRepository) CreateAlertHistory(ctx context.Context, history *AlertHistory) error {
	m.alertHistory[history.ID] = history
	return nil
}

func (m *MockRepository) GetAlertHistoryByID(ctx context.Context, id string) (*AlertHistory, error) {
	history, exists := m.alertHistory[id]
	if !exists {
		return nil, ErrNotFound
	}
	return history, nil
}

func (m *MockRepository) GetAlertHistoryByAlertID(ctx context.Context, alertID string, limit, offset int) ([]*AlertHistory, error) {
	var history []*AlertHistory
	for _, h := range m.alertHistory {
		if h.AlertID == alertID {
			history = append(history, h)
		}
	}
	return history, nil
}

func (m *MockRepository) UpdateAlertHistory(ctx context.Context, history *AlertHistory) error {
	if _, exists := m.alertHistory[history.ID]; !exists {
		return ErrNotFound
	}
	m.alertHistory[history.ID] = history
	return nil
}

func (m *MockRepository) DeleteOldAlertHistory(ctx context.Context, olderThan time.Time) error {
	for id, history := range m.alertHistory {
		if history.TriggeredAt.Before(olderThan) {
			delete(m.alertHistory, id)
		}
	}
	return nil
}

// Service operations
func (m *MockRepository) CreateService(ctx context.Context, service *Service) error {
	m.services[service.ID] = service
	return nil
}

func (m *MockRepository) GetServiceByID(ctx context.Context, id string) (*Service, error) {
	service, exists := m.services[id]
	if !exists {
		return nil, ErrNotFound
	}
	return service, nil
}

func (m *MockRepository) GetServiceByName(ctx context.Context, name string) (*Service, error) {
	for _, service := range m.services {
		if service.Name == name {
			return service, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateService(ctx context.Context, service *Service) error {
	if _, exists := m.services[service.ID]; !exists {
		return ErrNotFound
	}
	m.services[service.ID] = service
	return nil
}

func (m *MockRepository) DeleteService(ctx context.Context, id string) error {
	if _, exists := m.services[id]; !exists {
		return ErrNotFound
	}
	delete(m.services, id)
	return nil
}

func (m *MockRepository) ListServices(ctx context.Context, limit, offset int) ([]*Service, error) {
	services := make([]*Service, 0, len(m.services))
	for _, service := range m.services {
		services = append(services, service)
	}
	return services, nil
}

func (m *MockRepository) UpdateServiceHealth(ctx context.Context, id, health string) error {
	service, exists := m.services[id]
	if !exists {
		return ErrNotFound
	}
	service.Health = health
	service.UpdatedAt = time.Now()
	return nil
}

func (m *MockRepository) UpdateServiceLastSeen(ctx context.Context, id string) error {
	service, exists := m.services[id]
	if !exists {
		return ErrNotFound
	}
	service.LastSeen = time.Now()
	service.UpdatedAt = time.Now()
	return nil
}

// RefreshToken operations
func (m *MockRepository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	m.refreshTokens[token.ID] = token
	return nil
}

func (m *MockRepository) GetRefreshTokenByID(ctx context.Context, id string) (*RefreshToken, error) {
	token, exists := m.refreshTokens[id]
	if !exists {
		return nil, ErrNotFound
	}
	return token, nil
}

func (m *MockRepository) GetRefreshTokenByToken(ctx context.Context, token string) (*RefreshToken, error) {
	for _, t := range m.refreshTokens {
		if t.Token == token {
			return t, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) RevokeRefreshToken(ctx context.Context, id string) error {
	token, exists := m.refreshTokens[id]
	if !exists {
		return ErrNotFound
	}
	token.IsRevoked = true
	return nil
}

func (m *MockRepository) RevokeRefreshTokensByUser(ctx context.Context, userID string) error {
	for _, token := range m.refreshTokens {
		if token.UserID == userID {
			token.IsRevoked = true
		}
	}
	return nil
}

func (m *MockRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	now := time.Now()
	for id, token := range m.refreshTokens {
		if token.ExpiresAt.Before(now) {
			delete(m.refreshTokens, id)
		}
	}
	return nil
}

// Health check
func (m *MockRepository) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockRepository) Close() error {
	return nil
} 