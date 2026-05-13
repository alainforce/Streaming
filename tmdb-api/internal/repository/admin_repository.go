// internal/repository/admin_repository.go
package repository

import (
	"context"
	"fmt"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminRepository handles queries that only admins perform.
// Separating these from UserRepository keeps each file focused
// and makes it obvious which queries carry elevated privileges.
type AdminRepository struct {
	db *pgxpool.Pool
}

func NewAdminRepository(db *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{db: db}
}

// GetAllSavedMovies returns every favorite across all users, paginated.
// It JOINs favorites with users to include who saved each movie.
func (r *AdminRepository) GetAllSavedMovies(ctx context.Context, page, pageSize int) ([]models.SavedMovie, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM favorites`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("admin repository: failed to count favorites: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT
			f.id,
			f.user_id,
			u.email      AS user_email,
			f.movie_id,
			f.title,
			f.poster_path,
			f.vote_average,
			f.added_at
		FROM favorites f
		-- INNER JOIN: only rows where user_id matches a users.id row.
		-- Since we have ON DELETE CASCADE, orphaned favorites can't exist,
		-- but INNER JOIN makes the intent explicit.
		INNER JOIN users u ON u.id = f.user_id
		ORDER BY f.added_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("admin repository: failed to query favorites: %w", err)
	}
	defer rows.Close()

	movies := make([]models.SavedMovie, 0)
	for rows.Next() {
		var m models.SavedMovie
		err := rows.Scan(
			&m.ID, &m.UserID, &m.UserEmail,
			&m.MovieID, &m.Title, &m.PosterPath,
			&m.VoteAverage, &m.AddedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("admin repository: failed to scan favorite: %w", err)
		}
		movies = append(movies, m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("admin repository: row error: %w", err)
	}

	return movies, total, nil
}

// GetFavoriteStats returns aggregate data about favorites for the stats dashboard.
func (r *AdminRepository) GetFavoriteStats(ctx context.Context) (total, newLast7Days int, err error) {
	query := `
		SELECT
			COUNT(*)                                                        AS total,
			COUNT(*) FILTER (WHERE added_at > NOW() - INTERVAL '7 days')   AS new_last_7
		FROM favorites
	`
	err = r.db.QueryRow(ctx, query).Scan(&total, &newLast7Days)
	if err != nil {
		err = fmt.Errorf("admin repository: failed to get favorite stats: %w", err)
	}
	return
}

// GetTopSavedMovies returns the N most-saved movies across all users.
func (r *AdminRepository) GetTopSavedMovies(ctx context.Context, limit int) ([]models.TopMovie, error) {
	query := `
		SELECT
			movie_id,
			title,
			COUNT(*) AS save_count
		FROM favorites
		GROUP BY movie_id, title
		ORDER BY save_count DESC
		LIMIT $1
	`
	// GROUP BY movie_id + title gives us one row per unique movie.
	// COUNT(*) tells us how many users saved it.
	// ORDER BY save_count DESC puts the most popular first.

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("admin repository: failed to get top movies: %w", err)
	}
	defer rows.Close()

	movies := make([]models.TopMovie, 0)
	for rows.Next() {
		var m models.TopMovie
		if err := rows.Scan(&m.MovieID, &m.Title, &m.SaveCount); err != nil {
			return nil, fmt.Errorf("admin repository: failed to scan top movie: %w", err)
		}
		movies = append(movies, m)
	}

	return movies, rows.Err()
}
