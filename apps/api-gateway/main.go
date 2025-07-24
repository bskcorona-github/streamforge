package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/config"
	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/handlers"
	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/middleware"
	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/repository"
	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title StreamForge API
// @version 1.0
// @description StreamForge is a real-time observability platform that provides comprehensive monitoring, alerting, and analytics capabilities for distributed systems.
// @termsOfService https://github.com/bskcorona-github/streamforge
// @contact.name StreamForge Team
// @contact.url https://github.com/bskcorona-github/streamforge
// @contact.email team@streamforge.dev
// @license.name Apache 2.0
// @license.url https://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API Key for authentication
func main() {
	// 設定の読み込み
	cfg := config.Load()

	// ロガーの初期化
	logger := initLogger(cfg)
	defer logger.Sync()

	// OpenTelemetryの初期化
	tracer, err := initTracer(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	defer func() {
		if err := tracer.Shutdown(context.Background()); err != nil {
			logger.Error("Error shutting down tracer", zap.Error(err))
		}
	}()

	// データベース接続の初期化
	db, err := initDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Redis接続の初期化
	redisClient := initRedis(cfg)

	// リポジトリの初期化
	repo := repository.NewRepository(db, redisClient)

	// サービスの初期化
	svc := service.NewService(repo, logger)

	// ハンドラーの初期化
	handler := handlers.NewHandler(svc, logger)

	// 認証ミドルウェアの初期化
	authMiddleware := middleware.NewAuthMiddleware(&middleware.AuthConfig{
		JWTSecret:     cfg.JWTSecret,
		JWTExpiration: cfg.JWTExpiration,
		APIKeyHeader:  "X-API-Key",
		RedisClient:   redisClient,
		Logger:        logger,
	})

	// Ginルーターの設定
	router := gin.New()
	router.Use(gin.Recovery())

	// ミドルウェアの設定
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(cfg.RateLimit.Limit, cfg.RateLimit.Window))
	router.Use(middleware.Tracing(tracer))
	router.Use(middleware.RequestID())

	// ヘルスチェックエンドポイント
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   cfg.Version,
		})
	})

	// Prometheusメトリクスエンドポイント
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swaggerドキュメントエンドポイント
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 認証関連エンドポイント（認証不要）
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/login", handler.Login)
		auth.POST("/register", handler.Register)
		auth.POST("/refresh", handler.RefreshToken)
		auth.POST("/logout", authMiddleware.JWT(), handler.Logout)
	}

	// APIルートの設定（認証必要）
	api := router.Group("/api/v1")
	api.Use(authMiddleware.JWT()) // JWT認証を適用
	{
		// メトリクス関連
		metrics := api.Group("/metrics")
		{
			metrics.GET("", handler.GetMetrics)
			metrics.GET("/:name", handler.GetMetric)
			metrics.GET("/:name/timeseries", handler.GetTimeSeries)
			metrics.POST("", authMiddleware.RequireRole("admin", "write"), handler.CreateMetric)
			metrics.PUT("/:name", authMiddleware.RequireRole("admin", "write"), handler.UpdateMetric)
			metrics.DELETE("/:name", authMiddleware.RequireRole("admin"), handler.DeleteMetric)
		}

		// ログ関連
		logs := api.Group("/logs")
		{
			logs.GET("", handler.GetLogs)
			logs.GET("/search", handler.SearchLogs)
		}

		// トレース関連
		traces := api.Group("/traces")
		{
			traces.GET("", handler.GetTraces)
			traces.GET("/:traceId", handler.GetTrace)
			traces.GET("/:traceId/graph", handler.GetTraceGraph)
		}

		// アラート関連
		alerts := api.Group("/alerts")
		{
			alerts.GET("", handler.GetAlerts)
			alerts.GET("/:id", handler.GetAlert)
			alerts.POST("", authMiddleware.RequireRole("admin", "write"), handler.CreateAlert)
			alerts.PUT("/:id", authMiddleware.RequireRole("admin", "write"), handler.UpdateAlert)
			alerts.DELETE("/:id", authMiddleware.RequireRole("admin"), handler.DeleteAlert)
		}

		// ダッシュボード関連
		dashboards := api.Group("/dashboards")
		{
			dashboards.GET("", handler.GetDashboards)
			dashboards.GET("/:id", handler.GetDashboard)
			dashboards.POST("", authMiddleware.RequireRole("admin", "write"), handler.CreateDashboard)
			dashboards.PUT("/:id", authMiddleware.RequireRole("admin", "write"), handler.UpdateDashboard)
			dashboards.DELETE("/:id", authMiddleware.RequireRole("admin"), handler.DeleteDashboard)
		}

		// ユーザー管理（管理者のみ）
		users := api.Group("/users")
		users.Use(authMiddleware.RequireRole("admin"))
		{
			users.GET("", handler.GetUsers)
			users.GET("/:id", handler.GetUser)
			users.POST("", handler.CreateUser)
			users.PUT("/:id", handler.UpdateUser)
			users.DELETE("/:id", handler.DeleteUser)
		}

		// APIキー管理
		apikeys := api.Group("/apikeys")
		{
			apikeys.GET("", handler.GetAPIKeys)
			apikeys.POST("", authMiddleware.RequireRole("admin", "write"), handler.CreateAPIKey)
			apikeys.DELETE("/:id", authMiddleware.RequireRole("admin", "write"), handler.DeleteAPIKey)
		}
	}

	// サーバーの起動
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// グレースフルシャットダウンの設定
	go func() {
		logger.Info("Starting API Gateway server", zap.Int("port", cfg.Port))
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

func initLogger(cfg *config.Config) *zap.Logger {
	var logger *zap.Logger
	var err error

	if cfg.Environment == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatal("Failed to initialize logger", err)
	}

	return logger
}

func initTracer(cfg *config.Config) (*sdktrace.TracerProvider, error) {
	// Jaegerエクスポーターの設定
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.JaegerEndpoint)))
	if err != nil {
		return nil, err
	}

	// リソースの設定
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("streamforge-api-gateway"),
			semconv.ServiceVersion(cfg.Version),
		),
	)
	if err != nil {
		return nil, err
	}

	// TracerProviderの設定
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	// グローバルTracerProviderの設定
	otel.SetTracerProvider(tp)

	return tp, nil
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// マイグレーションの実行
	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 接続テスト
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis", err)
	}

	return client
}

func runMigrations(db *gorm.DB) error {
	// ここでデータベースマイグレーションを実行
	// 実際のプロジェクトでは、migrateライブラリを使用することを推奨
	return nil
} 