package security

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter manages rate limiters per IP address
type IPRateLimiter struct {
	limiters sync.Map // map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(requestsPerMinute, burstSize int) *IPRateLimiter {
	r := rate.Every(time.Minute / time.Duration(requestsPerMinute))
	return &IPRateLimiter{
		rate:  r,
		burst: burstSize,
	}
}

// GetLimiter returns the rate limiter for an IP, creating if needed
func (rl *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	limiter, exists := rl.limiters.Load(ip)
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters.Store(ip, limiter)
	}
	rateLimiter, ok := limiter.(*rate.Limiter)
	if !ok {
		// This should never happen as we always store *rate.Limiter
		rateLimiter = rate.NewLimiter(rl.rate, rl.burst)
	}
	return rateLimiter
}

// RefreshRateLimiter creates a rate limiting middleware for /auth/refresh endpoint
func RefreshRateLimiter(requestsPerMinute, burstSize int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(requestsPerMinute, burstSize)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.GetLimiter(clientIP).Allow() {
			AuditRateLimitExceeded(c, "/auth/refresh")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
