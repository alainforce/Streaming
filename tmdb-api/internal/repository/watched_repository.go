// internal/repository/watched_repository.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WatchedRepository struct {
	db *pgxpool.Pool
}

func NewWatchedRepository(db *pgxpool.Pool) *WatchedRepository {
	return &WatchedRepository{db: db}
}

// Save logs a movie as watched for the given user.
func (r *WatchedRepository) Save(ctx context.Context, userID string, req models.AddWatchedRequest) (*models.Watched, error) {
	// Handle optional watched_at: parse if provided, default to now.
	watchedAt := time.Now()
	if req.WatchedAt != nil && *req.WatchedAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.WatchedAt)
		if err != nil {
			return nil, ErrInvalidWatchedAt
		}
		// Reject future dates — you can't have watched a movie
		// that hasn't screened yet.
		if parsed.After(time.Now()) {
			return nil, ErrFutureWatchedAt
		}
		watchedAt = parsed
	}

	query := `
		INSERT INTO watched
			(user_id, movie_id, title, overview, poster_path, vote_average, personal_rating, watched_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, movie_id, title, overview, poster_path, vote_average, personal_rating, watched_at, added_at
	`

	var w models.Watched
	err := r.db.QueryRow(ctx, query,
		userID,
		req.MovieID,
		req.Title,
		req.Overview,
		req.PosterPath,
		req.VoteAverage,
		req.PersonalRating, // nil is fine — pgx maps Go nil to SQL NULL
		watchedAt,
	).Scan(
		&w.ID, &w.UserID, &w.MovieID, &w.Title,
		&w.Overview, &w.PosterPath, &w.VoteAverage,
		&w.PersonalRating, &w.WatchedAt, &w.AddedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAlreadyWatched
		}
		return nil, fmt.Errorf("repository: failed to save watched: %w", err)
	}

	return &w, nil
}

// GetAllByUser retrieves the user's full watched list, newest first.
func (r *WatchedRepository) GetAllByUser(ctx context.Context, userID string) ([]models.Watched, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, movie_id, title, overview, poster_path,
		       vote_average, personal_rating, watched_at, added_at
		FROM watched
		WHERE user_id = $1
		ORDER BY watched_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to query watched: %w", err)
	}
	defer rows.Close()

	watched := make([]models.Watched, 0)
	for rows.Next() {
		var w models.Watched
		if err := rows.Scan(
			&w.ID, &w.UserID, &w.MovieID, &w.Title,
			&w.Overview, &w.PosterPath, &w.VoteAverage,
			&w.PersonalRating, &w.WatchedAt, &w.AddedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: failed to scan watched: %w", err)
		}
		watched = append(watched, w)
	}

	return watched, rows.Err()
}

// Delete removes a movie from the user's watched list.
// Scoped by userID — users can only remove their own entries.
func (r *WatchedRepository) Delete(ctx context.Context, userID string, movieID int) error {
	result, err := r.db.Exec(ctx, `
		DELETE FROM watched
		WHERE user_id = $1 AND movie_id = $2
	`, userID, movieID)
	if err != nil {
		return fmt.Errorf("repository: failed to delete watched: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrWatchedNotFound
	}
	return nil
}

var (
	ErrAlreadyWatched   = errors.New("movie is already in your watched list")
	ErrWatchedNotFound  = errors.New("movie not found in your watched list")
	ErrInvalidWatchedAt = errors.New("watched_at must be a valid RFC3339 date e.g. 2024-03-15T20:30:00Z")
	ErrFutureWatchedAt  = errors.New("watched_at cannot be a future date")
)
