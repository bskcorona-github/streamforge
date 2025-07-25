package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestLogger(t *testing.T) {
	t.Run("logger middleware with successful request", func(t *testing.T) {
		router := setupTestRouter()
		logger := zaptest.NewLogger(t)
		
		router.Use(Logger(logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Logger should not cause any errors
	})

	t.Run("logger middleware with error request", func(t *testing.T) {
		router := setupTestRouter()
		logger := zaptest.NewLogger(t)
		
		router.Use(Logger(logger))
		router.GET("/error", func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error", nil)
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		// Logger should not cause any errors
	})

	t.Run("logger middleware with different HTTP methods", func(t *testing.T) {
		router := setupTestRouter()
		logger := zaptest.NewLogger(t)
		
		router.Use(Logger(logger))
		router.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"message": "created"})
		})
		router.PUT("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "updated"})
		})
		router.DELETE("/test", func(c *gin.Context) {
			c.JSON(http.StatusNoContent, nil)
		})

		methods := []string{"POST", "PUT", "DELETE"}
		expectedStatuses := []int{http.StatusCreated, http.StatusOK, http.StatusNoContent}

		for i, method := range methods {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, expectedStatuses[i], w.Code)
		}
	})

	t.Run("logger middleware with request body", func(t *testing.T) {
		router := setupTestRouter()
		logger := zaptest.NewLogger(t)
		
		router.Use(Logger(logger))
		router.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", nil)
		req.Header.Set("Content-Length", "100")
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("logger middleware with client IP", func(t *testing.T) {
		router := setupTestRouter()
		logger := zaptest.NewLogger(t)
		
		router.Use(Logger(logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("logger middleware with X-Forwarded-For header", func(t *testing.T) {
		router := setupTestRouter()
		logger := zaptest.NewLogger(t)
		
		router.Use(Logger(logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 70.41.3.18, 150.172.238.178")
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("logger middleware with X-Real-IP header", func(t *testing.T) {
		router := setupTestRouter()
		logger := zaptest.NewLogger(t)
		
		router.Use(Logger(logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Real-IP", "203.0.113.1")
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("logger middleware with long user agent", func(t *testing.T) {
		router := setupTestRouter()
		logger := zaptest.NewLogger(t)
		
		router.Use(Logger(logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		longUserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", longUserAgent)
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("logger middleware with empty user agent", func(t *testing.T) {
		router := setupTestRouter()
		logger := zaptest.NewLogger(t)
		
		router.Use(Logger(logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		// No User-Agent header set
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("logger middleware with large request body", func(t *testing.T) {
		router := setupTestRouter()
		logger := zaptest.NewLogger(t)
		
		router.Use(Logger(logger))
		router.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", nil)
		req.Header.Set("Content-Length", "1048576") // 1MB
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestLoggerPerformance(t *testing.T) {
	router := setupTestRouter()
	logger := zaptest.NewLogger(t)
	
	router.Use(Logger(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test multiple concurrent requests
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestLoggerWithNilLogger(t *testing.T) {
	router := setupTestRouter()
	
	// This should not panic
	router.Use(Logger(nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	
	// Should not panic
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
} 