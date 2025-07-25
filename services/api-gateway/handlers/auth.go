package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"streamforge/services/api-gateway/auth"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func Login(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	jwtManager := c.MustGet("jwt_manager").(*auth.JWTManager)

	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		logger.Error("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// TODO: Validate user credentials against database
	// For now, use mock validation
	if !validateUserCredentials(loginReq.Username, loginReq.Password) {
		logger.Warn("Invalid login attempt", zap.String("username", loginReq.Username))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	// Generate token pair
	tokenPair, err := jwtManager.GenerateTokenPair(
		"user-123", // TODO: Get actual user ID from database
		loginReq.Username,
		"user", // TODO: Get actual role from database
	)
	if err != nil {
		logger.Error("Failed to generate token pair", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authentication tokens",
		})
		return
	}

	logger.Info("User logged in successfully", zap.String("username", loginReq.Username))
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"tokens":  tokenPair,
	})
}

func Register(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	jwtManager := c.MustGet("jwt_manager").(*auth.JWTManager)

	var registerReq RegisterRequest
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		logger.Error("Invalid register request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Set default role if not provided
	if registerReq.Role == "" {
		registerReq.Role = "user"
	}

	// TODO: Check if user already exists in database
	// TODO: Hash password and store user in database

	// Generate token pair for new user
	tokenPair, err := jwtManager.GenerateTokenPair(
		"user-456", // TODO: Get actual user ID from database
		registerReq.Username,
		registerReq.Role,
	)
	if err != nil {
		logger.Error("Failed to generate token pair", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate authentication tokens",
		})
		return
	}

	logger.Info("User registered successfully", zap.String("username", registerReq.Username))
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"tokens":  tokenPair,
	})
}

func RefreshToken(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	jwtManager := c.MustGet("jwt_manager").(*auth.JWTManager)

	var refreshReq RefreshRequest
	if err := c.ShouldBindJSON(&refreshReq); err != nil {
		logger.Error("Invalid refresh request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Generate new token pair using refresh token
	tokenPair, err := jwtManager.RefreshToken(refreshReq.RefreshToken)
	if err != nil {
		logger.Error("Failed to refresh token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired refresh token",
		})
		return
	}

	logger.Info("Token refreshed successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"tokens":  tokenPair,
	})
}

func Logout(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	userID := c.MustGet("user_id").(string)

	// TODO: Invalidate refresh token in database
	// TODO: Add token to blacklist if needed

	logger.Info("User logged out", zap.String("user_id", userID))
	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

func GetProfile(c *gin.Context) {
	logger := c.MustGet("logger").(*zap.Logger)
	userID := c.MustGet("user_id").(string)
	username := c.MustGet("username").(string)
	role := c.MustGet("role").(string)

	// TODO: Get additional user information from database
	profile := gin.H{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		// Add more profile fields as needed
	}

	logger.Debug("Profile retrieved", zap.String("user_id", userID))
	c.JSON(http.StatusOK, gin.H{
		"profile": profile,
	})
}

// Mock function for user validation (replace with database lookup)
func validateUserCredentials(username, password string) bool {
	// TODO: Replace with actual database validation
	validUsers := map[string]string{
		"admin":  "$2a$10$example_hash", // bcrypt hash of "admin123"
		"user":   "$2a$10$example_hash", // bcrypt hash of "user123"
		"test":   "$2a$10$example_hash", // bcrypt hash of "test123"
	}

	storedHash, exists := validUsers[username]
	if !exists {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	return err == nil
} 