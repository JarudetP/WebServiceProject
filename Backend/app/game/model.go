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
	Timestamp      time.Time `json:"timestamp"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}