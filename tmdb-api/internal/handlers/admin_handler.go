// internal/handlers/admin_handler.go
package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/alainforce/streaming/tmdb-api/internal/repository"
	"github.com/alainforce/streaming/tmdb-api/internal/services"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// ListUsers handles GET /admin/users?page=1&page_size=20
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, pageSize := parsePagination(c)

	users, meta, err := h.adminService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"data":       users,
		"pagination": meta,
	})
}

// DeleteUser handles DELETE /admin/users/:id
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	if err := h.adminService.DeleteUser(c.Request.Context(), userID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.Status(http.StatusNoContent)
}

// BanUser handles PATCH /admin/users/:id/ban
func (h *AdminHandler) BanUser(c *gin.Context) {
	userID, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	if err := h.adminService.BanUser(c.Request.Context(), userID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ban user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "user has been banned"})
}

// UnbanUser handles PATCH /admin/users/:id/unban
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	userID, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	if err := h.adminService.UnbanUser(c.Request.Context(), userID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unban user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "user has been reinstated"})
}

// GetAllSavedMovies handles GET /admin/movies?page=1&page_size=20
func (h *AdminHandler) GetAllSavedMovies(c *gin.Context) {
	page, pageSize := parsePagination(c)

	movies, meta, err := h.adminService.GetAllSavedMovies(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve saved movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       movies,
		"pagination": meta,
		"status":     "success",
	})
}

// GetStats handles GET /admin/stats
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.adminService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// --- Private helpers ---

// parsePagination reads page and page_size query params with safe defaults.
// Centralising this prevents the same strconv boilerplate in every handler.
func parsePagination(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	return
}

// parseUserID extracts and validates the :id URL parameter.
// Returns false and writes the error response if invalid.
func parseUserID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return 0, false
	}
	return id, true
}
