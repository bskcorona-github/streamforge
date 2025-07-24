#!/usr/bin/env python3
"""
StreamForge ML Engine

Machine Learning engine for StreamForge observability platform.
Provides anomaly detection, forecasting, and pattern recognition capabilities.
"""

import asyncio
import logging
import signal
import sys
from typing import Optional

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
import uvicorn

from ml_engine.config import Config
from ml_engine.models import AnomalyDetector, Forecaster, PatternRecognizer
from ml_engine.api import router as api_router
from ml_engine.metrics import MetricsCollector
from ml_engine.storage import ModelStorage

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class MLEngine:
    """Main ML Engine class"""
    
    def __init__(self, config: Config):
        self.config = config
        self.app = FastAPI(
            title="StreamForge ML Engine",
            description="Machine Learning engine for StreamForge",
            version="0.1.0"
        )
        
        # Initialize components
        self.storage = ModelStorage(config.storage)
        self.metrics = MetricsCollector(config.metrics)
        
        # Initialize ML models
        self.anomaly_detector = AnomalyDetector(config.models.anomaly_detection)
        self.forecaster = Forecaster(config.models.forecasting)
        self.pattern_recognizer = PatternRecognizer(config.models.pattern_recognition)
        
        # Setup FastAPI
        self._setup_fastapi()
        
    def _setup_fastapi(self):
        """Setup FastAPI application"""
        # Add CORS middleware
        self.app.add_middleware(
            CORSMiddleware,
            allow_origins=self.config.api.cors_origins,
            allow_credentials=True,
            allow_methods=["*"],
            allow_headers=["*"],
        )
        
        # Include API router
        self.app.include_router(api_router, prefix="/api/v1")
        
        # Add health check endpoint
        @self.app.get("/health")
        async def health_check():
            return {
                "status": "healthy",
                "service": "ml-engine",
                "version": "0.1.0"
            }
        
        # Add metrics endpoint
        @self.app.get("/metrics")
        async def get_metrics():
            return await self.metrics.get_metrics()
    
    async def start(self):
        """Start the ML Engine"""
        logger.info("Starting StreamForge ML Engine...")
        
        try:
            # Initialize storage
            await self.storage.initialize()
            logger.info("Storage initialized")
            
            # Load models
            await self._load_models()
            logger.info("Models loaded")
            
            # Start metrics collection
            await self.metrics.start()
            logger.info("Metrics collection started")
            
            # Start FastAPI server
            config = uvicorn.Config(
                app=self.app,
                host=self.config.api.host,
                port=self.config.api.port,
                log_level="info"
            )
            server = uvicorn.Server(config)
            await server.serve()
            
        except Exception as e:
            logger.error(f"Failed to start ML Engine: {e}")
            raise
    
    async def stop(self):
        """Stop the ML Engine"""
        logger.info("Stopping StreamForge ML Engine...")
        
        try:
            # Stop metrics collection
            await self.metrics.stop()
            logger.info("Metrics collection stopped")
            
            # Save models
            await self._save_models()
            logger.info("Models saved")
            
            # Close storage
            await self.storage.close()
            logger.info("Storage closed")
            
        except Exception as e:
            logger.error(f"Error stopping ML Engine: {e}")
    
    async def _load_models(self):
        """Load ML models from storage"""
        try:
            await self.anomaly_detector.load(self.storage)
            await self.forecaster.load(self.storage)
            await self.pattern_recognizer.load(self.storage)
        except Exception as e:
            logger.warning(f"Failed to load some models: {e}")
    
    async def _save_models(self):
        """Save ML models to storage"""
        try:
            await self.anomaly_detector.save(self.storage)
            await self.forecaster.save(self.storage)
            await self.pattern_recognizer.save(self.storage)
        except Exception as e:
            logger.error(f"Failed to save models: {e}")

async def main():
    """Main entry point"""
    # Load configuration
    config = Config()
    
    # Create ML Engine instance
    engine = MLEngine(config)
    
    # Setup signal handlers
    def signal_handler(signum, frame):
        logger.info(f"Received signal {signum}")
        asyncio.create_task(engine.stop())
        sys.exit(0)
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    try:
        await engine.start()
    except KeyboardInterrupt:
        logger.info("Received keyboard interrupt")
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        sys.exit(1)
    finally:
        await engine.stop()

if __name__ == "__main__":
    asyncio.run(main()) 