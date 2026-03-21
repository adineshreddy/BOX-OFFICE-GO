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

UPDATE showtimes
SET base_price = 12.00,
    updated_at = NOW();
