package middleware

import (
	"go-template/internal/api/response"
	"go-template/pkg/auth"
	"go-template/pkg/errcode"
	"go-template/pkg/logger"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware validates JWT tokens and adds user information to the context
func JWTAuthMiddleware(config auth.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Err(c, errcode.UserUnauthorized, "Authorization header is required")
			c.Abort()
			return
		}

		// Extract the token from the header
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Err(c, errcode.AuthTokenInvalid, "Authorization format must be Bearer {token}")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := auth.ParseToken(tokenString, config.Secret)
		if err != nil {
			if err == auth.ErrExpiredToken {
				response.Err(c, errcode.AuthTokenExpired)
			} else {
				logger.Warnf("Invalid token: %v", err)
				response.Err(c, errcode.AuthTokenInvalid)
			}
			c.Abort()
			return
		}

		// Add claims to context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// RequireRole checks if the current user has the required role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by JWTAuthMiddleware)
		userRole, exists := c.Get("userRole")
		if !exists {
			response.Err(c, errcode.UserUnauthorized)
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		roleStr := userRole.(string)
		allowed := false
		for _, role := range roles {
			if role == roleStr {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Err(c, errcode.AuthAccessDenied, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}
