use std::sync::Arc;
use tokio::sync::Mutex;
use tonic::{transport::Server, Request, Response, Status};
use tracing::{info, warn, error};
use tracing_subscriber;

mod config;
mod processor;
mod storage;
mod metrics;
mod error;

use config::Config;
use processor::StreamProcessor;
use storage::StorageBackend;
use streamforge_v1::stream_processor_service_server::{StreamProcessorService, StreamProcessorServiceServer};
use streamforge_v1::{StartProcessingRequest, StartProcessingResponse, StopProcessingRequest, StopProcessingResponse, 
                     GetProcessingStatusRequest, GetProcessingStatusResponse, StreamData, StreamDataResponse, 
                     StreamResultsRequest, ProcessingResult};

pub mod streamforge_v1 {
    tonic::include_proto!("streamforge.v1");
}

#[derive(Debug)]
pub struct StreamProcessorService {
    processor: Arc<Mutex<StreamProcessor>>,
    config: Config,
}

#[tonic::async_trait]
impl StreamProcessorService for StreamProcessorService {
    async fn start_processing(
        &self,
        request: Request<StartProcessingRequest>,
    ) -> Result<Response<StartProcessingResponse>, Status> {
        let req = request.into_inner();
        info!("Starting processing for pipeline: {}", req.pipeline_id);

        let mut processor = self.processor.lock().await;
        match processor.start_pipeline(&req.pipeline_id, &req.config).await {
            Ok(job_id) => {
                info!("Processing started successfully with job_id: {}", job_id);
                Ok(Response::new(StartProcessingResponse {
                    success: true,
                    job_id,
                    message: "Processing started successfully".to_string(),
                }))
            }
            Err(e) => {
                error!("Failed to start processing: {}", e);
                Ok(Response::new(StartProcessingResponse {
                    success: false,
                    job_id: "".to_string(),
                    message: format!("Failed to start processing: {}", e),
                }))
            }
        }
    }

    async fn stop_processing(
        &self,
        request: Request<StopProcessingRequest>,
    ) -> Result<Response<StopProcessingResponse>, Status> {
        let req = request.into_inner();
        info!("Stopping processing for job: {}", req.job_id);

        let mut processor = self.processor.lock().await;
        match processor.stop_pipeline(&req.job_id).await {
            Ok(_) => {
                info!("Processing stopped successfully for job: {}", req.job_id);
                Ok(Response::new(StopProcessingResponse {
                    success: true,
                    message: "Processing stopped successfully".to_string(),
                }))
            }
            Err(e) => {
                error!("Failed to stop processing: {}", e);
                Ok(Response::new(StopProcessingResponse {
                    success: false,
                    message: format!("Failed to stop processing: {}", e),
                }))
            }
        }
    }

    async fn get_processing_status(
        &self,
        request: Request<GetProcessingStatusRequest>,
    ) -> Result<Response<GetProcessingStatusResponse>, Status> {
        let req = request.into_inner();
        
        let processor = self.processor.lock().await;
        match processor.get_pipeline_status(&req.job_id).await {
            Ok(status) => {
                Ok(Response::new(GetProcessingStatusResponse {
                    status: status.status,
                    job_id: req.job_id,
                    metrics: status.metrics,
                    start_time: Some(status.start_time.into()),
                    last_update: Some(status.last_update.into()),
                }))
            }
            Err(e) => {
                error!("Failed to get processing status: {}", e);
                Err(Status::internal(format!("Failed to get status: {}", e)))
            }
        }
    }

    type SendStreamDataStream = tokio_stream::wrappers::ReceiverStream<Result<StreamDataResponse, Status>>;

    async fn send_stream_data(
        &self,
        request: Request<tonic::Streaming<StreamData>>,
    ) -> Result<Response<Self::SendStreamDataStream>, Status> {
        let mut stream = request.into_inner();
        let (tx, rx) = tokio::sync::mpsc::channel(128);
        let processor = Arc::clone(&self.processor);

        tokio::spawn(async move {
            while let Some(data) = stream.message().await.unwrap_or(None) {
                let mut proc = processor.lock().await;
                match proc.process_stream_data(data).await {
                    Ok(_) => {
                        let _ = tx.send(Ok(StreamDataResponse {
                            success: true,
                            message: "Data processed successfully".to_string(),
                        })).await;
                    }
                    Err(e) => {
                        let _ = tx.send(Ok(StreamDataResponse {
                            success: false,
                            message: format!("Failed to process data: {}", e),
                        })).await;
                    }
                }
            }
        });

        Ok(Response::new(tokio_stream::wrappers::ReceiverStream::new(rx)))
    }

    type StreamResultsStream = tokio_stream::wrappers::ReceiverStream<Result<ProcessingResult, Status>>;

    async fn stream_results(
        &self,
        request: Request<StreamResultsRequest>,
    ) -> Result<Response<Self::StreamResultsStream>, Status> {
        let req = request.into_inner();
        let (tx, rx) = tokio::sync::mpsc::channel(128);
        let processor = Arc::clone(&self.processor);

        tokio::spawn(async move {
            let mut proc = processor.lock().await;
            if let Ok(mut result_stream) = proc.stream_results(&req.job_id).await {
                while let Some(result) = result_stream.next().await {
                    let _ = tx.send(Ok(result)).await;
                }
            }
        });

        Ok(Response::new(tokio_stream::wrappers::ReceiverStream::new(rx)))
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // ログの初期化
    tracing_subscriber::fmt::init();

    // 設定の読み込み
    let config = Config::load()?;
    info!("StreamForge Stream Processor starting with config: {:?}", config);

    // ストレージバックエンドの初期化
    let storage = StorageBackend::new(&config.storage).await?;
    info!("Storage backend initialized");

    // ストリームプロセッサーの初期化
    let processor = Arc::new(Mutex::new(StreamProcessor::new(config.clone(), storage).await?));
    info!("Stream processor initialized");

    // gRPCサービスの作成
    let service = StreamProcessorService {
        processor,
        config,
    };

    let addr = "[::1]:50051".parse()?;
    info!("Stream Processor server listening on {}", addr);

    // サーバーの起動
    Server::builder()
        .add_service(StreamProcessorServiceServer::new(service))
        .serve(addr)
        .await?;

    Ok(())
} 