package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sidot/backend/internal/middleware"
	"github.com/sidot/backend/internal/models"
	"github.com/sidot/backend/internal/repository"
	"github.com/sidot/backend/internal/services/audit"
	"github.com/sidot/backend/internal/services/auth"
)

var userRepo *repository.UserRepository

// SetUserRepository sets the user repository for handlers
func SetUserRepository(repo *repository.UserRepository) {
	userRepo = repo
}

// ListUsers returns users with pagination, search, and filtering (admin only)
// GET /api/v1/users
func ListUsers(c *gin.Context) {
	if userRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user repository not configured"})
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	search := c.Query("search")
	status := c.DefaultQuery("status", "all")

	// Validate per_page
	if perPage > 100 {
		perPage = 100
	}
	if perPage < 1 {
		perPage = 10
	}

	params := &models.UserListParams{
		Page:    page,
		PerPage: perPage,
		Search:  search,
		Status:  status,
	}

	result, err := userRepo.ListWithPagination(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	// Convert to response format
	response := make([]models.UserResponse, 0, len(result.Users))
	for _, u := range result.Users {
		response = append(response, u.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
		"meta": gin.H{
			"page":        result.Page,
			"per_page":    result.PerPage,
			"total":       result.Total,
			"total_pages": result.TotalPages,
		},
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

	// Validate mobile phone if provided
	if input.MobilePhone != nil && !models.ValidateMobilePhone(*input.MobilePhone) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid mobile phone format, must be in E.164 format (e.g., +5511999999999)",
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

	// Log audit event for user creation
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			models.ActionUsuarioCreate,
			"Usuario",
			user.ID.String(),
			nil, // no hospital_id for user operations
			models.SeverityInfo,
			map[string]interface{}{
				"email": user.Email,
				"role":  user.Role,
			},
			ipAddress,
			userAgent,
		)
	}

	c.JSON(http.StatusCreated, user.ToResponse())
}

// UpdateUser updates a user (admin only for management fields)
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

	// Only admin can update users through this endpoint
	if claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can manage users"})
		return
	}

	// Get existing user for audit comparison
	existingUser, err := userRepo.GetModelByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
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

	// Validate input
	validate := validator.New()
	if err := validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	// Validate mobile phone if provided
	if input.MobilePhone != nil && !models.ValidateMobilePhone(*input.MobilePhone) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid mobile phone format, must be in E.164 format (e.g., +5511999999999)",
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

	// Log audit event for user update
	if auditService != nil {
		userIDForAudit, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		detalhes := map[string]interface{}{}
		if input.Nome != nil {
			detalhes["nome_anterior"] = existingUser.Nome
			detalhes["nome_novo"] = *input.Nome
		}
		if input.Role != nil {
			detalhes["role_anterior"] = existingUser.Role
			detalhes["role_novo"] = *input.Role
		}

		auditService.LogEventWithUser(
			c.Request.Context(),
			userIDForAudit,
			actorName,
			models.ActionUsuarioUpdate,
			"Usuario",
			user.ID.String(),
			nil, // no hospital_id for user operations
			models.SeverityInfo,
			detalhes,
			ipAddress,
			userAgent,
		)
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

	// Get user info for audit before deactivation
	userToDeactivate, _ := userRepo.GetModelByID(c.Request.Context(), id)

	err = userRepo.DeactivateUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate user"})
		return
	}

	// Log audit event for user deactivation (WARN severity)
	if auditService != nil {
		userIDForAudit, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		detalhes := map[string]interface{}{}
		if userToDeactivate != nil {
			detalhes["email"] = userToDeactivate.Email
			detalhes["role"] = userToDeactivate.Role
		}

		auditService.LogEventWithUser(
			c.Request.Context(),
			userIDForAudit,
			actorName,
			models.ActionUsuarioDesativar,
			"Usuario",
			id.String(),
			nil, // no hospital_id for user operations
			models.SeverityWarn,
			detalhes,
			ipAddress,
			userAgent,
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deactivated successfully"})
}

// GetCurrentUser returns the currently authenticated user's profile
// GET /api/v1/users/me
func GetCurrentUser(c *gin.Context) {
	if userRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user repository not configured"})
		return
	}

	// Get current user claims
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID in token"})
		return
	}

	user, err := userRepo.GetModelByID(c.Request.Context(), userID)
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

// UpdateCurrentUser updates the currently authenticated user's profile
// PATCH /api/v1/users/me
func UpdateCurrentUser(c *gin.Context) {
	if userRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user repository not configured"})
		return
	}

	// Get current user claims
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID in token"})
		return
	}

	var input models.UpdateProfileInput
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

	// Handle password change
	var newPasswordHash *string
	if input.NewPassword != nil {
		// Current password is required to change password
		if input.CurrentPassword == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "current_password is required to change password",
			})
			return
		}

		// Verify current password
		user, err := userRepo.GetModelByID(c.Request.Context(), userID)
		if err != nil {
			if errors.Is(err, auth.ErrUserNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
			return
		}

		if err := auth.CheckPasswordHash(*input.CurrentPassword, user.PasswordHash); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "current password is incorrect",
			})
			return
		}

		// Validate new password strength
		if err := auth.ValidatePasswordStrength(*input.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Hash new password
		hash, err := auth.HashPassword(*input.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
			return
		}
		newPasswordHash = &hash
	}

	user, err := userRepo.UpdateProfile(c.Request.Context(), userID, &input, newPasswordHash)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}
