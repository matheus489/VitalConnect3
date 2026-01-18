package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireSuperAdmin is a middleware that requires the user to be a super admin.
// It checks the IsSuperAdmin field from UserClaims set by the AuthRequired middleware.
// Returns 403 Forbidden if the user is not a super admin.
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
				"code":  "AUTH_REQUIRED",
			})
			return
		}

		userClaims, ok := claims.(*UserClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "invalid claims format",
				"code":  "INVALID_CLAIMS",
			})
			return
		}

		if !userClaims.IsSuperAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "super admin access required",
				"code":  "SUPER_ADMIN_REQUIRED",
			})
			return
		}

		c.Next()
	}
}
