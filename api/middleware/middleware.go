package middleware

import (
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/pkg/config"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtService domain.JWTService
}

func NewAuthMiddleware(jwtService domain.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(config.GetStatusCode(config.ErrUnauthorized), gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			fmt.Printf("Empty token string after trimming Bearer prefix\n")
			c.JSON(config.GetStatusCode(config.ErrUnauthorized), gin.H{"error": "invalid token: empty token"})
			c.Abort()
			return
		}
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			fmt.Printf("Token validation failed: %v\n", err)
			c.JSON(config.GetStatusCode(config.ErrUnauthorized), gin.H{"error": fmt.Sprintf("invalid token: %v", err)})
			c.Abort()
			return
		}
		userID, ok := claims["id"].(float64)
		if !ok {
			fmt.Printf("Invalid user ID claim: %v\n", claims["id"])
			c.JSON(config.GetStatusCode(config.ErrUnauthorized), gin.H{"error": "invalid token claims: user ID missing or invalid"})
			c.Abort()
			return
		}
		role, ok := claims["role"].(string)
		if !ok {
			fmt.Printf("Invalid role claim: %v\n", claims["role"])
			c.JSON(config.GetStatusCode(config.ErrUnauthorized), gin.H{"error": "invalid token claims: role missing or invalid"})
			c.Abort()
			return
		}
		c.Set("user_id", uint(userID))
		c.Set("role", role)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(config.GetStatusCode(config.ErrForbidden), gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
