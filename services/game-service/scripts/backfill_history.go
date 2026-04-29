package main

import (
	"database/sql"
	"log"
	"math"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Connection string for local DB (Port 5437)
	connStr := "postgres://postgres:postgres@localhost:5437/game_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Starting backfill of history data...")

	rows, err := db.Query("SELECT id, total_players FROM games")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for rows.Next() {
		var id, totalPlayers int
		rows.Scan(&id, &totalPlayers)

		log.Printf("Backfilling for game ID %d...", id)
		
		// Clear existing history to avoid mess
		_, _ = db.Exec("DELETE FROM game_player_history WHERE game_id = $1", id)

		// Generate 7 days of data (every 30 mins)
		for d := 7; d >= 0; d-- {
			for h := 0; h < 24; h++ {
				for m := 0; m < 60; m += 30 {
					recordedAt := time.Now().AddDate(0, 0, -d).Add(time.Duration(h-time.Now().Hour()) * time.Hour).Add(time.Duration(m-time.Now().Minute()) * time.Minute)
					
					if recordedAt.After(time.Now()) {
						continue
					}

					hour := float64(recordedAt.Hour()) + float64(recordedAt.Minute())/60.0
					

					timeFactor := (hour - 15.0) * (2.0 * math.Pi / 24.0)
					wave1 := math.Sin(timeFactor)
					

					wave2 := 0.2 * math.Sin(hour * (2.0 * math.Pi / 6.0))
					

					wave3 := 0.1 * math.Sin(hour * (2.0 * math.Pi / 2.0))
					
					combinedWave := (wave1 + wave2 + wave3 + 1.3) / 2.6
					
					basePlayers := float64(totalPlayers) * 0.015
					jitter := 0.9 + rng.Float64()*0.2
					players := int(basePlayers * combinedWave * jitter)
					if players < 100 {
						players = 100 + rng.Intn(200)
					}

					_, err := db.Exec("INSERT INTO game_player_history (game_id, total_players, current_players, recorded_at) VALUES ($1, $2, $3, $4)", 
						id, totalPlayers, players, recordedAt)
					if err != nil {
						log.Printf("Error inserting history: %v", err)
					}
				}
			}
		}
	}

	log.Println("Backfill complete!")
}
