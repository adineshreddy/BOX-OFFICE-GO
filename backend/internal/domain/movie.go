package domain

import "time"

// Allowed sort fields for the movie list API.
const (
	MovieSortByReleaseDate = "release_date" // default
	MovieSortByTitle       = "title"
	MovieSortByRating      = "rating"
)

// ValidMovieSortFields is the allow-list used by validation.
var ValidMovieSortFields = map[string]bool{
	MovieSortByReleaseDate: true,
	MovieSortByTitle:       true,
	MovieSortByRating:      true,
}

// MovieListQuery carries validated, caller-supplied pagination/filter params.
type MovieListQuery struct {
	Page   int    // 1-based, default 1
	Limit  int    // rows per page, default 20, max 100
	SortBy string // one of the MovieSortBy* constants
	Title  string // ILIKE filter, optional
	Genre  string // ILIKE filter, optional
}

// MovieListResponse is the paginated envelope returned by GET /movies.
type MovieListResponse struct {
	Movies  []Movie `json:"movies"`
	Page    int     `json:"page"`
	Limit   int     `json:"limit"`
	Total   int     `json:"total"`
	HasNext bool    `json:"hasNext"`
}

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
