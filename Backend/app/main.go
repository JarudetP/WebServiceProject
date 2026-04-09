package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"gamedata/db"
	pkghandler "gamedata/pkg"
	userhandler "gamedata/user"
	gamehandler "gamedata/game"
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

	// Setup Gin router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// User routes
	users := r.Group("/api/users")
	{
		users.POST("/register", userH.Register)
		users.POST("/login", userH.Login)
		users.GET("/:id", userH.GetProfile)
		users.POST("/:id/topup", userH.TopUp)
		users.POST("/refresh" ,userH.Refresh)
	}

	// Package routes
	packages := r.Group("/api/packages")
	{
		packages.GET("", pkgH.ListPackages)
		packages.GET("/:id", pkgH.GetPackage)
		packages.POST("/purchase", pkgH.Purchase)
		packages.GET("/subscription", pkgH.GetActiveSubscription)
	}

	//Game routes
	games := r.Group("/api/games")
	{
		games.GET("", gamehandler.ListGames(db.DB))
		games.GET("/:id", gamehandler.GetGame(db.DB))
		games.POST("", gamehandler.CreateGame(db.DB))
		games.PUT("/:id", gamehandler.UpdateGame(db.DB))
		games.DELETE("/:id", gamehandler.DeleteGame(db.DB))
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
