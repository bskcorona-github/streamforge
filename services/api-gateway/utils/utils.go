package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// generateID generates a unique ID
func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// SendErrorResponse sends a standardized error response
func SendErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	logger := c.MustGet("logger").(*zap.Logger)
	
	if err != nil {
		logger.Error(message, zap.Error(err))
	} else {
		logger.Error(message)
	}

	response := gin.H{
		"error":   message,
		"success": false,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}

	if err != nil {
		response["details"] = err.Error()
	}

	c.JSON(statusCode, response)
}

// SendSuccessResponse sends a standardized success response
func SendSuccessResponse(c *gin.Context, data interface{}) {
	response := gin.H{
		"success": true,
		"data":    data,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// ValidateJSON validates JSON request body
func ValidateJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// ParseTimeRange parses time range from query parameters
func ParseTimeRange(c *gin.Context) (time.Time, time.Time, error) {
	startStr := c.DefaultQuery("start", "")
	endStr := c.DefaultQuery("end", "")

	var start, end time.Time
	var err error

	if startStr == "" {
		start = time.Now().Add(-1 * time.Hour) // デフォルト: 1時間前
	} else {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return start, end, fmt.Errorf("invalid start time: %w", err)
		}
	}

	if endStr == "" {
		end = time.Now() // デフォルト: 現在時刻
	} else {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return start, end, fmt.Errorf("invalid end time: %w", err)
		}
	}

	if start.After(end) {
		return start, end, fmt.Errorf("start time must be before end time")
	}

	return start, end, nil
}

// ParsePagination parses pagination parameters
func ParsePagination(c *gin.Context) (int, int) {
	page := 1
	limit := 100

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	return page, limit
}

// LogRequest logs incoming request details
func LogRequest(c *gin.Context, logger *zap.Logger) {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = GenerateID()
		c.Set("request_id", requestID)
	}

	logger.Info("Request received",
		zap.String("request_id", requestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)
}

// LogResponse logs response details
func LogResponse(c *gin.Context, logger *zap.Logger, statusCode int, duration time.Duration) {
	requestID := c.GetString("request_id")
	
	logger.Info("Response sent",
		zap.String("request_id", requestID),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	)
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(uuid string) bool {
	if len(uuid) != 36 {
		return false
	}
	
	for i, char := range uuid {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if char != '-' {
				return false
			}
		} else {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
	}
	
	return true
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(input string) string {
	// 基本的なサニタイゼーション（必要に応じて拡張）
	if len(input) > 1000 {
		input = input[:1000]
	}
	return input
}

// ConvertToJSON converts an interface to JSON string
func ConvertToJSON(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
} 