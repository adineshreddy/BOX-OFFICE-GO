-- Keep poster URLs aligned with app seed (real titles; Wikipedia-hosted artwork).
UPDATE movies
SET poster_url = 'https://upload.wikimedia.org/wikipedia/en/5/52/Dune_Part_Two_poster.jpeg',
    updated_at = NOW()
WHERE id = 'mov_001';

UPDATE movies
SET poster_url = 'https://upload.wikimedia.org/wikipedia/en/4/4a/Oppenheimer_%28film%29.jpg',
    updated_at = NOW()
WHERE id = 'mov_002';

UPDATE movies
SET poster_url = 'https://upload.wikimedia.org/wikipedia/en/0/0b/Barbie_2023_poster.jpg',
    updated_at = NOW()
WHERE id = 'mov_003';

UPDATE movies
SET poster_url = 'https://upload.wikimedia.org/wikipedia/en/b/b4/Spider-Man-_Across_the_Spider-Verse_poster.jpg',
    updated_at = NOW()
WHERE id = 'mov_004';

UPDATE movies
SET poster_url = 'https://upload.wikimedia.org/wikipedia/en/d/d0/John_Wick_-_Chapter_4_promotional_poster.jpg',
    updated_at = NOW()
WHERE id = 'mov_005';

UPDATE movies
SET poster_url = 'https://upload.wikimedia.org/wikipedia/en/1/1c/The_Dark_Knight_%282008_film%29.jpg',
    updated_at = NOW()
WHERE id = 'mov_006';

UPDATE movies
SET poster_url = 'https://upload.wikimedia.org/wikipedia/en/2/2e/Inception_%282010%29_theatrical_poster.jpg',
    updated_at = NOW()
WHERE id = 'mov_007';

UPDATE movies
SET poster_url = 'https://upload.wikimedia.org/wikipedia/en/1/1e/Everything_Everywhere_All_at_Once.jpg',
    updated_at = NOW()
WHERE id = 'mov_008';
