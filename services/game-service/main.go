package main

import (
	"database/sql"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"game-service/db"
	"game-service/game"
	"game-service/middleware"
)

func main() {
	// Load root .env
	_ = godotenv.Load("../../.env")

	// Connect to DB
	db.Connect()
	defer db.DB.Close()

	userSvcURL := os.Getenv("USER_SERVICE_URL")
	packageSvcURL := os.Getenv("PACKAGE_SERVICE_URL")
	if userSvcURL == "" {
		userSvcURL = "http://localhost:8081"
	}
	if packageSvcURL == "" {
		packageSvcURL = "http://localhost:8082"
	}

	// Wire up game module
	gameH := game.NewHandler(db.DB)

	// Initialize middleware
	mw := middleware.NewMiddleware(userSvcURL, packageSvcURL)

	// Start player simulator in background
	backfillHistoryIfEmpty(db.DB)
	go startPlayerSimulator(db.DB)

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
		c.JSON(200, gin.H{"status": "ok", "service": "game-service"})
	})

	// Static files (for game images)
	r.Static("/uploads", "./uploads")

	// Game routes
	api := r.Group("/api/games")
	{
		// Public API (Rate Limited by API Key)
		api.GET("", mw.AuthAPIKey(), gameH.ListGames)
		api.GET("/:id", mw.AuthAPIKey(), gameH.GetGame)
		api.GET("/:id/history", mw.AuthAPIKey(), gameH.GetGameHistory)

		// Feature-gated endpoints (API Key + package feature check)
		api.GET("/export", mw.AuthAPIKey(), mw.RequireFeature("has_bulk_export"), gameH.BulkExport)
		api.GET("/stream", mw.AuthAPIKey(), mw.RequireFeature("has_realtime_stream"), gameH.RealtimeStream)
		api.GET("/analytics/genre", mw.AuthAPIKey(), mw.RequireFeature("has_genre_analytics"), gameH.GenreAnalytics)
		api.GET("/analytics/revenue", mw.AuthAPIKey(), mw.RequireFeature("has_revenue_analytics"), gameH.RevenueAnalytics)
		api.GET("/analytics/region", mw.AuthAPIKey(), mw.RequireFeature("has_region_breakdown"), gameH.RegionBreakdown)

		// Admin API (Protected by JWT)
		admin := api.Group("")
		admin.Use(mw.RequireJWT(), mw.Admin())
		{
			admin.POST("", gameH.CreateGame)
			admin.PUT("/:id", gameH.UpdateGame)
			admin.DELETE("/:id", gameH.DeleteGame)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	log.Printf("Game Service running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func calculatePlayers(totalPlayers int, when time.Time, rng *rand.Rand) int {
	hour := float64(when.Hour()) + float64(when.Minute())/60.0
	// Base 24h wave (Peak at 9 PM)
	timeFactor := (hour - 15.0) * (2.0 * math.Pi / 24.0)
	wave1 := math.Sin(timeFactor)
	
	// Higher frequency wave for "messiness" (6h cycle)
	wave2 := 0.2 * math.Sin(hour * (2.0 * math.Pi / 6.0))
	
	// Even higher frequency (2h cycle)
	wave3 := 0.1 * math.Sin(hour * (2.0 * math.Pi / 2.0))
	
	combinedWave := (wave1 + wave2 + wave3 + 1.3) / 2.6 // Scale to ~0.0 to 1.0
	
	jitter := 0.95 + rng.Float64()*0.1 // ±5% noise
	
	players := int((30000.0 + (30000.0 * combinedWave)) * jitter)

	// Enforce strict boundaries bounds
	if players < 30000 {
		players = 30000 + rng.Intn(2000)
	} else if players > 60000 {
		players = 60000 - rng.Intn(2000)
	}
	return players
}

func backfillHistoryIfEmpty(dbConn *sql.DB) {
	var count int
	err := dbConn.QueryRow("SELECT COUNT(*) FROM game_player_history").Scan(&count)
	if err != nil {
		log.Printf("Backfill check error: %v", err)
		return
	}
	if count > 50 {
		log.Println("History already exists, skipping backfill.")
		return
	}

	log.Println("Backfilling player history with wavy data...")
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	rows, err := dbConn.Query("SELECT id, total_players FROM games")
	if err != nil {
		log.Printf("Backfill games query error: %v", err)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var id, totalPlayers int
		rows.Scan(&id, &totalPlayers)
		
		for d := 7; d >= 0; d-- {
			for h := 0; h < 24; h++ {
				for m := 0; m < 60; m += 30 {
					recordedAt := time.Now().AddDate(0, 0, -d).Add(time.Duration(h-time.Now().Hour()) * time.Hour).Add(time.Duration(m-time.Now().Minute()) * time.Minute)
					
					if recordedAt.After(time.Now()) {
						continue
					}
					
					players := calculatePlayers(totalPlayers, recordedAt, rng)
					dbConn.Exec("INSERT INTO game_player_history (game_id, total_players, current_players, recorded_at) VALUES ($1, $2, $3, $4)", id, totalPlayers, players, recordedAt)
				}
			}
		}
	}
	log.Println("Backfill complete.")
}

func startPlayerSimulator(dbConn *sql.DB) {
	log.Println("Starting Player Simulator with Enhanced Wavy Cycle...")
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	simulate := func() {
		rows, err := dbConn.Query("SELECT id, total_players FROM games")
		if err != nil {
			log.Printf("Simulator error: %v", err)
			return
		}
		defer rows.Close()

		now := time.Now()
		for rows.Next() {
			var id, totalPlayers int
			rows.Scan(&id, &totalPlayers)
			
			players := calculatePlayers(totalPlayers, now, rng)

			_, err = dbConn.Exec("UPDATE games SET current_players = $1 WHERE id = $2", players, id)
			if err != nil {
				log.Printf("Simulator update error for game %d: %v", id, err)
				continue
			}
			
			_, err = dbConn.Exec("INSERT INTO game_player_history (game_id, total_players, current_players, recorded_at) VALUES ($1, $2, $3, $4)", id, totalPlayers, players, now)
			if err != nil {
				log.Printf("Simulator history error for game %d: %v", id, err)
			}
		}
	}

	simulate()

	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		simulate()
	}
}
