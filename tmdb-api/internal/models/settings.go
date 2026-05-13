// internal/models/settings.go
package models

import "time"

// ProfileResponse is what GET /settings/profile returns.
// It combines the user's profile with their personal activity stats —
// one round trip gives the frontend everything it needs to render
// the settings page without a second API call.
type ProfileResponse struct {
	Profile UserResponse  `json:"profile"`
	Stats   PersonalStats `json:"stats"`
}

// PersonalStats contains activity numbers scoped to the requesting user.
// These are computed from the favorites and watched tables at query time.
type PersonalStats struct {
	TotalFavorites int `json:"total_favorites"`
	TotalWatched   int `json:"total_watched"`
	// MostRecentActivity is the timestamp of the latest action —
	// whichever is more recent between the last favorite added
	// and the last movie watched. Null if the user has no activity.
	MostRecentActivity *time.Time `json:"most_recent_activity"`
}

// UpdateEmailRequest is the body for PATCH /settings/email
type UpdateEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// UpdatePasswordRequest is the body for PATCH /settings/password
type UpdatePasswordRequest struct {
	// NewPassword only — no current password required per spec.
	// See security note in the handler.
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
