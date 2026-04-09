package game

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListGames(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var games []Game

		rows, err := db.Query("SELECT id, name, total_players, current_players, revenue, genre, region, platform, publisher, developer, timestamp, created_at, updated_at FROM games")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch games"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var game Game

			err := rows.Scan(
				&game.ID,
				&game.Name,
				&game.TotalPlayers,
				&game.CurrentPlayers,
				&game.Revenue,
				&game.Genre,
				&game.Region,
				&game.Platform,
				&game.Publisher,
				&game.Developer,
				&game.Timestamp,
				&game.CreatedAt,
				&game.UpdatedAt,
			)
			
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse game data"})
				return
			}
			games = append(games, game)
		}

		if err = rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating over games"})
			return
		}

		c.JSON(http.StatusOK, games)
	}
}
func GetGame(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var game Game

		err := db.QueryRow("SELECT id, name, total_players, current_players, revenue, genre, region, platform, publisher, developer, timestamp, created_at, updated_at FROM games WHERE id = $1", id).Scan(
			&game.ID,
			&game.Name,
			&game.TotalPlayers,
			&game.CurrentPlayers,
			&game.Revenue,
			&game.Genre,
			&game.Region,
			&game.Platform,
			&game.Publisher,
			&game.Developer,
			&game.Timestamp,
			&game.CreatedAt,
			&game.UpdatedAt,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch game"})
			return
		}

		c.JSON(http.StatusOK, game)
	}
}

func CreateGame(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var game Game
		if err := c.ShouldBindJSON(&game); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		err := db.QueryRow(
			"INSERT INTO games (name, total_players, current_players, revenue, genre, region, platform, publisher, developer) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
			game.Name,
			game.TotalPlayers,
			game.CurrentPlayers,
			game.Revenue,
			game.Genre,
			game.Region,
			game.Platform,
			game.Publisher,
			game.Developer,
		).Scan(&game.ID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game"})
			return
		}

		c.JSON(http.StatusCreated, game)
	}
}
func UpdateGame(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var game Game
		if err := c.ShouldBindJSON(&game); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		result, err := db.Exec(
			"UPDATE games SET name = $1, total_players = $2, current_players = $3, revenue = $4, genre = $5, region = $6, platform = $7, publisher = $8, developer = $9 WHERE id = $10",
			game.Name,
			game.TotalPlayers,
			game.CurrentPlayers,
			game.Revenue,
			game.Genre,
			game.Region,
			game.Platform,
			game.Publisher,
			game.Developer,
			id,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update game"})	
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update game"})
			return
		}

		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}

		c.JSON(http.StatusOK, game)
	}
}

func DeleteGame(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		result, err := db.Exec("DELETE FROM games WHERE id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete game"})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete game"})
			return
		}

		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Game deleted successfully"})
	}
}