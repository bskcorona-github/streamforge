package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/streamforge/streamforge/services/api-gateway/config"
	"github.com/streamforge/streamforge/services/api-gateway/handlers"
	"github.com/streamforge/streamforge/services/api-gateway/middleware"
	"go.uber.org/zap"
)

func main() {
	// ロガーの初期化
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 設定の読み込み
	cfg := config.Load()

	// Ginの設定
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ルーターの初期化
	router := gin.New()

	// ミドルウェアの設定
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.RequestID())

	// ヘルスチェック
	router.GET("/health", handlers.HealthCheck)

	// API v1 ルート
	v1 := router.Group("/api/v1")
	{
		// メトリクス関連
		metrics := v1.Group("/metrics")
		{
			metrics.POST("/", handlers.SendMetrics)
			metrics.GET("/", handlers.GetMetrics)
			metrics.GET("/timeseries", handlers.GetTimeSeries)
			metrics.GET("/stream", handlers.StreamMetrics)
		}

		// ログ関連
		logs := v1.Group("/logs")
		{
			logs.POST("/", handlers.SendLogs)
			logs.GET("/", handlers.GetLogs)
			logs.GET("/stream", handlers.StreamLogs)
		}

		// トレース関連
		traces := v1.Group("/traces")
		{
			traces.POST("/", handlers.SendTrace)
			traces.GET("/:trace_id", handlers.GetTrace)
			traces.GET("/stream", handlers.StreamTraces)
		}

		// アラート関連
		alerts := v1.Group("/alerts")
		{
			alerts.POST("/", handlers.CreateAlert)
			alerts.GET("/", handlers.ListAlerts)
			alerts.GET("/:alert_id", handlers.GetAlert)
			alerts.PUT("/:alert_id", handlers.UpdateAlert)
			alerts.DELETE("/:alert_id", handlers.DeleteAlert)
			alerts.GET("/stream", handlers.StreamAlerts)
		}

		// サービス管理
		services := v1.Group("/services")
		{
			services.POST("/", handlers.RegisterService)
			services.GET("/", handlers.ListServices)
			services.PUT("/:service_name", handlers.UpdateService)
			services.DELETE("/:service_name", handlers.UnregisterService)
			services.POST("/:service_name/heartbeat", handlers.Heartbeat)
		}
	}

	// サーバーの設定
	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// グレースフルシャットダウンの設定
	go func() {
		logger.Info("Starting API Gateway server", zap.String("address", cfg.Server.Address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// グレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
} 