// internal/middleware/admin.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireAdmin checks that the authenticated user has role = "admin".
// This middleware MUST be chained AFTER RequireAuth — it relies on the
// "userRole" value that RequireAuth injects into the context.
// If you use RequireAdmin without RequireAuth before it, it will always
// return 403 because the context key will never be set.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "access denied",
			})
			return
		}

		if role.(string) != "admin" {
			// 403 Forbidden — the user is authenticated but not authorised.
			// 401 Unauthorized means "not authenticated" (no/bad token).
			// 403 Forbidden means "authenticated but not allowed".
			// These are different — using them correctly matters.
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			return
		}

		c.Next()
	}
}
