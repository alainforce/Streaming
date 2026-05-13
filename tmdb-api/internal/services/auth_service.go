// internal/services/auth_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alainforce/streaming/tmdb-api/internal/cache"
	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/alainforce/streaming/tmdb-api/internal/repository"
	jwtpkg "github.com/alainforce/streaming/tmdb-api/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *jwtpkg.Manager
	blacklist  *cache.TokenBlacklist
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *jwtpkg.Manager, blacklist *cache.TokenBlacklist) *AuthService {
	return &AuthService{userRepo: userRepo, jwtManager: jwtManager, blacklist: blacklist}
}

// SeedAdmin creates the admin account if ADMIN_EMAIL is configured and
// the account doesn't already exist. Safe to call on every startup.
func (s *AuthService) SeedAdmin(ctx context.Context, email, password string) error {
	if email == "" || password == "" {
		return nil // No admin configured — skip silently
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("auth: failed to hash admin password: %w", err)
	}

	// Pass role = "admin" explicitly.
	_, err = s.userRepo.Create(ctx, email, string(hash), "admin")
	if err != nil {
		// ErrDuplicateEmail means the admin already exists — that's fine.
		// On every restart after the first, this will trigger. Not an error.
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil
		}
		return fmt.Errorf("auth: failed to seed admin: %w", err)
	}

	return nil
}

// Signup creates a regular user account. Role is always "user" here.
func (s *AuthService) Signup(ctx context.Context, req models.SignupRequest) (*models.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to hash password: %w", err)
	}

	// Role is always "user" for self-registered accounts.
	// Admins can only be created via the seeder.
	user, err := s.userRepo.Create(ctx, req.Email, string(hash), "user")
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, repository.ErrDuplicateEmail
		}
		return nil, fmt.Errorf("auth: failed to create user: %w", err)
	}

	return s.buildAuthResponse(user)
}

// Login verifies credentials, checks ban status, and issues a JWT.
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("auth: failed to find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check ban AFTER verifying credentials.
	// If we checked before, we'd reveal that the email exists
	// (different error = user enumeration). Check credentials first,
	// then gate on status.
	if user.Status == "banned" {
		return nil, ErrAccountBanned
	}

	return s.buildAuthResponse(user)
}

// DeleteAccount removes the calling user's own account.
func (s *AuthService) DeleteAccount(ctx context.Context, userID string) error {
	return s.userRepo.Delete(ctx, userID)
}

// buildAuthResponse is a private helper to build the JWT + UserResponse.
// Both Signup and Login return the same shape — DRY.
func (s *AuthService) buildAuthResponse(user *models.User) (*models.AuthResponse, error) {
	token, err := s.jwtManager.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to generate token: %w", err)
	}

	return &models.AuthResponse{
		Token: token,
		User: models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

// Logout revokes the current token by adding its jti to the Redis blacklist.
// After this call, the token is permanently invalid even if it hasn't expired.
func (s *AuthService) Logout(ctx context.Context, jti string, remainingTTL time.Duration) error {
	if err := s.blacklist.BlacklistToken(ctx, jti, remainingTTL); err != nil {
		return fmt.Errorf("auth: failed to revoke token: %w", err)
	}
	return nil
}

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountBanned      = errors.New("this account has been suspended")
)
