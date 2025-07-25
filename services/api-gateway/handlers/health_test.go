package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Setup logger
	logger, _ := zap.NewDevelopment()
	router.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Next()
	})
	
	return router
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()
	router.GET("/health", HealthCheck)

	t.Run("successful health check", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check required fields
		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "api-gateway", response["service"])
		assert.Equal(t, "1.0.0", response["version"])
		
		// Check timestamp format
		timestamp, ok := response["timestamp"].(string)
		assert.True(t, ok)
		_, err = time.Parse(time.RFC3339, timestamp)
		assert.NoError(t, err)
	})

	t.Run("health check with logger", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	})
}

func TestHealthCheckResponseStructure(t *testing.T) {
	router := setupTestRouter()
	router.GET("/health", HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify all expected fields are present
	expectedFields := []string{"status", "timestamp", "service", "version"}
	for _, field := range expectedFields {
		assert.Contains(t, response, field, "Response should contain field: %s", field)
	}

	// Verify field types
	assert.IsType(t, "", response["status"])
	assert.IsType(t, "", response["timestamp"])
	assert.IsType(t, "", response["service"])
	assert.IsType(t, "", response["version"])
}

func TestHealthCheckTimestampFormat(t *testing.T) {
	router := setupTestRouter()
	router.GET("/health", HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	timestamp, ok := response["timestamp"].(string)
	assert.True(t, ok, "Timestamp should be a string")

	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	assert.NoError(t, err, "Timestamp should be in RFC3339 format")

	// Verify timestamp is recent (within last minute)
	now := time.Now()
	diff := now.Sub(parsedTime)
	assert.True(t, diff < time.Minute, "Timestamp should be recent")
}

func TestHealthCheckServiceInfo(t *testing.T) {
	router := setupTestRouter()
	router.GET("/health", HealthCheck)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "api-gateway", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
} 