package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AuthConfig は認証設定を定義します
type AuthConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
	APIKeyHeader  string
	RedisClient   *redis.Client
	Logger        *zap.Logger
}

// JWTClaims はJWTトークンのクレームを定義します
type JWTClaims struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	TenantID string   `json:"tenant_id"`
	jwt.RegisteredClaims
}

// AuthMiddleware は認証ミドルウェアを提供します
type AuthMiddleware struct {
	config *AuthConfig
}

// NewAuthMiddleware は新しい認証ミドルウェアを作成します
func NewAuthMiddleware(config *AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{config: config}
}

// JWT はJWTトークン認証を実装します
func (am *AuthMiddleware) JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Bearer トークンの形式をチェック
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := am.validateJWT(tokenString)
		if err != nil {
			am.config.Logger.Warn("JWT validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// トークンのブラックリストチェック
		if am.isTokenBlacklisted(tokenString) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		// コンテキストにユーザー情報を設定
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("tenant_id", claims.TenantID)

		c.Next()
	}
}

// APIKey はAPIキー認証を実装します
func (am *AuthMiddleware) APIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(am.config.APIKeyHeader)
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		// APIキーの検証
		valid, userInfo, err := am.validateAPIKey(apiKey)
		if err != nil {
			am.config.Logger.Error("API key validation error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication service error"})
			c.Abort()
			return
		}

		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// コンテキストにユーザー情報を設定
		c.Set("user_id", userInfo.UserID)
		c.Set("email", userInfo.Email)
		c.Set("roles", userInfo.Roles)
		c.Set("tenant_id", userInfo.TenantID)

		c.Next()
	}
}

// RequireRole は特定のロールを要求するミドルウェアを実装します
func (am *AuthMiddleware) RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User roles not found"})
			c.Abort()
			return
		}

		roles, ok := userRoles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid user roles"})
			c.Abort()
			return
		}

		// 必要なロールのいずれかを持っているかチェック
		hasRole := false
		for _, requiredRole := range requiredRoles {
			for _, userRole := range roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTenant は特定のテナントを要求するミドルウェアを実装します
func (am *AuthMiddleware) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		if !exists || tenantID == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tenant access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateJWT はJWTトークンを検証します
func (am *AuthMiddleware) validateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(am.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// isTokenBlacklisted はトークンがブラックリストに含まれているかチェックします
func (am *AuthMiddleware) isTokenBlacklisted(tokenString string) bool {
	if am.config.RedisClient == nil {
		return false
	}

	ctx := context.Background()
	key := fmt.Sprintf("blacklist:token:%s", tokenString)
	exists, err := am.config.RedisClient.Exists(ctx, key).Result()
	if err != nil {
		am.config.Logger.Error("Failed to check token blacklist", zap.Error(err))
		return false
	}

	return exists > 0
}

// validateAPIKey はAPIキーを検証します
func (am *AuthMiddleware) validateAPIKey(apiKey string) (bool, *UserInfo, error) {
	if am.config.RedisClient == nil {
		return false, nil, fmt.Errorf("Redis client not available")
	}

	ctx := context.Background()
	key := fmt.Sprintf("api_key:%s", apiKey)
	
	// RedisからAPIキー情報を取得
	userData, err := am.config.RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return false, nil, err
	}

	if len(userData) == 0 {
		return false, nil, nil
	}

	// APIキーの有効期限をチェック
	expiresAt, exists := userData["expires_at"]
	if exists && expiresAt != "" {
		expTime, err := time.Parse(time.RFC3339, expiresAt)
		if err == nil && time.Now().After(expTime) {
			return false, nil, nil
		}
	}

	// ユーザー情報を構築
	userInfo := &UserInfo{
		UserID:   userData["user_id"],
		Email:    userData["email"],
		TenantID: userData["tenant_id"],
	}

	// ロールを解析
	if roles, exists := userData["roles"]; exists {
		userInfo.Roles = strings.Split(roles, ",")
	}

	return true, userInfo, nil
}

// UserInfo はユーザー情報を定義します
type UserInfo struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	TenantID string   `json:"tenant_id"`
}

// GenerateJWT は新しいJWTトークンを生成します
func (am *AuthMiddleware) GenerateJWT(userInfo *UserInfo) (string, error) {
	claims := &JWTClaims{
		UserID:   userInfo.UserID,
		Email:    userInfo.Email,
		Roles:    userInfo.Roles,
		TenantID: userInfo.TenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(am.config.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(am.config.JWTSecret))
}

// RevokeToken はトークンを無効化します
func (am *AuthMiddleware) RevokeToken(tokenString string) error {
	if am.config.RedisClient == nil {
		return fmt.Errorf("Redis client not available")
	}

	// トークンの有効期限を取得
	claims, err := am.validateJWT(tokenString)
	if err != nil {
		return err
	}

	// ブラックリストに追加
	ctx := context.Background()
	key := fmt.Sprintf("blacklist:token:%s", tokenString)
	expiration := time.Until(claims.ExpiresAt.Time)
	
	return am.config.RedisClient.Set(ctx, key, "revoked", expiration).Err()
} 