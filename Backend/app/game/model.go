package game

import "time"


type Game struct {
	id          int     `json:"id"`
	name        string  `json:"name"`
	total_players int     `json:"total_players"`
	current_players int     `json:"current_players"`
	revenue	 float64 `json:"revenue"`
	genre 	 string  `json:"genre"`
	region 	 string  `json:"region"`
	platform string  `json:"platform"`
	publisher string  `json:"publisher"`
	developer string  `json:"developer"`
	created_at time.Time `json:"created_at"`
	updated_at time.Time `json:"updated_at"`
}