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
        'https://image.tmdb.org/t/p/w500/gEU2QniE6E77NI6lCU6MxlNBvIx.jpg',
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
        'https://image.tmdb.org/t/p/w500/arw2vcBveWOVZr6pxd9XTd1TdQa.jpg',
        TRUE,
        NOW(),
        NOW()
    )
ON CONFLICT (id) DO NOTHING;
