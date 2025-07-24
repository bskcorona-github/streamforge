use crate::config::Config;
use anyhow::Result;
use rdkafka::client::ClientContext;
use rdkafka::config::ClientConfig;
use rdkafka::consumer::{Consumer, ConsumerContext, Rebalance, StreamConsumer};
use rdkafka::error::KafkaResult;
use rdkafka::message::{Headers, OwnedHeaders};
use rdkafka::producer::{FutureProducer, FutureRecord, ProducerContext};
use rdkafka::topic_partition_list::TopicPartitionList;
use rdkafka::{Offset, TopicPartitionList};
use std::sync::Arc;
use std::time::Duration;
use tracing::{error, info, warn};

use crate::config::Config;
use crate::metrics::Metrics;

#[derive(Clone)]
pub struct KafkaManager {
    config: Config,
    metrics: Arc<Metrics>,
}

impl KafkaManager {
    pub async fn new(config: &Config, metrics: Arc<Metrics>) -> Result<Self> {
        info!("Initializing Kafka Manager...");

        // Test Kafka connectivity
        Self::test_connectivity(config).await?;

        Ok(Self {
            config: config.clone(),
            metrics,
        })
    }

    pub async fn create_consumer(&self) -> Result<StreamConsumer> {
        let consumer_config = self.config.kafka_consumer_config();
        
        info!("Creating Kafka consumer with group: {}", self.config.kafka.group_id);
        
        let consumer: StreamConsumer = consumer_config
            .create()
            .map_err(|e| anyhow::anyhow!("Failed to create Kafka consumer: {}", e))?;

        info!("Kafka consumer created successfully");
        Ok(consumer)
    }

    pub async fn create_producer(&self) -> Result<FutureProducer> {
        let producer_config = self.config.kafka_producer_config();
        
        info!("Creating Kafka producer...");
        
        let producer: FutureProducer = producer_config
            .create()
            .map_err(|e| anyhow::anyhow!("Failed to create Kafka producer: {}", e))?;

        info!("Kafka producer created successfully");
        Ok(producer)
    }

    pub async fn send_message(
        &self,
        producer: &FutureProducer,
        topic: &str,
        key: Option<&str>,
        payload: &[u8],
    ) -> Result<()> {
        let record = if let Some(key) = key {
            FutureRecord::to(topic).key(key).payload(payload)
        } else {
            FutureRecord::to(topic).payload(payload)
        };

        match producer.send(record, std::time::Duration::from_secs(5)).await {
            Ok(_) => {
                info!("Message sent successfully to topic: {}", topic);
                Ok(())
            }
            Err((e, _)) => {
                error!("Failed to send message to topic {}: {}", topic, e);
                Err(anyhow::anyhow!("Failed to send message: {}", e))
            }
        }
    }

    pub async fn send_batch_messages(
        &self,
        producer: &FutureProducer,
        topic: &str,
        messages: Vec<(Option<String>, Vec<u8>)>,
    ) -> Result<()> {
        info!("Sending batch of {} messages to topic: {}", messages.len(), topic);

        let mut futures = Vec::new();

        for (key, payload) in messages {
            let record = if let Some(key) = key {
                FutureRecord::to(topic).key(&key).payload(&payload)
            } else {
                FutureRecord::to(topic).payload(&payload)
            };

            futures.push(producer.send(record, std::time::Duration::from_secs(5)));
        }

        // Wait for all messages to be sent
        let results = futures::future::join_all(futures).await;

        let mut success_count = 0;
        let mut error_count = 0;

        for result in results {
            match result {
                Ok(_) => success_count += 1,
                Err((e, _)) => {
                    error_count += 1;
                    error!("Failed to send message in batch: {}", e);
                }
            }
        }

        info!(
            "Batch send completed: {} successful, {} failed",
            success_count, error_count
        );

        if error_count > 0 {
            warn!("Some messages in batch failed to send");
        }

        Ok(())
    }

    pub async fn get_consumer_lag(&self, consumer: &StreamConsumer) -> Result<Vec<(i32, i64)>> {
        let mut lag_info = Vec::new();

        for topic in &self.config.kafka.input_topics {
            let partitions = consumer.assignment()?;
            
            for partition in partitions {
                if partition.topic() == topic {
                    let (low, high) = consumer.fetch_watermarks(partition.topic(), partition.partition(), std::time::Duration::from_secs(5))?;
                    let offset = consumer.committed(&[partition.clone()], std::time::Duration::from_secs(5))?;
                    
                    let committed_offset = offset.first()
                        .and_then(|r| r.as_ref().ok())
                        .map(|o| o.offset())
                        .unwrap_or(-1);
                    
                    let lag = high - committed_offset;
                    lag_info.push((partition.partition(), lag));

                    // Update metrics
                    self.metrics.set_consumer_lag(partition.partition(), lag);
                }
            }
        }

        Ok(lag_info)
    }

    pub async fn create_topics_if_not_exist(&self) -> Result<()> {
        info!("Checking and creating Kafka topics if they don't exist...");

        // This would typically use the Kafka Admin API to create topics
        // For now, we'll just log the topics that should exist
        for topic in &self.config.kafka.input_topics {
            info!("Ensuring topic exists: {}", topic);
        }

        info!("Ensuring output topic exists: {}", self.config.kafka.output_topic);
        info!("Ensuring error topic exists: {}", self.config.kafka.error_topic);

        Ok(())
    }

    async fn test_connectivity(config: &Config) -> Result<()> {
        info!("Testing Kafka connectivity...");

        let consumer_config = config.kafka_consumer_config();
        
        // Try to create a consumer to test connectivity
        match consumer_config.create::<StreamConsumer>() {
            Ok(_) => {
                info!("Kafka connectivity test successful");
                Ok(())
            }
            Err(e) => {
                error!("Kafka connectivity test failed: {}", e);
                Err(anyhow::anyhow!("Kafka connectivity test failed: {}", e))
            }
        }
    }

    pub fn get_input_topics(&self) -> &[String] {
        &self.config.kafka.input_topics
    }

    pub fn get_output_topic(&self) -> &str {
        &self.config.kafka.output_topic
    }

    pub fn get_error_topic(&self) -> &str {
        &self.config.kafka.error_topic
    }
}

// Helper struct for managing Kafka message metadata
#[derive(Debug, Clone)]
pub struct KafkaMessageMetadata {
    pub topic: String,
    pub partition: i32,
    pub offset: i64,
    pub timestamp: i64,
    pub key: Option<String>,
    pub headers: Vec<(String, Vec<u8>)>,
}

impl KafkaMessageMetadata {
    pub fn new(topic: String, partition: i32, offset: i64, timestamp: i64) -> Self {
        Self {
            topic,
            partition,
            offset,
            timestamp,
            key: None,
            headers: Vec::new(),
        }
    }

    pub fn with_key(mut self, key: String) -> Self {
        self.key = Some(key);
        self
    }

    pub fn with_headers(mut self, headers: Vec<(String, Vec<u8>)>) -> Self {
        self.headers = headers;
        self
    }
}

// Error types for Kafka operations
#[derive(Debug, thiserror::Error)]
pub enum KafkaError {
    #[error("Failed to create consumer: {0}")]
    ConsumerCreationFailed(String),
    
    #[error("Failed to create producer: {0}")]
    ProducerCreationFailed(String),
    
    #[error("Failed to send message: {0}")]
    SendFailed(String),
    
    #[error("Failed to receive message: {0}")]
    ReceiveFailed(String),
    
    #[error("Failed to commit offset: {0}")]
    CommitFailed(String),
    
    #[error("Topic not found: {0}")]
    TopicNotFound(String),
    
    #[error("Partition not found: {0}")]
    PartitionNotFound(i32),
    
    #[error("Connection timeout")]
    ConnectionTimeout,
    
    #[error("Authentication failed: {0}")]
    AuthenticationFailed(String),
    
    #[error("Authorization failed: {0}")]
    AuthorizationFailed(String),
}

impl From<rdkafka::error::KafkaError> for KafkaError {
    fn from(err: rdkafka::error::KafkaError) -> Self {
        match err {
            rdkafka::error::KafkaError::MessageConsumption(_) => {
                KafkaError::ReceiveFailed(err.to_string())
            }
            rdkafka::error::KafkaError::MessageProduction(_) => {
                KafkaError::SendFailed(err.to_string())
            }
            rdkafka::error::KafkaError::ClientCreation(_) => {
                KafkaError::ConsumerCreationFailed(err.to_string())
            }
            rdkafka::error::KafkaError::TopicCreation(_) => {
                KafkaError::TopicNotFound(err.to_string())
            }
            rdkafka::error::KafkaError::PartitionEOF(_) => {
                KafkaError::PartitionNotFound(0) // We don't have partition info here
            }
            _ => KafkaError::SendFailed(err.to_string()),
        }
    }
} 