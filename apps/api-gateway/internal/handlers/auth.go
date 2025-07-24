package handlers

import (
	"net/http"
	"time"

	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/middleware"
	"github.com/bskcorona-github/streamforge/apps/api-gateway/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler は認証関連のハンドラーを提供します
type AuthHandler struct {
	service       *service.Service
	logger        *zap.Logger
	authMiddleware *middleware.AuthMiddleware
}

// NewAuthHandler は新しい認証ハンドラーを作成します
func NewAuthHandler(svc *service.Service, logger *zap.Logger, authMiddleware *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		service:        svc,
		logger:         logger,
		authMiddleware: authMiddleware,
	}
}

// LoginRequest はログインリクエストを定義します
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest は登録リクエストを定義します
type RegisterRequest struct {
	Email     string   `json:"email" binding:"required,email"`
	Password  string   `json:"password" binding:"required,min=6"`
	Name      string   `json:"name" binding:"required"`
	Roles     []string `json:"roles"`
	TenantID  string   `json:"tenant_id"`
}

// RefreshTokenRequest はトークン更新リクエストを定義します
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse は認証レスポンスを定義します
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo はユーザー情報を定義します
type UserInfo struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Roles    []string `json:"roles"`
	TenantID string   `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Login はユーザーログインを処理します
// @Summary ユーザーログイン
// @Description ユーザーの認証を行い、JWTトークンを返します
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "ログイン情報"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// ユーザーの検証
	user, err := h.service.GetUserByEmail(req.Email)
	if err != nil {
		h.logger.Warn("Login failed: user not found", zap.String("email", req.Email))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Authentication failed",
			Message: "Invalid email or password",
		})
		return
	}

	// パスワードの検証
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.logger.Warn("Login failed: invalid password", zap.String("email", req.Email))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Authentication failed",
			Message: "Invalid email or password",
		})
		return
	}

	// JWTトークンの生成
	userInfo := &middleware.UserInfo{
		UserID:   user.ID,
		Email:    user.Email,
		Roles:    user.Roles,
		TenantID: user.TenantID,
	}

	accessToken, err := h.authMiddleware.GenerateJWT(userInfo)
	if err != nil {
		h.logger.Error("Failed to generate access token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to generate token",
		})
		return
	}

	// リフレッシュトークンの生成
	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	// リフレッシュトークンをRedisに保存
	err = h.service.StoreRefreshToken(refreshToken, user.ID, expiresAt)
	if err != nil {
		h.logger.Error("Failed to store refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to store refresh token",
		})
		return
	}

	// レスポンスの構築
	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User: UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Roles:     user.Roles,
			TenantID:  user.TenantID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	h.logger.Info("User logged in successfully", zap.String("email", user.Email))
	c.JSON(http.StatusOK, response)
}

// Register はユーザー登録を処理します
// @Summary ユーザー登録
// @Description 新しいユーザーを登録します
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "登録情報"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// ユーザーが既に存在するかチェック
	existingUser, _ := h.service.GetUserByEmail(req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "User already exists",
			Message: "A user with this email already exists",
		})
		return
	}

	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to process registration",
		})
		return
	}

	// デフォルトロールの設定
	if len(req.Roles) == 0 {
		req.Roles = []string{"user"}
	}

	// ユーザーの作成
	user := &service.User{
		ID:       uuid.New().String(),
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		Roles:    req.Roles,
		TenantID: req.TenantID,
	}

	err = h.service.CreateUser(user)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to create user",
		})
		return
	}

	// JWTトークンの生成
	userInfo := &middleware.UserInfo{
		UserID:   user.ID,
		Email:    user.Email,
		Roles:    user.Roles,
		TenantID: user.TenantID,
	}

	accessToken, err := h.authMiddleware.GenerateJWT(userInfo)
	if err != nil {
		h.logger.Error("Failed to generate access token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to generate token",
		})
		return
	}

	// リフレッシュトークンの生成
	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	// リフレッシュトークンをRedisに保存
	err = h.service.StoreRefreshToken(refreshToken, user.ID, expiresAt)
	if err != nil {
		h.logger.Error("Failed to store refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to store refresh token",
		})
		return
	}

	// レスポンスの構築
	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User: UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Roles:     user.Roles,
			TenantID:  user.TenantID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	h.logger.Info("User registered successfully", zap.String("email", user.Email))
	c.JSON(http.StatusCreated, response)
}

// RefreshToken はトークンの更新を処理します
// @Summary トークン更新
// @Description リフレッシュトークンを使用してアクセストークンを更新します
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "リフレッシュトークン"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// リフレッシュトークンの検証
	userID, err := h.service.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("Invalid refresh token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid refresh token",
			Message: "The refresh token is invalid or expired",
		})
		return
	}

	// ユーザー情報の取得
	user, err := h.service.GetUserByID(userID)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to get user information",
		})
		return
	}

	// 新しいJWTトークンの生成
	userInfo := &middleware.UserInfo{
		UserID:   user.ID,
		Email:    user.Email,
		Roles:    user.Roles,
		TenantID: user.TenantID,
	}

	accessToken, err := h.authMiddleware.GenerateJWT(userInfo)
	if err != nil {
		h.logger.Error("Failed to generate access token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to generate token",
		})
		return
	}

	// 新しいリフレッシュトークンの生成
	newRefreshToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	// 古いリフレッシュトークンを削除
	h.service.DeleteRefreshToken(req.RefreshToken)

	// 新しいリフレッシュトークンをRedisに保存
	err = h.service.StoreRefreshToken(newRefreshToken, user.ID, expiresAt)
	if err != nil {
		h.logger.Error("Failed to store refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to store refresh token",
		})
		return
	}

	// レスポンスの構築
	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		User: UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Roles:     user.Roles,
			TenantID:  user.TenantID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	h.logger.Info("Token refreshed successfully", zap.String("user_id", user.ID))
	c.JSON(http.StatusOK, response)
}

// Logout はユーザーログアウトを処理します
// @Summary ユーザーログアウト
// @Description ユーザーをログアウトし、トークンを無効化します
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 現在のトークンを取得
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "No authorization header",
		})
		return
	}

	// Bearer トークンの形式をチェック
	parts := gin.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "Invalid authorization header format",
		})
		return
	}

	tokenString := parts[1]

	// トークンを無効化
	err := h.authMiddleware.RevokeToken(tokenString)
	if err != nil {
		h.logger.Error("Failed to revoke token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal server error",
			Message: "Failed to logout",
		})
		return
	}

	h.logger.Info("User logged out successfully")
	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Logged out successfully",
	})
}

// ErrorResponse はエラーレスポンスを定義します
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse は成功レスポンスを定義します
type SuccessResponse struct {
	Message string `json:"message"`
} 