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

func (s *MovieService) ListMovies(ctx context.Context, titleQuery string, genreQuery string) ([]domain.Movie, error) {
	return s.movieRepository.ListActive(ctx, titleQuery, genreQuery)
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
		parsed, parseErr := time.Parse("2006-01-02", showDate)
		if parseErr != nil {
			return domain.MovieTheaterListResponse{}, fmt.Errorf("date must be in YYYY-MM-DD format")
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
