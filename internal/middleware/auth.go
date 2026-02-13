package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kloset/backend/internal/utils"
)

// AuthMiddleware validates JWT token from Authorization header
func AuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Authorization header required", nil)
		c.Abort()
		return
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Invalid Authorization header format", nil)
		c.Abort()
		return
	}

	token := parts[1]
	claims, err := utils.VerifyToken(token)
	if err != nil {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Invalid or expired token", err)
		c.Abort()
		return
	}

	// Store user info in context
	c.Set("userID", claims.UserID)
	c.Set("email", claims.Email)
	c.Next()
}

// OptionalAuthMiddleware extracts user info if token is provided, but doesn't require it
func OptionalAuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			claims, err := utils.VerifyToken(token)
			if err == nil {
				c.Set("userID", claims.UserID)
				c.Set("email", claims.Email)
				c.Set("authenticated", true)
			}
		}
	}
	c.Next()
}

// AdminMiddleware checks if user is admin (can be extended based on user model)
func AdminMiddleware(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.HTTPErrorHandler(c, http.StatusUnauthorized, "Admin access required", nil)
		c.Abort()
		return
	}

	// TODO: Check if user is admin from database
	// For now, admin check is simplified - in production, check user.IsAdmin or user.Role
	c.Set("userID", userID)
	c.Next()
}
