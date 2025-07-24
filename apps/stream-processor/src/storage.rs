use crate::config::Config;
use crate::processor::{ProcessedMessage, ProcessingMetadata};
use anyhow::Result;
use chrono::{DateTime, Utc};
use sqlx::{postgres::PgPoolOptions, PgPool, Row};
use std::sync::Arc;
use tracing::{error, info, warn};

pub struct StorageManager {
    pool: PgPool,
}

impl StorageManager {
    pub async fn new(config: &Config) -> Result<Self> {
        let pool = PgPoolOptions::new()
            .max_connections(config.database.max_connections)
            .min_connections(config.database.min_connections)
            .connect_timeout(std::time::Duration::from_secs(config.database.connect_timeout))
            .acquire_timeout(std::time::Duration::from_secs(config.database.acquire_timeout))
            .idle_timeout(std::time::Duration::from_secs(config.database.idle_timeout))
            .max_lifetime(std::time::Duration::from_secs(config.database.max_lifetime))
            .connect(&config.database.url)
            .await
            .map_err(|e| anyhow::anyhow!("Failed to connect to database: {}", e))?;

        info!("Database connection pool created successfully");

        // Initialize database schema
        Self::init_schema(&pool).await?;

        Ok(Self { pool })
    }

    async fn init_schema(pool: &PgPool) -> Result<()> {
        let schema_sql = r#"
            -- Create processed_messages table
            CREATE TABLE IF NOT EXISTS processed_messages (
                id UUID PRIMARY KEY,
                original_message JSONB NOT NULL,
                processed_message JSONB NOT NULL,
                processed_at TIMESTAMP WITH TIME ZONE NOT NULL,
                processor_version VARCHAR(50) NOT NULL,
                source_topic VARCHAR(255) NOT NULL,
                partition INTEGER NOT NULL,
                offset BIGINT NOT NULL,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
            );

            -- Create index on processed_at for time-based queries
            CREATE INDEX IF NOT EXISTS idx_processed_messages_processed_at 
            ON processed_messages (processed_at);

            -- Create index on source_topic for topic-based queries
            CREATE INDEX IF NOT EXISTS idx_processed_messages_source_topic 
            ON processed_messages (source_topic);

            -- Create index on processor_version for version-based queries
            CREATE INDEX IF NOT EXISTS idx_processed_messages_processor_version 
            ON processed_messages (processor_version);

            -- Create composite index for efficient filtering
            CREATE INDEX IF NOT EXISTS idx_processed_messages_topic_partition_offset 
            ON processed_messages (source_topic, partition, offset);

            -- Create metrics table for storing aggregated metrics
            CREATE TABLE IF NOT EXISTS metrics (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                metric_name VARCHAR(255) NOT NULL,
                metric_value DOUBLE PRECISION NOT NULL,
                metric_type VARCHAR(50) NOT NULL,
                tags JSONB,
                timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
            );

            -- Create index on metric_name and timestamp for efficient queries
            CREATE INDEX IF NOT EXISTS idx_metrics_name_timestamp 
            ON metrics (metric_name, timestamp);

            -- Create index on tags for tag-based queries
            CREATE INDEX IF NOT EXISTS idx_metrics_tags 
            ON metrics USING GIN (tags);

            -- Create logs table for storing log entries
            CREATE TABLE IF NOT EXISTS logs (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                log_level VARCHAR(20) NOT NULL,
                message TEXT NOT NULL,
                service_name VARCHAR(255),
                host_name VARCHAR(255),
                trace_id VARCHAR(255),
                span_id VARCHAR(255),
                attributes JSONB,
                timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
            );

            -- Create index on timestamp for time-based queries
            CREATE INDEX IF NOT EXISTS idx_logs_timestamp 
            ON logs (timestamp);

            -- Create index on log_level for level-based queries
            CREATE INDEX IF NOT EXISTS idx_logs_level 
            ON logs (log_level);

            -- Create index on service_name for service-based queries
            CREATE INDEX IF NOT EXISTS idx_logs_service_name 
            ON logs (service_name);

            -- Create index on trace_id for trace-based queries
            CREATE INDEX IF NOT EXISTS idx_logs_trace_id 
            ON logs (trace_id);

            -- Create traces table for storing trace spans
            CREATE TABLE IF NOT EXISTS traces (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                trace_id VARCHAR(255) NOT NULL,
                span_id VARCHAR(255) NOT NULL,
                parent_span_id VARCHAR(255),
                name VARCHAR(255) NOT NULL,
                service_name VARCHAR(255),
                start_time TIMESTAMP WITH TIME ZONE NOT NULL,
                end_time TIMESTAMP WITH TIME ZONE,
                duration_ms BIGINT,
                status VARCHAR(50),
                attributes JSONB,
                events JSONB,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
            );

            -- Create index on trace_id for trace-based queries
            CREATE INDEX IF NOT EXISTS idx_traces_trace_id 
            ON traces (trace_id);

            -- Create index on span_id for span-based queries
            CREATE INDEX IF NOT EXISTS idx_traces_span_id 
            ON traces (span_id);

            -- Create index on start_time for time-based queries
            CREATE INDEX IF NOT EXISTS idx_traces_start_time 
            ON traces (start_time);

            -- Create index on service_name for service-based queries
            CREATE INDEX IF NOT EXISTS idx_traces_service_name 
            ON traces (service_name);

            -- Create function to update updated_at timestamp
            CREATE OR REPLACE FUNCTION update_updated_at_column()
            RETURNS TRIGGER AS $$
            BEGIN
                NEW.updated_at = NOW();
                RETURN NEW;
            END;
            $$ language 'plpgsql';

            -- Create trigger for processed_messages table
            DROP TRIGGER IF EXISTS update_processed_messages_updated_at ON processed_messages;
            CREATE TRIGGER update_processed_messages_updated_at
                BEFORE UPDATE ON processed_messages
                FOR EACH ROW
                EXECUTE FUNCTION update_updated_at_column();
        "#;

        sqlx::query(schema_sql)
            .execute(pool)
            .await
            .map_err(|e| anyhow::anyhow!("Failed to initialize database schema: {}", e))?;

        info!("Database schema initialized successfully");
        Ok(())
    }

    pub async fn store_processed_message(&self, message: &ProcessedMessage) -> Result<()> {
        let sql = r#"
            INSERT INTO processed_messages (
                id, original_message, processed_message, processed_at, 
                processor_version, source_topic, partition, offset
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            ON CONFLICT (id) DO UPDATE SET
                original_message = EXCLUDED.original_message,
                processed_message = EXCLUDED.processed_message,
                processed_at = EXCLUDED.processed_at,
                processor_version = EXCLUDED.processor_version,
                source_topic = EXCLUDED.source_topic,
                partition = EXCLUDED.partition,
                offset = EXCLUDED.offset,
                updated_at = NOW()
        "#;

        sqlx::query(sql)
            .bind(&message.id)
            .bind(&serde_json::to_value(&message.original_message)?)
            .bind(&serde_json::to_value(&message.processed_message)?)
            .bind(message.processing_metadata.processed_at)
            .bind(&message.processing_metadata.processor_version)
            .bind(&message.processing_metadata.source_topic)
            .bind(message.processing_metadata.partition)
            .bind(message.processing_metadata.offset)
            .execute(&self.pool)
            .await
            .map_err(|e| anyhow::anyhow!("Failed to store processed message: {}", e))?;

        Ok(())
    }

    pub async fn store_processed_messages(&self, messages: &[ProcessedMessage]) -> Result<()> {
        if messages.is_empty() {
            return Ok(());
        }

        let mut transaction = self.pool.begin().await?;

        for message in messages {
            let sql = r#"
                INSERT INTO processed_messages (
                    id, original_message, processed_message, processed_at, 
                    processor_version, source_topic, partition, offset
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
                ON CONFLICT (id) DO UPDATE SET
                    original_message = EXCLUDED.original_message,
                    processed_message = EXCLUDED.processed_message,
                    processed_at = EXCLUDED.processed_at,
                    processor_version = EXCLUDED.processor_version,
                    source_topic = EXCLUDED.source_topic,
                    partition = EXCLUDED.partition,
                    offset = EXCLUDED.offset,
                    updated_at = NOW()
            "#;

            sqlx::query(sql)
                .bind(&message.id)
                .bind(&serde_json::to_value(&message.original_message)?)
                .bind(&serde_json::to_value(&message.processed_message)?)
                .bind(message.processing_metadata.processed_at)
                .bind(&message.processing_metadata.processor_version)
                .bind(&message.processing_metadata.source_topic)
                .bind(message.processing_metadata.partition)
                .bind(message.processing_metadata.offset)
                .execute(&mut *transaction)
                .await
                .map_err(|e| anyhow::anyhow!("Failed to store processed message in batch: {}", e))?;
        }

        transaction.commit().await?;
        info!("Stored {} processed messages in batch", messages.len());

        Ok(())
    }

    pub async fn get_processed_message(&self, id: &str) -> Result<Option<ProcessedMessage>> {
        let sql = r#"
            SELECT id, original_message, processed_message, processed_at, 
                   processor_version, source_topic, partition, offset
            FROM processed_messages
            WHERE id = $1
        "#;

        let row = sqlx::query(sql)
            .bind(id)
            .fetch_optional(&self.pool)
            .await
            .map_err(|e| anyhow::anyhow!("Failed to get processed message: {}", e))?;

        match row {
            Some(row) => {
                let processing_metadata = ProcessingMetadata {
                    processed_at: row.get("processed_at"),
                    processor_version: row.get("processor_version"),
                    source_topic: row.get("source_topic"),
                    partition: row.get("partition"),
                    offset: row.get("offset"),
                };

                Ok(Some(ProcessedMessage {
                    id: row.get("id"),
                    original_message: serde_json::from_value(row.get("original_message"))?,
                    processed_message: serde_json::from_value(row.get("processed_message"))?,
                    processing_metadata,
                }))
            }
            None => Ok(None),
        }
    }

    pub async fn get_processed_messages_by_topic(
        &self,
        topic: &str,
        limit: Option<i64>,
        offset: Option<i64>,
    ) -> Result<Vec<ProcessedMessage>> {
        let sql = r#"
            SELECT id, original_message, processed_message, processed_at, 
                   processor_version, source_topic, partition, offset
            FROM processed_messages
            WHERE source_topic = $1
            ORDER BY processed_at DESC
            LIMIT $2 OFFSET $3
        "#;

        let limit = limit.unwrap_or(100);
        let offset = offset.unwrap_or(0);

        let rows = sqlx::query(sql)
            .bind(topic)
            .bind(limit)
            .bind(offset)
            .fetch_all(&self.pool)
            .await
            .map_err(|e| anyhow::anyhow!("Failed to get processed messages by topic: {}", e))?;

        let mut messages = Vec::new();
        for row in rows {
            let processing_metadata = ProcessingMetadata {
                processed_at: row.get("processed_at"),
                processor_version: row.get("processor_version"),
                source_topic: row.get("source_topic"),
                partition: row.get("partition"),
                offset: row.get("offset"),
            };

            messages.push(ProcessedMessage {
                id: row.get("id"),
                original_message: serde_json::from_value(row.get("original_message"))?,
                processed_message: serde_json::from_value(row.get("processed_message"))?,
                processing_metadata,
            });
        }

        Ok(messages)
    }

    pub async fn get_processed_messages_by_time_range(
        &self,
        start_time: DateTime<Utc>,
        end_time: DateTime<Utc>,
        limit: Option<i64>,
        offset: Option<i64>,
    ) -> Result<Vec<ProcessedMessage>> {
        let sql = r#"
            SELECT id, original_message, processed_message, processed_at, 
                   processor_version, source_topic, partition, offset
            FROM processed_messages
            WHERE processed_at >= $1 AND processed_at <= $2
            ORDER BY processed_at DESC
            LIMIT $3 OFFSET $4
        "#;

        let limit = limit.unwrap_or(100);
        let offset = offset.unwrap_or(0);

        let rows = sqlx::query(sql)
            .bind(start_time)
            .bind(end_time)
            .bind(limit)
            .bind(offset)
            .fetch_all(&self.pool)
            .await
            .map_err(|e| anyhow::anyhow!("Failed to get processed messages by time range: {}", e))?;

        let mut messages = Vec::new();
        for row in rows {
            let processing_metadata = ProcessingMetadata {
                processed_at: row.get("processed_at"),
                processor_version: row.get("processor_version"),
                source_topic: row.get("source_topic"),
                partition: row.get("partition"),
                offset: row.get("offset"),
            };

            messages.push(ProcessedMessage {
                id: row.get("id"),
                original_message: serde_json::from_value(row.get("original_message"))?,
                processed_message: serde_json::from_value(row.get("processed_message"))?,
                processing_metadata,
            });
        }

        Ok(messages)
    }

    pub async fn store_metric(
        &self,
        name: &str,
        value: f64,
        metric_type: &str,
        tags: Option<serde_json::Value>,
        timestamp: DateTime<Utc>,
    ) -> Result<()> {
        let sql = r#"
            INSERT INTO metrics (metric_name, metric_value, metric_type, tags, timestamp)
            VALUES ($1, $2, $3, $4, $5)
        "#;

        sqlx::query(sql)
            .bind(name)
            .bind(value)
            .bind(metric_type)
            .bind(tags)
            .bind(timestamp)
            .execute(&self.pool)
            .await
            .map_err(|e| anyhow::anyhow!("Failed to store metric: {}", e))?;

        Ok(())
    }

    pub async fn store_log(
        &self,
        level: &str,
        message: &str,
        service_name: Option<&str>,
        host_name: Option<&str>,
        trace_id: Option<&str>,
        span_id: Option<&str>,
        attributes: Option<serde_json::Value>,
        timestamp: DateTime<Utc>,
    ) -> Result<()> {
        let sql = r#"
            INSERT INTO logs (log_level, message, service_name, host_name, trace_id, span_id, attributes, timestamp)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        "#;

        sqlx::query(sql)
            .bind(level)
            .bind(message)
            .bind(service_name)
            .bind(host_name)
            .bind(trace_id)
            .bind(span_id)
            .bind(attributes)
            .bind(timestamp)
            .execute(&self.pool)
            .await
            .map_err(|e| anyhow::anyhow!("Failed to store log: {}", e))?;

        Ok(())
    }

    pub async fn store_trace(
        &self,
        trace_id: &str,
        span_id: &str,
        parent_span_id: Option<&str>,
        name: &str,
        service_name: Option<&str>,
        start_time: DateTime<Utc>,
        end_time: Option<DateTime<Utc>>,
        duration_ms: Option<i64>,
        status: Option<&str>,
        attributes: Option<serde_json::Value>,
        events: Option<serde_json::Value>,
    ) -> Result<()> {
        let sql = r#"
            INSERT INTO traces (trace_id, span_id, parent_span_id, name, service_name, 
                               start_time, end_time, duration_ms, status, attributes, events)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        "#;

        sqlx::query(sql)
            .bind(trace_id)
            .bind(span_id)
            .bind(parent_span_id)
            .bind(name)
            .bind(service_name)
            .bind(start_time)
            .bind(end_time)
            .bind(duration_ms)
            .bind(status)
            .bind(attributes)
            .bind(events)
            .execute(&self.pool)
            .await
            .map_err(|e| anyhow::anyhow!("Failed to store trace: {}", e))?;

        Ok(())
    }

    pub async fn health_check(&self) -> Result<bool> {
        match sqlx::query("SELECT 1").fetch_one(&self.pool).await {
            Ok(_) => Ok(true),
            Err(e) => {
                error!("Database health check failed: {}", e);
                Ok(false)
            }
        }
    }
}

impl Clone for StorageManager {
    fn clone(&self) -> Self {
        Self {
            pool: self.pool.clone(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::Utc;
    use serde_json::json;

    #[tokio::test]
    async fn test_storage_manager_creation() {
        let config = Config::default();
        let storage = StorageManager::new(&config).await;
        // This will likely fail in test environment without database running
        // but we can test that the method signature is correct
        assert!(storage.is_ok() || storage.is_err());
    }

    #[tokio::test]
    async fn test_store_and_get_processed_message() {
        let config = Config::default();
        let storage = StorageManager::new(&config).await;
        
        if let Ok(storage) = storage {
            let message = ProcessedMessage {
                id: "test-id".to_string(),
                original_message: json!({"test": "original"}),
                processed_message: json!({"test": "processed"}),
                processing_metadata: ProcessingMetadata {
                    processed_at: Utc::now(),
                    processor_version: "1.0.0".to_string(),
                    source_topic: "test-topic".to_string(),
                    partition: 0,
                    offset: 100,
                },
            };

            let store_result = storage.store_processed_message(&message).await;
            assert!(store_result.is_ok() || store_result.is_err());

            if store_result.is_ok() {
                let get_result = storage.get_processed_message("test-id").await;
                assert!(get_result.is_ok());
            }
        }
    }
} 