package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
	logger        *zap.Logger
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func NewJWTManager(secretKey string, tokenDuration time.Duration, logger *zap.Logger) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
		logger:        logger,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTManager) GenerateTokenPair(userID, username, role string) (*TokenPair, error) {
	now := time.Now()
	
	// Access token claims
	accessClaims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "streamforge-api-gateway",
			Subject:   userID,
		},
	}

	// Refresh token claims (longer expiration)
	refreshClaims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.tokenDuration * 7)), // 7 times longer
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "streamforge-api-gateway",
			Subject:   userID,
		},
	}

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secretKey)
	if err != nil {
		j.logger.Error("Failed to sign access token", zap.Error(err))
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.secretKey)
	if err != nil {
		j.logger.Error("Failed to sign refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	tokenPair := &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(j.tokenDuration.Seconds()),
	}

	j.logger.Info("Generated token pair", 
		zap.String("user_id", userID),
		zap.String("username", username),
		zap.String("role", role),
	)

	return tokenPair, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		j.logger.Error("Failed to parse token", zap.Error(err))
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		j.logger.Error("Invalid token claims")
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// RefreshToken generates a new token pair using a refresh token
func (j *JWTManager) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		j.logger.Error("Failed to validate refresh token", zap.Error(err))
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new token pair
	tokenPair, err := j.GenerateTokenPair(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		j.logger.Error("Failed to generate new token pair", zap.Error(err))
		return nil, fmt.Errorf("failed to generate new token pair: %w", err)
	}

	j.logger.Info("Refreshed token pair", 
		zap.String("user_id", claims.UserID),
		zap.String("username", claims.Username),
	)

	return tokenPair, nil
}

// ExtractUserFromToken extracts user information from token without validation
func (j *JWTManager) ExtractUserFromToken(tokenString string) (*Claims, error) {
	parser := jwt.Parser{SkipClaimsValidation: true}
	token, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
} 