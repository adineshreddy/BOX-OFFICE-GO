package repository

import (
	"context"

	"box-office-go/backend/internal/domain"
)

type MovieRepository interface {
	ListActive(ctx context.Context, titleQuery string, genreQuery string) ([]domain.Movie, error)
}
