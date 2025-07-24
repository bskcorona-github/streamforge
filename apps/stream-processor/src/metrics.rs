use anyhow::Result;
use prometheus::{
    Counter, Gauge, Histogram, HistogramOpts, IntCounter, IntGauge, Opts, Registry,
};
use std::sync::Arc;
use tokio::net::TcpListener;
use tracing::{error, info};

pub struct Metrics {
    registry: Registry,
    
    // Kafka metrics
    pub kafka_messages_received: IntCounter,
    pub kafka_messages_processed: IntCounter,
    pub kafka_messages_failed: IntCounter,
    pub kafka_consumer_lag: IntGauge,
    
    // Processing metrics
    pub processing_duration: Histogram,
    pub processing_batch_size: Histogram,
    pub processing_errors: IntCounter,
    pub processing_retries: IntCounter,
    
    // Database metrics
    pub database_operations: IntCounter,
    pub database_errors: IntCounter,
    pub database_connection_pool_size: IntGauge,
    pub database_connection_pool_available: IntGauge,
    
    // Stream processing metrics
    pub stream_watermark: IntGauge,
    pub stream_window_count: IntCounter,
    pub stream_late_records: IntCounter,
    
    // System metrics
    pub memory_usage_bytes: IntGauge,
    pub cpu_usage_percent: Gauge,
    pub active_tasks: IntGauge,
}

impl Metrics {
    pub fn new() -> Result<Self> {
        let registry = Registry::new();
        
        // Kafka metrics
        let kafka_messages_received = IntCounter::new(
            "kafka_messages_received_total",
            "Total number of Kafka messages received",
        )?;
        
        let kafka_messages_processed = IntCounter::new(
            "kafka_messages_processed_total",
            "Total number of Kafka messages processed successfully",
        )?;
        
        let kafka_messages_failed = IntCounter::new(
            "kafka_messages_failed_total",
            "Total number of Kafka messages that failed processing",
        )?;
        
        let kafka_consumer_lag = IntGauge::new(
            "kafka_consumer_lag",
            "Current consumer lag for each partition",
        )?;
        
        // Processing metrics
        let processing_duration = Histogram::with_opts(HistogramOpts::new(
            "processing_duration_seconds",
            "Time spent processing messages",
        ))?;
        
        let processing_batch_size = Histogram::with_opts(HistogramOpts::new(
            "processing_batch_size",
            "Size of processing batches",
        ))?;
        
        let processing_errors = IntCounter::new(
            "processing_errors_total",
            "Total number of processing errors",
        )?;
        
        let processing_retries = IntCounter::new(
            "processing_retries_total",
            "Total number of processing retries",
        )?;
        
        // Database metrics
        let database_operations = IntCounter::new(
            "database_operations_total",
            "Total number of database operations",
        )?;
        
        let database_errors = IntCounter::new(
            "database_errors_total",
            "Total number of database errors",
        )?;
        
        let database_connection_pool_size = IntGauge::new(
            "database_connection_pool_size",
            "Total number of database connections in pool",
        )?;
        
        let database_connection_pool_available = IntGauge::new(
            "database_connection_pool_available",
            "Number of available database connections in pool",
        )?;
        
        // Stream processing metrics
        let stream_watermark = IntGauge::new(
            "stream_watermark_timestamp",
            "Current watermark timestamp for stream processing",
        )?;
        
        let stream_window_count = IntCounter::new(
            "stream_window_count_total",
            "Total number of stream processing windows",
        )?;
        
        let stream_late_records = IntCounter::new(
            "stream_late_records_total",
            "Total number of late records in stream processing",
        )?;
        
        // System metrics
        let memory_usage_bytes = IntGauge::new(
            "memory_usage_bytes",
            "Current memory usage in bytes",
        )?;
        
        let cpu_usage_percent = Gauge::new(
            "cpu_usage_percent",
            "Current CPU usage percentage",
        )?;
        
        let active_tasks = IntGauge::new(
            "active_tasks",
            "Number of currently active processing tasks",
        )?;
        
        // Register all metrics
        registry.register(Box::new(kafka_messages_received.clone()))?;
        registry.register(Box::new(kafka_messages_processed.clone()))?;
        registry.register(Box::new(kafka_messages_failed.clone()))?;
        registry.register(Box::new(kafka_consumer_lag.clone()))?;
        registry.register(Box::new(processing_duration.clone()))?;
        registry.register(Box::new(processing_batch_size.clone()))?;
        registry.register(Box::new(processing_errors.clone()))?;
        registry.register(Box::new(processing_retries.clone()))?;
        registry.register(Box::new(database_operations.clone()))?;
        registry.register(Box::new(database_errors.clone()))?;
        registry.register(Box::new(database_connection_pool_size.clone()))?;
        registry.register(Box::new(database_connection_pool_available.clone()))?;
        registry.register(Box::new(stream_watermark.clone()))?;
        registry.register(Box::new(stream_window_count.clone()))?;
        registry.register(Box::new(stream_late_records.clone()))?;
        registry.register(Box::new(memory_usage_bytes.clone()))?;
        registry.register(Box::new(cpu_usage_percent.clone()))?;
        registry.register(Box::new(active_tasks.clone()))?;
        
        Ok(Self {
            registry,
            kafka_messages_received,
            kafka_messages_processed,
            kafka_messages_failed,
            kafka_consumer_lag,
            processing_duration,
            processing_batch_size,
            processing_errors,
            processing_retries,
            database_operations,
            database_errors,
            database_connection_pool_size,
            database_connection_pool_available,
            stream_watermark,
            stream_window_count,
            stream_late_records,
            memory_usage_bytes,
            cpu_usage_percent,
            active_tasks,
        })
    }
    
    pub fn registry(&self) -> &Registry {
        &self.registry
    }
    
    // Helper methods for common metric operations
    pub fn increment_messages_received(&self, count: u64) {
        self.kafka_messages_received.inc_by(count);
    }
    
    pub fn increment_messages_processed(&self, count: u64) {
        self.kafka_messages_processed.inc_by(count);
    }
    
    pub fn increment_messages_failed(&self, count: u64) {
        self.kafka_messages_failed.inc_by(count);
    }
    
    pub fn set_consumer_lag(&self, partition: i32, lag: i64) {
        self.kafka_consumer_lag.with_label_values(&[&partition.to_string()]).set(lag);
    }
    
    pub fn observe_processing_duration(&self, duration: f64) {
        self.processing_duration.observe(duration);
    }
    
    pub fn observe_batch_size(&self, size: f64) {
        self.processing_batch_size.observe(size);
    }
    
    pub fn increment_processing_errors(&self) {
        self.processing_errors.inc();
    }
    
    pub fn increment_processing_retries(&self) {
        self.processing_retries.inc();
    }
    
    pub fn increment_database_operations(&self) {
        self.database_operations.inc();
    }
    
    pub fn increment_database_errors(&self) {
        self.database_errors.inc();
    }
    
    pub fn set_connection_pool_size(&self, size: i64) {
        self.database_connection_pool_size.set(size);
    }
    
    pub fn set_connection_pool_available(&self, available: i64) {
        self.database_connection_pool_available.set(available);
    }
    
    pub fn set_watermark(&self, timestamp: i64) {
        self.stream_watermark.set(timestamp);
    }
    
    pub fn increment_window_count(&self) {
        self.stream_window_count.inc();
    }
    
    pub fn increment_late_records(&self) {
        self.stream_late_records.inc();
    }
    
    pub fn set_memory_usage(&self, bytes: i64) {
        self.memory_usage_bytes.set(bytes);
    }
    
    pub fn set_cpu_usage(&self, percent: f64) {
        self.cpu_usage_percent.set(percent);
    }
    
    pub fn set_active_tasks(&self, count: i64) {
        self.active_tasks.set(count);
    }
}

pub async fn start_server(addr: &str, metrics: Arc<Metrics>) -> Result<()> {
    let listener = TcpListener::bind(addr).await?;
    info!("Metrics server listening on {}", addr);
    
    loop {
        let (stream, _) = listener.accept().await?;
        let metrics = metrics.clone();
        
        tokio::spawn(async move {
            if let Err(e) = handle_metrics_request(stream, metrics).await {
                error!("Error handling metrics request: {}", e);
            }
        });
    }
}

async fn handle_metrics_request(
    stream: tokio::net::TcpStream,
    metrics: Arc<Metrics>,
) -> Result<()> {
    use tokio::io::{AsyncReadExt, AsyncWriteExt};
    
    let mut buffer = [0; 1024];
    let n = stream.readable().await?;
    let n = stream.try_read(&mut buffer)?;
    
    if n > 0 {
        let request = String::from_utf8_lossy(&buffer[..n]);
        
        if request.contains("GET /metrics") {
            let response = format!(
                "HTTP/1.1 200 OK\r\nContent-Type: text/plain; version=0.0.4\r\nContent-Length: {}\r\n\r\n{}",
                prometheus::TextEncoder::new().encode_to_string(metrics.registry())?.len(),
                prometheus::TextEncoder::new().encode_to_string(metrics.registry())?
            );
            
            stream.writable().await?;
            stream.try_write(response.as_bytes())?;
        } else {
            let response = "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n";
            stream.writable().await?;
            stream.try_write(response.as_bytes())?;
        }
    }
    
    Ok(())
} 