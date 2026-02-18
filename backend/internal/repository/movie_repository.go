package repository

import (
	"context"
	"errors"
	"time"

	"box-office-go/backend/internal/domain"
)

var ErrMovieNotFound = errors.New("movie not found")

type MovieRepository interface {
	ListActive(ctx context.Context, titleQuery string, genreQuery string) ([]domain.Movie, error)
	GetByID(ctx context.Context, movieID string) (domain.Movie, error)
	ListMovieShowtimeRecords(ctx context.Context, movieID string, showDate *time.Time) ([]domain.MovieShowtimeRecord, error)
}
