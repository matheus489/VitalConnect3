package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sidot/backend/internal/middleware"
	"github.com/sidot/backend/internal/models"
	"github.com/sidot/backend/internal/repository"
	"github.com/sidot/backend/internal/services/audit"
	"github.com/sidot/backend/internal/services/auth"
)

var adminUserRepo *repository.AdminUserRepository
var impersonateService *auth.ImpersonationService

// SetAdminUserRepository sets the admin user repository for handlers
func SetAdminUserRepository(repo *repository.AdminUserRepository) {
	adminUserRepo = repo
}

// SetImpersonateService sets the impersonation service for handlers
func SetImpersonateService(svc *auth.ImpersonationService) {
	impersonateService = svc
}

// AdminListUsers returns all users with pagination (cross-tenant, optional tenant filter)
// GET /api/v1/admin/users
func AdminListUsers(c *gin.Context) {
	if adminUserRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin user repository not configured"})
		return
	}

	var params repository.AdminUserListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Parse tenant_id from query if provided
	tenantIDStr := c.Query("tenant_id")
	if tenantIDStr != "" {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
			return
		}
		params.TenantID = &tenantID
	}

	result, err := adminUserRepo.ListAllUsers(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	// Convert to response format
	response := make([]models.UserWithTenantResponse, 0, len(result.Users))
	for _, u := range result.Users {
		response = append(response, u.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        response,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
	})
}

// AdminGetUser returns a single user by ID (cross-tenant)
// GET /api/v1/admin/users/:id
func AdminGetUser(c *gin.Context) {
	if adminUserRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin user repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	user, err := adminUserRepo.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAdminUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// AdminImpersonateUser generates a temporary JWT to impersonate a user
// POST /api/v1/admin/users/:id/impersonate
func AdminImpersonateUser(c *gin.Context) {
	if impersonateService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "impersonation service not configured"})
		return
	}

	idParam := c.Param("id")
	targetUserID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Get current admin user claims
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	adminUserID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid admin user ID"})
		return
	}

	// Prevent self-impersonation
	if adminUserID == targetUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot impersonate yourself"})
		return
	}

	// Extract request info for audit
	ipAddress, userAgent := audit.ExtractRequestInfo(c)

	// Generate impersonation token
	result, err := impersonateService.GenerateImpersonationToken(
		c.Request.Context(),
		adminUserID,
		targetUserID,
		ipAddress,
		userAgent,
	)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "target user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate impersonation token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "impersonation token generated",
		"access_token": result.AccessToken,
		"expires_at":   result.ExpiresAt,
		"expires_in":   result.ExpiresIn,
		"user": gin.H{
			"id":    result.User.ID,
			"email": result.User.Email,
			"role":  result.User.Role,
		},
		"warning": "This action has been logged for audit purposes",
	})
}

// AdminUpdateUserRole updates a user's role and/or super admin status
// PUT /api/v1/admin/users/:id/role
func AdminUpdateUserRole(c *gin.Context) {
	if adminUserRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin user repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	var input models.AdminUpdateUserRoleInput
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

	// Get existing user for audit
	existingUser, err := adminUserRepo.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAdminUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// Update user role
	user, err := adminUserRepo.UpdateUserRole(c.Request.Context(), id, input.Role, input.IsSuperAdmin)
	if err != nil {
		if errors.Is(err, repository.ErrAdminUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role"})
		return
	}

	// Log audit event
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		detalhes := map[string]interface{}{
			"target_user_id":    id.String(),
			"target_user_email": existingUser.Email,
		}
		if input.Role != nil {
			detalhes["role_anterior"] = existingUser.Role
			detalhes["role_novo"] = *input.Role
		}
		if input.IsSuperAdmin != nil {
			detalhes["is_super_admin_anterior"] = existingUser.IsSuperAdmin
			detalhes["is_super_admin_novo"] = *input.IsSuperAdmin
		}

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			auth.ActionUserRoleChange,
			"User",
			id.String(),
			nil,
			models.SeverityWarn,
			detalhes,
			ipAddress,
			userAgent,
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user role updated successfully",
		"user":    user.ToResponse(),
	})
}

// AdminBanUser bans or unbans a user (deactivates/activates)
// PUT /api/v1/admin/users/:id/ban
func AdminBanUser(c *gin.Context) {
	if adminUserRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin user repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	var input models.AdminBanUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get current admin user claims
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Prevent self-ban
	if claims.UserID == id.String() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot ban yourself"})
		return
	}

	// Get existing user for audit
	existingUser, err := adminUserRepo.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAdminUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// Determine new status (banned = inactive)
	newActive := !input.Banned

	// Update user status
	user, err := adminUserRepo.UpdateUserStatus(c.Request.Context(), id, newActive, input.BanReason)
	if err != nil {
		if errors.Is(err, repository.ErrAdminUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user status"})
		return
	}

	// Log audit event
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		action := auth.ActionUserBan
		if !input.Banned {
			action = auth.ActionUserUnban
		}

		detalhes := map[string]interface{}{
			"target_user_id":    id.String(),
			"target_user_email": existingUser.Email,
			"ativo_anterior":    existingUser.Ativo,
			"ativo_novo":        newActive,
		}
		if input.BanReason != nil {
			detalhes["ban_reason"] = *input.BanReason
		}

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			action,
			"User",
			id.String(),
			nil,
			models.SeverityWarn,
			detalhes,
			ipAddress,
			userAgent,
		)
	}

	message := "user banned successfully"
	if !input.Banned {
		message = "user unbanned successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"user":    user.ToResponse(),
	})
}

// AdminResetPassword resets a user's password to a temporary password
// POST /api/v1/admin/users/:id/reset-password
func AdminResetPassword(c *gin.Context) {
	if adminUserRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin user repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Get existing user for audit
	existingUser, err := adminUserRepo.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAdminUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// Generate temporary password
	tempPassword := generateTempPassword()

	// Hash the temporary password
	hashedPassword, err := auth.HashPassword(tempPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate password"})
		return
	}

	// Update user password
	err = adminUserRepo.ResetUserPassword(c.Request.Context(), id, hashedPassword)
	if err != nil {
		if errors.Is(err, repository.ErrAdminUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
		return
	}

	// Log audit event
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		detalhes := map[string]interface{}{
			"target_user_id":    id.String(),
			"target_user_email": existingUser.Email,
		}

		detalhesJSON, _ := json.Marshal(detalhes)

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			auth.ActionUserResetPwd,
			"User",
			id.String(),
			nil,
			models.SeverityWarn,
			map[string]interface{}{
				"target_user_id":    id.String(),
				"target_user_email": existingUser.Email,
			},
			ipAddress,
			userAgent,
		)

		_ = detalhesJSON // suppress unused variable
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "password reset successfully",
		"temporary_password": tempPassword,
		"warning":            "Please share this password securely with the user and advise them to change it immediately",
	})
}

// generateTempPassword generates a secure temporary password
func generateTempPassword() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
