// internal/services/favorite_service.go
package services

import (
	"context"
	"errors"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/alainforce/streaming/tmdb-api/internal/repository"
)

// FavoriteService contains business logic for the favorites feature.
type FavoriteService struct {
	repo *repository.FavoriteRepository
}

// NewFavoriteService is the constructor.
func NewFavoriteService(repo *repository.FavoriteRepository) *FavoriteService {
	return &FavoriteService{repo: repo}
}

// AddFavorite saves a movie to favorites.
// It translates the repository-level ErrDuplicateFavorite into a service-level
// concern, so the handler can decide the HTTP response without knowing
// anything about database error codes.
func (s *FavoriteService) AddFavorite(ctx context.Context, req models.AddFavoriteRequest, userID string) (*models.Favorite, error) {
	favorite, err := s.repo.Save(ctx, req, userID)
	if err != nil {
		// We surface the duplicate error as-is so the handler can
		// respond with 409 Conflict instead of 500 Internal Server Error.
		if errors.Is(err, repository.ErrDuplicateFavorite) {
			return nil, err
		}
		return nil, err
	}
	return favorite, nil
}

// GetFavorites retrieves all saved favorites.
func (s *FavoriteService) GetFavorites(ctx context.Context) ([]models.Favorite, error) {
	return s.repo.GetAll(ctx)
}

func (s *FavoriteService) GetFavoriteByUser(ctx context.Context, userID string) ([]models.Favorite, error) {
	return s.repo.Get(ctx, userID)
}

func (s *FavoriteService) RemoveFavorite(ctx context.Context, userID string, movieID int) error {
	return s.repo.Delete(ctx, userID, movieID)
}
