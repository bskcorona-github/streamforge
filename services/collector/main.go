package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/streamforge/streamforge/services/collector/collectors"
	"github.com/streamforge/streamforge/services/collector/config"
	"github.com/streamforge/streamforge/services/collector/processor"
	"github.com/streamforge/streamforge/services/collector/storage"
	"go.uber.org/zap"
)

func main() {
	// ロガーの初期化
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 設定の読み込み
	cfg := config.Load()

	logger.Info("Starting StreamForge Collector", 
		zap.String("environment", cfg.Environment),
		zap.String("collector_id", cfg.Collector.ID))

	// ストレージの初期化
	storageClient, err := storage.NewClient(cfg.Storage)
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}
	defer storageClient.Close()

	// プロセッサーの初期化
	proc := processor.NewProcessor(cfg.Processor, storageClient, logger)

	// コレクターの初期化
	collector := collectors.NewCollector(cfg.Collector, proc, logger)

	// コンテキストの作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// コレクターの開始
	if err := collector.Start(ctx); err != nil {
		logger.Fatal("Failed to start collector", zap.Error(err))
	}

	logger.Info("Collector started successfully")

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down collector...")

	// グレースフルシャットダウン
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := collector.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	logger.Info("Collector stopped")
} 