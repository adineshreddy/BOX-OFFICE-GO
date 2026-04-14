-- Posters that match seeded titles (fictional movies; TMDB IDs were unrelated films).
UPDATE movies
SET poster_url = 'https://placehold.co/500x750/0c1222/7dd3fc/png?text=Starlight+Horizon',
    updated_at = NOW()
WHERE id = 'mov_001';

UPDATE movies
SET poster_url = 'https://placehold.co/500x750/0f1f1a/86efac/png?text=Monsoon+Diaries',
    updated_at = NOW()
WHERE id = 'mov_002';

UPDATE movies
SET poster_url = 'https://placehold.co/500x750/1a0f14/fbcfe8/png?text=The+Last+Ticket',
    updated_at = NOW()
WHERE id = 'mov_003';
