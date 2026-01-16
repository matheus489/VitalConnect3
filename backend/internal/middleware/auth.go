package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vitalconnect/backend/internal/services/auth"
)

// UserClaims represents the JWT claims for a user
type UserClaims struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	HospitalID string `json:"hospital_id,omitempty"`
}

// contextKey is the key used to store JWT service in context
const jwtServiceKey = "jwt_service"

// SetJWTService stores the JWT service in the Gin context for use by middlewares
func SetJWTService(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(jwtServiceKey, jwtService)
		c.Next()
	}
}

// GetJWTService retrieves the JWT service from context
func GetJWTService(c *gin.Context) (*auth.JWTService, bool) {
	service, exists := c.Get(jwtServiceKey)
	if !exists {
		return nil, false
	}
	jwtService, ok := service.(*auth.JWTService)
	return jwtService, ok
}

// AuthRequired is a middleware that requires a valid JWT token
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get JWT service from context
		jwtService, ok := GetJWTService(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "authentication service not configured",
			})
			return
		}

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token required",
			})
			return
		}

		// Validate the access token
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			switch err {
			case auth.ErrExpiredToken:
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "token has expired",
					"code":  "TOKEN_EXPIRED",
				})
			case auth.ErrInvalidToken, auth.ErrInvalidClaims:
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "invalid token",
					"code":  "INVALID_TOKEN",
				})
			default:
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "authentication failed",
				})
			}
			return
		}

		// Set user claims in context
		userClaims := &UserClaims{
			UserID:     claims.UserID,
			Email:      claims.Email,
			Role:       claims.Role,
			HospitalID: claims.HospitalID,
		}
		c.Set("user_claims", userClaims)

		c.Next()
	}
}

// RequireRole is a middleware that requires specific roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		userClaims, ok := claims.(*UserClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "invalid claims format",
			})
			return
		}

		// Check if user role is in allowed roles
		roleAllowed := false
		for _, role := range roles {
			if userClaims.Role == role {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":         "insufficient permissions",
				"required_role": roles,
				"user_role":     userClaims.Role,
			})
			return
		}

		c.Next()
	}
}

// GetUserClaims extracts user claims from context
func GetUserClaims(c *gin.Context) (*UserClaims, bool) {
	claims, exists := c.Get("user_claims")
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*UserClaims)
	return userClaims, ok
}

// OptionalAuth is a middleware that optionally validates JWT token
// It doesn't block the request if no token is provided
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get JWT service from context
		jwtService, ok := GetJWTService(c)
		if !ok {
			c.Next()
			return
		}

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.Next()
			return
		}

		// Validate the access token
		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Set user claims in context
		userClaims := &UserClaims{
			UserID:     claims.UserID,
			Email:      claims.Email,
			Role:       claims.Role,
			HospitalID: claims.HospitalID,
		}
		c.Set("user_claims", userClaims)

		c.Next()
	}
}
