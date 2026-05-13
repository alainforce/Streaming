// internal/repository/user_repository.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user with a specific role.
// The role parameter lets the seeder create an admin without a separate query.
func (r *UserRepository) Create(ctx context.Context, email, passwordHash, role string) (*models.User, error) {
	query := `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, role, status, created_at
	`

	var u models.User
	err := r.db.QueryRow(ctx, query, email, passwordHash, role).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("repository: failed to create user: %w", err)
	}

	return &u, nil
}

// FindByEmail retrieves a user by email for login.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, created_at
		FROM users
		WHERE email = $1
	`

	var u models.User
	err := r.db.QueryRow(ctx, query, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: failed to find user: %w", err)
	}

	return &u, nil
}

// Delete permanently removes a user.
// ON DELETE CASCADE in the DB handles favorites cleanup automatically.
func (r *UserRepository) Delete(ctx context.Context, userID string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("repository: failed to delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// --- Admin-only methods below this line ---

// ListAll returns a paginated list of all users.
// We return both the slice AND a total count so the caller can
// compute pagination metadata without a second query.
func (r *UserRepository) ListAll(ctx context.Context, page, pageSize int) ([]models.AdminUserView, int, error) {
	// Get total count first — needed for pagination math.
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: failed to count users: %w", err)
	}

	// LIMIT + OFFSET is the standard SQL pagination pattern.
	// OFFSET = (page - 1) * pageSize
	// e.g. page 2, size 10 → skip first 10 rows, return next 10.
	offset := (page - 1) * pageSize
	query := `
		SELECT id, email, role, status, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: failed to list users: %w", err)
	}

	defer rows.Close()

	users := make([]models.AdminUserView, 0)
	for rows.Next() {
		var u models.AdminUserView
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.Status, &u.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("repository: failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository: row error: %w", err)
	}

	return users, total, nil
}

// SetStatus updates a user's status to 'active' or 'banned'.
// Using a dedicated method instead of a generic Update prevents callers
// from accidentally changing fields they shouldn't (like role or email).
func (r *UserRepository) SetStatus(ctx context.Context, userID, status string) error {
	result, err := r.db.Exec(ctx,
		`UPDATE users SET status = $1 WHERE id = $2`,
		status, userID,
	)
	if err != nil {
		return fmt.Errorf("repository: failed to set status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// GetStats returns aggregate numbers for the stats dashboard.
func (r *UserRepository) GetStats(ctx context.Context) (total, active, banned, newLast7Days int, err error) {
	// A single query with conditional aggregation is much faster
	// than four separate COUNT queries — one round trip to the DB.
	query := `
		SELECT
			COUNT(*)                                             AS total,
			COUNT(*) FILTER (WHERE status = 'active')           AS active,
			COUNT(*) FILTER (WHERE status = 'banned')           AS banned,
			COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') AS new_last_7
		FROM users
	`
	err = r.db.QueryRow(ctx, query).Scan(&total, &active, &banned, &newLast7Days)
	if err != nil {
		err = fmt.Errorf("repository: failed to get user stats: %w", err)
	}
	return
}

// ComputeTotalPages is a helper to avoid this math being duplicated in services.
func ComputeTotalPages(total, pageSize int) int {
	return int(math.Ceil(float64(total) / float64(pageSize)))
}

// Sentinel errors (ErrDuplicateEmail, ErrUserNotFound stay the same)
var (
	ErrDuplicateEmail = errors.New("email is already registered")
	ErrUserNotFound   = errors.New("user not found")
)
