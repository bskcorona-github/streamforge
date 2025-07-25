use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};
use tokio::sync::mpsc;
use tokio_tungstenite::{connect_async, tungstenite::protocol::Message};
use url::Url;

/// StreamForge client configuration
#[derive(Debug, Clone)]
pub struct Config {
    pub api_url: String,
    pub ws_url: String,
    pub api_key: Option<String>,
    pub timeout: std::time::Duration,
    pub retries: u32,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            api_url: "http://localhost:8080".to_string(),
            ws_url: "ws://localhost:8080".to_string(),
            api_key: None,
            timeout: std::time::Duration::from_secs(30),
            retries: 3,
        }
    }
}

/// Metric data point
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Metric {
    pub name: String,
    pub value: f64,
    pub unit: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub labels: Option<HashMap<String, String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub timestamp: Option<u64>,
}

/// Log entry
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LogEntry {
    pub level: String,
    pub message: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub fields: Option<HashMap<String, serde_json::Value>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub timestamp: Option<u64>,
}

/// Alert
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Alert {
    pub id: String,
    pub severity: String,
    pub message: String,
    pub timestamp: u64,
    pub service: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<HashMap<String, serde_json::Value>>,
}

/// Service status
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServiceStatus {
    pub name: String,
    pub status: String,
    pub uptime: f64,
    pub response_time: f64,
    pub requests_per_second: f64,
    pub last_check: u64,
}

/// Health check response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthCheck {
    pub status: String,
    pub timestamp: u64,
}

/// StreamForge API error
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StreamForgeError {
    pub message: String,
    pub status_code: u16,
    pub code: Option<String>,
}

impl std::fmt::Display for StreamForgeError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "StreamForge error: {} (status: {}, code: {:?})",
            self.message, self.status_code, self.code
        )
    }
}

impl std::error::Error for StreamForgeError {}

/// WebSocket message types
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum WebSocketMessage {
    #[serde(rename = "metrics")]
    Metrics { metrics: Vec<Metric> },
    #[serde(rename = "alerts")]
    Alerts { alerts: Vec<Alert> },
    #[serde(rename = "service_status")]
    ServiceStatus { services: Vec<ServiceStatus> },
}

/// WebSocket callbacks
pub struct WebSocketCallbacks {
    pub on_metrics: Option<Box<dyn Fn(Vec<Metric>) + Send + Sync>>,
    pub on_alerts: Option<Box<dyn Fn(Vec<Alert>) + Send + Sync>>,
    pub on_service_status: Option<Box<dyn Fn(Vec<ServiceStatus>) + Send + Sync>>,
    pub on_error: Option<Box<dyn Fn(String) + Send + Sync>>,
    pub on_connect: Option<Box<dyn Fn() + Send + Sync>>,
    pub on_disconnect: Option<Box<dyn Fn() + Send + Sync>>,
}

/// StreamForge client
pub struct Client {
    config: Config,
    http_client: reqwest::Client,
}

impl Client {
    /// Create a new StreamForge client
    pub fn new(config: Config) -> Self {
        let http_client = reqwest::Client::builder()
            .timeout(config.timeout)
            .build()
            .expect("Failed to create HTTP client");

        Self {
            config,
            http_client,
        }
    }

    /// Send metrics to the StreamForge API
    pub async fn send_metrics(&self, metrics: Vec<Metric>) -> Result<(), StreamForgeError> {
        let payload = serde_json::json!({
            "metrics": metrics
        });

        self.make_request("POST", "/api/v1/metrics", Some(payload)).await?;
        Ok(())
    }

    /// Send logs to the StreamForge API
    pub async fn send_logs(&self, logs: Vec<LogEntry>) -> Result<(), StreamForgeError> {
        let payload = serde_json::json!({
            "logs": logs
        });

        self.make_request("POST", "/api/v1/logs", Some(payload)).await?;
        Ok(())
    }

    /// Get metrics from the StreamForge API
    pub async fn get_metrics(
        &self,
        filters: Option<HashMap<String, String>>,
    ) -> Result<Vec<Metric>, StreamForgeError> {
        let mut url = format!("{}/api/v1/metrics", self.config.api_url);
        
        if let Some(filters) = filters {
            let params: Vec<String> = filters
                .iter()
                .map(|(k, v)| format!("{}={}", k, v))
                .collect();
            if !params.is_empty() {
                url.push_str(&format!("?{}", params.join("&")));
            }
        }

        let response: serde_json::Value = self.make_request("GET", &url, None).await?;
        
        let metrics = response["metrics"]
            .as_array()
            .ok_or_else(|| StreamForgeError {
                message: "Invalid response format".to_string(),
                status_code: 0,
                code: None,
            })?
            .iter()
            .filter_map(|m| serde_json::from_value(m.clone()).ok())
            .collect();

        Ok(metrics)
    }

    /// Get alerts from the StreamForge API
    pub async fn get_alerts(
        &self,
        filters: Option<HashMap<String, String>>,
    ) -> Result<Vec<Alert>, StreamForgeError> {
        let mut url = format!("{}/api/v1/alerts", self.config.api_url);
        
        if let Some(filters) = filters {
            let params: Vec<String> = filters
                .iter()
                .map(|(k, v)| format!("{}={}", k, v))
                .collect();
            if !params.is_empty() {
                url.push_str(&format!("?{}", params.join("&")));
            }
        }

        let response: serde_json::Value = self.make_request("GET", &url, None).await?;
        
        let alerts = response["alerts"]
            .as_array()
            .ok_or_else(|| StreamForgeError {
                message: "Invalid response format".to_string(),
                status_code: 0,
                code: None,
            })?
            .iter()
            .filter_map(|a| serde_json::from_value(a.clone()).ok())
            .collect();

        Ok(alerts)
    }

    /// Get service status from the StreamForge API
    pub async fn get_service_status(&self) -> Result<Vec<ServiceStatus>, StreamForgeError> {
        let response: serde_json::Value = self
            .make_request("GET", "/api/v1/services/status", None)
            .await?;
        
        let services = response["services"]
            .as_array()
            .ok_or_else(|| StreamForgeError {
                message: "Invalid response format".to_string(),
                status_code: 0,
                code: None,
            })?
            .iter()
            .filter_map(|s| serde_json::from_value(s.clone()).ok())
            .collect();

        Ok(services)
    }

    /// Perform a health check
    pub async fn health_check(&self) -> Result<HealthCheck, StreamForgeError> {
        let response: serde_json::Value = self.make_request("GET", "/api/v1/health", None).await?;
        
        serde_json::from_value(response).map_err(|_| StreamForgeError {
            message: "Failed to parse health check response".to_string(),
            status_code: 0,
            code: None,
        })
    }

    /// Connect to WebSocket
    pub async fn connect_websocket(
        &self,
        callbacks: WebSocketCallbacks,
    ) -> Result<(), Box<dyn std::error::Error>> {
        let mut ws_url = self.config.ws_url.clone();
        if let Some(api_key) = &self.config.api_key {
            ws_url.push_str(&format!("?api_key={}", api_key));
        }

        let url = Url::parse(&ws_url)?;
        let (ws_stream, _) = connect_async(url).await?;
        let (write, read) = ws_stream.split();

        // Call on_connect callback
        if let Some(on_connect) = callbacks.on_connect {
            on_connect();
        }

        // Handle incoming messages
        tokio::spawn(async move {
            use futures_util::StreamExt;
            use futures_util::SinkExt;

            let mut read = read;
            let mut write = write;

            while let Some(msg) = read.next().await {
                match msg {
                    Ok(Message::Text(text)) => {
                        match serde_json::from_str::<WebSocketMessage>(&text) {
                            Ok(message) => {
                                match message {
                                    WebSocketMessage::Metrics { metrics } => {
                                        if let Some(on_metrics) = &callbacks.on_metrics {
                                            on_metrics(metrics);
                                        }
                                    }
                                    WebSocketMessage::Alerts { alerts } => {
                                        if let Some(on_alerts) = &callbacks.on_alerts {
                                            on_alerts(alerts);
                                        }
                                    }
                                    WebSocketMessage::ServiceStatus { services } => {
                                        if let Some(on_service_status) = &callbacks.on_service_status {
                                            on_service_status(services);
                                        }
                                    }
                                }
                            }
                            Err(e) => {
                                if let Some(on_error) = &callbacks.on_error {
                                    on_error(format!("Failed to parse WebSocket message: {}", e));
                                }
                            }
                        }
                    }
                    Ok(Message::Close(_)) => {
                        if let Some(on_disconnect) = callbacks.on_disconnect {
                            on_disconnect();
                        }
                        break;
                    }
                    Err(e) => {
                        if let Some(on_error) = &callbacks.on_error {
                            on_error(format!("WebSocket error: {}", e));
                        }
                        break;
                    }
                    _ => {}
                }
            }
        });

        Ok(())
    }

    async fn make_request(
        &self,
        method: &str,
        path: &str,
        payload: Option<serde_json::Value>,
    ) -> Result<serde_json::Value, StreamForgeError> {
        let url = if path.starts_with("http") {
            path.to_string()
        } else {
            format!("{}{}", self.config.api_url, path)
        };

        let mut request = self.http_client.request(
            method.parse().unwrap(),
            &url,
        );

        request = request.header("Content-Type", "application/json");
        
        if let Some(api_key) = &self.config.api_key {
            request = request.header("Authorization", format!("Bearer {}", api_key));
        }

        if let Some(payload) = payload {
            request = request.json(&payload);
        }

        let response = request.send().await.map_err(|e| StreamForgeError {
            message: format!("Request failed: {}", e),
            status_code: 0,
            code: None,
        })?;

        let status = response.status();
        let body = response.json::<serde_json::Value>().await.map_err(|e| StreamForgeError {
            message: format!("Failed to parse response: {}", e),
            status_code: status.as_u16(),
            code: None,
        })?;

        if status.is_success() {
            Ok(body)
        } else {
            Err(StreamForgeError {
                message: body["message"]
                    .as_str()
                    .unwrap_or("Unknown error")
                    .to_string(),
                status_code: status.as_u16(),
                code: body["code"].as_str().map(|s| s.to_string()),
            })
        }
    }
}

/// Helper functions
pub fn create_metric(
    name: String,
    value: f64,
    unit: String,
    labels: Option<HashMap<String, String>>,
) -> Metric {
    let timestamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs();

    Metric {
        name,
        value,
        unit,
        labels,
        timestamp: Some(timestamp),
    }
}

pub fn create_log_entry(
    level: String,
    message: String,
    fields: Option<HashMap<String, serde_json::Value>>,
) -> LogEntry {
    let timestamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs();

    LogEntry {
        level,
        message,
        fields,
        timestamp: Some(timestamp),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_create_metric() {
        let metric = create_metric(
            "test_metric".to_string(),
            42.0,
            "count".to_string(),
            None,
        );
        assert_eq!(metric.name, "test_metric");
        assert_eq!(metric.value, 42.0);
        assert_eq!(metric.unit, "count");
        assert!(metric.timestamp.is_some());
    }

    #[test]
    fn test_create_log_entry() {
        let log = create_log_entry(
            "info".to_string(),
            "test message".to_string(),
            None,
        );
        assert_eq!(log.level, "info");
        assert_eq!(log.message, "test message");
        assert!(log.timestamp.is_some());
    }
} 