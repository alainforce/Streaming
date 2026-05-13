// internal/repository/settings_repository.go
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SettingsRepository handles all DB operations for user settings.
// It is separate from UserRepository intentionally — UserRepository
// serves auth and admin concerns, SettingsRepository serves the
// user's own self-service operations. Keeping them apart makes each
// file's purpose immediately obvious.
type SettingsRepository struct {
	db *pgxpool.Pool
}

func NewSettingsRepository(db *pgxpool.Pool) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// GetProfile fetches the user's profile row.
func (r *SettingsRepository) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, created_at
		FROM users
		WHERE id = $1
	`
	var u models.User
	err := r.db.QueryRow(ctx, query, userID).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("settings: failed to get profile: %w", err)
	}
	return &u, nil
}

// GetPersonalStats fetches activity counts and the most recent
// activity timestamp for the given user in a single query.
//
// We use a FULL OUTER JOIN between the favorites and watched
// aggregate subqueries so that if the user has entries in only
// one table, we still get a result row with the other as zero/null.
// COALESCE fills in 0 for missing counts.
func (r *SettingsRepository) GetPersonalStats(ctx context.Context, userID string) (*models.PersonalStats, error) {
	query := `
		SELECT
			COALESCE(f.total, 0)        AS total_favorites,
			COALESCE(w.total, 0)        AS total_watched,
			-- GREATEST picks the most recent timestamp between the two lists.
			-- If one is NULL (no activity in that list), GREATEST ignores it.
			GREATEST(f.last_added, w.last_watched) AS most_recent_activity
		FROM
			(SELECT COUNT(*) AS total, MAX(added_at) AS last_added
			 FROM favorites WHERE user_id = $1) f
		FULL OUTER JOIN
			(SELECT COUNT(*) AS total, MAX(watched_at) AS last_watched
			 FROM watched WHERE user_id = $1) w
		ON true
	`
	// "ON true" in a FULL OUTER JOIN is intentional — we're joining
	// two single-row subqueries, so there's no real join key.
	// This gives us one result row with all four values.

	var stats models.PersonalStats
	err := r.db.QueryRow(ctx, query, userID).
		Scan(&stats.TotalFavorites, &stats.TotalWatched, &stats.MostRecentActivity)
	if err != nil {
		return nil, fmt.Errorf("settings: failed to get personal stats: %w", err)
	}
	return &stats, nil
}

// UpdateEmail changes the user's email address.
// Returns ErrDuplicateEmail if the new email is already taken.
func (r *SettingsRepository) UpdateEmail(ctx context.Context, userID, newEmail string) (*models.User, error) {
	query := `
		UPDATE users
		SET email = $1
		WHERE id = $2
		RETURNING id, email, password_hash, role, status, created_at
	`
	var u models.User
	err := r.db.QueryRow(ctx, query, newEmail, userID).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateEmail
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("settings: failed to update email: %w", err)
	}
	return &u, nil
}

// UpdatePassword replaces the user's password hash.
// The caller is responsible for hashing the new password before
// passing it here — the repository never sees raw passwords.
func (r *SettingsRepository) UpdatePassword(ctx context.Context, userID, newHash string) error {
	result, err := r.db.Exec(ctx, `
		UPDATE users SET password_hash = $1 WHERE id = $2
	`, newHash, userID)
	if err != nil {
		return fmt.Errorf("settings: failed to update password: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
