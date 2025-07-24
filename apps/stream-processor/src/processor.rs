use anyhow::Result;
use futures::StreamExt;
use rdkafka::consumer::{Consumer, StreamConsumer};
use rdkafka::producer::{FutureProducer, FutureRecord};
use rdkafka::Message;
use sqlx::PgPool;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::mpsc;
use tokio::time::timeout;
use tracing::{error, info, warn};

use crate::config::Config;
use crate::kafka::KafkaManager;
use crate::metrics::Metrics;
use crate::processing::MessageProcessor;
use crate::storage::DatabaseManager;

pub struct StreamProcessor {
    config: Config,
    metrics: Arc<Metrics>,
    kafka_manager: KafkaManager,
    database_manager: DatabaseManager,
    message_processor: MessageProcessor,
}

impl StreamProcessor {
    pub async fn new(config: Config, metrics: Arc<Metrics>) -> Result<Self> {
        info!("Initializing Stream Processor...");

        // Initialize Kafka manager
        let kafka_manager = KafkaManager::new(&config, metrics.clone()).await?;
        info!("Kafka manager initialized");

        // Initialize database manager
        let database_manager = DatabaseManager::new(&config, metrics.clone()).await?;
        info!("Database manager initialized");

        // Initialize message processor
        let message_processor = MessageProcessor::new(&config, metrics.clone());
        info!("Message processor initialized");

        Ok(Self {
            config,
            metrics,
            kafka_manager,
            database_manager,
            message_processor,
        })
    }

    pub async fn run(&self) -> Result<()> {
        info!("Starting stream processor...");

        // Create channels for communication between components
        let (tx, rx) = mpsc::channel(1000);

        // Start Kafka consumer
        let consumer_handle = self.start_kafka_consumer(tx.clone()).await?;

        // Start message processing workers
        let worker_handles = self.start_processing_workers(rx).await?;

        // Start database writer
        let db_writer_handle = self.start_database_writer().await?;

        // Start metrics collection
        let metrics_handle = self.start_metrics_collection().await?;

        // Wait for all components to complete
        tokio::select! {
            _ = consumer_handle => {
                info!("Kafka consumer stopped");
            }
            _ = futures::future::join_all(worker_handles) => {
                info!("All processing workers stopped");
            }
            _ = db_writer_handle => {
                info!("Database writer stopped");
            }
            _ = metrics_handle => {
                info!("Metrics collection stopped");
            }
        }

        info!("Stream processor stopped");
        Ok(())
    }

    async fn start_kafka_consumer(&self, tx: mpsc::Sender<KafkaMessage>) -> Result<tokio::task::JoinHandle<()>> {
        let config = self.config.clone();
        let metrics = self.metrics.clone();
        let kafka_manager = self.kafka_manager.clone();

        let handle = tokio::spawn(async move {
            if let Err(e) = Self::run_kafka_consumer(config, metrics, kafka_manager, tx).await {
                error!("Kafka consumer error: {}", e);
            }
        });

        Ok(handle)
    }

    async fn run_kafka_consumer(
        config: Config,
        metrics: Arc<Metrics>,
        kafka_manager: KafkaManager,
        tx: mpsc::Sender<KafkaMessage>,
    ) -> Result<()> {
        let consumer: StreamConsumer = kafka_manager.create_consumer().await?;
        
        // Subscribe to topics
        consumer.subscribe(&config.kafka.input_topics)?;
        info!("Subscribed to topics: {:?}", config.kafka.input_topics);

        let mut message_stream = consumer.stream();

        while let Some(message_result) = message_stream.next().await {
            match message_result {
                Ok(message) => {
                    let topic = message.topic().to_string();
                    let partition = message.partition();
                    let offset = message.offset();
                    
                    info!("Received message from topic: {}, partition: {}, offset: {}", 
                          topic, partition, offset);

                    // Update metrics
                    metrics.increment_messages_received(1);

                    // Create Kafka message
                    let kafka_message = KafkaMessage {
                        topic,
                        partition,
                        offset,
                        payload: message.payload().unwrap_or_default().to_vec(),
                        timestamp: message.timestamp().to_millis(),
                    };

                    // Send to processing channel
                    if let Err(e) = tx.send(kafka_message).await {
                        error!("Failed to send message to processing channel: {}", e);
                        metrics.increment_messages_failed(1);
                    }
                }
                Err(e) => {
                    error!("Error receiving Kafka message: {}", e);
                    metrics.increment_messages_failed(1);
                }
            }
        }

        Ok(())
    }

    async fn start_processing_workers(&self, rx: mpsc::Receiver<KafkaMessage>) -> Result<Vec<tokio::task::JoinHandle<()>>> {
        let mut handles = Vec::new();
        let worker_count = self.config.processing.max_concurrent_tasks;

        for worker_id in 0..worker_count {
            let rx = rx.clone();
            let message_processor = self.message_processor.clone();
            let metrics = self.metrics.clone();
            let config = self.config.clone();

            let handle = tokio::spawn(async move {
                if let Err(e) = Self::run_processing_worker(
                    worker_id,
                    rx,
                    message_processor,
                    metrics,
                    config,
                ).await {
                    error!("Processing worker {} error: {}", worker_id, e);
                }
            });

            handles.push(handle);
        }

        info!("Started {} processing workers", worker_count);
        Ok(handles)
    }

    async fn run_processing_worker(
        worker_id: usize,
        mut rx: mpsc::Receiver<KafkaMessage>,
        message_processor: MessageProcessor,
        metrics: Arc<Metrics>,
        config: Config,
    ) -> Result<()> {
        info!("Processing worker {} started", worker_id);

        let mut batch = Vec::new();
        let batch_timeout = config.batch_timeout();

        while let Some(message) = rx.recv().await {
            batch.push(message);

            // Process batch if it's full or timeout reached
            if batch.len() >= config.processing.batch_size {
                if let Err(e) = Self::process_batch(&batch, &message_processor, &metrics).await {
                    error!("Worker {} failed to process batch: {}", worker_id, e);
                }
                batch.clear();
            } else {
                // Wait for more messages or timeout
                match timeout(batch_timeout, rx.recv()).await {
                    Ok(Some(message)) => {
                        batch.push(message);
                    }
                    Ok(None) => break, // Channel closed
                    Err(_) => {
                        // Timeout reached, process current batch
                        if !batch.is_empty() {
                            if let Err(e) = Self::process_batch(&batch, &message_processor, &metrics).await {
                                error!("Worker {} failed to process batch: {}", worker_id, e);
                            }
                            batch.clear();
                        }
                    }
                }
            }
        }

        // Process remaining messages
        if !batch.is_empty() {
            if let Err(e) = Self::process_batch(&batch, &message_processor, &metrics).await {
                error!("Worker {} failed to process final batch: {}", worker_id, e);
            }
        }

        info!("Processing worker {} stopped", worker_id);
        Ok(())
    }

    async fn process_batch(
        batch: &[KafkaMessage],
        message_processor: &MessageProcessor,
        metrics: &Arc<Metrics>,
    ) -> Result<()> {
        let start_time = Instant::now();
        
        info!("Processing batch of {} messages", batch.len());
        metrics.observe_batch_size(batch.len() as f64);

        for message in batch {
            match message_processor.process_message(message).await {
                Ok(_) => {
                    metrics.increment_messages_processed(1);
                }
                Err(e) => {
                    error!("Failed to process message: {}", e);
                    metrics.increment_messages_failed(1);
                    metrics.increment_processing_errors();
                }
            }
        }

        let duration = start_time.elapsed();
        metrics.observe_processing_duration(duration.as_secs_f64());
        
        info!("Batch processed in {:?}", duration);
        Ok(())
    }

    async fn start_database_writer(&self) -> Result<tokio::task::JoinHandle<()>> {
        let database_manager = self.database_manager.clone();
        let metrics = self.metrics.clone();

        let handle = tokio::spawn(async move {
            if let Err(e) = Self::run_database_writer(database_manager, metrics).await {
                error!("Database writer error: {}", e);
            }
        });

        Ok(handle)
    }

    async fn run_database_writer(
        database_manager: DatabaseManager,
        metrics: Arc<Metrics>,
    ) -> Result<()> {
        info!("Database writer started");

        // This would typically handle writing processed data to the database
        // For now, we'll just keep it running
        loop {
            tokio::time::sleep(Duration::from_secs(1)).await;
            
            // Update connection pool metrics
            let pool_size = database_manager.pool_size().await;
            let pool_available = database_manager.pool_available().await;
            
            metrics.set_connection_pool_size(pool_size);
            metrics.set_connection_pool_available(pool_available);
        }
    }

    async fn start_metrics_collection(&self) -> Result<tokio::task::JoinHandle<()>> {
        let metrics = self.metrics.clone();

        let handle = tokio::spawn(async move {
            if let Err(e) = Self::run_metrics_collection(metrics).await {
                error!("Metrics collection error: {}", e);
            }
        });

        Ok(handle)
    }

    async fn run_metrics_collection(metrics: Arc<Metrics>) -> Result<()> {
        info!("Metrics collection started");

        loop {
            tokio::time::sleep(Duration::from_secs(30)).await;

            // Update system metrics
            // In a real implementation, you would collect actual system metrics
            metrics.set_memory_usage(1024 * 1024 * 100); // 100MB example
            metrics.set_cpu_usage(25.0); // 25% example
            metrics.set_active_tasks(5); // 5 active tasks example
        }
    }
}

#[derive(Debug, Clone)]
pub struct KafkaMessage {
    pub topic: String,
    pub partition: i32,
    pub offset: i64,
    pub payload: Vec<u8>,
    pub timestamp: i64,
}

// Re-export modules for easier access
pub mod kafka;
pub mod processing;
pub mod storage; 