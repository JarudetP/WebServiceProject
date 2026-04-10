package middleware

import (
	"gamedata/pkg"
	"gamedata/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Middleware struct {
	userRepo *user.Repository
	pkgRepo  *pkg.Repository
}

func NewMiddleware(userRepo *user.Repository, pkgRepo *pkg.Repository) *Middleware {
	return &Middleware{
		userRepo: userRepo,
		pkgRepo:  pkgRepo,
	}
}

// Auth checks for X-API-Key header and validates it
func (m *Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required (X-API-Key header)"})
			c.Abort()
			return
		}

		userID, apiKeyID, role, err := m.userRepo.FindByAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or inactive API key"})
			c.Abort()
			return
		}

		// Store in context for subsequent handlers/middleware
		c.Set("userID", userID)
		c.Set("apiKeyID", apiKeyID)
		c.Set("role", role)
		
		// Debug log
		// fmt.Printf("DEBUG: Auth middleware set role: %s for userID: %d\n", role, userID)
		c.Next()
	}
}

// Admin checks if the user has the 'admin' role
func (m *Middleware) Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		role, _ := roleVal.(string)
		
		// fmt.Printf("DEBUG: Admin middleware checking role: '%s' (exists: %v)\n", role, exists)
		
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimit enforces the request limit based on user's purchased package
func (m *Middleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("userID")
		role := c.GetString("role")

		// Admins bypass rate limiting and subscription checks
		if role == "admin" {
			c.Next()
			return
		}

		apiKeyID := c.GetInt("apiKeyID")

		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User context missing"})
			c.Abort()
			return
		}

		// 1. Get user's active subscription
		sub, err := m.pkgRepo.GetActiveSubscription(userID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Active subscription required to access this API"})
			c.Abort()
			return
		}

		// 2. Get the package details to find the limit
		p, err := m.pkgRepo.FindPackageByID(sub.PackageID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch package details"})
			c.Abort()
			return
		}

		// 3. Check current usage in the refresh interval
		// Note: p.RequestLimit == -1 means unlimited
		if p.RequestLimit != -1 {
			usage, err := m.pkgRepo.GetUsageCountInInterval(userID, p.RefreshIntervalMinutes)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check usage logs"})
				c.Abort()
				return
			}

			if usage >= p.RequestLimit {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":         "Rate limit exceeded",
					"limit":         p.RequestLimit,
					"interval_min":  p.RefreshIntervalMinutes,
					"current_usage": usage,
				})
				c.Abort()
				return
			}
		}

		// 4. Log the request before proceeding (or after if you prefer, but here we log entry)
		// We'll record the actual status code in a real implementation by wrapping the writer,
		// but for now we log that the attempt was made and was within quota.
		_ = m.pkgRepo.LogAPIUsage(userID, apiKeyID, c.Request.URL.Path, c.Request.Method, 200)

		c.Next()
	}
}
