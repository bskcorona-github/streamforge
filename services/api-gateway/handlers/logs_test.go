package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendLogs(t *testing.T) {
	router := setupTestRouter()
	router.POST("/logs", SendLogs)

	t.Run("successful log submission", func(t *testing.T) {
		logData := LogData{
			Level:   "info",
			Service: "test-service",
			Message: "Test log message",
			TraceID: "trace-123",
			SpanID:  "span-456",
			Attributes: map[string]string{
				"user_id": "user-123",
				"action":  "login",
			},
		}

		jsonData, _ := json.Marshal(logData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/logs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "accepted", response["status"])
		assert.Contains(t, response, "id")
		assert.Contains(t, response, "log")
	})

	t.Run("log submission without timestamp", func(t *testing.T) {
		logData := LogData{
			Level:   "error",
			Service: "test-service",
			Message: "Error occurred",
		}

		jsonData, _ := json.Marshal(logData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/logs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check that timestamp was auto-generated
		log, ok := response["log"].(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, log, "timestamp")
		
		timestamp, ok := log["timestamp"].(string)
		assert.True(t, ok)
		_, err = time.Parse(time.RFC3339, timestamp)
		assert.NoError(t, err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidJSON := `{"level": "info", "service": "test-service"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/logs", bytes.NewBufferString(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response, "details")
	})

	t.Run("missing required fields", func(t *testing.T) {
		logData := map[string]interface{}{
			"level": "info",
			// Missing service and message
		}

		jsonData, _ := json.Marshal(logData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/logs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response, "details")
	})

	t.Run("empty service name", func(t *testing.T) {
		logData := LogData{
			Level:   "info",
			Service: "",
			Message: "Test message",
		}

		jsonData, _ := json.Marshal(logData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/logs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty message", func(t *testing.T) {
		logData := LogData{
			Level:   "info",
			Service: "test-service",
			Message: "",
		}

		jsonData, _ := json.Marshal(logData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/logs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid log level", func(t *testing.T) {
		logData := LogData{
			Level:   "invalid-level",
			Service: "test-service",
			Message: "Test message",
		}

		jsonData, _ := json.Marshal(logData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/logs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code) // Should accept any level
	})

	t.Run("with trace and span IDs", func(t *testing.T) {
		logData := LogData{
			Level:   "debug",
			Service: "test-service",
			Message: "Debug message",
			TraceID: "trace-789",
			SpanID:  "span-012",
		}

		jsonData, _ := json.Marshal(logData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/logs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		log, ok := response["log"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "trace-789", log["trace_id"])
		assert.Equal(t, "span-012", log["span_id"])
	})

	t.Run("with attributes", func(t *testing.T) {
		logData := LogData{
			Level:   "warn",
			Service: "test-service",
			Message: "Warning message",
			Attributes: map[string]string{
				"user_id":    "user-456",
				"session_id": "session-789",
				"ip_address": "192.168.1.1",
			},
		}

		jsonData, _ := json.Marshal(logData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/logs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		log, ok := response["log"].(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, log, "attributes")
	})
}

func TestSendLogsResponseStructure(t *testing.T) {
	router := setupTestRouter()
	router.POST("/logs", SendLogs)

	logData := LogData{
		Level:   "info",
		Service: "test-service",
		Message: "Test message",
		TraceID: "trace-123",
		SpanID:  "span-456",
		Attributes: map[string]string{
			"user_id": "user-123",
		},
	}

	jsonData, _ := json.Marshal(logData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/logs", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "id")
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "log")

	// Verify ID is a string
	id, ok := response["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, id)

	// Verify status
	assert.Equal(t, "accepted", response["status"])

	// Verify log object
	log, ok := response["log"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "info", log["level"])
	assert.Equal(t, "test-service", log["service"])
	assert.Equal(t, "Test message", log["message"])
	assert.Equal(t, "trace-123", log["trace_id"])
	assert.Equal(t, "span-456", log["span_id"])
}

func TestSendLogsContentType(t *testing.T) {
	router := setupTestRouter()
	router.POST("/logs", SendLogs)

	logData := LogData{
		Level:   "info",
		Service: "test-service",
		Message: "Test message",
	}

	jsonData, _ := json.Marshal(logData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/logs", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestSendLogsValidLevels(t *testing.T) {
	router := setupTestRouter()
	router.POST("/logs", SendLogs)

	validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}

	for _, level := range validLevels {
		t.Run("level_"+level, func(t *testing.T) {
			logData := LogData{
				Level:   level,
				Service: "test-service",
				Message: "Test message",
			}

			jsonData, _ := json.Marshal(logData)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/logs", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)
		})
	}
} 