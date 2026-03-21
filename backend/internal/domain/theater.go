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

type ShowDetails struct {
	ShowtimeID       string    `json:"showtimeId"`
	MovieID          string    `json:"movieId"`
	MovieTitle       string    `json:"movieTitle"`
	TheaterID        string    `json:"theaterId"`
	TheaterName      string    `json:"theaterName"`
	City             string    `json:"city"`
	AddressLine1     string    `json:"addressLine1"`
	ScreenName       string    `json:"screenName"`
	StartTime        time.Time `json:"startTime"`
	Language         string    `json:"language"`
	Format           string    `json:"format"`
	BasePrice        float64   `json:"basePrice"`
	DurationMinutes  int       `json:"durationMinutes"`
	AvailableSeats   int       `json:"availableSeats"`
	TotalSeats       int       `json:"totalSeats"`
	UnavailableSeats int       `json:"unavailableSeats"`
}

type SeatItem struct {
	SeatNumber  string  `json:"seatNumber"`
	RowLabel    string  `json:"rowLabel"`
	SeatIndex   int     `json:"seatIndex"`
	SeatType    string  `json:"seatType"`
	PriceFactor float64 `json:"priceFactor"`
	IsAvailable bool    `json:"isAvailable"`
	IsHeld      bool    `json:"isHeld"`
}

type SeatRow struct {
	RowLabel string     `json:"rowLabel"`
	Seats    []SeatItem `json:"seats"`
}

type SeatMapResponse struct {
	ShowtimeID       string    `json:"showtimeId"`
	MovieTitle       string    `json:"movieTitle"`
	TheaterName      string    `json:"theaterName"`
	ScreenName       string    `json:"screenName"`
	ShowTime         time.Time `json:"showTime"`
	TotalSeats       int       `json:"totalSeats"`
	AvailableSeats   int       `json:"availableSeats"`
	UnavailableSeats int       `json:"unavailableSeats"`
	Rows             []SeatRow `json:"rows"`
}

type SeatAvailabilityResponse struct {
	ShowtimeID       string    `json:"showtimeId"`
	MovieTitle       string    `json:"movieTitle"`
	TheaterName      string    `json:"theaterName"`
	ShowTime         time.Time `json:"showTime"`
	TotalSeats       int       `json:"totalSeats"`
	AvailableSeats   int       `json:"availableSeats"`
	UnavailableSeats int       `json:"unavailableSeats"`
	LastRefreshedAt  time.Time `json:"lastRefreshedAt"`
}
