package game

import "time"

type Game struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	TotalPlayers   int       `json:"total_players"`
	CurrentPlayers int       `json:"current_players"`
	Revenue        float64   `json:"revenue"`
	Genre          string    `json:"genre"`
	Region         string    `json:"region"`
	Platform       string    `json:"platform"`
	Publisher      string    `json:"publisher"`
	Developer      string    `json:"developer"`
	ImageURL       string    `json:"image_url"`
	Timestamp      time.Time `json:"timestamp"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type GameHistory struct {
	GameID         int       `json:"game_id"`
	TotalPlayers   int       `json:"total_players"`
	CurrentPlayers int       `json:"current_players"`
	RecordedAt     time.Time `json:"recorded_at"`
}

type GenreAnalytic struct {
	Genre          string  `json:"genre"`
	GameCount      int     `json:"game_count"`
	TotalPlayers   int64   `json:"total_players"`
	CurrentPlayers int64   `json:"current_players"`
	TotalRevenue   float64 `json:"total_revenue"`
}

type RegionAnalytic struct {
	Region         string  `json:"region"`
	GameCount      int     `json:"game_count"`
	TotalPlayers   int64   `json:"total_players"`
	CurrentPlayers int64   `json:"current_players"`
	TotalRevenue   float64 `json:"total_revenue"`
}

type RevenueEntry struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Genre    string  `json:"genre"`
	Region   string  `json:"region"`
	Platform string  `json:"platform"`
	Revenue  float64 `json:"revenue"`
}
