package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendMetrics(t *testing.T) {
	router := setupTestRouter()
	router.POST("/metrics", SendMetrics)

	t.Run("successful metric submission", func(t *testing.T) {
		metricData := MetricData{
			Service:   "test-service",
			Metric:    "cpu_usage",
			Value:     75.5,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Labels: map[string]string{
				"instance": "server-1",
				"region":   "us-west-1",
			},
		}

		jsonData, _ := json.Marshal(metricData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "accepted", response["status"])
		assert.Contains(t, response, "id")
		assert.Contains(t, response, "metric")
	})

	t.Run("metric submission without timestamp", func(t *testing.T) {
		metricData := MetricData{
			Service: "test-service",
			Metric:  "memory_usage",
			Value:   85.2,
			Labels: map[string]string{
				"instance": "server-2",
			},
		}

		jsonData, _ := json.Marshal(metricData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check that timestamp was auto-generated
		metric, ok := response["metric"].(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, metric, "timestamp")
		
		timestamp, ok := metric["timestamp"].(string)
		assert.True(t, ok)
		_, err = time.Parse(time.RFC3339, timestamp)
		assert.NoError(t, err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidJSON := `{"service": "test-service", "metric": "cpu_usage", "value": "invalid"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/metrics", bytes.NewBufferString(invalidJSON))
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
		metricData := map[string]interface{}{
			"service": "test-service",
			// Missing metric and value
		}

		jsonData, _ := json.Marshal(metricData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
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
		metricData := MetricData{
			Service: "",
			Metric:  "cpu_usage",
			Value:   75.5,
		}

		jsonData, _ := json.Marshal(metricData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty metric name", func(t *testing.T) {
		metricData := MetricData{
			Service: "test-service",
			Metric:  "",
			Value:   75.5,
		}

		jsonData, _ := json.Marshal(metricData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("negative value", func(t *testing.T) {
		metricData := MetricData{
			Service: "test-service",
			Metric:  "cpu_usage",
			Value:   -10.5,
		}

		jsonData, _ := json.Marshal(metricData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code) // Negative values should be allowed
	})

	t.Run("very large value", func(t *testing.T) {
		metricData := MetricData{
			Service: "test-service",
			Metric:  "cpu_usage",
			Value:   1e308, // Max float64
		}

		jsonData, _ := json.Marshal(metricData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("large labels object", func(t *testing.T) {
		labels := make(map[string]string)
		for i := 0; i < 100; i++ {
			labels[fmt.Sprintf("label_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		metricData := MetricData{
			Service: "test-service",
			Metric:  "cpu_usage",
			Value:   75.5,
			Labels:  labels,
		}

		jsonData, _ := json.Marshal(metricData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestSendMetricsResponseStructure(t *testing.T) {
	router := setupTestRouter()
	router.POST("/metrics", SendMetrics)

	metricData := MetricData{
		Service: "test-service",
		Metric:  "cpu_usage",
		Value:   75.5,
		Labels: map[string]string{
			"instance": "server-1",
		},
	}

	jsonData, _ := json.Marshal(metricData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "id")
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "metric")

	// Verify ID is a string
	id, ok := response["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, id)

	// Verify status
	assert.Equal(t, "accepted", response["status"])

	// Verify metric object
	metric, ok := response["metric"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test-service", metric["service"])
	assert.Equal(t, "cpu_usage", metric["metric"])
	assert.Equal(t, 75.5, metric["value"])
}

func TestSendMetricsContentType(t *testing.T) {
	router := setupTestRouter()
	router.POST("/metrics", SendMetrics)

	metricData := MetricData{
		Service: "test-service",
		Metric:  "cpu_usage",
		Value:   75.5,
	}

	jsonData, _ := json.Marshal(metricData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/metrics", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
} 