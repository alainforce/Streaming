// internal/models/admin.go
package models

import (
	"time"
)

// AdminUserView is what the admin sees when listing users.
// It includes role and status — fields a regular user doesn't need to see
// about themselves or others.
type AdminUserView struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// SavedMovie represents a favorite saved by any user — used in admin view.
// Includes who saved it, unlike the regular Favorite model.
type SavedMovie struct {
	ID          int       `json:"id"`
	UserID      string    `json:"user_id"`
	UserEmail   string    `json:"user_email"`
	MovieID     int       `json:"movie_id"`
	Title       string    `json:"title"`
	PosterPath  string    `json:"poster_path"`
	VoteAverage float64   `json:"vote_average"`
	AddedAt     time.Time `json:"added_at"`
}

// TopMovie is used in the stats response to show most-saved movies.
type TopMovie struct {
	MovieID   int    `json:"movie_id"`
	Title     string `json:"title"`
	SaveCount int    `json:"save_count"`
}

// AppStats is the full stats payload returned by GET /admin/stats.
type AppStats struct {
	TotalUsers            int        `json:"total_users"`
	ActiveUsers           int        `json:"active_users"`
	BannedUsers           int        `json:"banned_users"`
	NewUsersLast7Days     int        `json:"new_users_last_7_days"`
	TotalFavoritesSaved   int        `json:"total_favorites_saved"`
	NewFavoritesLast7Days int        `json:"new_favorites_last_7_days"`
	TopSavedMovies        []TopMovie `json:"top_saved_movies"`
}

// PaginationMeta is returned alongside list endpoints so clients
// know how to fetch the next page.
type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}
