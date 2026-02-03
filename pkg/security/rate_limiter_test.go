package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRefreshRateLimiter_AllowsUnderLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Rate limiter: 10 requests per minute
	router.POST("/test", RefreshRateLimiter(10, 10), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 9 requests - all should succeed
	for i := 0; i < 9; i++ {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}
}

func TestRefreshRateLimiter_BlocksOverLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Rate limiter: 10 requests per minute
	router.POST("/test", RefreshRateLimiter(10, 10), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 10 requests - all should succeed
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 11th request should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "11th request should be rate limited")
	assert.Contains(t, w.Body.String(), "Rate limit exceeded")
	assert.Contains(t, w.Body.String(), "retry_after")
}

func TestRefreshRateLimiter_DifferentIPsIndependent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Rate limiter: 5 requests per minute
	router.POST("/test", RefreshRateLimiter(5, 5), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 5 requests from IP 1
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 6th request from IP 1 should be blocked
	req1 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusTooManyRequests, w1.Code, "6th request from IP 1 should be blocked")

	// Request from different IP should succeed
	req2 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req2.RemoteAddr = "192.168.1.2:5678"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code, "Request from different IP should succeed")
}

func TestRefreshRateLimiter_ResetAfterWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Rate limiter: 2 requests per 100ms for faster test
	router.POST("/test", RefreshRateLimiter(1200, 2), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 2 requests - should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 3rd request should be blocked
	req1 := httptest.NewRequest(http.MethodPost, "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusTooManyRequests, w1.Code, "3rd request should be blocked")

	// Wait for rate limiter to reset (slightly more than the window)
	time.Sleep(60 * time.Millisecond)

	// Request should now succeed
	req2 := httptest.NewRequest(http.MethodPost, "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code, "Request after window should succeed")
}

func TestIPRateLimiter_GetLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(10, 10)

	// Get limiter for IP 1
	limiter1 := limiter.GetLimiter("192.168.1.1")
	assert.NotNil(t, limiter1)

	// Get limiter for same IP should return same instance (same pointer)
	limiter1Again := limiter.GetLimiter("192.168.1.1")
	assert.Same(t, limiter1, limiter1Again, "Same IP should return same limiter instance")

	// Get limiter for different IP should return different instance (different pointer)
	limiter2 := limiter.GetLimiter("192.168.1.2")
	assert.NotNil(t, limiter2)
	assert.NotSame(t, limiter1, limiter2, "Different IPs should return different limiter instances")
}
