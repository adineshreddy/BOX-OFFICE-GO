CREATE TABLE IF NOT EXISTS theaters (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    address_line1 TEXT NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'Asia/Kolkata',
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
        'Regal Lakeside',
        'Chicago',
        '455 N Cityfront Plaza Dr',
        'America/Chicago',
        5,
        TRUE,
        NOW(),
        NOW()
    ),
    (
        'th_003',
        'Cinemark Harbor Point',
        'Seattle',
        '600 Pine St, Downtown',
        'America/Los_Angeles',
        3,
        TRUE,
        NOW(),
        NOW()
    ),
    (
        'th_004',
        'Alamo Riverwalk',
        'San Antonio',
        '849 E Commerce St',
        'America/Chicago',
        4,
        TRUE,
        NOW(),
        NOW()
    ),
    (
        'th_005',
        'Landmark Atlantic Station',
        'Atlanta',
        '261 19th St NW',
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
VALUES
    ('st_001', 'mov_001', 'th_001', 'Screen 1', NOW() + INTERVAL '2 hours', 'English', '2D', 220.00, TRUE, NOW(), NOW()),
    ('st_002', 'mov_001', 'th_001', 'Screen 1', NOW() + INTERVAL '5 hours', 'English', '2D', 220.00, TRUE, NOW(), NOW()),
    ('st_003', 'mov_001', 'th_002', 'Screen 2', NOW() + INTERVAL '3 hours', 'English', 'IMAX', 420.00, TRUE, NOW(), NOW()),
    ('st_004', 'mov_001', 'th_002', 'Screen 2', NOW() + INTERVAL '6 hours 30 minutes', 'English', 'IMAX', 420.00, TRUE, NOW(), NOW()),
    ('st_005', 'mov_001', 'th_003', 'Screen A', NOW() + INTERVAL '4 hours', 'English', '2D', 250.00, TRUE, NOW(), NOW()),
    ('st_006', 'mov_001', 'th_003', 'Screen A', NOW() + INTERVAL '7 hours', 'English', '2D', 250.00, TRUE, NOW(), NOW()),
    ('st_007', 'mov_002', 'th_001', 'Screen 2', NOW() + INTERVAL '2 hours 30 minutes', 'Hindi', '2D', 180.00, TRUE, NOW(), NOW()),
    ('st_008', 'mov_002', 'th_001', 'Screen 2', NOW() + INTERVAL '5 hours', 'Hindi', '2D', 180.00, TRUE, NOW(), NOW()),
    ('st_009', 'mov_002', 'th_004', 'Screen 3', NOW() + INTERVAL '3 hours', 'Hindi', '2D', 210.00, TRUE, NOW(), NOW()),
    ('st_010', 'mov_002', 'th_004', 'Screen 3', NOW() + INTERVAL '6 hours', 'Hindi', '2D', 210.00, TRUE, NOW(), NOW()),
    ('st_011', 'mov_002', 'th_005', 'Screen 1', NOW() + INTERVAL '4 hours', 'Hindi', '2D', 190.00, TRUE, NOW(), NOW()),
    ('st_012', 'mov_003', 'th_002', 'Screen 4', NOW() + INTERVAL '3 hours 15 minutes', 'Telugu', '2D', 200.00, TRUE, NOW(), NOW()),
    ('st_013', 'mov_003', 'th_002', 'Screen 4', NOW() + INTERVAL '6 hours 30 minutes', 'Telugu', '2D', 200.00, TRUE, NOW(), NOW()),
    ('st_014', 'mov_003', 'th_003', 'Screen B', NOW() + INTERVAL '4 hours 30 minutes', 'Telugu', '2D', 220.00, TRUE, NOW(), NOW()),
    ('st_015', 'mov_003', 'th_003', 'Screen B', NOW() + INTERVAL '8 hours', 'Telugu', '2D', 220.00, TRUE, NOW(), NOW()),
    ('st_016', 'mov_003', 'th_005', 'Screen 2', NOW() + INTERVAL '5 hours', 'Telugu', '2D', 210.00, TRUE, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
