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

type movieRepoExtraStub struct {
	listActiveErr          error
	getByIDErr             error
	showDetailsErr         error
	seatMapErr             error
	seatAvailabilityErr    error
	listShowtimeRecordsErr error
}

func (s movieRepoExtraStub) ListActive(_ context.Context, q domain.MovieListQuery) (domain.MovieListResponse, error) {
	if s.listActiveErr != nil {
		return domain.MovieListResponse{}, s.listActiveErr
	}
	return domain.MovieListResponse{Movies: []domain.Movie{{ID: "mov_1", Title: "Movie"}}, Page: q.Page, Limit: q.Limit, Total: 1}, nil
}
func (s movieRepoExtraStub) GetByID(_ context.Context, movieID string) (domain.Movie, error) {
	if s.getByIDErr != nil {
		return domain.Movie{}, s.getByIDErr
	}
	return domain.Movie{ID: movieID, Title: "Movie", DurationMinutes: 100}, nil
}
func (s movieRepoExtraStub) ListMovieShowtimeRecords(_ context.Context, _ string, _ *time.Time) ([]domain.MovieShowtimeRecord, error) {
	if s.listShowtimeRecordsErr != nil {
		return nil, s.listShowtimeRecordsErr
	}
	return []domain.MovieShowtimeRecord{}, nil
}
func (s movieRepoExtraStub) GetShowDetailsBySelection(_ context.Context, _, _ string, _ time.Time) (domain.ShowDetails, error) {
	if s.showDetailsErr != nil {
		return domain.ShowDetails{}, s.showDetailsErr
	}
	return domain.ShowDetails{ShowtimeID: "st_1"}, nil
}
func (s movieRepoExtraStub) GetSeatMapBySelection(_ context.Context, _, _ string, _ time.Time) (domain.SeatMapResponse, error) {
	if s.seatMapErr != nil {
		return domain.SeatMapResponse{}, s.seatMapErr
	}
	return domain.SeatMapResponse{ShowtimeID: "st_1"}, nil
}
func (s movieRepoExtraStub) GetSeatAvailabilityBySelection(_ context.Context, _, _ string, _ time.Time) (domain.SeatAvailabilityResponse, error) {
	if s.seatAvailabilityErr != nil {
		return domain.SeatAvailabilityResponse{}, s.seatAvailabilityErr
	}
	return domain.SeatAvailabilityResponse{ShowtimeID: "st_1"}, nil
}

func TestMovieHandlerListMovies_MethodNotAllowed(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoExtraStub{}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/movies", nil)
	rec := httptest.NewRecorder()

	h.ListMovies(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestMovieHandlerListMovies_InternalError(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoExtraStub{listActiveErr: errors.New("db")}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies", nil)
	rec := httptest.NewRecorder()

	h.ListMovies(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestMovieHandlerGetMovie_MissingPathParam(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoExtraStub{}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies/", nil)
	rec := httptest.NewRecorder()

	h.GetMovie(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestMovieHandlerGetShowDetails_MethodNotAllowed(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoExtraStub{}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shows/details", nil)
	rec := httptest.NewRecorder()

	h.GetShowDetailsBySelection(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestMovieHandlerGetShowDetails_NotFound(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoExtraStub{showDetailsErr: repository.ErrShowtimeNotFound}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/shows/details?movieId=mov_1&theaterId=th_1&showTime=2026-03-25T15:00:00Z", nil)
	rec := httptest.NewRecorder()

	h.GetShowDetailsBySelection(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestMovieHandlerGetSeatMap_InternalError(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoExtraStub{seatMapErr: errors.New("db")}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/shows/seat-map?movieId=mov_1&theaterId=th_1&showTime=2026-03-25T15:00:00Z", nil)
	rec := httptest.NewRecorder()

	h.GetSeatMapBySelection(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestMovieHandlerRefreshSeatAvailability_InvalidQuery(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoExtraStub{}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/shows/seat-map/availability?movieId=&theaterId=th_1&showTime=2026-03-25T15:00:00Z", nil)
	rec := httptest.NewRecorder()

	h.RefreshSeatAvailability(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestMovieHandlerListTheatersByMovie_MethodNotAllowed(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoExtraStub{}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/movies/mov_1/theaters", nil)
	req.SetPathValue("movieId", "mov_1")
	rec := httptest.NewRecorder()

	h.ListTheatersByMovie(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestMovieHandlerListTheatersByMovie_MovieNotFound(t *testing.T) {
	h := NewMovieHandler(service.NewMovieService(movieRepoExtraStub{getByIDErr: repository.ErrMovieNotFound}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/movies/missing/theaters", nil)
	req.SetPathValue("movieId", "missing")
	rec := httptest.NewRecorder()

	h.ListTheatersByMovie(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
