package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Logger はリクエストログを記録するミドルウェアです
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.String("client_ip", param.ClientIP),
			zap.Int("status_code", param.StatusCode),
			zap.String("latency", param.Latency.String()),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.String("error", param.ErrorMessage),
		)
		return ""
	})
}

// CORS はCORSヘッダーを設定するミドルウェアです
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimit はレート制限を実装するミドルウェアです
func RateLimit(limit int, window int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// Redisクライアントの取得（実際の実装では依存性注入を使用）
		rdb := getRedisClient()
		if rdb == nil {
			c.Next()
			return
		}

		ctx := context.Background()
		current, err := rdb.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
			c.Abort()
			return
		}

		if current >= limit {
			c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Duration(window)*time.Second).Unix(), 10))
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		pipe := rdb.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Duration(window)*time.Second)
		cmds, err := pipe.Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit update failed"})
			c.Abort()
			return
		}

		newCount := cmds[0].(*redis.IntCmd).Val()
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(limit-newCount, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Duration(window)*time.Second).Unix(), 10))

		c.Next()
	}
}

// Tracing はOpenTelemetryトレーシングを実装するミドルウェアです
func Tracing(tracer trace.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// トレースコンテキストの抽出
		ctx := c.Request.Context()
		spanCtx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		defer span.End()

		// リクエストIDの生成
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// コンテキストにリクエストIDを追加
		ctx = context.WithValue(spanCtx, "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		// レスポンスヘッダーにリクエストIDを追加
		c.Header("X-Request-ID", requestID)

		// リクエスト処理
		c.Next()

		// レスポンス情報をスパンに追加
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// エラーの場合
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("error", c.Errors.String()))
		}
	}
}

// Auth はJWT認証を実装するミドルウェアです
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Bearerトークンの抽出
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// JWTトークンの検証（実際の実装では適切なJWTライブラリを使用）
		claims, err := validateJWT(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// コンテキストにユーザー情報を追加
		ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequestID はリクエストIDを生成・管理するミドルウェアです
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// Metrics はPrometheusメトリクスを収集するミドルウェアです
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		// メトリクスの記録（実際の実装ではPrometheusクライアントを使用）
		recordHTTPMetrics(c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	}
}

// Recovery はパニックを回復するミドルウェアです
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logger.Error("Panic recovered",
				zap.String("error", err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	})
}

// ヘルパー関数

// getRedisClient はRedisクライアントを取得します（実際の実装では依存性注入を使用）
func getRedisClient() *redis.Client {
	// 実際の実装では、アプリケーションコンテキストからRedisクライアントを取得
	return nil
}

// validateJWT はJWTトークンを検証します（実際の実装では適切なJWTライブラリを使用）
func validateJWT(tokenString, secret string) (*JWTClaims, error) {
	// 実際の実装では、JWTライブラリを使用してトークンを検証
	return &JWTClaims{
		UserID: "user123",
		Email:  "user@example.com",
	}, nil
}

// recordHTTPMetrics はHTTPメトリクスを記録します（実際の実装ではPrometheusクライアントを使用）
func recordHTTPMetrics(method, path string, statusCode int, duration float64) {
	// 実際の実装では、Prometheusクライアントを使用してメトリクスを記録
}

// JWTClaims はJWTクレームを表します
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
} 