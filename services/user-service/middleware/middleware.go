package middleware

import (
	"user-service/user"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	userRepo *user.Repository
}

func NewMiddleware(userRepo *user.Repository) *Middleware {
	return &Middleware{
		userRepo: userRepo,
	}
}

// Admin checks if the user has the 'admin' role
func (m *Middleware) Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		role, _ := roleVal.(string)
		
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireJWT validates the Bearer token
func (m *Middleware) RequireJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		token, err := jwt.ParseWithClaims(tokenString, &user.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*user.CustomClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		c.Set("jwt_user_id", claims.UserID)
		c.Set("jwt_username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireSelf ensures the :id URL param matches the logged in user
func (m *Middleware) RequireSelf() gin.HandlerFunc {
	return func(c *gin.Context) {
		paramIDStr := c.Param("id")
		if paramIDStr == "" {
			c.Next()
			return
		}
		
		paramID, err := strconv.Atoi(paramIDStr)
		if err != nil {
			c.Next()
			return
		}

		jwtUserID := c.GetInt("jwt_user_id")
		if jwtUserID != 0 && jwtUserID != paramID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Cannot access other users' data"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}
