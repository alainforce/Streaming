// internal/database/migrate.go
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(pool *pgxpool.Pool) error {
	migrations := []string{
		// Enable pgcrypto for gen_random_uuid().
		// gen_random_uuid() generates a UUID v4 — random and unguessable.
		// This is a PostgreSQL built-in since v13, pgcrypto covers older versions.
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE EXTENSION IF NOT EXISTS citext`,

		// 001 — users table with UUID primary key
		`CREATE TABLE IF NOT EXISTS users (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email         CITEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role          TEXT NOT NULL DEFAULT 'user'
			              CHECK (role IN ('user', 'admin')),
			status        TEXT NOT NULL DEFAULT 'active'
			              CHECK (status IN ('active', 'banned')),
			created_at    TIMESTAMPTZ DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,

		// 002 — favorites table
		// user_id is now UUID to match users.id
		`CREATE TABLE IF NOT EXISTS favorites (
			id           SERIAL PRIMARY KEY,
			user_id      UUID NOT NULL
			             REFERENCES users(id) ON DELETE CASCADE,
			movie_id     INTEGER NOT NULL,
			title        TEXT NOT NULL,
			overview     TEXT,
			poster_path  TEXT,
			vote_average NUMERIC(3, 1),
			added_at     TIMESTAMPTZ DEFAULT NOW()
		)`,

		// Composite unique — one user cannot save the same movie twice
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_favorites_user_movie
			ON favorites(user_id, movie_id)`,

		`CREATE INDEX IF NOT EXISTS idx_favorites_user_id
			ON favorites(user_id)`,

		`CREATE INDEX IF NOT EXISTS idx_favorites_movie_id
			ON favorites(movie_id)`,

		// 003 — watched list table
		// Independent from favorites — a movie can be in both lists.
		// personal_rating is 1–10, stored as SMALLINT (1 byte, enough for 0-32767).
		// watched_at defaults to NOW() but can be overridden by the user
		// (they might log a movie they watched last week).
		`CREATE TABLE IF NOT EXISTS watched (
			id              SERIAL PRIMARY KEY,
			user_id         UUID NOT NULL
			                REFERENCES users(id) ON DELETE CASCADE,
			movie_id        INTEGER NOT NULL,
			title           TEXT NOT NULL,
			overview        TEXT,
			poster_path     TEXT,
			vote_average    NUMERIC(3, 1),
			personal_rating SMALLINT
			                CHECK (personal_rating BETWEEN 1 AND 10),
			watched_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			added_at        TIMESTAMPTZ DEFAULT NOW()
		)`,

		// A user cannot log the same movie twice in their watched list
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_watched_user_movie
			ON watched(user_id, movie_id)`,

		`CREATE INDEX IF NOT EXISTS idx_watched_user_id
			ON watched(user_id)`,
	}

	ctx := context.Background()
	for i, query := range migrations {
		if _, err := pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("migrate: failed on migration %d: %w", i+1, err)
		}
	}

	return nil
}
