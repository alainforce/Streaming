// internal/repository/favorite_repository.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FavoriteRepository handles all database operations for favorites.
type FavoriteRepository struct {
	db *pgxpool.Pool
}

// NewFavoriteRepository is the constructor.
func NewFavoriteRepository(db *pgxpool.Pool) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

// Save inserts a new favorite into the database.
// It returns the fully-populated Favorite (with the DB-generated id and added_at).
func (r *FavoriteRepository) Save(ctx context.Context, req models.AddFavoriteRequest, userID string) (*models.Favorite, error) {
	query := `
		INSERT INTO favorites (user_id, movie_id, title, overview, poster_path, vote_average)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, movie_id, title, overview, poster_path, vote_average, added_at
	`
	// $1, $2... are parameterized placeholders. NEVER use fmt.Sprintf to
	// build SQL strings with user input — that is a SQL injection vulnerability.
	// pgx will safely escape and bind these values.

	var f models.Favorite
	err := r.db.QueryRow(ctx, query,
		userID,
		req.MovieID,
		req.Title,
		req.Overview,
		req.PosterPath,
		req.VoteAverage,
	).Scan(&f.ID, &f.UserID, &f.MovieID, &f.Title, &f.Overview, &f.PosterPath, &f.VoteAverage, &f.AddedAt)

	if err != nil {
		// pgx exposes a specific error type for constraint violations.
		// A UNIQUE constraint failure means the movie is already in favorites.
		// We check for this specifically so we can return a clean 409 Conflict
		// instead of a generic 500 Internal Server Error.
		if isUniqueViolation(err) {
			return nil, ErrDuplicateFavorite
		}
		return nil, fmt.Errorf("repository: failed to save favorite: %w", err)
	}

	return &f, nil
}

// GetAll retrieves every saved favorite, ordered by most recently added.
func (r *FavoriteRepository) GetAll(ctx context.Context) ([]models.Favorite, error) {
	query := `
		SELECT id, user_id, movie_id, title, overview, poster_path, vote_average, added_at
		FROM favorites
		ORDER BY added_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to query favorites: %w", err)
	}
	// Like http.Response.Body, pgx rows must be closed after use.
	defer rows.Close()

	// Pre-allocate the slice. If there are no rows, we return an empty
	// slice [] instead of null in JSON — much friendlier for API clients.
	favorites := make([]models.Favorite, 0)

	for rows.Next() {
		var f models.Favorite
		err := rows.Scan(
			&f.ID, &f.MovieID, &f.Title,
			&f.Overview, &f.PosterPath, &f.VoteAverage, &f.AddedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("repository: failed to scan row: %w", err)
		}
		favorites = append(favorites, f)
	}

	// rows.Err() must be checked after the loop. It captures any error
	// that occurred during iteration that wasn't caught by rows.Next().
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: row iteration error: %w", err)
	}

	return favorites, nil
}

// Get retrieves every saved favorite by a specific user, ordered by most recently added.
func (r *FavoriteRepository) Get(ctx context.Context, userID string) ([]models.Favorite, error) {
	query := `
		SELECT id, user_id, movie_id, title, overview, poster_path, vote_average, added_at
		FROM favorites
		WHERE user_id = $1
		ORDER BY added_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to query favorites: %w", err)
	}
	// Like http.Response.Body, pgx rows must be closed after use.
	defer rows.Close()

	// Pre-allocate the slice. If there are no rows, we return an empty
	// slice [] instead of null in JSON — much friendlier for API clients.
	favorites := make([]models.Favorite, 0)

	for rows.Next() {
		var f models.Favorite
		err := rows.Scan(
			&f.ID, &f.UserID, &f.MovieID, &f.Title,
			&f.Overview, &f.PosterPath, &f.VoteAverage, &f.AddedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("repository: failed to scan row: %w", err)
		}
		favorites = append(favorites, f)
	}

	// rows.Err() must be checked after the loop. It captures any error
	// that occurred during iteration that wasn't caught by rows.Next().
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: row iteration error: %w", err)
	}

	return favorites, nil
}

// Delete removes a specific movie from a user's favorites.
// We scope by BOTH userID AND movieID — a user can only delete their own entries.
// Without the userID check, user A could delete user B's favorites
// just by knowing the movieID.
func (r *FavoriteRepository) Delete(ctx context.Context, userID string, movieID int) error {
	result, err := r.db.Exec(ctx, `
		DELETE FROM favorites
		WHERE user_id = $1 AND movie_id = $2
	`, userID, movieID)
	if err != nil {
		return fmt.Errorf("repository: failed to delete favorite: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrFavoriteNotFound
	}
	return nil
}

// --- Sentinel errors and helpers ---

// ErrDuplicateFavorite is a sentinel error — a package-level error variable
// with a specific identity. The service layer can check for this exact error
// using errors.Is() to decide what HTTP status code to return.

// isUniqueViolation checks if a pgx error is a PostgreSQL unique constraint
// violation (error code 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	// errors.As unwraps the error chain to find a *pgx.PgError
	if errors.As(err, &pgErr) {
		// Constraint violation specific to unique
		return pgErr.Code == "23505"
	}

	// fallback: check the string
	_ = time.Now() // suppress unused import if pgx.PgError approach works
	return false
}

var (
	ErrDuplicateFavorite = errors.New("movie is already in your favorites")
	ErrFavoriteNotFound  = errors.New("movie not found in your favorites")
)
