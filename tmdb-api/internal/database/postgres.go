// internal/database/postgres.go
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool creates and validates a PostgreSQL connection pool.
// We return *pgxpool.Pool, which is safe for concurrent use by multiple
// goroutines — one pool is shared by your entire application.
func NewPostgresPool(databaseURL string) (*pgxpool.Pool, error) {
	// pgxpool.Config lets us tune pool behavior.
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("database: failed to parse config: %w", err)
	}

	// These are sensible defaults for a small-to-medium application.
	config.MaxConns = 10               // Maximum open connections in the pool.
	config.MinConns = 2                // Keep 2 connections warm and ready.
	config.MaxConnLifetime = time.Hour // Recycle connections after 1 hour.
	config.MaxConnIdleTime = 30 * time.Minute

	// Create the pool. This does NOT open connections yet.
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("database: failed to create pool: %w", err)
	}

	// Ping the DB to verify the connection is actually reachable.
	// Fail fast at startup rather than failing on the first request.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database: failed to ping: %w", err)
	}

	return pool, nil
}