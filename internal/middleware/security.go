package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles CORS with credentials
func CORSMiddleware(c *gin.Context) {
	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000,http://localhost:3001"
	}

	origins := strings.Split(corsOrigin, ",")
	requestOrigin := c.GetHeader("Origin")

	// Check if origin is allowed
	for _, o := range origins {
		if strings.TrimSpace(o) == requestOrigin {
			c.Writer.Header().Set("Access-Control-Allow-Origin", requestOrigin)
			break
		}
	}

	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
		return
	}

	c.Next()
}

// SecurityHeadersMiddleware adds security headers using Helmet approach
func SecurityHeadersMiddleware(c *gin.Context) {
	c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
	c.Writer.Header().Set("X-Frame-Options", "DENY")
	c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
	c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'")
	c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	c.Next()
}

// ResponseTimeMiddleware adds response time header
func ResponseTimeMiddleware(c *gin.Context) {
	start := time.Now()
	c.Next()
	duration := time.Since(start)
	c.Writer.Header().Set("X-Response-Time", duration.String())
}

// RequestIDMiddleware generates unique request ID
func RequestIDMiddleware(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}
	c.Set("requestID", requestID)
	c.Writer.Header().Set("X-Request-ID", requestID)
	c.Next()
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + string(rune(time.Now().UnixNano()%10000))
}
