package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"box-office-go/backend/internal/domain"
	"box-office-go/backend/internal/repository"
)

type MovieRepository struct {
	db *sql.DB
}

func NewMovieRepository(db *sql.DB) *MovieRepository {
	return &MovieRepository{db: db}
}

func (r *MovieRepository) ListActive(ctx context.Context, titleQuery string, genreQuery string) ([]domain.Movie, error) {
	baseQuery := `
	SELECT
		id,
		title,
		description,
		genre,
		language,
		duration_minutes,
		release_date,
		rating,
		poster_url,
		is_active,
		created_at,
		updated_at
	FROM movies
	WHERE is_active = TRUE
	`

	clauses := make([]string, 0)
	args := make([]any, 0)
	argIndex := 1

	if trimmed := strings.TrimSpace(titleQuery); trimmed != "" {
		clauses = append(clauses, fmt.Sprintf("title ILIKE $%d", argIndex))
		args = append(args, "%"+trimmed+"%")
		argIndex++
	}

	if trimmed := strings.TrimSpace(genreQuery); trimmed != "" {
		clauses = append(clauses, fmt.Sprintf("genre ILIKE $%d", argIndex))
		args = append(args, "%"+trimmed+"%")
		argIndex++
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(baseQuery)

	if len(clauses) > 0 {
		queryBuilder.WriteString(" AND ")
		queryBuilder.WriteString(strings.Join(clauses, " AND "))
	}

	queryBuilder.WriteString(" ORDER BY release_date DESC, title ASC")

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := make([]domain.Movie, 0)
	for rows.Next() {
		var movie domain.Movie
		var posterURL sql.NullString

		if scanErr := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.Description,
			&movie.Genre,
			&movie.Language,
			&movie.DurationMinutes,
			&movie.ReleaseDate,
			&movie.Rating,
			&posterURL,
			&movie.IsActive,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		); scanErr != nil {
			return nil, scanErr
		}

		if posterURL.Valid {
			movie.PosterURL = &posterURL.String
		}

		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (r *MovieRepository) GetByID(ctx context.Context, movieID string) (domain.Movie, error) {
	query := `
	SELECT
		id,
		title,
		description,
		genre,
		language,
		duration_minutes,
		release_date,
		rating,
		poster_url,
		is_active,
		created_at,
		updated_at
	FROM movies
	WHERE id = $1
	`

	var movie domain.Movie
	var posterURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(movieID)).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Description,
		&movie.Genre,
		&movie.Language,
		&movie.DurationMinutes,
		&movie.ReleaseDate,
		&movie.Rating,
		&posterURL,
		&movie.IsActive,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Movie{}, repository.ErrMovieNotFound
		}
		return domain.Movie{}, err
	}

	if posterURL.Valid {
		movie.PosterURL = &posterURL.String
	}

	return movie, nil
}

func (r *MovieRepository) ListMovieShowtimeRecords(ctx context.Context, movieID string, showDate *time.Time) ([]domain.MovieShowtimeRecord, error) {
	baseQuery := `
	SELECT
		m.id,
		m.title,
		m.duration_minutes,
		t.id,
		t.name,
		t.city,
		t.address_line1,
		t.timezone,
		s.id,
		s.screen_name,
		s.start_time,
		s.language,
		s.format,
		s.base_price
	FROM showtimes s
	JOIN movies m ON m.id = s.movie_id
	JOIN theaters t ON t.id = s.theater_id
	WHERE m.id = $1
	  AND m.is_active = TRUE
	  AND t.is_active = TRUE
	  AND s.is_active = TRUE
	  AND (s.start_time AT TIME ZONE 'America/New_York')::date >= (NOW() AT TIME ZONE 'America/New_York')::date
	`

	args := []any{strings.TrimSpace(movieID)}
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(baseQuery)

	if showDate != nil {
		queryBuilder.WriteString(" AND (s.start_time AT TIME ZONE 'America/New_York')::date = $2::date")
		args = append(args, showDate.Format("2006-01-02"))
	}

	queryBuilder.WriteString(" ORDER BY t.name ASC, s.start_time ASC")

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]domain.MovieShowtimeRecord, 0)
	for rows.Next() {
		var record domain.MovieShowtimeRecord

		if scanErr := rows.Scan(
			&record.MovieID,
			&record.MovieTitle,
			&record.MovieDuration,
			&record.TheaterID,
			&record.TheaterName,
			&record.City,
			&record.AddressLine1,
			&record.Timezone,
			&record.ShowtimeID,
			&record.ScreenName,
			&record.StartTime,
			&record.Language,
			&record.Format,
			&record.BasePrice,
		); scanErr != nil {
			return nil, scanErr
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *MovieRepository) GetShowDetailsBySelection(ctx context.Context, movieID string, theaterID string, showTime time.Time) (domain.ShowDetails, error) {
	query := `
	SELECT
		s.id,
		m.id,
		m.title,
		t.id,
		t.name,
		t.city,
		t.address_line1,
		s.screen_name,
		s.start_time,
		s.language,
		s.format,
		s.base_price,
		m.duration_minutes,
		COUNT(si.seat_number) AS total_seats,
		COUNT(si.seat_number) FILTER (WHERE si.is_available = TRUE AND si.is_held = FALSE) AS available_seats
	FROM showtimes s
	JOIN movies m ON m.id = s.movie_id
	JOIN theaters t ON t.id = s.theater_id
	LEFT JOIN seat_inventory si ON si.showtime_id = s.id
	WHERE m.id = $1
	  AND t.id = $2
	  AND s.start_time = $3
	  AND m.is_active = TRUE
	  AND t.is_active = TRUE
	  AND s.is_active = TRUE
	GROUP BY s.id, m.id, m.title, t.id, t.name, t.city, t.address_line1, s.screen_name, s.start_time, s.language, s.format, s.base_price, m.duration_minutes
	LIMIT 1
	`

	var details domain.ShowDetails

	err := r.db.QueryRowContext(
		ctx,
		query,
		strings.TrimSpace(movieID),
		strings.TrimSpace(theaterID),
		showTime.UTC(),
	).Scan(
		&details.ShowtimeID,
		&details.MovieID,
		&details.MovieTitle,
		&details.TheaterID,
		&details.TheaterName,
		&details.City,
		&details.AddressLine1,
		&details.ScreenName,
		&details.StartTime,
		&details.Language,
		&details.Format,
		&details.BasePrice,
		&details.DurationMinutes,
		&details.TotalSeats,
		&details.AvailableSeats,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ShowDetails{}, repository.ErrShowtimeNotFound
		}
		return domain.ShowDetails{}, err
	}

	details.UnavailableSeats = details.TotalSeats - details.AvailableSeats
	return details, nil
}

func (r *MovieRepository) GetSeatMapBySelection(ctx context.Context, movieID string, theaterID string, showTime time.Time) (domain.SeatMapResponse, error) {
	query := `
	SELECT
		s.id,
		m.title,
		t.name,
		s.screen_name,
		s.start_time,
		si.seat_number,
		si.row_label,
		si.seat_index,
		si.seat_type,
		si.price_multiplier,
		si.is_available,
		si.is_held
	FROM showtimes s
	JOIN movies m ON m.id = s.movie_id
	JOIN theaters t ON t.id = s.theater_id
	JOIN seat_inventory si ON si.showtime_id = s.id
	WHERE m.id = $1
	  AND t.id = $2
	  AND s.start_time = $3
	  AND m.is_active = TRUE
	  AND t.is_active = TRUE
	  AND s.is_active = TRUE
	ORDER BY si.row_label ASC, si.seat_index ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		strings.TrimSpace(movieID),
		strings.TrimSpace(theaterID),
		showTime.UTC(),
	)
	if err != nil {
		return domain.SeatMapResponse{}, err
	}
	defer rows.Close()

	result := domain.SeatMapResponse{}
	rowMap := make(map[string][]domain.SeatItem)
	rowOrder := make([]string, 0)

	foundAny := false
	for rows.Next() {
		foundAny = true

		var seat domain.SeatItem
		var rowLabel string

		if scanErr := rows.Scan(
			&result.ShowtimeID,
			&result.MovieTitle,
			&result.TheaterName,
			&result.ScreenName,
			&result.ShowTime,
			&seat.SeatNumber,
			&rowLabel,
			&seat.SeatIndex,
			&seat.SeatType,
			&seat.PriceFactor,
			&seat.IsAvailable,
			&seat.IsHeld,
		); scanErr != nil {
			return domain.SeatMapResponse{}, scanErr
		}

		seat.RowLabel = rowLabel

		if _, ok := rowMap[rowLabel]; !ok {
			rowMap[rowLabel] = make([]domain.SeatItem, 0)
			rowOrder = append(rowOrder, rowLabel)
		}

		rowMap[rowLabel] = append(rowMap[rowLabel], seat)
		result.TotalSeats++
		if seat.IsAvailable && !seat.IsHeld {
			result.AvailableSeats++
		}
	}

	if err := rows.Err(); err != nil {
		return domain.SeatMapResponse{}, err
	}

	if !foundAny {
		return domain.SeatMapResponse{}, repository.ErrShowtimeNotFound
	}

	sort.Strings(rowOrder)
	result.Rows = make([]domain.SeatRow, 0, len(rowOrder))
	for _, label := range rowOrder {
		seats := rowMap[label]
		sort.Slice(seats, func(i, j int) bool {
			return seats[i].SeatIndex < seats[j].SeatIndex
		})
		result.Rows = append(result.Rows, domain.SeatRow{
			RowLabel: label,
			Seats:    seats,
		})
	}

	result.UnavailableSeats = result.TotalSeats - result.AvailableSeats
	return result, nil
}

func (r *MovieRepository) GetSeatAvailabilityBySelection(ctx context.Context, movieID string, theaterID string, showTime time.Time) (domain.SeatAvailabilityResponse, error) {
	query := `
	SELECT
		s.id,
		m.title,
		t.name,
		s.start_time,
		COUNT(si.seat_number) AS total_seats,
		COUNT(si.seat_number) FILTER (WHERE si.is_available = TRUE AND si.is_held = FALSE) AS available_seats
	FROM showtimes s
	JOIN movies m ON m.id = s.movie_id
	JOIN theaters t ON t.id = s.theater_id
	JOIN seat_inventory si ON si.showtime_id = s.id
	WHERE m.id = $1
	  AND t.id = $2
	  AND s.start_time = $3
	  AND m.is_active = TRUE
	  AND t.is_active = TRUE
	  AND s.is_active = TRUE
	GROUP BY s.id, m.title, t.name, s.start_time
	LIMIT 1
	`

	var result domain.SeatAvailabilityResponse
	err := r.db.QueryRowContext(
		ctx,
		query,
		strings.TrimSpace(movieID),
		strings.TrimSpace(theaterID),
		showTime.UTC(),
	).Scan(
		&result.ShowtimeID,
		&result.MovieTitle,
		&result.TheaterName,
		&result.ShowTime,
		&result.TotalSeats,
		&result.AvailableSeats,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SeatAvailabilityResponse{}, repository.ErrShowtimeNotFound
		}
		return domain.SeatAvailabilityResponse{}, err
	}

	result.UnavailableSeats = result.TotalSeats - result.AvailableSeats
	result.LastRefreshedAt = time.Now().UTC()

	return result, nil
}

var _ repository.MovieRepository = (*MovieRepository)(nil)
