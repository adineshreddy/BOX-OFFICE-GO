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

	createSeatInventoryTableQuery := `
	CREATE TABLE IF NOT EXISTS seat_inventory (
		showtime_id TEXT NOT NULL REFERENCES showtimes(id) ON DELETE CASCADE,
		seat_number TEXT NOT NULL,
		row_label TEXT NOT NULL,
		seat_index INTEGER NOT NULL CHECK (seat_index > 0),
		seat_type TEXT NOT NULL DEFAULT 'regular',
		price_multiplier NUMERIC(6,2) NOT NULL DEFAULT 1.00 CHECK (price_multiplier > 0),
		is_available BOOLEAN NOT NULL DEFAULT TRUE,
		is_held BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (showtime_id, seat_number)
	);
	`

	createBookingHoldsTableQuery := `
	CREATE TABLE IF NOT EXISTS booking_holds (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		showtime_id TEXT NOT NULL REFERENCES showtimes(id) ON DELETE CASCADE,
		status TEXT NOT NULL CHECK (status IN ('HELD', 'CONFIRMED', 'EXPIRED', 'CANCELLED')),
		hold_expires_at TIMESTAMPTZ NOT NULL,
		total_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	createBookingHoldSeatsTableQuery := `
	CREATE TABLE IF NOT EXISTS booking_hold_seats (
		hold_id TEXT NOT NULL REFERENCES booking_holds(id) ON DELETE CASCADE,
		showtime_id TEXT NOT NULL REFERENCES showtimes(id) ON DELETE CASCADE,
		seat_number TEXT NOT NULL,
		price_at_hold NUMERIC(10,2) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (hold_id, seat_number)
	);
	`

	createBookingsTableQuery := `
	CREATE TABLE IF NOT EXISTS bookings (
		id TEXT PRIMARY KEY,
		hold_id TEXT NOT NULL UNIQUE REFERENCES booking_holds(id) ON DELETE CASCADE,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		showtime_id TEXT NOT NULL REFERENCES showtimes(id) ON DELETE CASCADE,
		status TEXT NOT NULL CHECK (status IN ('CONFIRMED', 'CANCELLED')),
		total_amount NUMERIC(10,2) NOT NULL,
		confirmed_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	createAuthSessionsTableQuery := `
	CREATE TABLE IF NOT EXISTS auth_sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_id TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMPTZ NOT NULL,
		revoked_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	createPaymentTransactionsTableQuery := `
	CREATE TABLE IF NOT EXISTS payment_transactions (
		id TEXT PRIMARY KEY,
		hold_id TEXT NOT NULL REFERENCES booking_holds(id) ON DELETE CASCADE,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		amount NUMERIC(10,2) NOT NULL,
		payment_method TEXT NOT NULL,
		status TEXT NOT NULL CHECK (status IN ('SUCCESS', 'FAILED', 'PENDING')),
		gateway_txn_id TEXT,
		failure_reason TEXT,
		idempotency_key TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_txn_idempotency ON payment_transactions(idempotency_key);
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

	if _, err := db.ExecContext(ctx, createSeatInventoryTableQuery); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, createBookingHoldsTableQuery); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, createBookingHoldSeatsTableQuery); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, createBookingsTableQuery); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, createAuthSessionsTableQuery); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, createPaymentTransactionsTableQuery); err != nil {
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
			'https://placehold.co/500x750/0c1222/7dd3fc/png?text=Starlight+Horizon',
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
			'https://placehold.co/500x750/0f1f1a/86efac/png?text=Monsoon+Diaries',
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
			'https://placehold.co/500x750/1a0f14/fbcfe8/png?text=The+Last+Ticket',
			TRUE,
			NOW(),
			NOW()
		)
	ON CONFLICT (id) DO NOTHING;
	`

	if _, err := db.ExecContext(ctx, seedMoviesQuery); err != nil {
		return err
	}

	patchPosters := []struct {
		id  string
		url string
	}{
		{"mov_001", "https://placehold.co/500x750/0c1222/7dd3fc/png?text=Starlight+Horizon"},
		{"mov_002", "https://placehold.co/500x750/0f1f1a/86efac/png?text=Monsoon+Diaries"},
		{"mov_003", "https://placehold.co/500x750/1a0f14/fbcfe8/png?text=The+Last+Ticket"},
	}
	for _, p := range patchPosters {
		if _, err := db.ExecContext(ctx,
			`UPDATE movies SET poster_url = $2, updated_at = NOW() WHERE id = $1`,
			p.id, p.url,
		); err != nil {
			return err
		}
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
				('mov_001', 'th_001', 'Screen 1', '11 hours'::interval, 'English', '2D', 12.00::numeric),
				('mov_001', 'th_002', 'Screen 2', '18 hours'::interval, 'English', 'IMAX', 12.00::numeric),
				('mov_002', 'th_001', 'Screen 2', '10 hours 30 minutes'::interval, 'Hindi', '2D', 12.00::numeric),
				('mov_002', 'th_004', 'Screen 3', '17 hours 30 minutes'::interval, 'Hindi', '2D', 12.00::numeric),
				('mov_003', 'th_003', 'Screen B', '12 hours 15 minutes'::interval, 'Telugu', '2D', 12.00::numeric),
				('mov_003', 'th_005', 'Screen 2', '19 hours 15 minutes'::interval, 'Telugu', '2D', 12.00::numeric)
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

	forceBasePriceQuery := `
	UPDATE showtimes
	SET base_price = 12.00,
		updated_at = NOW();
	`

	if _, err := db.ExecContext(ctx, forceBasePriceQuery); err != nil {
		return err
	}

	seedSeatInventoryQuery := `
	WITH row_labels AS (
		SELECT unnest(ARRAY['A','B','C','D','E','F','G','H','I','J']) AS row_label
	),
	seat_numbers AS (
		SELECT generate_series(1, 12) AS seat_index
	),
	base_layout AS (
		SELECT
			rl.row_label,
			sn.seat_index,
			rl.row_label || LPAD(sn.seat_index::text, 2, '0') AS seat_number,
			CASE
				WHEN rl.row_label IN ('A', 'B') THEN 'premium'
				WHEN rl.row_label IN ('C', 'D', 'E', 'F', 'G', 'H') THEN 'regular'
				ELSE 'recliner'
			END AS seat_type,
			CASE
				WHEN rl.row_label IN ('A', 'B') THEN 1.40
				WHEN rl.row_label IN ('C', 'D', 'E', 'F', 'G', 'H') THEN 1.00
				ELSE 1.25
			END AS price_multiplier
		FROM row_labels rl
		CROSS JOIN seat_numbers sn
	),
	all_rows AS (
		SELECT
			s.id AS showtime_id,
			b.seat_number,
			b.row_label,
			b.seat_index,
			b.seat_type,
			b.price_multiplier,
			TRUE AS is_available,
			FALSE AS is_held,
			NOW() AS created_at,
			NOW() AS updated_at
		FROM showtimes s
		CROSS JOIN base_layout b
	)
	INSERT INTO seat_inventory (
		showtime_id,
		seat_number,
		row_label,
		seat_index,
		seat_type,
		price_multiplier,
		is_available,
		is_held,
		created_at,
		updated_at
	)
	SELECT
		showtime_id,
		seat_number,
		row_label,
		seat_index,
		seat_type,
		price_multiplier,
		is_available,
		is_held,
		created_at,
		updated_at
	FROM all_rows
	ON CONFLICT (showtime_id, seat_number) DO NOTHING;
	`

	if _, err := db.ExecContext(ctx, seedSeatInventoryQuery); err != nil {
		return err
	}

	return nil
}
