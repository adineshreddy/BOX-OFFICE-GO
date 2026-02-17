package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := ensureSchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ensure schema: %w", err)
	}

	return db, nil
}

func ensureSchema(ctx context.Context, db *sql.DB) error {
	createUsersTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		phone TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		is_admin BOOLEAN NOT NULL DEFAULT FALSE,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		email_verified BOOLEAN NOT NULL DEFAULT FALSE,
		last_login_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	createMoviesTableQuery := `
	CREATE TABLE IF NOT EXISTS movies (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		genre TEXT NOT NULL,
		language TEXT NOT NULL,
		duration_minutes INTEGER NOT NULL,
		release_date DATE NOT NULL,
		rating NUMERIC(3,1) NOT NULL DEFAULT 0.0,
		poster_url TEXT,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.ExecContext(ctx, createUsersTableQuery); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, createMoviesTableQuery); err != nil {
		return err
	}

	alterUsersTableQueries := []string{
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;`,
		`UPDATE users SET updated_at = created_at WHERE updated_at IS NULL;`,
		`ALTER TABLE users ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;`,
		`ALTER TABLE users ALTER COLUMN updated_at SET NOT NULL;`,
	}

	for _, query := range alterUsersTableQueries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	seedMoviesQuery := `
	INSERT INTO movies (
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
	)
	VALUES
		(
			'mov_001',
			'Starlight Horizon',
			'A deep-space rescue mission turns into a fight for humanity’s survival.',
			'Sci-Fi',
			'English',
			128,
			'2024-06-14',
			8.4,
			'https://example.com/posters/starlight-horizon.jpg',
			TRUE,
			NOW(),
			NOW()
		),
		(
			'mov_002',
			'Monsoon Diaries',
			'An indie filmmaker uncovers family secrets during a monsoon season.',
			'Drama',
			'Hindi',
			112,
			'2023-11-03',
			7.6,
			'https://example.com/posters/monsoon-diaries.jpg',
			TRUE,
			NOW(),
			NOW()
		),
		(
			'mov_003',
			'The Last Ticket',
			'A train journey forces two strangers to rewrite their destinies.',
			'Romance',
			'Telugu',
			121,
			'2024-02-09',
			8.1,
			'https://example.com/posters/the-last-ticket.jpg',
			TRUE,
			NOW(),
			NOW()
		)
	ON CONFLICT (id) DO NOTHING;
	`

	if _, err := db.ExecContext(ctx, seedMoviesQuery); err != nil {
		return err
	}

	return nil
}
