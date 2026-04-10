package game

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	_ "golang.org/x/image/webp"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	maxImageSize = 5 * 1024 * 1024 // 5 MB
	uploadDir    = "./uploads/games"
)

// generateFilename creates a random filename
func generateFilename(ext string) string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b) + ext
}

// saveAndCompressImage decodes an uploaded image, compresses as JPEG to ≤ 5 MB, and saves it.
// Returns the URL path (e.g. "/uploads/games/<name>.jpg") or an error.
func saveAndCompressImage(fileBytes []byte) (string, error) {
	// 1. Check MIME type
	contentType := http.DetectContentType(fileBytes)
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	if !allowedTypes[contentType] {
		return "", fmt.Errorf("unsupported file type: %s. Only jpeg, png, and webp are allowed", contentType)
	}

	// 2. Decode the uploaded image
	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload dir: %w", err)
	}

	filename := generateFilename(".jpg")
	filePath := filepath.Join(uploadDir, filename)

	// Try encoding with decreasing quality until ≤ 5 MB
	quality := 90
	for quality >= 10 {
		var buf bytes.Buffer
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return "", fmt.Errorf("failed to encode image: %w", err)
		}

		if buf.Len() <= maxImageSize {
			if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
				return "", fmt.Errorf("failed to save file: %w", err)
			}
			return "/uploads/games/" + filename, nil
		}

		quality -= 10
	}

	// Last resort: save with lowest quality
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 5}); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	return "/uploads/games/" + filename, nil
}

func ListGames(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var games []Game

		rows, err := db.Query("SELECT id, name, total_players, current_players, revenue, genre, region, platform, publisher, developer, image_url, timestamp, created_at, updated_at FROM games")
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
				&game.ImageURL,
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

		err := db.QueryRow("SELECT id, name, total_players, current_players, revenue, genre, region, platform, publisher, developer, image_url, timestamp, created_at, updated_at FROM games WHERE id = $1", id).Scan(
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
			&game.ImageURL,
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
		// Parse form fields
		name := c.PostForm("name")
		totalPlayers, _ := strconv.Atoi(c.PostForm("total_players"))
		currentPlayers, _ := strconv.Atoi(c.PostForm("current_players"))
		revenue, _ := strconv.ParseFloat(c.PostForm("revenue"), 64)
		genre := c.PostForm("genre")
		region := c.PostForm("region")
		platform := c.PostForm("platform")
		publisher := c.PostForm("publisher")
		developer := c.PostForm("developer")

		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		// Handle image upload (optional)
		var imageURL string
		file, err := c.FormFile("image")
		if err == nil && file != nil {
			f, err := file.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
				return
			}
			defer f.Close()

			fileBytes := make([]byte, file.Size)
			if _, err := f.Read(fileBytes); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
				return
			}

			imageURL, err = saveAndCompressImage(fileBytes)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Image processing failed: %v", err)})
				return
			}
		}

		var game Game
		err = db.QueryRow(
			"INSERT INTO games (name, total_players, current_players, revenue, genre, region, platform, publisher, developer, image_url) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id",
			name, totalPlayers, currentPlayers, revenue, genre, region, platform, publisher, developer, imageURL,
		).Scan(&game.ID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create game: %v", err)})
			return
		}

		game.Name = name
		game.TotalPlayers = totalPlayers
		game.CurrentPlayers = currentPlayers
		game.Revenue = revenue
		game.Genre = genre
		game.Region = region
		game.Platform = platform
		game.Publisher = publisher
		game.Developer = developer
		game.ImageURL = imageURL

		c.JSON(http.StatusCreated, game)
	}
}

func UpdateGame(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Parse form fields
		name := c.PostForm("name")
		totalPlayers, _ := strconv.Atoi(c.PostForm("total_players"))
		currentPlayers, _ := strconv.Atoi(c.PostForm("current_players"))
		revenue, _ := strconv.ParseFloat(c.PostForm("revenue"), 64)
		genre := c.PostForm("genre")
		region := c.PostForm("region")
		platform := c.PostForm("platform")
		publisher := c.PostForm("publisher")
		developer := c.PostForm("developer")

		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		// Handle optional image upload
		var imageURL string
		file, err := c.FormFile("image")
		if err == nil && file != nil {
			f, err := file.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
				return
			}
			defer f.Close()

			fileBytes := make([]byte, file.Size)
			if _, err := f.Read(fileBytes); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
				return
			}

			imageURL, err = saveAndCompressImage(fileBytes)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Image processing failed: %v", err)})
				return
			}
		}

		// Build query based on whether image was uploaded
		var result sql.Result
		if imageURL != "" {
			result, err = db.Exec(
				"UPDATE games SET name = $1, total_players = $2, current_players = $3, revenue = $4, genre = $5, region = $6, platform = $7, publisher = $8, developer = $9, image_url = $10 WHERE id = $11",
				name, totalPlayers, currentPlayers, revenue, genre, region, platform, publisher, developer, imageURL, id,
			)
		} else {
			result, err = db.Exec(
				"UPDATE games SET name = $1, total_players = $2, current_players = $3, revenue = $4, genre = $5, region = $6, platform = $7, publisher = $8, developer = $9 WHERE id = $10",
				name, totalPlayers, currentPlayers, revenue, genre, region, platform, publisher, developer, id,
			)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update game"})
			return
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

		c.JSON(http.StatusOK, gin.H{
			"message":   "Game updated successfully",
			"image_url": imageURL,
		})
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