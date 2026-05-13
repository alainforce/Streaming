// internal/handlers/auth_handler.go
package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/alainforce/streaming/tmdb-api/internal/models"
	"github.com/alainforce/streaming/tmdb-api/internal/repository"
	"github.com/alainforce/streaming/tmdb-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Signup handles POST /auth/signup
func (h *AuthHandler) Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.authService.Signup(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "an account with this email already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create account",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			// 401 Unauthorized is the correct code for bad credentials.
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid email or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to log in",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// internal/handlers/auth_handler.go
// Add this method to the existing AuthHandler — everything else stays the same.

// Logout handles POST /auth/logout
// The RequireAuth middleware has already validated the token and injected
// the jti and expiry into the Gin context — we just read them here.
func (h *AuthHandler) Logout(c *gin.Context) {
	// Read the jti injected by RequireAuth middleware.
	jti, exists := c.Get("jti")
	if !exists {
		// This should never happen if RequireAuth ran correctly,
		// but we guard against it defensively.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process logout"})
		return
	}

	// Read the token expiry injected by RequireAuth.
	// We use it to compute the remaining TTL for the Redis blacklist entry.
	tokenExp, exists := c.Get("tokenExp")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process logout"})
		return
	}

	// Compute remaining lifetime so the blacklist entry expires
	// exactly when the token would have expired.
	expTime := tokenExp.(*jwt.NumericDate)
	remainingTTL := time.Until(expTime.Time)

	if err := h.authService.Logout(c.Request.Context(), jti.(string), remainingTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log out"})
		return
	}

	// 204 No Content — logged out, nothing to return.
	c.Status(http.StatusNoContent)
}

// DeleteAccount handles DELETE /auth/account
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	// The auth middleware sets "userID" in the Gin context.
	// We retrieve it here — no need to re-parse the JWT.
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.authService.DeleteAccount(c.Request.Context(), userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete account",
		})
		return
	}

	// 204 No Content — success, but nothing to return.
	// The account is gone. Don't return a body.
	c.Status(http.StatusNoContent)
}
