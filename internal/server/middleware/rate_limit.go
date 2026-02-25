package middleware

import (
	"net/http"
	"sync"
	"time"
	
	"github.com/gin-gonic/gin"
	
	"github.com/looplj/axonhub/internal/log"
)

// RateLimiter stores rate limit info per IP
type RateLimiter struct {
	sync.RWMutex
	visitors map[string]*visitor
}

type visitor struct {
	lastSeen time.Time
	count    int
	blocked  bool
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BlockDuration     time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
	}
	
	// Cleanup goroutine
	go rl.cleanup(config.BlockDuration)
	
	return rl
}

// RateLimit middleware for basic rate limiting
func (rl *RateLimiter) RateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		rl.Lock()
		v, exists := rl.visitors[ip]
		if !exists {
			v = &visitor{
				lastSeen: time.Now(),
				count:    0,
				blocked:  false,
			}
			rl.visitors[ip] = v
		}
		
		// Check if blocked
		if v.blocked {
			rl.Unlock()
			log.Warn(c.Request.Context(), "Blocked request from IP", log.String("ip", ip))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		
		// Check rate limit
		if time.Since(v.lastSeen) > time.Minute {
			v.count = 0
			v.lastSeen = time.Now()
		}
		
		v.count++
		
		// Block if exceeded
		if v.count > config.RequestsPerMinute {
			v.blocked = true
			rl.Unlock()
			log.Warn(c.Request.Context(), "Rate limit exceeded", log.String("ip", ip), log.Int("count", v.count))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}
		
		rl.Unlock()
		c.Next()
	}
}

// cleanup removes old entries
func (rl *RateLimiter) cleanup(blockDuration time.Duration) {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > blockDuration {
				delete(rl.visitors, ip)
			}
		}
		rl.Unlock()
	}
}

// WithRateLimit creates a rate limiting middleware
func WithRateLimit(requestsPerMinute int) gin.HandlerFunc {
	config := RateLimitConfig{
		RequestsPerMinute: requestsPerMinute,
		BlockDuration:     time.Minute * 5,
	}
	limiter := NewRateLimiter(config)
	return limiter.RateLimit(config)
}

// WithAPIRateLimit creates API-specific rate limiting
func WithAPIRateLimit() gin.HandlerFunc {
	// 60 requests per minute for API calls
	return WithRateLimit(60)
}

// WithAuthRateLimit creates auth-specific rate limiting (stricter)
func WithAuthRateLimit() gin.HandlerFunc {
	// 10 requests per minute for auth endpoints
	return WithRateLimit(10)
}
