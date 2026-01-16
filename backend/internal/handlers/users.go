package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/middleware"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
	"github.com/vitalconnect/backend/internal/services/auth"
)

var userRepo *repository.UserRepository

// SetUserRepository sets the user repository for handlers
func SetUserRepository(repo *repository.UserRepository) {
	userRepo = repo
}

// ListUsers returns all users (admin only)
// GET /api/v1/users
func ListUsers(c *gin.Context) {
	if userRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user repository not configured"})
		return
	}

	users, err := userRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	// Convert to response format
	response := make([]models.UserResponse, 0, len(users))
	for _, u := range users {
		response = append(response, u.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"total": len(response),
	})
}

// GetUser returns a user by ID
// GET /api/v1/users/:id
func GetUser(c *gin.Context) {
	if userRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Get current user claims
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check authorization: admin can view any user, others can only view themselves
	if claims.Role != "admin" && claims.UserID != id.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	user, err := userRepo.GetModelByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// CreateUser creates a new user (admin only)
// POST /api/v1/users
func CreateUser(c *gin.Context) {
	if userRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user repository not configured"})
		return
	}

	var input models.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate input
	validate := validator.New()
	if err := validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	// Validate password strength
	if err := auth.ValidatePasswordStrength(input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	user, err := userRepo.CreateUser(c.Request.Context(), &input, passwordHash)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "user with this email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user.ToResponse())
}

// UpdateUser updates a user
// PATCH /api/v1/users/:id
func UpdateUser(c *gin.Context) {
	if userRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Get current user claims
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check authorization
	isAdmin := claims.Role == "admin"
	isSelf := claims.UserID == id.String()

	if !isAdmin && !isSelf {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var input models.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Non-admins cannot change role or ativo status
	if !isAdmin {
		if input.Role != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins can change user roles"})
			return
		}
		if input.Ativo != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins can change user active status"})
			return
		}
	}

	// Validate input
	validate := validator.New()
	if err := validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	// Handle password update
	var passwordHash *string
	if input.Password != nil {
		if err := auth.ValidatePasswordStrength(*input.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		hash, err := auth.HashPassword(*input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
			return
		}
		passwordHash = &hash
	}

	user, err := userRepo.UpdateUser(c.Request.Context(), id, &input, passwordHash)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, repository.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "user with this email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// DeleteUser deactivates a user (admin only)
// DELETE /api/v1/users/:id
func DeleteUser(c *gin.Context) {
	if userRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Get current user claims
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Prevent self-deletion
	if claims.UserID == id.String() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot deactivate your own account"})
		return
	}

	err = userRepo.DeactivateUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deactivated successfully"})
}
