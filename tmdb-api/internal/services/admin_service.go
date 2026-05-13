// internal/services/admin_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/alainforce/streaming/tmdb-api/internal/cache"
	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/alainforce/streaming/tmdb-api/internal/repository"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	topMoviesLimit  = 10
)

// AdminService contains all admin business logic.
type AdminService struct {
	userRepo  *repository.UserRepository
	adminRepo *repository.AdminRepository
	blacklist *cache.TokenBlacklist
}

func NewAdminService(userRepo *repository.UserRepository, adminRepo *repository.AdminRepository, blacklist *cache.TokenBlacklist) *AdminService {
	return &AdminService{
		userRepo:  userRepo,
		adminRepo: adminRepo,
		blacklist: blacklist,
	}
}

// ListUsers returns a paginated list of all users with metadata.
func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int) ([]models.AdminUserView, models.PaginationMeta, error) {
	// Clamp pageSize to prevent abuse — nobody needs 10,000 rows per page.
	pageSize = clampPageSize(pageSize)
	if page < 1 {
		page = 1
	}

	users, total, err := s.userRepo.ListAll(ctx, page, pageSize)
	if err != nil {
		log.Printf("ERROR: failed to list users: %v", err)
		return nil, models.PaginationMeta{}, err
	}

	meta := models.PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: repository.ComputeTotalPages(total, pageSize),
	}

	return users, meta, nil
}

// BanUser sets a user's status to 'banned'.
// A banned user's JWT is still technically valid (we'll discuss this below),
// but the login endpoint will block them from getting new tokens.
func (s *AdminService) BanUser(ctx context.Context, userID string) error {
	// 1. Update the database — blocks future logins
	if err := s.userRepo.SetStatus(ctx, userID, "banned"); err != nil {
		return err
	}

	// 2. Write the Redis ban marker — immediately blocks active tokens.
	// If the Redis write fails we still return nil because the DB is the
	// source of truth. The user will be blocked on their next login attempt.
	// Log the failure so it can be investigated.
	if err := s.blacklist.BanUser(ctx, userID); err != nil {
		log.Printf("WARN: ban DB succeeded but Redis marker failed for userID=%s: %v", userID, err)
	}

	return nil
}

// UnbanUser restores a user's status to 'active'.
func (s *AdminService) UnbanUser(ctx context.Context, userID string) error {
	if err := s.userRepo.SetStatus(ctx, userID, "active"); err != nil {
		return err
	}

	// 2. Remove the Redis ban marker — immediately restores access.
	if err := s.blacklist.UnbanUser(ctx, userID); err != nil {
		log.Printf("WARN: unban DB succeeded but Redis marker failed for userID=%s: %v", userID, err)
	}

	return nil
}

// DeleteUser permanently removes a user account.
// Favorites are removed automatically by ON DELETE CASCADE.
func (s *AdminService) DeleteUser(ctx context.Context, userID string) error {
	return s.userRepo.Delete(ctx, userID)
}

// GetAllSavedMovies returns a paginated list of all favorites across all users.
func (s *AdminService) GetAllSavedMovies(ctx context.Context, page, pageSize int) ([]models.SavedMovie, models.PaginationMeta, error) {
	pageSize = clampPageSize(pageSize)
	if page < 1 {
		page = 1
	}

	movies, total, err := s.adminRepo.GetAllSavedMovies(ctx, page, pageSize)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	meta := models.PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: repository.ComputeTotalPages(total, pageSize),
	}

	return movies, meta, nil
}

// GetStats assembles the full stats payload from multiple repository calls.
func (s *AdminService) GetStats(ctx context.Context) (*models.AppStats, error) {
	totalUsers, activeUsers, bannedUsers, newUsers, err := s.userRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("admin service: failed to get user stats: %w", err)
	}

	totalFavs, newFavs, err := s.adminRepo.GetFavoriteStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("admin service: failed to get favorite stats: %w", err)
	}

	topMovies, err := s.adminRepo.GetTopSavedMovies(ctx, topMoviesLimit)
	if err != nil {
		return nil, fmt.Errorf("admin service: failed to get top movies: %w", err)
	}

	return &models.AppStats{
		TotalUsers:            totalUsers,
		ActiveUsers:           activeUsers,
		BannedUsers:           bannedUsers,
		NewUsersLast7Days:     newUsers,
		TotalFavoritesSaved:   totalFavs,
		NewFavoritesLast7Days: newFavs,
		TopSavedMovies:        topMovies,
	}, nil
}

func clampPageSize(pageSize int) int {
	if pageSize < 1 {
		return defaultPageSize
	}
	if pageSize > maxPageSize {
		return maxPageSize
	}
	return pageSize
}

// ErrCannotBanAdmin prevents admins from banning other admins accidentally.
var ErrCannotBanAdmin = errors.New("cannot ban an admin account")
