use serde::{Deserialize, Serialize};
use std::time::Duration;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    pub kafka: KafkaConfig,
    pub database: DatabaseConfig,
    pub metrics: MetricsConfig,
    pub telemetry: TelemetryConfig,
    pub processing: ProcessingConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct KafkaConfig {
    pub bootstrap_servers: String,
    pub group_id: String,
    pub topics: Vec<String>,
    pub auto_offset_reset: String,
    pub enable_auto_commit: bool,
    pub session_timeout_ms: i32,
    pub heartbeat_interval_ms: i32,
    pub max_poll_records: i32,
    pub fetch_max_wait_ms: i32,
    pub fetch_min_bytes: i32,
    pub fetch_max_bytes: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DatabaseConfig {
    pub url: String,
    pub max_connections: u32,
    pub min_connections: u32,
    pub connect_timeout: Duration,
    pub idle_timeout: Duration,
    pub max_lifetime: Duration,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MetricsConfig {
    pub enabled: bool,
    pub port: u16,
    pub host: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TelemetryConfig {
    pub jaeger_endpoint: String,
    pub service_name: String,
    pub service_version: String,
    pub environment: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProcessingConfig {
    pub batch_size: usize,
    pub batch_timeout: Duration,
    pub max_concurrent_tasks: usize,
    pub retry_attempts: u32,
    pub retry_delay: Duration,
    pub dead_letter_queue_topic: String,
}

impl Config {
    pub fn load(path: &str) -> crate::Result<Self> {
        let config = config::Config::builder()
            .add_source(config::File::with_name(path).required(false))
            .add_source(config::Environment::with_prefix("STREAM_PROCESSOR"))
            .build()?;

        let config: Config = config.try_deserialize()?;
        Ok(config)
    }

    pub fn default() -> Self {
        Self {
            kafka: KafkaConfig::default(),
            database: DatabaseConfig::default(),
            metrics: MetricsConfig::default(),
            telemetry: TelemetryConfig::default(),
            processing: ProcessingConfig::default(),
        }
    }
}

impl Default for KafkaConfig {
    fn default() -> Self {
        Self {
            bootstrap_servers: "localhost:9092".to_string(),
            group_id: "stream-processor-group".to_string(),
            topics: vec!["metrics".to_string(), "logs".to_string(), "traces".to_string()],
            auto_offset_reset: "earliest".to_string(),
            enable_auto_commit: false,
            session_timeout_ms: 30000,
            heartbeat_interval_ms: 3000,
            max_poll_records: 500,
            fetch_max_wait_ms: 500,
            fetch_min_bytes: 1,
            fetch_max_bytes: 52428800, // 50MB
        }
    }
}

impl Default for DatabaseConfig {
    fn default() -> Self {
        Self {
            url: "postgresql://streamforge:password@localhost:5432/streamforge".to_string(),
            max_connections: 10,
            min_connections: 2,
            connect_timeout: Duration::from_secs(30),
            idle_timeout: Duration::from_secs(300),
            max_lifetime: Duration::from_secs(3600),
        }
    }
}

impl Default for MetricsConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            port: 9090,
            host: "0.0.0.0".to_string(),
        }
    }
}

impl Default for TelemetryConfig {
    fn default() -> Self {
        Self {
            jaeger_endpoint: "http://localhost:14268/api/traces".to_string(),
            service_name: "stream-processor".to_string(),
            service_version: env!("CARGO_PKG_VERSION").to_string(),
            environment: "development".to_string(),
        }
    }
}

impl Default for ProcessingConfig {
    fn default() -> Self {
        Self {
            batch_size: 1000,
            batch_timeout: Duration::from_secs(5),
            max_concurrent_tasks: 10,
            retry_attempts: 3,
            retry_delay: Duration::from_secs(1),
            dead_letter_queue_topic: "dlq".to_string(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::NamedTempFile;
    use std::fs;

    #[test]
    fn test_config_load() {
        let config_content = r#"
[app]
name = "test-processor"
version = "1.0.0"
environment = "test"
log_level = "debug"

[kafka]
bootstrap_servers = "localhost:9092"
group_id = "test-group"
input_topics = ["test.input"]
output_topics = ["test.output"]
auto_offset_reset = "earliest"
enable_auto_commit = false
session_timeout_ms = 30000
heartbeat_interval_ms = 3000
max_poll_records = 500
fetch_max_wait_ms = 500
fetch_min_bytes = 1
fetch_max_bytes = 52428800
compression_type = "snappy"
acks = "all"
retries = 3
batch_size = 16384
linger_ms = 5
buffer_memory = 33554432

[database]
url = "postgresql://test:test@localhost:5432/test"
max_connections = 5
min_connections = 1
connect_timeout = 30
acquire_timeout = 30
idle_timeout = 300
max_lifetime = 3600

[metrics]
enabled = true
bind_address = "0.0.0.0"
port = 9090
path = "/metrics"

[tracing]
enabled = false
jaeger_endpoint = "http://localhost:14268/api/traces"
service_name = "test-processor"
service_version = "1.0.0"

[processing]
batch_size = 100
batch_timeout_ms = 500
max_concurrent_tasks = 5
buffer_size = 1000
retry_attempts = 2
retry_delay_ms = 500
circuit_breaker_threshold = 3
circuit_breaker_timeout_ms = 30000
"#;

        let temp_file = NamedTempFile::new().unwrap();
        fs::write(&temp_file, config_content).unwrap();

        let config = Config::load(temp_file.path()).unwrap();

        assert_eq!(config.app.name, "test-processor");
        assert_eq!(config.kafka.bootstrap_servers, "localhost:9092");
        assert_eq!(config.database.url, "postgresql://test:test@localhost:5432/test");
        assert_eq!(config.processing.batch_size, 100);
    }

    #[test]
    fn test_config_default() {
        let config = Config::default();

        assert_eq!(config.app.name, "stream-processor");
        assert_eq!(config.kafka.bootstrap_servers, "localhost:9092");
        assert_eq!(config.metrics.enabled, true);
        assert_eq!(config.tracing.enabled, false);
    }
} 