package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

var _ repository.MovieRepository = (*MovieRepository)(nil)
