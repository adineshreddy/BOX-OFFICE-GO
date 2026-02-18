package domain

import "time"

type Theater struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	City         string    `json:"city"`
	AddressLine1 string    `json:"addressLine1"`
	Timezone     string    `json:"timezone"`
	TotalScreens int       `json:"totalScreens"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type MovieShowtimeRecord struct {
	MovieID       string
	MovieTitle    string
	MovieDuration int
	TheaterID     string
	TheaterName   string
	City          string
	AddressLine1  string
	Timezone      string
	ShowtimeID    string
	ScreenName    string
	StartTime     time.Time
	Language      string
	Format        string
	BasePrice     float64
}

type ShowtimeItem struct {
	ShowtimeID string    `json:"showtimeId"`
	ScreenName string    `json:"screenName"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
	Language   string    `json:"language"`
	Format     string    `json:"format"`
	BasePrice  float64   `json:"basePrice"`
}

type TheaterSchedule struct {
	TheaterID    string         `json:"theaterId"`
	TheaterName  string         `json:"theaterName"`
	City         string         `json:"city"`
	AddressLine1 string         `json:"addressLine1"`
	Timezone     string         `json:"timezone"`
	Showtimes    []ShowtimeItem `json:"showtimes"`
}

type MovieTheaterListResponse struct {
	MovieID         string            `json:"movieId"`
	MovieTitle      string            `json:"movieTitle"`
	DurationMinutes int               `json:"durationMinutes"`
	Theaters        []TheaterSchedule `json:"theaters"`
}
