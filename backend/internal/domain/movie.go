package domain

import "time"

type Movie struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Genre           string    `json:"genre"`
	Language        string    `json:"language"`
	DurationMinutes int       `json:"durationMinutes"`
	ReleaseDate     time.Time `json:"releaseDate"`
	Rating          float64   `json:"rating"`
	PosterURL       *string   `json:"posterUrl,omitempty"`
	IsActive        bool      `json:"isActive"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
