-- internal/database/migrations/001_create_favorites.sql
CREATE TABLE IF NOT EXISTS favorites (
    id          SERIAL PRIMARY KEY,
    movie_id    INTEGER NOT NULL UNIQUE,  -- UNIQUE prevents saving same movie twice
    title       TEXT NOT NULL,
    overview    TEXT,
    poster_path TEXT,
    vote_average NUMERIC(3, 1),           -- e.g. 8.4 — 3 digits total, 1 decimal
    added_at    TIMESTAMPTZ DEFAULT NOW() -- TIMESTAMPTZ stores timezone-aware time
);