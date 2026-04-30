package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
)

type MovieService struct {
	movieRepository repository.MovieRepository
}

func NewMovieService(movieRepository repository.MovieRepository) *MovieService {
	return &MovieService{movieRepository: movieRepository}
}

const (
	defaultMoviePage  = 1
	defaultMovieLimit = 20
	maxMovieLimit     = 100
)

func (s *MovieService) ListMovies(ctx context.Context, q domain.MovieListQuery) (domain.MovieListResponse, error) {
	// Apply defaults
	if q.Page < 1 {
		q.Page = defaultMoviePage
	}
	if q.Limit < 1 {
		q.Limit = defaultMovieLimit
	}
	if q.Limit > maxMovieLimit {
		q.Limit = maxMovieLimit
	}
	if !domain.ValidMovieSortFields[q.SortBy] {
		q.SortBy = domain.MovieSortByReleaseDate
	}

	return s.movieRepository.ListActive(ctx, q)
}

func (s *MovieService) GetMovieByID(ctx context.Context, movieID string) (domain.Movie, error) {
	return s.movieRepository.GetByID(ctx, strings.TrimSpace(movieID))
}

func (s *MovieService) GetShowDetailsBySelection(ctx context.Context, movieID string, theaterID string, showTime string) (domain.ShowDetails, error) {
	parsedShowTime, err := parseSelectionInputs(movieID, theaterID, showTime)
	if err != nil {
		return domain.ShowDetails{}, err
	}

	return s.movieRepository.GetShowDetailsBySelection(ctx, movieID, theaterID, parsedShowTime)
}

func (s *MovieService) GetSeatMapBySelection(ctx context.Context, movieID string, theaterID string, showTime string) (domain.SeatMapResponse, error) {
	parsedShowTime, err := parseSelectionInputs(movieID, theaterID, showTime)
	if err != nil {
		return domain.SeatMapResponse{}, err
	}

	return s.movieRepository.GetSeatMapBySelection(ctx, movieID, theaterID, parsedShowTime)
}

func (s *MovieService) GetSeatAvailabilityBySelection(ctx context.Context, movieID string, theaterID string, showTime string) (domain.SeatAvailabilityResponse, error) {
	parsedShowTime, err := parseSelectionInputs(movieID, theaterID, showTime)
	if err != nil {
		return domain.SeatAvailabilityResponse{}, err
	}

	return s.movieRepository.GetSeatAvailabilityBySelection(ctx, movieID, theaterID, parsedShowTime)
}

func (s *MovieService) ListTheatersByMovie(ctx context.Context, movieID string, showDate string) (domain.MovieTheaterListResponse, error) {
	trimmedMovieID := strings.TrimSpace(movieID)
	if trimmedMovieID == "" {
		return domain.MovieTheaterListResponse{}, fmt.Errorf("movie id is required")
	}

	movie, err := s.movieRepository.GetByID(ctx, trimmedMovieID)
	if err != nil {
		return domain.MovieTheaterListResponse{}, err
	}

	var parsedDate *time.Time
	if strings.TrimSpace(showDate) != "" {
		loc := time.Local
		parsed, parseErr := time.ParseInLocation("2006-01-02", showDate, loc)
		if parseErr != nil {
			return domain.MovieTheaterListResponse{}, fmt.Errorf("date must be in YYYY-MM-DD format")
		}
		now := time.Now().In(loc)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		latest := today.AddDate(0, 0, 14)
		chosen := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, loc)
		if chosen.Before(today) {
			return domain.MovieTheaterListResponse{}, fmt.Errorf("date cannot be in the past")
		}
		if chosen.After(latest) {
			return domain.MovieTheaterListResponse{}, fmt.Errorf("bookings are only available within the next 14 days")
		}
		parsedDate = &parsed
	}

	records, err := s.movieRepository.ListMovieShowtimeRecords(ctx, trimmedMovieID, parsedDate)
	if err != nil {
		return domain.MovieTheaterListResponse{}, err
	}

	theaterMap := make(map[string]*domain.TheaterSchedule)
	theaterOrder := make([]string, 0)

	for _, record := range records {
		theaterSchedule, exists := theaterMap[record.TheaterID]
		if !exists {
			theaterSchedule = &domain.TheaterSchedule{
				TheaterID:    record.TheaterID,
				TheaterName:  record.TheaterName,
				City:         record.City,
				AddressLine1: record.AddressLine1,
				Timezone:     record.Timezone,
				Showtimes:    make([]domain.ShowtimeItem, 0),
			}
			theaterMap[record.TheaterID] = theaterSchedule
			theaterOrder = append(theaterOrder, record.TheaterID)
		}

		showtime := domain.ShowtimeItem{
			ShowtimeID: record.ShowtimeID,
			ScreenName: record.ScreenName,
			StartTime:  record.StartTime,
			EndTime:    record.StartTime.Add(time.Duration(movie.DurationMinutes) * time.Minute),
			Language:   record.Language,
			Format:     record.Format,
			BasePrice:  record.BasePrice,
		}

		theaterSchedule.Showtimes = append(theaterSchedule.Showtimes, showtime)
	}

	theaters := make([]domain.TheaterSchedule, 0, len(theaterOrder))
	for _, theaterID := range theaterOrder {
		theaterSchedule := theaterMap[theaterID]
		sort.Slice(theaterSchedule.Showtimes, func(i, j int) bool {
			return theaterSchedule.Showtimes[i].StartTime.Before(theaterSchedule.Showtimes[j].StartTime)
		})

		for index := 1; index < len(theaterSchedule.Showtimes); index++ {
			previous := theaterSchedule.Showtimes[index-1]
			current := theaterSchedule.Showtimes[index]

			if current.StartTime.Sub(previous.StartTime) <= time.Duration(movie.DurationMinutes)*time.Minute {
				return domain.MovieTheaterListResponse{}, fmt.Errorf(
					"invalid showtime configuration for theater %s: gaps between start times must be greater than movie runtime",
					theaterSchedule.TheaterName,
				)
			}
		}

		theaters = append(theaters, *theaterSchedule)
	}

	response := domain.MovieTheaterListResponse{
		MovieID:         movie.ID,
		MovieTitle:      movie.Title,
		DurationMinutes: movie.DurationMinutes,
		Theaters:        theaters,
	}

	return response, nil
}

func parseSelectionInputs(movieID string, theaterID string, showTime string) (time.Time, error) {
	if strings.TrimSpace(movieID) == "" {
		return time.Time{}, fmt.Errorf("movieId query parameter is required")
	}

	if strings.TrimSpace(theaterID) == "" {
		return time.Time{}, fmt.Errorf("theaterId query parameter is required")
	}

	trimmedShowTime := strings.TrimSpace(showTime)
	if trimmedShowTime == "" {
		return time.Time{}, fmt.Errorf("showTime query parameter is required")
	}

	parsedShowTime, err := time.Parse(time.RFC3339, trimmedShowTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("showTime must be in RFC3339 format, e.g. 2026-03-25T15:00:00Z")
	}

	return parsedShowTime.UTC(), nil
}
