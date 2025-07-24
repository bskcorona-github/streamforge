package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bskcorona-github/streamforge/apps/collector/internal/collector"
	"github.com/bskcorona-github/streamforge/apps/collector/internal/config"
	"github.com/bskcorona-github/streamforge/apps/collector/internal/telemetry"
	"go.uber.org/zap"
)

func main() {
	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ロガーの初期化
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// テレメトリの初期化
	tp, err := telemetry.InitTracer(cfg.Telemetry.Endpoint)
	if err != nil {
		logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("Error shutting down tracer provider", zap.Error(err))
		}
	}()

	// コレクターの初期化
	c, err := collector.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create collector", zap.Error(err))
	}

	// コンテキストの作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// コレクターの開始
	go func() {
		if err := c.Start(ctx); err != nil {
			logger.Error("Collector failed", zap.Error(err))
			cancel()
		}
	}()

	logger.Info("StreamForge Collector started",
		zap.String("version", "0.1.0"),
		zap.String("endpoint", cfg.API.Endpoint),
		zap.Duration("interval", cfg.Collection.Interval),
	)

	// シグナル待機
	select {
	case sig := <-sigChan:
		logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	}

	// グレースフルシャットダウン
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := c.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	logger.Info("StreamForge Collector stopped")
} 