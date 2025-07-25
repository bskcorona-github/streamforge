#!/usr/bin/env python3
"""
StreamForge ML Engine - 異常検知と機械学習パイプライン
"""

import asyncio
import logging
import signal
import sys
from typing import Dict, List, Optional, Any
from dataclasses import dataclass
from datetime import datetime
import json

import numpy as np
import pandas as pd
from sklearn.ensemble import IsolationForest, RandomForestClassifier
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import train_test_split
from sklearn.metrics import classification_report, confusion_matrix
import joblib
import redis
import grpc
from concurrent import futures

# Protocol Buffers
import streamforge_pb2
import streamforge_pb2_grpc

# 設定
@dataclass
class Config:
    redis_host: str = "localhost"
    redis_port: int = 6379
    grpc_port: int = 50052
    model_storage_path: str = "/tmp/models"
    log_level: str = "INFO"
    max_workers: int = 10

class AnomalyDetector:
    """異常検知エンジン"""
    
    def __init__(self, config: Config):
        self.config = config
        self.models: Dict[str, Any] = {}
        self.scalers: Dict[str, StandardScaler] = {}
        self.redis_client = redis.Redis(
            host=config.redis_host, 
            port=config.redis_port, 
            decode_responses=True
        )
        
    async def train_anomaly_model(self, model_id: str, data: pd.DataFrame) -> bool:
        """異常検知モデルの訓練"""
        try:
            logging.info(f"Training anomaly detection model: {model_id}")
            
            # データの前処理
            scaler = StandardScaler()
            scaled_data = scaler.fit_transform(data)
            
            # Isolation Forestモデルの訓練
            model = IsolationForest(
                contamination=0.1,
                random_state=42,
                n_estimators=100
            )
            model.fit(scaled_data)
            
            # モデルの保存
            self.models[model_id] = model
            self.scalers[model_id] = scaler
            
            # Redisにモデル情報を保存
            model_info = {
                "model_id": model_id,
                "trained_at": datetime.now().isoformat(),
                "data_shape": data.shape,
                "contamination": 0.1
            }
            self.redis_client.hset(f"model:{model_id}", mapping=model_info)
            
            logging.info(f"Anomaly detection model {model_id} trained successfully")
            return True
            
        except Exception as e:
            logging.error(f"Failed to train anomaly model {model_id}: {e}")
            return False
    
    async def detect_anomalies(self, model_id: str, data: pd.DataFrame) -> List[bool]:
        """異常検知の実行"""
        try:
            if model_id not in self.models:
                raise ValueError(f"Model {model_id} not found")
            
            model = self.models[model_id]
            scaler = self.scalers[model_id]
            
            # データの前処理
            scaled_data = scaler.transform(data)
            
            # 異常検知
            predictions = model.predict(scaled_data)
            # -1: 異常, 1: 正常
            anomalies = predictions == -1
            
            logging.info(f"Detected {sum(anomalies)} anomalies out of {len(data)} samples")
            return anomalies.tolist()
            
        except Exception as e:
            logging.error(f"Failed to detect anomalies with model {model_id}: {e}")
            return []
    
    async def get_model_info(self, model_id: str) -> Optional[Dict[str, Any]]:
        """モデル情報の取得"""
        try:
            model_data = self.redis_client.hgetall(f"model:{model_id}")
            if model_data:
                return model_data
            return None
        except Exception as e:
            logging.error(f"Failed to get model info for {model_id}: {e}")
            return None

class MLPipeline:
    """機械学習パイプライン"""
    
    def __init__(self, config: Config):
        self.config = config
        self.anomaly_detector = AnomalyDetector(config)
        self.pipelines: Dict[str, Any] = {}
        
    async def create_pipeline(self, pipeline_id: str, config: Dict[str, Any]) -> bool:
        """パイプラインの作成"""
        try:
            logging.info(f"Creating ML pipeline: {pipeline_id}")
            
            pipeline_config = {
                "pipeline_id": pipeline_id,
                "config": config,
                "created_at": datetime.now().isoformat(),
                "status": "created"
            }
            
            self.pipelines[pipeline_id] = pipeline_config
            
            # Redisにパイプライン情報を保存
            redis_client = redis.Redis(
                host=self.config.redis_host,
                port=self.config.redis_port,
                decode_responses=True
            )
            redis_client.hset(f"pipeline:{pipeline_id}", mapping=pipeline_config)
            
            logging.info(f"ML pipeline {pipeline_id} created successfully")
            return True
            
        except Exception as e:
            logging.error(f"Failed to create pipeline {pipeline_id}: {e}")
            return False
    
    async def execute_pipeline(self, pipeline_id: str, data: pd.DataFrame) -> Dict[str, Any]:
        """パイプラインの実行"""
        try:
            if pipeline_id not in self.pipelines:
                raise ValueError(f"Pipeline {pipeline_id} not found")
            
            pipeline_config = self.pipelines[pipeline_id]
            results = {
                "pipeline_id": pipeline_id,
                "executed_at": datetime.now().isoformat(),
                "data_shape": data.shape,
                "results": {}
            }
            
            # パイプラインの設定に基づいて処理を実行
            if "anomaly_detection" in pipeline_config["config"]:
                model_id = pipeline_config["config"]["anomaly_detection"]["model_id"]
                anomalies = await self.anomaly_detector.detect_anomalies(model_id, data)
                results["results"]["anomalies"] = anomalies
                results["results"]["anomaly_count"] = sum(anomalies)
            
            logging.info(f"Pipeline {pipeline_id} executed successfully")
            return results
            
        except Exception as e:
            logging.error(f"Failed to execute pipeline {pipeline_id}: {e}")
            return {"error": str(e)}

class MLEngineService(streamforge_pb2_grpc.MLEngineServiceServicer):
    """gRPC ML Engine サービス"""
    
    def __init__(self, config: Config):
        self.config = config
        self.pipeline = MLPipeline(config)
        
    async def TrainModel(self, request, context):
        """モデルの訓練"""
        try:
            model_id = request.model_id
            data = pd.DataFrame(request.data)
            
            success = await self.pipeline.anomaly_detector.train_anomaly_model(model_id, data)
            
            return streamforge_pb2.TrainModelResponse(
                success=success,
                model_id=model_id,
                message="Model trained successfully" if success else "Failed to train model"
            )
            
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return streamforge_pb2.TrainModelResponse(
                success=False,
                model_id=request.model_id,
                message=f"Error: {str(e)}"
            )
    
    async def DetectAnomalies(self, request, context):
        """異常検知の実行"""
        try:
            model_id = request.model_id
            data = pd.DataFrame(request.data)
            
            anomalies = await self.pipeline.anomaly_detector.detect_anomalies(model_id, data)
            
            return streamforge_pb2.DetectAnomaliesResponse(
                success=True,
                anomalies=anomalies,
                anomaly_count=sum(anomalies),
                message="Anomaly detection completed"
            )
            
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return streamforge_pb2.DetectAnomaliesResponse(
                success=False,
                anomalies=[],
                anomaly_count=0,
                message=f"Error: {str(e)}"
            )
    
    async def CreatePipeline(self, request, context):
        """パイプラインの作成"""
        try:
            pipeline_id = request.pipeline_id
            config = json.loads(request.config)
            
            success = await self.pipeline.create_pipeline(pipeline_id, config)
            
            return streamforge_pb2.CreatePipelineResponse(
                success=success,
                pipeline_id=pipeline_id,
                message="Pipeline created successfully" if success else "Failed to create pipeline"
            )
            
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return streamforge_pb2.CreatePipelineResponse(
                success=False,
                pipeline_id=request.pipeline_id,
                message=f"Error: {str(e)}"
            )
    
    async def ExecutePipeline(self, request, context):
        """パイプラインの実行"""
        try:
            pipeline_id = request.pipeline_id
            data = pd.DataFrame(request.data)
            
            results = await self.pipeline.execute_pipeline(pipeline_id, data)
            
            return streamforge_pb2.ExecutePipelineResponse(
                success="error" not in results,
                pipeline_id=pipeline_id,
                results=json.dumps(results),
                message="Pipeline executed successfully" if "error" not in results else f"Error: {results['error']}"
            )
            
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return streamforge_pb2.ExecutePipelineResponse(
                success=False,
                pipeline_id=request.pipeline_id,
                results="{}",
                message=f"Error: {str(e)}"
            )

async def serve(config: Config):
    """gRPCサーバーの起動"""
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=config.max_workers))
    streamforge_pb2_grpc.add_MLEngineServiceServicer_to_server(
        MLEngineService(config), server
    )
    
    listen_addr = f"[::]:{config.grpc_port}"
    server.add_insecure_port(listen_addr)
    
    logging.info(f"ML Engine server starting on {listen_addr}")
    await server.start()
    
    # シグナルハンドリング
    async def shutdown(signum, frame):
        logging.info("Shutting down ML Engine server...")
        await server.stop(0)
    
    # シグナルハンドラーの設定
    for sig in (signal.SIGINT, signal.SIGTERM):
        signal.signal(sig, lambda s, f: asyncio.create_task(shutdown(s, f)))
    
    await server.wait_for_termination()

def main():
    """メイン関数"""
    # ログの設定
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    # 設定の読み込み
    config = Config()
    
    try:
        # gRPCサーバーの起動
        asyncio.run(serve(config))
    except KeyboardInterrupt:
        logging.info("ML Engine server stopped by user")
    except Exception as e:
        logging.error(f"ML Engine server error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main() 