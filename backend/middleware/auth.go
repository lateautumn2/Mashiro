package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/config"
	"backend/db"
	"backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var JwtKey = config.JWTSecret()

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

const (
	maxTimestampSkewSeconds = 300 // 5-minute anti-replay window
)

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

// AgentAuthRequired validates the X-Agent-Token header against the database,
// and verifies the HMAC request signature to prevent replay and tampering.
func AgentAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Agent-Token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent token required"})
			c.Abort()
			return
		}

		// Look up server by token in the database.
		var server models.Server
		if err := db.DB.Where("auth_token = ?", token).First(&server).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid server token"})
			c.Abort()
			return
		}

		// Verify HMAC request signature to prevent replay and tampering.
		if !verifyRequestSignature(c, []byte(token)) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid request signature"})
			c.Abort()
			return
		}

		c.Set("agent_token", token)
		c.Set("agent_server_id", server.ID)
		c.Next()
	}
}

// verifyRequestSignature validates the X-Timestamp and X-Signature headers.
// The signature is HMAC-SHA256 of (HTTP-method + URL-path + timestamp + body), keyed with the token.
// The timestamp must be within the allowed skew window to prevent replay attacks.
func verifyRequestSignature(c *gin.Context, token []byte) bool {
	sigHeader := c.GetHeader("X-Signature")
	tsHeader := c.GetHeader("X-Timestamp")

	if sigHeader == "" || tsHeader == "" {
		return false
	}

	ts, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		return false
	}

	now := time.Now().Unix()
	diff := now - ts
	if diff < 0 {
		diff = -diff
	}
	if diff > maxTimestampSkewSeconds {
		return false
	}

	// Read request body for HMAC computation.  Gin caches the body
	// so downstream handlers can still call c.ShouldBindJSON etc.
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		bodyBytes = []byte{}
	}
	// Restore the body for downstream consumption.
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	method := c.Request.Method
	path := c.Request.URL.Path

	mac := hmac.New(sha256.New, token)
	mac.Write([]byte(method))
	mac.Write([]byte(path))
	mac.Write([]byte(tsHeader))
	mac.Write(bodyBytes)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sigHeader), []byte(expected))
}
