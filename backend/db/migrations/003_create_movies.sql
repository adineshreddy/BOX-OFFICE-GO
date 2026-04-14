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
