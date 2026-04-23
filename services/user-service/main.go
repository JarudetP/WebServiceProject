package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/joho/godotenv"

	"user-service/db"
	"user-service/middleware"
	"user-service/user"
)

func main() {
	// Attempt to load .env for local Go development (ignores errors if it doesn't exist, like in Docker)
	_ = godotenv.Load("../../.env")

	// Connect to DB
	db.Connect()
	defer db.DB.Close()

	// Wire up user module
	userRepo := user.NewRepository(db.DB)
	userSvc := user.NewService(userRepo)
	userH := user.NewHandler(userSvc)

	// Initialize middleware
	mw := middleware.NewMiddleware(userRepo)

	// Setup Gin router
	r := gin.Default()

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key", "Accept"},
		ExposeHeaders:   []string{"Content-Length"},
		MaxAge:          12 * time.Hour,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-service"})
	})

	// User routes (Public)
	api := r.Group("/api/users")
	{
		api.POST("/register", userH.Register)
		api.POST("/login", userH.Login)
		api.POST("/refresh", userH.Refresh)

		// Protected user routes
		protected := api.Group("")
		protected.Use(mw.RequireJWT(), mw.RequireSelf())
		{
			protected.GET("/:id", userH.GetProfile)
			protected.POST("/:id/topup", userH.TopUp)
			protected.POST("/:id/keys", userH.GenerateAPIKey)
			protected.GET("/:id/keys", userH.ListAPIKeys)
			protected.DELETE("/:id/keys/:key", userH.DeleteAPIKey)
			protected.GET("/:id/stats", userH.GetUsageStats)
		}
	}

	// Internal routes (for other microservices, no JWT required usually)
	internal := r.Group("/internal")
	{
		internal.POST("/users/:id/deduct", userH.InternalDeductBalance)
		internal.GET("/keys/:key/validate", userH.InternalValidateKey)
		internal.GET("/usage/count", userH.InternalGetUsageCount)
		internal.POST("/usage/log", userH.InternalLogUsage)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("User Service running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
