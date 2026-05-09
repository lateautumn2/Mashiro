package middleware

import (
	"net/http"
	"strings"

	"backend/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var JwtKey = config.JWTSecret()

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

func AgentAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Agent-Token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent token required"})
			c.Abort()
			return
		}

		c.Set("agent_token", token)
		c.Next()
	}
}
