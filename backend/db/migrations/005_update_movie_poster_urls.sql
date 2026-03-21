-- Replace placeholder example.com poster URLs with working TMDB CDN URLs.
UPDATE movies
SET poster_url = 'https://image.tmdb.org/t/p/w500/gEU2QniE6E77NI6lCU6MxlNBvIx.jpg',
    updated_at = NOW()
WHERE id = 'mov_001';

UPDATE movies
SET poster_url = 'https://image.tmdb.org/t/p/w500/q6y0Go1tsGEsmtFryDOJo3dEmqu.jpg',
    updated_at = NOW()
WHERE id = 'mov_002';

UPDATE movies
SET poster_url = 'https://image.tmdb.org/t/p/w500/arw2vcBveWOVZr6pxd9XTd1TdQa.jpg',
    updated_at = NOW()
WHERE id = 'mov_003';
