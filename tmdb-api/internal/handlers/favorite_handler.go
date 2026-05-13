// internal/handlers/favorite_handler.go
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

// FavoriteHandler handles HTTP requests for the favorites feature.
type FavoriteHandler struct {
	favoriteService *services.FavoriteService
}

// NewFavoriteHandler is the constructor.
func NewFavoriteHandler(favoriteService *services.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favoriteService: favoriteService}
}

// AddFavorite handles POST /favorites
func (h *FavoriteHandler) AddFavorite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req models.AddFavoriteRequest

	// ShouldBindJSON parses the request body AND validates binding tags.
	// If movie_id or title are missing, it returns a 400 automatically.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(), // Safe to expose — this is a validation error,
			// not an internal system error.
		})
		return
	}

	favorite, err := h.favoriteService.AddFavorite(c.Request.Context(), req, userID.(string))
	if err != nil {
		// Check for the specific duplicate error and return 409 Conflict.
		// All other errors are internal — return 500.
		if errors.Is(err, repository.ErrDuplicateFavorite) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "this movie is already in your favorites",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save favorite",
		})
		return
	}

	// 201 Created is the correct status for a successful resource creation.
	// 200 OK is for successful reads. Using the right codes matters.
	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   favorite,
	})
}

// GetFavorites handles GET /favorites
func (h *FavoriteHandler) GetFavorites(c *gin.Context) {

	favorites, err := h.favoriteService.GetFavorites(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve favorites",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(favorites),
		"data":   favorites,
	})
}

func (h *FavoriteHandler) GetFavoriteByUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	favorites, err := h.favoriteService.GetFavoriteByUser(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve user favortites",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(favorites),
		"data":   favorites,
	})
}

// DeleteFavorite handles DELETE /favorites/:movie_id
func (h *FavoriteHandler) DeleteFavorite(c *gin.Context) {
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

	if err := h.favoriteService.RemoveFavorite(c.Request.Context(), userID.(string), movieID); err != nil {
		if errors.Is(err, repository.ErrFavoriteNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found in your favorites"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove favorite"})
		return
	}

	// 204 No Content — deleted successfully, nothing to return
	c.Status(http.StatusNoContent)
}
