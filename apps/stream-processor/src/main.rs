use anyhow::Result;
use clap::Parser;
use std::sync::Arc;
use tokio::signal;
use tracing::{error, info, warn};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use stream_processor::config::Config;
use stream_processor::processor::StreamProcessor;
use stream_processor::telemetry;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Configuration file path
    #[arg(short, long, default_value = "config/config.toml")]
    config: String,
    
    /// Log level
    #[arg(short, long, default_value = "info")]
    log_level: String,
}

#[tokio::main]
async fn main() -> Result<()> {
    // Parse command line arguments
    let args = Args::parse();
    
    // Initialize logging and tracing
    telemetry::init(&args.log_level)?;
    
    info!("Starting StreamForge Stream Processor");
    info!("Version: {}", env!("CARGO_PKG_VERSION"));
    info!("Configuration file: {}", args.config);
    
    // Load configuration
    let config = Config::load(&args.config)?;
    info!("Configuration loaded successfully");
    
    // Create stream processor
    let processor = Arc::new(StreamProcessor::new(config).await?);
    info!("Stream processor initialized");
    
    // Start the processor
    let processor_handle = {
        let processor = Arc::clone(&processor);
        tokio::spawn(async move {
            if let Err(e) = processor.start().await {
                error!("Stream processor error: {}", e);
            }
        })
    };
    
    // Wait for shutdown signal
    match signal::ctrl_c().await {
        Ok(()) => {
            info!("Received shutdown signal");
        }
        Err(err) => {
            error!("Unable to listen for shutdown signal: {}", err);
        }
    }
    
    // Graceful shutdown
    info!("Initiating graceful shutdown...");
    
    // Stop the processor
    processor.stop().await;
    info!("Stream processor stopped");
    
    // Wait for processor to finish
    if let Err(e) = processor_handle.await {
        warn!("Processor task error during shutdown: {}", e);
    }
    
    info!("StreamForge Stream Processor shutdown complete");
    Ok(())
} 