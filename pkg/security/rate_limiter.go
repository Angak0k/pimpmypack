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
	limiters      sync.Map // map[string]*rate.Limiter
	rate          rate.Limit
	burst         int
	windowMinutes int
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(requestsPerWindow, burstSize, windowMinutes int) *IPRateLimiter {
	window := time.Duration(windowMinutes) * time.Minute
	r := rate.Every(window / time.Duration(requestsPerWindow))
	return &IPRateLimiter{
		rate:          r,
		burst:         burstSize,
		windowMinutes: windowMinutes,
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

// RetryAfterSeconds returns the retry_after value in seconds based on the configured window
func (rl *IPRateLimiter) RetryAfterSeconds() int {
	return rl.windowMinutes * 60
}

// NewEndpointRateLimiter creates a rate limiting middleware for any endpoint
func NewEndpointRateLimiter(endpoint string, requestsPerWindow, burstSize, windowMinutes int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(requestsPerWindow, burstSize, windowMinutes)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.GetLimiter(clientIP).Allow() {
			AuditRateLimitExceeded(c, endpoint)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": limiter.RetryAfterSeconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
