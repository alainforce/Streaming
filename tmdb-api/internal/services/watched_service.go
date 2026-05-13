// internal/services/watched_service.go
package services

import (
	"context"
	"errors"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/alainforce/streaming/tmdb-api/internal/repository"
)

type WatchedService struct {
	repo *repository.WatchedRepository
}

func NewWatchedService(repo *repository.WatchedRepository) *WatchedService {
	return &WatchedService{repo: repo}
}

func (s *WatchedService) AddWatched(ctx context.Context, userID string, req models.AddWatchedRequest) (*models.Watched, error) {
	watched, err := s.repo.Save(ctx, userID, req)
	if err != nil {
		// Surface known errors so the handler can return the right HTTP status
		if errors.Is(err, repository.ErrAlreadyWatched) ||
			errors.Is(err, repository.ErrInvalidWatchedAt) ||
			errors.Is(err, repository.ErrFutureWatchedAt) {
			return nil, err
		}
		return nil, err
	}
	return watched, nil
}

func (s *WatchedService) GetWatched(ctx context.Context, userID string) ([]models.Watched, error) {
	return s.repo.GetAllByUser(ctx, userID)
}

func (s *WatchedService) RemoveWatched(ctx context.Context, userID string, movieID int) error {
	return s.repo.Delete(ctx, userID, movieID)
}
