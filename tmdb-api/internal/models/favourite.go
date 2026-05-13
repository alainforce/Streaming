// internal/models/favorite.go
package models

import "time"

type Favorite struct {
	ID          int       `json:"id"`
	UserID      string    `json:"user_id"` // string, not int
	MovieID     int       `json:"movie_id"`
	Title       string    `json:"title"`
	Overview    string    `json:"overview"`
	PosterPath  string    `json:"poster_path"`
	VoteAverage float64   `json:"vote_average"`
	AddedAt     time.Time `json:"added_at"`
}

// AddFavoriteRequest — userID still comes from JWT, never from body
type AddFavoriteRequest struct {
	MovieID     int     `json:"movie_id"     binding:"required"`
	Title       string  `json:"title"        binding:"required"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	VoteAverage float64 `json:"vote_average"`
}
