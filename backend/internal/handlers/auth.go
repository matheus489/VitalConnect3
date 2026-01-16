package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitalconnect/backend/internal/middleware"
	"github.com/vitalconnect/backend/internal/services/auth"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *auth.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=1"`
}

// RefreshRequest represents the refresh token request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest represents the logout request body
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Login handles user authentication
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid email or password",
			})
		case auth.ErrUserInactive:
			c.JSON(http.StatusForbidden, gin.H{
				"error": "user account is inactive",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "authentication failed",
			})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	result, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case auth.ErrExpiredToken:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "refresh token has expired",
				"code":  "TOKEN_EXPIRED",
			})
		case auth.ErrInvalidToken, auth.ErrInvalidClaims:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid refresh token",
				"code":  "INVALID_TOKEN",
			})
		case auth.ErrTokenRevoked:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "refresh token has been revoked",
				"code":  "TOKEN_REVOKED",
			})
		case auth.ErrUserInactive:
			c.JSON(http.StatusForbidden, gin.H{
				"error": "user account is inactive",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "token refresh failed",
			})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Logout always returns success even if token is invalid
	// This prevents information leakage
	_ = h.authService.Logout(c.Request.Context(), req.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// Me returns the current authenticated user
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), claims.UserID)
	if err != nil {
		switch err {
		case auth.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get user information",
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// Global handlers using singleton pattern for backwards compatibility with existing routes

var globalAuthHandler *AuthHandler

// SetGlobalAuthHandler sets the global auth handler
func SetGlobalAuthHandler(handler *AuthHandler) {
	globalAuthHandler = handler
}

// Login is the global handler for login
func Login(c *gin.Context) {
	if globalAuthHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "auth handler not configured",
		})
		return
	}
	globalAuthHandler.Login(c)
}

// RefreshToken is the global handler for token refresh
func RefreshToken(c *gin.Context) {
	if globalAuthHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "auth handler not configured",
		})
		return
	}
	globalAuthHandler.RefreshToken(c)
}

// Logout is the global handler for logout
func Logout(c *gin.Context) {
	if globalAuthHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "auth handler not configured",
		})
		return
	}
	globalAuthHandler.Logout(c)
}

// Me is the global handler for getting current user
func Me(c *gin.Context) {
	if globalAuthHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "auth handler not configured",
		})
		return
	}
	globalAuthHandler.Me(c)
}
