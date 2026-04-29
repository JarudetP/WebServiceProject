package game

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	_ "golang.org/x/image/webp"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxImageSize = 5 * 1024 * 1024 // 5 MB
	uploadDir    = "./uploads/games"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func generateFilename(ext string) string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b) + ext
}

// saveAndCompressImage decodes an uploaded image, compresses as JPEG to ≤ 5 MB, and saves it.
func saveAndCompressImage(fileBytes []byte) (string, error) {
	contentType := http.DetectContentType(fileBytes)
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	if !allowedTypes[contentType] {
		return "", fmt.Errorf("unsupported file type: %s", contentType)
	}

	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload dir: %w", err)
	}

	filename := generateFilename(".jpg")
	filePath := filepath.Join(uploadDir, filename)

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

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 5})
	os.WriteFile(filePath, buf.Bytes(), 0644)
	return "/uploads/games/" + filename, nil
}

// GET /api/games
func (h *Handler) ListGames(c *gin.Context) {
	var games []Game
	rows, err := h.db.Query("SELECT id, name, total_players, current_players, revenue, genre, region, platform, publisher, developer, image_url, timestamp, created_at, updated_at FROM games")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch games"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var g Game
		err := rows.Scan(&g.ID, &g.Name, &g.TotalPlayers, &g.CurrentPlayers, &g.Revenue, &g.Genre, &g.Region, &g.Platform, &g.Publisher, &g.Developer, &g.ImageURL, &g.Timestamp, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse game data"})
			return
		}
		games = append(games, g)
	}
	c.JSON(http.StatusOK, games)
}

// GET /api/games/:id
func (h *Handler) GetGame(c *gin.Context) {
	id := c.Param("id")
	var g Game
	err := h.db.QueryRow("SELECT id, name, total_players, current_players, revenue, genre, region, platform, publisher, developer, image_url, timestamp, created_at, updated_at FROM games WHERE id = $1", id).Scan(
		&g.ID, &g.Name, &g.TotalPlayers, &g.CurrentPlayers, &g.Revenue, &g.Genre, &g.Region, &g.Platform, &g.Publisher, &g.Developer, &g.ImageURL, &g.Timestamp, &g.CreatedAt, &g.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch game"})
		return
	}
	c.JSON(http.StatusOK, g)
}

// POST /api/games
func (h *Handler) CreateGame(c *gin.Context) {
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

	var imageURL string
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		f, _ := file.Open()
		defer f.Close()
		fileBytes := make([]byte, file.Size)
		f.Read(fileBytes)
		imageURL, _ = saveAndCompressImage(fileBytes)
	}

	var gameID int
	err = h.db.QueryRow(
		"INSERT INTO games (name, total_players, current_players, revenue, genre, region, platform, publisher, developer, image_url) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id",
		name, totalPlayers, currentPlayers, revenue, genre, region, platform, publisher, developer, imageURL,
	).Scan(&gameID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create game: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": gameID, "message": "Game created"})
}

// PUT /api/games/:id
func (h *Handler) UpdateGame(c *gin.Context) {
	id := c.Param("id")
	name := c.PostForm("name")
	totalPlayers, _ := strconv.Atoi(c.PostForm("total_players"))
	currentPlayers, _ := strconv.Atoi(c.PostForm("current_players"))
	revenue, _ := strconv.ParseFloat(c.PostForm("revenue"), 64)
	genre := c.PostForm("genre")
	region := c.PostForm("region")
	platform := c.PostForm("platform")
	publisher := c.PostForm("publisher")
	developer := c.PostForm("developer")

	var imageURL string
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		f, _ := file.Open()
		defer f.Close()
		fileBytes := make([]byte, file.Size)
		f.Read(fileBytes)
		imageURL, _ = saveAndCompressImage(fileBytes)
	}

	var result sql.Result
	if imageURL != "" {
		result, err = h.db.Exec(
			"UPDATE games SET name = $1, total_players = $2, current_players = $3, revenue = $4, genre = $5, region = $6, platform = $7, publisher = $8, developer = $9, image_url = $10 WHERE id = $11",
			name, totalPlayers, currentPlayers, revenue, genre, region, platform, publisher, developer, imageURL, id,
		)
	} else {
		result, err = h.db.Exec(
			"UPDATE games SET name = $1, total_players = $2, current_players = $3, revenue = $4, genre = $5, region = $6, platform = $7, publisher = $8, developer = $9 WHERE id = $10",
			name, totalPlayers, currentPlayers, revenue, genre, region, platform, publisher, developer, id,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update game"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Game updated successfully"})
}

// DELETE /api/games/:id
func (h *Handler) DeleteGame(c *gin.Context) {
	id := c.Param("id")
	result, err := h.db.Exec("DELETE FROM games WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete game"})
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Game deleted successfully"})
}

// GET /api/games/:id/history
func (h *Handler) GetGameHistory(c *gin.Context) {
	id := c.Param("id")

	days := 7
	if val, exists := c.Get("historical_data_days"); exists {
		if d, ok := val.(int); ok && d > 0 {
			days = d
		}
	}

	rows, err := h.db.Query(`
		SELECT game_id, total_players, current_players, recorded_at
		FROM game_player_history
		WHERE game_id = $1
		AND recorded_at >= NOW() - make_interval(days => $2)
		ORDER BY recorded_at ASC
	`, id, days)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}
	defer rows.Close()

	var history []GameHistory
	for rows.Next() {
		var hi GameHistory
		err := rows.Scan(&hi.GameID, &hi.TotalPlayers, &hi.CurrentPlayers, &hi.RecordedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse history"})
			return
		}
		history = append(history, hi)
	}
	c.JSON(http.StatusOK, history)
}

// GET /api/games/analytics/genre  (requires has_genre_analytics)
func (h *Handler) GenreAnalytics(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT genre, COUNT(*) AS game_count,
		       SUM(total_players) AS total_players,
		       SUM(current_players) AS current_players,
		       SUM(revenue) AS total_revenue
		FROM games
		GROUP BY genre
		ORDER BY total_players DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch genre analytics"})
		return
	}
	defer rows.Close()

	var results []GenreAnalytic
	for rows.Next() {
		var g GenreAnalytic
		if err := rows.Scan(&g.Genre, &g.GameCount, &g.TotalPlayers, &g.CurrentPlayers, &g.TotalRevenue); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse genre data"})
			return
		}
		results = append(results, g)
	}
	c.JSON(http.StatusOK, results)
}

// GET /api/games/analytics/revenue  (requires has_revenue_analytics)
func (h *Handler) RevenueAnalytics(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, genre, region, platform, revenue
		FROM games
		ORDER BY revenue DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revenue analytics"})
		return
	}
	defer rows.Close()

	var results []RevenueEntry
	for rows.Next() {
		var r RevenueEntry
		if err := rows.Scan(&r.ID, &r.Name, &r.Genre, &r.Region, &r.Platform, &r.Revenue); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse revenue data"})
			return
		}
		results = append(results, r)
	}
	c.JSON(http.StatusOK, results)
}

// GET /api/games/analytics/region  (requires has_region_breakdown)
func (h *Handler) RegionBreakdown(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT region, COUNT(*) AS game_count,
		       SUM(total_players) AS total_players,
		       SUM(current_players) AS current_players,
		       SUM(revenue) AS total_revenue
		FROM games
		GROUP BY region
		ORDER BY total_players DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch region breakdown"})
		return
	}
	defer rows.Close()

	var results []RegionAnalytic
	for rows.Next() {
		var r RegionAnalytic
		if err := rows.Scan(&r.Region, &r.GameCount, &r.TotalPlayers, &r.CurrentPlayers, &r.TotalRevenue); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse region data"})
			return
		}
		results = append(results, r)
	}
	c.JSON(http.StatusOK, results)
}

// GET /api/games/export  (requires has_bulk_export)
func (h *Handler) BulkExport(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, total_players, current_players, revenue, genre, region,
		       platform, publisher, developer, image_url, timestamp, created_at, updated_at
		FROM games
		ORDER BY id ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export games"})
		return
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		if err := rows.Scan(&g.ID, &g.Name, &g.TotalPlayers, &g.CurrentPlayers, &g.Revenue,
			&g.Genre, &g.Region, &g.Platform, &g.Publisher, &g.Developer,
			&g.ImageURL, &g.Timestamp, &g.CreatedAt, &g.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse game data"})
			return
		}
		games = append(games, g)
	}

	c.Header("Content-Disposition", "attachment; filename=games_export.json")
	c.JSON(http.StatusOK, games)
}

// GET /api/games/stream  (requires has_realtime_stream)
// Server-Sent Events — pushes live game data every 30 seconds.
func (h *Handler) RealtimeStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	clientGone := c.Request.Context().Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	sendEvent := func() bool {
		rows, err := h.db.Query(`
			SELECT id, name, total_players, current_players, revenue, genre, region,
			       platform, publisher, developer, image_url, timestamp, created_at, updated_at
			FROM games
		`)
		if err != nil {
			return false
		}
		defer rows.Close()

		var games []Game
		for rows.Next() {
			var g Game
			if err := rows.Scan(&g.ID, &g.Name, &g.TotalPlayers, &g.CurrentPlayers, &g.Revenue,
				&g.Genre, &g.Region, &g.Platform, &g.Publisher, &g.Developer,
				&g.ImageURL, &g.Timestamp, &g.CreatedAt, &g.UpdatedAt); err != nil {
				return false
			}
			games = append(games, g)
		}

		data, _ := json.Marshal(games)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		c.Writer.(http.Flusher).Flush()
		return true
	}

	if !sendEvent() {
		return
	}

	for {
		select {
		case <-clientGone:
			return
		case <-ticker.C:
			if !sendEvent() {
				return
			}
		}
	}
}
