package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kloset/backend/internal/utils"
)

// RateLimiter stores rate limit info per IP
type RateLimiter struct {
	requests map[string]*RequestInfo
	mu       sync.RWMutex
	window   time.Duration
	maxReqs  int
}

// RequestInfo stores request count and timestamp
type RequestInfo struct {
	count     int
	resetTime time.Time
}

var rateLimiter *RateLimiter

// InitRateLimiter initializes the rate limiter
func InitRateLimiter() {
	rateLimiter = &RateLimiter{
		requests: make(map[string]*RequestInfo),
		window:   15 * time.Minute,
		maxReqs:  100,
	}

	// Start cleanup goroutine to remove old entries
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			rateLimiter.cleanup()
		}
	}()
}

// RateLimitMiddleware enforces rate limiting per IP
func RateLimitMiddleware(c *gin.Context) {
	if rateLimiter == nil {
		InitRateLimiter()
	}

	clientIP := getClientIP(c)
	if !rateLimiter.allow(clientIP) {
		utils.HTTPErrorHandler(c, http.StatusTooManyRequests, "Rate limit exceeded", nil)
		c.Abort()
		return
	}

	c.Next()
}

// allow checks if the IP is allowed to make a request
func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.requests[ip]

	if !exists || now.After(info.resetTime) {
		// New window
		rl.requests[ip] = &RequestInfo{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	// Same window
	if info.count < rl.maxReqs {
		info.count++
		return true
	}

	return false
}

// cleanup removes expired entries
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, info := range rl.requests {
		if now.After(info.resetTime) {
			delete(rl.requests, ip)
		}
	}
}

// getClientIP extracts client IP from request
func getClientIP(c *gin.Context) string {
	// Try X-Forwarded-For header first (for proxies)
	if xForwardedFor := c.GetHeader("X-Forwarded-For"); xForwardedFor != "" {
		return xForwardedFor
	}

	// Try X-Real-IP header
	if xRealIP := c.GetHeader("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	// Fall back to remote address
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ip
}
