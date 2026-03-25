package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
	"box-office-go/backend/internal/service"
)

type movieRepoHandlerStub struct{}

func (movieRepoHandlerStub) ListActive(_ context.Context, _, _ string) ([]domain.Movie, error) {
	return []domain.Movie{{ID: "mov_1", Title: "Movie"}}, nil
}
func (movieRepoHandlerStub) GetByID(_ context.Context, movieID string) (domain.Movie, error) {
	if movieID == "missing" {
		return domain.Movie{}, repository.ErrMovieNotFound
	}
	return domain.Movie{ID: movieID, Title: "Movie", DurationMinutes: 100}, nil
}
func (movieRepoHandlerStub) ListMovieShowtimeRecords(_ context.Context, _ string, _ *time.Time) ([]domain.MovieShowtimeRecord, error) {
	return []domain.MovieShowtimeRecord{}, nil
}
func (movieRepoHandlerStub) GetShowDetailsBySelection(_ context.Context, _, _ string, _ time.Time) (domain.ShowDetails, error) {
	return domain.ShowDetails{}, errors.New("not implemented")
}
func (movieRepoHandlerStub) GetSeatMapBySelection(_ context.Context, _, _ string, _ time.Time) (domain.SeatMapResponse, error) {
	return domain.SeatMapResponse{}, errors.New("not implemented")
}
func (movieRepoHandlerStub) GetSeatAvailabilityBySelection(_ context.Context, _, _ string, _ time.Time) (domain.SeatAvailabilityResponse, error) {
	return domain.SeatAvailabilityResponse{}, errors.New("not implemented")
}

func TestMovieHandlerListMovies_Success(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoHandlerStub{}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies", nil)
	rec := httptest.NewRecorder()

	h.ListMovies(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMovieHandlerGetMovie_NotFound(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoHandlerStub{}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies/missing", nil)
	req.SetPathValue("movieId", "missing")
	rec := httptest.NewRecorder()

	h.GetMovie(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestMovieHandlerGetShowDetails_MissingQuery(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoHandlerStub{}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/shows/details", nil)
	rec := httptest.NewRecorder()

	h.GetShowDetailsBySelection(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
