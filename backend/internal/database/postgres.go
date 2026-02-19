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

	createTheatersTableQuery := `
	CREATE TABLE IF NOT EXISTS theaters (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		city TEXT NOT NULL,
		address_line1 TEXT NOT NULL,
		timezone TEXT NOT NULL DEFAULT 'America/New_York',
		total_screens INTEGER NOT NULL DEFAULT 1 CHECK (total_screens > 0),
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	createShowtimesTableQuery := `
	CREATE TABLE IF NOT EXISTS showtimes (
		id TEXT PRIMARY KEY,
		movie_id TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
		theater_id TEXT NOT NULL REFERENCES theaters(id) ON DELETE CASCADE,
		screen_name TEXT NOT NULL,
		start_time TIMESTAMPTZ NOT NULL,
		language TEXT NOT NULL,
		format TEXT NOT NULL,
		base_price NUMERIC(10,2) NOT NULL CHECK (base_price >= 0),
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(theater_id, screen_name, start_time)
	);
	`

	if _, err := db.ExecContext(ctx, createUsersTableQuery); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, createMoviesTableQuery); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, createTheatersTableQuery); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, createShowtimesTableQuery); err != nil {
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

	seedTheatersQuery := `
	INSERT INTO theaters (
		id,
		name,
		city,
		address_line1,
		timezone,
		total_screens,
		is_active,
		created_at,
		updated_at
	)
	VALUES
		(
			'th_001',
			'AMC Downtown 12',
			'New York',
			'123 W 42nd St, Manhattan',
			'America/New_York',
			4,
			TRUE,
			NOW(),
			NOW()
		),
		(
			'th_002',
			'Regal Back Bay',
			'Boston',
			'401 Park Dr',
			'America/New_York',
			5,
			TRUE,
			NOW(),
			NOW()
		),
		(
			'th_003',
			'Cinemark Buckhead',
			'Atlanta',
			'3393 Peachtree Rd NE',
			'America/New_York',
			3,
			TRUE,
			NOW(),
			NOW()
		),
		(
			'th_004',
			'Alamo South Beach',
			'Miami',
			'1212 Lincoln Rd',
			'America/New_York',
			4,
			TRUE,
			NOW(),
			NOW()
		),
		(
			'th_005',
			'Landmark Capitol View',
			'Washington',
			'700 Pennsylvania Ave NW',
			'America/New_York',
			3,
			TRUE,
			NOW(),
			NOW()
		)
	ON CONFLICT (id) DO UPDATE
	SET
		name = EXCLUDED.name,
		city = EXCLUDED.city,
		address_line1 = EXCLUDED.address_line1,
		timezone = EXCLUDED.timezone,
		total_screens = EXCLUDED.total_screens,
		is_active = EXCLUDED.is_active,
		updated_at = NOW();
	`

	if _, err := db.ExecContext(ctx, seedTheatersQuery); err != nil {
		return err
	}

	seedShowtimesQuery := `
	WITH bounds AS (
		SELECT
			date_trunc('day', NOW() AT TIME ZONE 'America/New_York') AS start_date,
			date_trunc('day', (NOW() AT TIME ZONE 'America/New_York') + INTERVAL '6 months') AS end_date
	),
	days AS (
		SELECT gs::timestamp AS show_day
		FROM bounds b,
		generate_series(b.start_date, b.end_date, INTERVAL '1 day') gs
	),
	templates AS (
		SELECT *
		FROM (
			VALUES
				('mov_001', 'th_001', 'Screen 1', '11 hours'::interval, 'English', '2D', 220.00::numeric),
				('mov_001', 'th_002', 'Screen 2', '18 hours'::interval, 'English', 'IMAX', 420.00::numeric),
				('mov_002', 'th_001', 'Screen 2', '10 hours 30 minutes'::interval, 'Hindi', '2D', 180.00::numeric),
				('mov_002', 'th_004', 'Screen 3', '17 hours 30 minutes'::interval, 'Hindi', '2D', 210.00::numeric),
				('mov_003', 'th_003', 'Screen B', '12 hours 15 minutes'::interval, 'Telugu', '2D', 220.00::numeric),
				('mov_003', 'th_005', 'Screen 2', '19 hours 15 minutes'::interval, 'Telugu', '2D', 210.00::numeric)
		) AS t(movie_id, theater_id, screen_name, time_of_day, language, format, base_price)
	),
	generated AS (
		SELECT
			'st_' || substr(md5(t.movie_id || '|' || t.theater_id || '|' || t.screen_name || '|' || to_char(d.show_day, 'YYYYMMDD') || '|' || t.time_of_day::text), 1, 20) AS id,
			t.movie_id,
			t.theater_id,
			t.screen_name,
			(d.show_day + t.time_of_day) AT TIME ZONE 'America/New_York' AS start_time,
			t.language,
			t.format,
			t.base_price
		FROM days d
		CROSS JOIN templates t
	)
	INSERT INTO showtimes (
		id,
		movie_id,
		theater_id,
		screen_name,
		start_time,
		language,
		format,
		base_price,
		is_active,
		created_at,
		updated_at
	)
	SELECT
		g.id,
		g.movie_id,
		g.theater_id,
		g.screen_name,
		g.start_time,
		g.language,
		g.format,
		g.base_price,
		TRUE,
		NOW(),
		NOW()
	FROM generated g
	ON CONFLICT (theater_id, screen_name, start_time) DO UPDATE
	SET
		movie_id = EXCLUDED.movie_id,
		language = EXCLUDED.language,
		format = EXCLUDED.format,
		base_price = EXCLUDED.base_price,
		is_active = EXCLUDED.is_active,
		updated_at = NOW();
	`

	if _, err := db.ExecContext(ctx, seedShowtimesQuery); err != nil {
		return err
	}

	return nil
}
