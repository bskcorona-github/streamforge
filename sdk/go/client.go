package streamforge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// Config represents the StreamForge client configuration
type Config struct {
	APIURL   string
	WSURL    string
	APIKey   string
	Timeout  time.Duration
	Retries  int
}

// Metric represents a metric data point
type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
}

// Alert represents an alert
type Alert struct {
	ID        string                 `json:"id"`
	Severity  string                 `json:"severity"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	Uptime          float64   `json:"uptime"`
	ResponseTime    float64   `json:"responseTime"`
	RequestsPerSec  float64   `json:"requestsPerSecond"`
	LastCheck       time.Time `json:"lastCheck"`
}

// HealthCheck represents a health check response
type HealthCheck struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// StreamForgeError represents a StreamForge API error
type StreamForgeError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Code       string `json:"code"`
}

func (e *StreamForgeError) Error() string {
	return fmt.Sprintf("StreamForge error: %s (status: %d, code: %s)", e.Message, e.StatusCode, e.Code)
}

// Client represents a StreamForge client
type Client struct {
	config     *Config
	httpClient *http.Client
	wsConn     *websocket.Conn
}

// NewClient creates a new StreamForge client
func NewClient(config *Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.Retries == 0 {
		config.Retries = 3
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// SendMetrics sends metrics to the StreamForge API
func (c *Client) SendMetrics(ctx context.Context, metrics []Metric) error {
	payload := map[string]interface{}{
		"metrics": metrics,
	}

	return c.makeRequest(ctx, "POST", "/api/v1/metrics", payload, nil)
}

// SendLogs sends logs to the StreamForge API
func (c *Client) SendLogs(ctx context.Context, logs []LogEntry) error {
	payload := map[string]interface{}{
		"logs": logs,
	}

	return c.makeRequest(ctx, "POST", "/api/v1/logs", payload, nil)
}

// GetMetrics retrieves metrics from the StreamForge API
func (c *Client) GetMetrics(ctx context.Context, filters map[string]string) ([]Metric, error) {
	params := url.Values{}
	for key, value := range filters {
		params.Add(key, value)
	}

	var response struct {
		Metrics []Metric `json:"metrics"`
	}

	err := c.makeRequest(ctx, "GET", "/api/v1/metrics?"+params.Encode(), nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Metrics, nil
}

// GetAlerts retrieves alerts from the StreamForge API
func (c *Client) GetAlerts(ctx context.Context, filters map[string]string) ([]Alert, error) {
	params := url.Values{}
	for key, value := range filters {
		params.Add(key, value)
	}

	var response struct {
		Alerts []Alert `json:"alerts"`
	}

	err := c.makeRequest(ctx, "GET", "/api/v1/alerts?"+params.Encode(), nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Alerts, nil
}

// GetServiceStatus retrieves service status from the StreamForge API
func (c *Client) GetServiceStatus(ctx context.Context) ([]ServiceStatus, error) {
	var response struct {
		Services []ServiceStatus `json:"services"`
	}

	err := c.makeRequest(ctx, "GET", "/api/v1/services/status", nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Services, nil
}

// HealthCheck performs a health check
func (c *Client) HealthCheck(ctx context.Context) (*HealthCheck, error) {
	var health HealthCheck

	err := c.makeRequest(ctx, "GET", "/api/v1/health", nil, &health)
	if err != nil {
		return nil, err
	}

	return &health, nil
}

// ConnectWebSocket connects to the WebSocket endpoint
func (c *Client) ConnectWebSocket(callbacks WebSocketCallbacks) error {
	wsURL := c.config.WSURL
	if c.config.APIKey != "" {
		wsURL += "?api_key=" + c.config.APIKey
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect WebSocket: %w", err)
	}

	c.wsConn = conn

	// Start listening for messages
	go c.listenWebSocket(callbacks)

	return nil
}

// DisconnectWebSocket disconnects from the WebSocket
func (c *Client) DisconnectWebSocket() error {
	if c.wsConn != nil {
		return c.wsConn.Close()
	}
	return nil
}

// WebSocketCallbacks defines callbacks for WebSocket events
type WebSocketCallbacks struct {
	OnMetrics       func([]Metric)
	OnAlerts        func([]Alert)
	OnServiceStatus func([]ServiceStatus)
	OnError         func(error)
	OnConnect       func()
	OnDisconnect    func()
}

func (c *Client) listenWebSocket(callbacks WebSocketCallbacks) {
	defer func() {
		if callbacks.OnDisconnect != nil {
			callbacks.OnDisconnect()
		}
	}()

	for {
		_, message, err := c.wsConn.ReadMessage()
		if err != nil {
			if callbacks.OnError != nil {
				callbacks.OnError(fmt.Errorf("WebSocket read error: %w", err))
			}
			return
		}

		var data map[string]interface{}
		if err := json.Unmarshal(message, &data); err != nil {
			if callbacks.OnError != nil {
				callbacks.OnError(fmt.Errorf("failed to parse WebSocket message: %w", err))
			}
			continue
		}

		messageType, ok := data["type"].(string)
		if !ok {
			continue
		}

		switch messageType {
		case "metrics":
			if callbacks.OnMetrics != nil {
				var metrics []Metric
				if metricsData, ok := data["metrics"].([]interface{}); ok {
					metricsBytes, _ := json.Marshal(metricsData)
					json.Unmarshal(metricsBytes, &metrics)
					callbacks.OnMetrics(metrics)
				}
			}
		case "alerts":
			if callbacks.OnAlerts != nil {
				var alerts []Alert
				if alertsData, ok := data["alerts"].([]interface{}); ok {
					alertsBytes, _ := json.Marshal(alertsData)
					json.Unmarshal(alertsBytes, &alerts)
					callbacks.OnAlerts(alerts)
				}
			}
		case "service_status":
			if callbacks.OnServiceStatus != nil {
				var services []ServiceStatus
				if servicesData, ok := data["services"].([]interface{}); ok {
					servicesBytes, _ := json.Marshal(servicesData)
					json.Unmarshal(servicesBytes, &services)
					callbacks.OnServiceStatus(services)
				}
			}
		}
	}
}

func (c *Client) makeRequest(ctx context.Context, method, path string, payload interface{}, response interface{}) error {
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			return fmt.Errorf("failed to encode payload: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.config.APIURL+path, &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiError StreamForgeError
		if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		apiError.StatusCode = resp.StatusCode
		return &apiError
	}

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Helper functions
func CreateMetric(name string, value float64, unit string, labels map[string]string) Metric {
	return Metric{
		Name:      name,
		Value:     value,
		Unit:      unit,
		Labels:    labels,
		Timestamp: time.Now(),
	}
}

func CreateLogEntry(level, message string, fields map[string]interface{}) LogEntry {
	return LogEntry{
		Level:     level,
		Message:   message,
		Fields:    fields,
		Timestamp: time.Now(),
	}
} 