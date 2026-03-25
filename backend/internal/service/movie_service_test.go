package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
)

type movieRepoStub struct {
	listActiveFn            func(ctx context.Context, titleQuery string, genreQuery string) ([]domain.Movie, error)
	getByIDFn               func(ctx context.Context, movieID string) (domain.Movie, error)
	listShowtimeRecordsFn   func(ctx context.Context, movieID string, showDate *time.Time) ([]domain.MovieShowtimeRecord, error)
	getShowDetailsFn        func(ctx context.Context, movieID string, theaterID string, showTime time.Time) (domain.ShowDetails, error)
	getSeatMapFn            func(ctx context.Context, movieID string, theaterID string, showTime time.Time) (domain.SeatMapResponse, error)
	getSeatAvailabilityFn   func(ctx context.Context, movieID string, theaterID string, showTime time.Time) (domain.SeatAvailabilityResponse, error)
}

func (s *movieRepoStub) ListActive(ctx context.Context, titleQuery string, genreQuery string) ([]domain.Movie, error) {
	return s.listActiveFn(ctx, titleQuery, genreQuery)
}
func (s *movieRepoStub) GetByID(ctx context.Context, movieID string) (domain.Movie, error) {
	return s.getByIDFn(ctx, movieID)
}
func (s *movieRepoStub) ListMovieShowtimeRecords(ctx context.Context, movieID string, showDate *time.Time) ([]domain.MovieShowtimeRecord, error) {
	return s.listShowtimeRecordsFn(ctx, movieID, showDate)
}
func (s *movieRepoStub) GetShowDetailsBySelection(ctx context.Context, movieID string, theaterID string, showTime time.Time) (domain.ShowDetails, error) {
	return s.getShowDetailsFn(ctx, movieID, theaterID, showTime)
}
func (s *movieRepoStub) GetSeatMapBySelection(ctx context.Context, movieID string, theaterID string, showTime time.Time) (domain.SeatMapResponse, error) {
	return s.getSeatMapFn(ctx, movieID, theaterID, showTime)
}
func (s *movieRepoStub) GetSeatAvailabilityBySelection(ctx context.Context, movieID string, theaterID string, showTime time.Time) (domain.SeatAvailabilityResponse, error) {
	return s.getSeatAvailabilityFn(ctx, movieID, theaterID, showTime)
}

func TestNewMovieService(t *testing.T) {
	svc := NewMovieService(&movieRepoStub{})
	if svc == nil {
		t.Fatal("expected service instance, got nil")
	}
}

func TestMovieServiceListMovies(t *testing.T) {
	svc := NewMovieService(&movieRepoStub{
		listActiveFn: func(_ context.Context, _, _ string) ([]domain.Movie, error) {
			return []domain.Movie{{ID: "mov_1", Title: "X"}}, nil
		},
	})

	movies, err := svc.ListMovies(context.Background(), "", "")
	if err != nil || len(movies) != 1 {
		t.Fatalf("unexpected result movies=%v err=%v", movies, err)
	}
}

func TestMovieServiceGetMovieByID_TrimsInput(t *testing.T) {
	received := ""
	svc := NewMovieService(&movieRepoStub{
		getByIDFn: func(_ context.Context, movieID string) (domain.Movie, error) {
			received = movieID
			return domain.Movie{ID: movieID}, nil
		},
	})

	_, _ = svc.GetMovieByID(context.Background(), "  mov_1  ")
	if received != "mov_1" {
		t.Fatalf("expected trimmed movie id, got %q", received)
	}
}

func TestMovieServiceSelectionEndpoints_ValidateInput(t *testing.T) {
	svc := NewMovieService(&movieRepoStub{})

	if _, err := svc.GetShowDetailsBySelection(context.Background(), "", "th_1", "2026-03-25T15:00:00Z"); err == nil {
		t.Fatal("expected validation error for movieId")
	}
	if _, err := svc.GetSeatMapBySelection(context.Background(), "mov_1", "", "2026-03-25T15:00:00Z"); err == nil {
		t.Fatal("expected validation error for theaterId")
	}
	if _, err := svc.GetSeatAvailabilityBySelection(context.Background(), "mov_1", "th_1", "bad-time"); err == nil {
		t.Fatal("expected validation error for showTime")
	}
}

func TestMovieServiceGetShowDetailsBySelection_Success(t *testing.T) {
	called := false
	svc := NewMovieService(&movieRepoStub{
		getShowDetailsFn: func(_ context.Context, movieID string, theaterID string, showTime time.Time) (domain.ShowDetails, error) {
			called = true
			return domain.ShowDetails{MovieID: movieID, TheaterID: theaterID, StartTime: showTime}, nil
		},
	})

	result, err := svc.GetShowDetailsBySelection(context.Background(), "mov_1", "th_1", "2026-03-25T15:00:00Z")
	if err != nil || !called || result.MovieID != "mov_1" {
		t.Fatalf("unexpected result=%+v err=%v called=%v", result, err, called)
	}
}

func TestMovieServiceListTheatersByMovie_MissingMovieID(t *testing.T) {
	svc := NewMovieService(&movieRepoStub{})
	if _, err := svc.ListTheatersByMovie(context.Background(), "", ""); err == nil {
		t.Fatal("expected validation error for missing movie id")
	}
}

func TestMovieServiceListTheatersByMovie_InvalidDate(t *testing.T) {
	svc := NewMovieService(&movieRepoStub{
		getByIDFn: func(_ context.Context, _ string) (domain.Movie, error) {
			return domain.Movie{ID: "mov_1", Title: "X", DurationMinutes: 100}, nil
		},
	})

	if _, err := svc.ListTheatersByMovie(context.Background(), "mov_1", "03/25/2026"); err == nil {
		t.Fatal("expected invalid date error")
	}
}

func TestMovieServiceListTheatersByMovie_ShowtimeGapValidation(t *testing.T) {
	svc := NewMovieService(&movieRepoStub{
		getByIDFn: func(_ context.Context, _ string) (domain.Movie, error) {
			return domain.Movie{ID: "mov_1", Title: "X", DurationMinutes: 120}, nil
		},
		listShowtimeRecordsFn: func(_ context.Context, _ string, _ *time.Time) ([]domain.MovieShowtimeRecord, error) {
			start := time.Date(2026, 3, 25, 15, 0, 0, 0, time.UTC)
			return []domain.MovieShowtimeRecord{
				{TheaterID: "th_1", TheaterName: "T", ShowtimeID: "st_1", ScreenName: "S1", StartTime: start, Language: "EN", Format: "2D", BasePrice: 12},
				{TheaterID: "th_1", TheaterName: "T", ShowtimeID: "st_2", ScreenName: "S1", StartTime: start.Add(90 * time.Minute), Language: "EN", Format: "2D", BasePrice: 12},
			}, nil
		},
	})

	_, err := svc.ListTheatersByMovie(context.Background(), "mov_1", "")
	if err == nil {
		t.Fatal("expected showtime gap validation error")
	}
}

func TestMovieServiceListTheatersByMovie_Success(t *testing.T) {
	svc := NewMovieService(&movieRepoStub{
		getByIDFn: func(_ context.Context, _ string) (domain.Movie, error) {
			return domain.Movie{ID: "mov_1", Title: "X", DurationMinutes: 100}, nil
		},
		listShowtimeRecordsFn: func(_ context.Context, _ string, _ *time.Time) ([]domain.MovieShowtimeRecord, error) {
			start := time.Date(2026, 3, 25, 15, 0, 0, 0, time.UTC)
			return []domain.MovieShowtimeRecord{
				{TheaterID: "th_1", TheaterName: "T", City: "NY", AddressLine1: "A", Timezone: "UTC", ShowtimeID: "st_1", ScreenName: "S1", StartTime: start, Language: "EN", Format: "2D", BasePrice: 12},
				{TheaterID: "th_1", TheaterName: "T", City: "NY", AddressLine1: "A", Timezone: "UTC", ShowtimeID: "st_2", ScreenName: "S1", StartTime: start.Add(3 * time.Hour), Language: "EN", Format: "2D", BasePrice: 12},
			}, nil
		},
	})

	resp, err := svc.ListTheatersByMovie(context.Background(), "mov_1", "")
	if err != nil {
		t.Fatalf("expected success, got err=%v", err)
	}
	if resp.MovieID != "mov_1" || len(resp.Theaters) != 1 || len(resp.Theaters[0].Showtimes) != 2 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestMovieServiceListTheatersByMovie_RepoError(t *testing.T) {
	svc := NewMovieService(&movieRepoStub{
		getByIDFn: func(_ context.Context, _ string) (domain.Movie, error) {
			return domain.Movie{}, repository.ErrMovieNotFound
		},
	})

	_, err := svc.ListTheatersByMovie(context.Background(), "mov_404", "")
	if !errors.Is(err, repository.ErrMovieNotFound) {
		t.Fatalf("expected ErrMovieNotFound, got %v", err)
	}
}
