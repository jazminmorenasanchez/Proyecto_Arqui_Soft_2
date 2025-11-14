package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Store claims in context for role checking
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Validate expiration (exp is numeric unix)
			expF, ok := claims["exp"].(float64)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token expiration"})
				c.Abort()
				return
			}

			if time.Now().Unix() > int64(expF) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
				c.Abort()
				return
			}

			// subject (sub) should be numeric user id
			if subF, ok := claims["sub"].(float64); ok {
				c.Set("userId", uint64(subF))
			}
			if r, ok := claims["rol"].(string); ok {
				c.Set("role", r)
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authentication"})
			c.Abort()
			return
		}

		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireUser ensures the requester has a non-admin user role (or admin can also be accepted depending on rules)
func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authentication"})
			c.Abort()
			return
		}

		rs, _ := role.(string)
		if rs != "user" && rs != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "user role required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
