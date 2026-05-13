// internal/handlers/settings_handler.go
package handlers

import (
	"errors"
	"net/http"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/alainforce/streaming/tmdb-api/internal/repository"
	"github.com/alainforce/streaming/tmdb-api/internal/services"
	"github.com/gin-gonic/gin"
)

// SettingsHandler handles all user self-service settings endpoints.
type SettingsHandler struct {
	settingsService *services.SettingsService
}

func NewSettingsHandler(settingsService *services.SettingsService) *SettingsHandler {
	return &SettingsHandler{settingsService: settingsService}
}

// GetProfile handles GET /settings/profile
// Returns the user's profile info and personal activity stats
// in a single response — one call, everything the settings page needs.
func (h *SettingsHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	profile, err := h.settingsService.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// This should never happen if the JWT is valid —
			// it would mean the user was deleted after their token was issued.
			// We handle it anyway so the frontend gets a clean 404
			// rather than a mysterious 500.
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user account no longer exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   profile,
	})
}

// UpdateEmail handles PATCH /settings/email
// Updates the authenticated user's email address.
func (h *SettingsHandler) UpdateEmail(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req models.UpdateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	updatedUser, err := h.settingsService.UpdateEmail(
		c.Request.Context(),
		userID.(string),
		req.Email,
	)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "this email address is already in use",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update email",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "email updated successfully",
		"data":    updatedUser,
	})
}

// UpdatePassword handles PATCH /settings/password
// Updates the authenticated user's password.
func (h *SettingsHandler) UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req models.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.settingsService.UpdatePassword(
		c.Request.Context(),
		userID.(string),
		req.NewPassword,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update password",
		})
		return
	}

	// 200 with a message — no sensitive data in the response.
	// Do NOT return the user object here, there's nothing to update
	// on the client side after a password change.
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "password updated successfully",
	})
}
