// internal/services/settings_service.go
package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/alainforce/streaming/tmdb-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// SettingsService handles business logic for user self-service settings.
type SettingsService struct {
	settingsRepo *repository.SettingsRepository
}

func NewSettingsService(settingsRepo *repository.SettingsRepository) *SettingsService {
	return &SettingsService{settingsRepo: settingsRepo}
}

// GetProfile fetches the user's profile and personal stats in parallel.
// We use goroutines so both DB queries run simultaneously instead of
// sequentially — the response time is max(query1, query2) rather than
// query1 + query2. For a settings page this is a meaningful improvement.
func (s *SettingsService) GetProfile(ctx context.Context, userID string) (*models.ProfileResponse, error) {
	// Channels to receive results from concurrent goroutines.
	// Each channel is buffered with capacity 1 so the goroutine
	// can send without blocking even if we haven't read yet.
	type profileResult struct {
		user *models.User
		err  error
	}
	type statsResult struct {
		stats *models.PersonalStats
		err   error
	}

	profileCh := make(chan profileResult, 1)
	statsCh := make(chan statsResult, 1)

	// Launch both queries concurrently.
	go func() {
		user, err := s.settingsRepo.GetProfile(ctx, userID)
		profileCh <- profileResult{user, err}
	}()

	go func() {
		stats, err := s.settingsRepo.GetPersonalStats(ctx, userID)
		statsCh <- statsResult{stats, err}
	}()

	// Collect both results. We wait for both regardless of errors
	// so goroutines don't leak.
	pr := <-profileCh
	sr := <-statsCh

	if pr.err != nil {
		return nil, fmt.Errorf("settings: failed to get profile: %w", pr.err)
	}
	if sr.err != nil {
		return nil, fmt.Errorf("settings: failed to get stats: %w", sr.err)
	}

	return &models.ProfileResponse{
		Profile: models.UserResponse{
			ID:        pr.user.ID,
			Email:     pr.user.Email,
			Role:      pr.user.Role,
			Status:    pr.user.Status,
			CreatedAt: pr.user.CreatedAt,
		},
		Stats: *sr.stats,
	}, nil
}

// UpdateEmail changes the user's email address.
// If the new email is already registered to another account,
// we return ErrDuplicateEmail so the handler can respond with 409.
func (s *SettingsService) UpdateEmail(ctx context.Context, userID, newEmail string) (*models.UserResponse, error) {
	user, err := s.settingsRepo.UpdateEmail(ctx, userID, newEmail)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, repository.ErrDuplicateEmail
		}
		return nil, fmt.Errorf("settings: %w", err)
	}

	return &models.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}, nil
}

// UpdatePassword hashes the new password and persists it.
// Hashing lives in the service layer — the repository only ever
// sees hashed values, never raw passwords.
func (s *SettingsService) UpdatePassword(ctx context.Context, userID, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("settings: failed to hash password: %w", err)
	}

	return s.settingsRepo.UpdatePassword(ctx, userID, string(hash))
}
