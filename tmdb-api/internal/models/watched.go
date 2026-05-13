// internal/models/watched.go
package models

import "time"

// Watched represents a movie the user has logged as seen.
// It is completely independent from Favorite — a movie can exist
// in both lists simultaneously with no constraints between them.
type Watched struct {
	ID          int     `json:"id"`
	UserID      string  `json:"user_id"`
	MovieID     int     `json:"movie_id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	VoteAverage float64 `json:"vote_average"`
	// PersonalRating is the user's own score, 1–10.
	// Using *int (pointer) so we can distinguish "not rated" (nil/null)
	// from "rated 0". A missing rating serializes to null in JSON,
	// which is more honest than serializing to 0.
	PersonalRating *int `json:"personal_rating"`
	// WatchedAt is when the user watched the movie.
	// Defaults to now but can be set to a past date.
	WatchedAt time.Time `json:"watched_at"`
	AddedAt   time.Time `json:"added_at"`
}

// AddWatchedRequest is the JSON body for POST /watched
type AddWatchedRequest struct {
	MovieID     int     `json:"movie_id"         binding:"required"`
	Title       string  `json:"title"            binding:"required"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	VoteAverage float64 `json:"vote_average"`
	// PersonalRating is optional — pointer so nil means "not provided"
	PersonalRating *int `json:"personal_rating"`
	// WatchedAt is optional — if omitted, the server uses NOW().
	// Format: RFC3339 e.g. "2024-03-15T20:30:00Z"
	WatchedAt *string `json:"watched_at"`
}
