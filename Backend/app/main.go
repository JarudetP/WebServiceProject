package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"gamedata/db"
	pkghandler "gamedata/pkg"
	userhandler "gamedata/user"
	gamehandler "gamedata/game"
	"gamedata/middleware"
	"github.com/gin-contrib/cors"
)

func main() {
	// Load .env — try current dir first, then parent dir
	if err := godotenv.Load(".env"); err != nil {
		if err2 := godotenv.Load("../.env"); err2 != nil {
			log.Println("No .env file found, using system environment")
		}
	}

	// Connect to DB
	db.Connect()
	defer db.DB.Close()

	// Wire up user module
	userRepo := userhandler.NewRepository(db.DB)
	userSvc := userhandler.NewService(userRepo)
	userH := userhandler.NewHandler(userSvc)

	// Wire up package module (passes userRepo for balance deduction)
	pkgRepo := pkghandler.NewRepository(db.DB)
	pkgSvc := pkghandler.NewService(pkgRepo, userRepo)
	pkgH := pkghandler.NewHandler(pkgSvc)

	// Initialize middleware
	mw := middleware.NewMiddleware(userRepo, pkgRepo)

	// Start player simulator background worker
	go startPlayerSimulator(db.DB)

	// Setup Gin router
	r := gin.Default()

	// Official CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
	}))

	// Fallback to explicitly catch OPTIONS requests if Not Found
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization, X-API-Key, Accept")
			c.AbortWithStatus(204)
			return
		}
		c.JSON(404, gin.H{"error": "Not Found"})
	})


	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Serve uploaded files
	r.Static("/uploads", "./uploads")

	// User routes
	users := r.Group("/api/users")
	{
		users.POST("/register", userH.Register)
		users.POST("/login", userH.Login)
		users.POST("/refresh", userH.Refresh)
		
		// Protected user routes
		protectedUsers := users.Group("")
		protectedUsers.Use(mw.RequireJWT(), mw.RequireSelf())
		{
			protectedUsers.GET("/:id", userH.GetProfile)
			protectedUsers.POST("/:id/topup", userH.TopUp)
			protectedUsers.POST("/:id/keys", userH.GenerateAPIKey)
			protectedUsers.GET("/:id/keys", userH.ListAPIKeys)
			protectedUsers.DELETE("/:id/keys/:key", userH.DeleteAPIKey)
		}
	}

	// Package routes
	packages := r.Group("/api/packages")
	{
		packages.GET("", pkgH.ListPackages)
		packages.GET("/:id", pkgH.GetPackage)
		packages.POST("/purchase", pkgH.Purchase)
		packages.GET("/subscription", pkgH.GetActiveSubscription)
	}

	//Game routes (Protected by Auth and RateLimit)
	games := r.Group("/api/games")
	games.Use(mw.Auth(), mw.RateLimit())
	{
		games.GET("", gamehandler.ListGames(db.DB))
		games.GET("/:id", gamehandler.GetGame(db.DB))
		
		// Admin only routes
		adminGames := games.Group("")
		adminGames.Use(mw.Admin())
		{
			adminGames.POST("", gamehandler.CreateGame(db.DB))
			adminGames.PUT("/:id", gamehandler.UpdateGame(db.DB))
			adminGames.DELETE("/:id", gamehandler.DeleteGame(db.DB))
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func startPlayerSimulator(db *sql.DB) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Update EACH game with a DIFFERENT random number between 100,000 and 250,000 using SQL
		result, err := db.Exec("UPDATE games SET current_players = floor(random() * 150001 + 100000)")
		if err != nil {
			log.Printf("Simulator error: %v", err)
			continue
		}
		rows, _ := result.RowsAffected()
		log.Printf("Simulator: Updated %d games with unique player counts", rows)
	}
}
