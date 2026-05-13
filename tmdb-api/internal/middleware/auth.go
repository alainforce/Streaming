// internal/middleware/auth.go
package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alainforce/streaming/tmdb-api/internal/cache"
	jwtpkg "github.com/alainforce/streaming/tmdb-api/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// RequireAuth validates the JWT and enforces blacklist + ban checks.
// Three gates, in order:
//  1. Is the JWT signature valid and unexpired?
//  2. Has this specific token been revoked (logged out)?
//  3. Has this user been banned since the token was issued?
func RequireAuth(jwtManager *jwtpkg.Manager, blacklist *cache.TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- Gate 1: Parse and validate the JWT ---
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header format must be: Bearer <token>",
			})
			return
		}

		claims, err := jwtManager.Validate(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Use a short context timeout for Redis checks.
		// We don't want a slow Redis to hold up every single request.
		// 2 seconds is generous — Redis typically responds in <1ms.
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// --- Gate 2: Check if this token has been revoked ---
		blacklisted, err := blacklist.IsTokenBlacklisted(ctx, claims.JTI)
		if err != nil {
			// Redis is down or slow. This is a security-sensitive check —
			// we FAIL CLOSED: reject the request rather than letting
			// a potentially revoked token through.
			// This is the opposite of what we did for the cache in Step 4,
			// where we failed open (served live data if Redis was down).
			// Security checks fail closed. Performance optimisations fail open.
			log.Printf("WARN: blacklist check failed for jti=%s: %v", claims.JTI, err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "authentication service temporarily unavailable",
			})
			return
		}
		if blacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token has been revoked, please log in again",
			})
			return
		}

		// --- Gate 3: Check if the user has been banned ---
		banned, err := blacklist.IsUserBanned(ctx, claims.UserID)
		if err != nil {
			log.Printf("WARN: ban check failed for userID=%s: %v", claims.UserID, err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "authentication service temporarily unavailable",
			})
			return
		}
		if banned {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "your account has been suspended",
			})
			return
		}

		// All gates passed — inject claims into context for downstream handlers.
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("userRole", claims.Role)
		c.Set("jti", claims.JTI)            // Needed by the logout handler
		c.Set("tokenExp", claims.ExpiresAt) // Needed to compute remaining TTL

		c.Next()
	}
}
