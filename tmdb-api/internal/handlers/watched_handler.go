// internal/handlers/watched_handler.go
package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/alainforce/streaming/tmdb-api/internal/repository"
	"github.com/alainforce/streaming/tmdb-api/internal/services"
	"github.com/gin-gonic/gin"
)

type WatchedHandler struct {
	watchedService *services.WatchedService
}

func NewWatchedHandler(watchedService *services.WatchedService) *WatchedHandler {
	return &WatchedHandler{watchedService: watchedService}
}

// AddWatched handles POST /watched
func (h *WatchedHandler) AddWatched(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req models.AddWatchedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	// Validate personal_rating range if provided
	if req.PersonalRating != nil && (*req.PersonalRating < 1 || *req.PersonalRating > 10) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "personal_rating must be between 1 and 10"})
		return
	}

	watched, err := h.watchedService.AddWatched(c.Request.Context(), userID.(string), req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAlreadyWatched):
			c.JSON(http.StatusConflict, gin.H{"error": "this movie is already in your watched list"})
		case errors.Is(err, repository.ErrInvalidWatchedAt):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, repository.ErrFutureWatchedAt):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add to watched list"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": watched})
}

// GetWatched handles GET /watched
func (h *WatchedHandler) GetWatched(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	watched, err := h.watchedService.GetWatched(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve watched list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "count": len(watched), "data": watched})
}

// DeleteWatched handles DELETE /watched/:movie_id
func (h *WatchedHandler) DeleteWatched(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	movieID, err := strconv.Atoi(c.Param("movie_id"))
	if err != nil || movieID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie_id"})
		return
	}

	if err := h.watchedService.RemoveWatched(c.Request.Context(), userID.(string), movieID); err != nil {
		if errors.Is(err, repository.ErrWatchedNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found in your watched list"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove from watched list"})
		return
	}

	c.Status(http.StatusNoContent)
}
