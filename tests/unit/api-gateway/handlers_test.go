package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/streamforge/streamforge/services/api-gateway/handlers"
	"github.com/streamforge/streamforge/services/api-gateway/middleware"
	"github.com/streamforge/streamforge/services/api-gateway/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock services
type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) SendMetrics(metrics []models.Metric) error {
	args := m.Called(metrics)
	return args.Error(0)
}

func (m *MockMetricsService) GetMetrics(filters map[string]string) ([]models.Metric, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Metric), args.Error(1)
}

type MockLogsService struct {
	mock.Mock
}

func (m *MockLogsService) SendLogs(logs []models.LogEntry) error {
	args := m.Called(logs)
	return args.Error(0)
}

func (m *MockLogsService) GetLogs(filters map[string]string) ([]models.LogEntry, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.LogEntry), args.Error(1)
}

type MockAlertsService struct {
	mock.Mock
}

func (m *MockAlertsService) GetAlerts(filters map[string]string) ([]models.Alert, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Alert), args.Error(1)
}

func (m *MockAlertsService) CreateAlert(alert models.Alert) error {
	args := m.Called(alert)
	return args.Error(0)
}

// Test setup
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(nil))
	router.Use(middleware.Recovery(nil))
	return router
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()
	handler := handlers.NewHealthHandler()
	router.GET("/health", handler.HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestSendMetrics(t *testing.T) {
	router := setupTestRouter()
	
	mockMetricsService := new(MockMetricsService)
	handler := handlers.NewMetricsHandler(mockMetricsService)
	
	router.POST("/api/v1/metrics", handler.SendMetrics)

	// Test data
	metrics := []models.Metric{
		{
			Name:  "cpu_usage",
			Value: 75.5,
			Unit:  "percent",
			Labels: map[string]string{
				"service": "api-gateway",
				"instance": "pod-1",
			},
		},
	}

	// Setup mock expectations
	mockMetricsService.On("SendMetrics", metrics).Return(nil)

	// Create request
	payload, _ := json.Marshal(map[string]interface{}{
		"metrics": metrics,
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/metrics", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockMetricsService.AssertExpectations(t)
}

func TestSendMetricsInvalidPayload(t *testing.T) {
	router := setupTestRouter()
	
	mockMetricsService := new(MockMetricsService)
	handler := handlers.NewMetricsHandler(mockMetricsService)
	
	router.POST("/api/v1/metrics", handler.SendMetrics)

	// Invalid JSON payload
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/metrics", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMetrics(t *testing.T) {
	router := setupTestRouter()
	
	mockMetricsService := new(MockMetricsService)
	handler := handlers.NewMetricsHandler(mockMetricsService)
	
	router.GET("/api/v1/metrics", handler.GetMetrics)

	// Expected metrics
	expectedMetrics := []models.Metric{
		{
			Name:  "cpu_usage",
			Value: 75.5,
			Unit:  "percent",
			Labels: map[string]string{
				"service": "api-gateway",
			},
		},
	}

	// Setup mock expectations
	mockMetricsService.On("GetMetrics", map[string]string{
		"service": "api-gateway",
	}).Return(expectedMetrics, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/metrics?service=api-gateway", nil)
	
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	metrics := response["metrics"].([]interface{})
	assert.Len(t, metrics, 1)
	
	mockMetricsService.AssertExpectations(t)
}

func TestSendLogs(t *testing.T) {
	router := setupTestRouter()
	
	mockLogsService := new(MockLogsService)
	handler := handlers.NewLogsHandler(mockLogsService)
	
	router.POST("/api/v1/logs", handler.SendLogs)

	// Test data
	logs := []models.LogEntry{
		{
			Level:   "info",
			Message: "API request processed",
			Fields: map[string]interface{}{
				"method": "GET",
				"path":   "/api/v1/metrics",
			},
		},
	}

	// Setup mock expectations
	mockLogsService.On("SendLogs", logs).Return(nil)

	// Create request
	payload, _ := json.Marshal(map[string]interface{}{
		"logs": logs,
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/logs", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockLogsService.AssertExpectations(t)
}

func TestGetLogs(t *testing.T) {
	router := setupTestRouter()
	
	mockLogsService := new(MockLogsService)
	handler := handlers.NewLogsHandler(mockLogsService)
	
	router.GET("/api/v1/logs", handler.GetLogs)

	// Expected logs
	expectedLogs := []models.LogEntry{
		{
			Level:   "info",
			Message: "API request processed",
			Fields: map[string]interface{}{
				"method": "GET",
				"path":   "/api/v1/metrics",
			},
		},
	}

	// Setup mock expectations
	mockLogsService.On("GetLogs", map[string]string{
		"level": "info",
	}).Return(expectedLogs, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/logs?level=info", nil)
	
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	logs := response["logs"].([]interface{})
	assert.Len(t, logs, 1)
	
	mockLogsService.AssertExpectations(t)
}

func TestGetAlerts(t *testing.T) {
	router := setupTestRouter()
	
	mockAlertsService := new(MockAlertsService)
	handler := handlers.NewAlertsHandler(mockAlertsService)
	
	router.GET("/api/v1/alerts", handler.GetAlerts)

	// Expected alerts
	expectedAlerts := []models.Alert{
		{
			ID:       "alert-1",
			Severity: "warning",
			Message:  "High CPU usage detected",
			Service:  "api-gateway",
		},
	}

	// Setup mock expectations
	mockAlertsService.On("GetAlerts", map[string]string{
		"severity": "warning",
	}).Return(expectedAlerts, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/alerts?severity=warning", nil)
	
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	alerts := response["alerts"].([]interface{})
	assert.Len(t, alerts, 1)
	
	mockAlertsService.AssertExpectations(t)
}

func TestCreateAlert(t *testing.T) {
	router := setupTestRouter()
	
	mockAlertsService := new(MockAlertsService)
	handler := handlers.NewAlertsHandler(mockAlertsService)
	
	router.POST("/api/v1/alerts", handler.CreateAlert)

	// Test data
	alert := models.Alert{
		Severity: "critical",
		Message:  "Service down",
		Service:  "api-gateway",
	}

	// Setup mock expectations
	mockAlertsService.On("CreateAlert", mock.MatchedBy(func(a models.Alert) bool {
		return a.Severity == alert.Severity && a.Message == alert.Message && a.Service == alert.Service
	})).Return(nil)

	// Create request
	payload, _ := json.Marshal(alert)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/alerts", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockAlertsService.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkHealthCheck(b *testing.B) {
	router := setupTestRouter()
	handler := handlers.NewHealthHandler()
	router.GET("/health", handler.HealthCheck)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkSendMetrics(b *testing.B) {
	router := setupTestRouter()
	
	mockMetricsService := new(MockMetricsService)
	handler := handlers.NewMetricsHandler(mockMetricsService)
	
	router.POST("/api/v1/metrics", handler.SendMetrics)

	metrics := []models.Metric{
		{
			Name:  "cpu_usage",
			Value: 75.5,
			Unit:  "percent",
		},
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"metrics": metrics,
	})

	mockMetricsService.On("SendMetrics", metrics).Return(nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/metrics", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
} 