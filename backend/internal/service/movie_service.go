package service

import (
	"context"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
)

type MovieService struct {
	movieRepository repository.MovieRepository
}

func NewMovieService(movieRepository repository.MovieRepository) *MovieService {
	return &MovieService{movieRepository: movieRepository}
}

func (s *MovieService) ListMovies(ctx context.Context, titleQuery string, genreQuery string) ([]domain.Movie, error) {
	return s.movieRepository.ListActive(ctx, titleQuery, genreQuery)
}
