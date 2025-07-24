package routes

import (
	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/handlers"
	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// SetupRoutes はルーティングを設定します
func SetupRoutes(router *gin.Engine, handler *handlers.Handler, logger *zap.Logger) {
	// ミドルウェアを適用
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.RequestID())
	router.Use(middleware.Tracing())

	// ヘルスチェック
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "streamforge-api-gateway",
		})
	})

	// メトリクスエンドポイント（Prometheus）
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 グループ
	v1 := router.Group("/api/v1")
	{
		// メトリクス関連
		metrics := v1.Group("/metrics")
		{
			metrics.GET("", handler.GetMetrics)
			metrics.GET("/:name", handler.GetMetric)
			metrics.GET("/:name/timeseries", handler.GetTimeSeries)
		}

		// ログ関連
		logs := v1.Group("/logs")
		{
			logs.GET("", handler.GetLogs)
			logs.GET("/search", handler.SearchLogs)
		}

		// トレース関連
		traces := v1.Group("/traces")
		{
			traces.GET("", handler.GetTraces)
			traces.GET("/:traceId", handler.GetTrace)
			traces.GET("/:traceId/graph", handler.GetTraceGraph)
		}

		// アラート関連
		alerts := v1.Group("/alerts")
		{
			alerts.GET("", handler.GetAlerts)
			alerts.GET("/:id", handler.GetAlert)
			alerts.POST("", handler.CreateAlert)
			alerts.PUT("/:id", handler.UpdateAlert)
			alerts.DELETE("/:id", handler.DeleteAlert)
		}

		// ダッシュボード関連
		dashboards := v1.Group("/dashboards")
		{
			dashboards.GET("", handler.GetDashboards)
			dashboards.GET("/:id", handler.GetDashboard)
			dashboards.POST("", handler.CreateDashboard)
			dashboards.PUT("/:id", handler.UpdateDashboard)
			dashboards.DELETE("/:id", handler.DeleteDashboard)
		}
	}

	// GraphQL エンドポイント（将来の実装）
	router.GET("/graphql", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "GraphQL endpoint - coming soon",
		})
	})

	// gRPC-Web エンドポイント（将来の実装）
	router.GET("/grpc", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "gRPC-Web endpoint - coming soon",
		})
	})

	// 404 ハンドラー
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"message": "The requested resource was not found",
			"path": c.Request.URL.Path,
		})
	})
} 