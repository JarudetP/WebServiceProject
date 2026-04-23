package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"package-service/db"
	"package-service/middleware"
	"package-service/pkg"
)

func main() {
	// Load root .env
	_ = godotenv.Load("../../.env")

	// Connect to DB
	db.Connect()
	defer db.DB.Close()

	userSvcURL := os.Getenv("USER_SERVICE_URL")
	if userSvcURL == "" {
		userSvcURL = "http://localhost:8081"
	}

	// Wire up package module
	pkgRepo := pkg.NewRepository(db.DB)
	pkgSvc := pkg.NewService(pkgRepo, userSvcURL)
	pkgH := pkg.NewHandler(pkgSvc)

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
		c.JSON(200, gin.H{"status": "ok", "service": "package-service"})
	})

	// Package routes
	api := r.Group("/api/packages")
	api.Use(middleware.RequireJWT())
	{
		api.GET("", pkgH.ListPackages)
		api.GET("/:id", pkgH.GetPackage)
		api.POST("/purchase", pkgH.Purchase)
		api.GET("/subscription", pkgH.GetActiveSubscription)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Package Service running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
