package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Middleware struct {
	userSvcURL    string
	packageSvcURL string
}

func NewMiddleware(userSvcURL, packageSvcURL string) *Middleware {
	return &Middleware{
		userSvcURL:    userSvcURL,
		packageSvcURL: packageSvcURL,
	}
}

// RequireJWT validates the Bearer token (for Admin/Web users)
func (m *Middleware) RequireJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, _ := token.Claims.(*CustomClaims)
		c.Set("jwt_user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// Admin checks if the user has the 'admin' role
func (m *Middleware) Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, _ := c.Get("role")
		if roleVal != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuthAPIKey validates X-API-Key and enforces rate limits
func (m *Middleware) AuthAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API Key is required"})
			c.Abort()
			return
		}

		// 1. Validate API Key via User Service
		valURL := fmt.Sprintf("%s/internal/keys/%s/validate", m.userSvcURL, apiKey)
		resp, err := http.Get(valURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
			c.Abort()
			return
		}
		var keyData struct {
			UserID   int    `json:"user_id"`
			APIKeyID int    `json:"api_key_id"`
			Role     string `json:"role"`
		}
		json.NewDecoder(resp.Body).Decode(&keyData)
		resp.Body.Close()

		c.Set("user_id", keyData.UserID)
		c.Set("api_key_id", keyData.APIKeyID)

		// 2. Rate Limit Check
		// Need subscription -> package details -> current usage
		if err := m.checkRateLimit(c, keyData.UserID); err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Next()

		// 3. Log Usage after request completes
		m.logUsage(keyData.UserID, keyData.APIKeyID, c.Request.URL.Path, c.Request.Method, c.Writer.Status())
	}
}

func (m *Middleware) checkRateLimit(c *gin.Context, userID int) error {
	subURL := fmt.Sprintf("%s/api/packages/subscription?user_id=%d", m.packageSvcURL, userID)
	resp, err := http.Get(subURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to check subscription")
	}
	var sub map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sub)
	resp.Body.Close()

	if sub == nil {
		return fmt.Errorf("no active subscription found")
	}

	packageID := int(sub["package_id"].(float64))

	// b. Get Package Details
	pkgURL := fmt.Sprintf("%s/api/packages/%d", m.packageSvcURL, packageID)
	resp, err = http.Get(pkgURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get package details")
	}
	var pkg map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&pkg)
	resp.Body.Close()

	limit := int(pkg["request_limit"].(float64))
	interval := int(pkg["refresh_interval_minutes"].(float64))

	if limit == -1 {
		return nil
	}

	usageURL := fmt.Sprintf("%s/internal/usage/count?user_id=%d&minutes=%d", m.userSvcURL, userID, interval)
	resp, err = http.Get(usageURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to check usage count")
	}
	var usageData struct {
		Count int `json:"count"`
	}
	json.NewDecoder(resp.Body).Decode(&usageData)
	resp.Body.Close()

	if usageData.Count >= limit {
		return fmt.Errorf("rate limit exceeded (Limit: %d requests per %d minutes)", limit, interval)
	}

	return nil
}

func (m *Middleware) logUsage(userID, apiKeyID int, endpoint, method string, status int) {
	logURL := fmt.Sprintf("%s/internal/usage/log", m.userSvcURL)
	body, _ := json.Marshal(map[string]interface{}{
		"user_id":     userID,
		"api_key_id":  apiKeyID,
		"endpoint":    endpoint,
		"method":      method,
		"status_code": status,
	})
	http.Post(logURL, "application/json", bytes.NewBuffer(body))
}
