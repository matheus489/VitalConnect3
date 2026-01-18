package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
	"github.com/vitalconnect/backend/internal/services/audit"
)

var adminSettingsRepo *repository.AdminSettingsRepository

// SetAdminSettingsRepository sets the admin settings repository for handlers
func SetAdminSettingsRepository(repo *repository.AdminSettingsRepository) {
	adminSettingsRepo = repo
}

// AdminListSettings returns all system settings
// Encrypted values are masked in the response
// GET /api/v1/admin/settings
func AdminListSettings(c *gin.Context) {
	if adminSettingsRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin settings repository not configured"})
		return
	}

	settings, err := adminSettingsRepo.GetAllSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to list settings",
			"details": err.Error(),
		})
		return
	}

	// Convert to masked response format
	response := make([]models.SystemSettingMaskedResponse, 0, len(settings))
	for _, s := range settings {
		response = append(response, s.ToMaskedResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"total": len(response),
	})
}

// AdminGetSetting returns a single setting by key
// Encrypted values are masked in the response
// GET /api/v1/admin/settings/:key
func AdminGetSetting(c *gin.Context) {
	if adminSettingsRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin settings repository not configured"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "setting key is required"})
		return
	}

	setting, err := adminSettingsRepo.GetSettingByKey(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, repository.ErrAdminSettingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get setting"})
		return
	}

	// Return masked response
	c.JSON(http.StatusOK, setting.ToMaskedResponse())
}

// AdminUpsertSetting creates or updates a system setting
// PUT /api/v1/admin/settings/:key
func AdminUpsertSetting(c *gin.Context) {
	if adminSettingsRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin settings repository not configured"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "setting key is required"})
		return
	}

	var input models.CreateSystemSettingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Override key from URL parameter
	input.Key = key

	// Validate input
	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	// Check if this is a create or update for audit logging
	isCreate := false
	_, err := adminSettingsRepo.GetSettingByKey(c.Request.Context(), key)
	if errors.Is(err, repository.ErrAdminSettingNotFound) {
		isCreate = true
	}

	setting, err := adminSettingsRepo.UpsertSetting(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to save setting",
			"details": err.Error(),
		})
		return
	}

	// Log audit event
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		action := "admin.setting.update"
		if isCreate {
			action = "admin.setting.create"
		}

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			action,
			"SystemSetting",
			key,
			nil,
			models.SeverityInfo,
			map[string]interface{}{
				"key":          key,
				"is_encrypted": setting.IsEncrypted,
			},
			ipAddress,
			userAgent,
		)
	}

	// Return masked response
	c.JSON(http.StatusOK, gin.H{
		"message": "setting saved successfully",
		"setting": setting.ToMaskedResponse(),
	})
}

// AdminDeleteSetting deletes a system setting
// DELETE /api/v1/admin/settings/:key
func AdminDeleteSetting(c *gin.Context) {
	if adminSettingsRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin settings repository not configured"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "setting key is required"})
		return
	}

	// Get the setting first for audit logging
	setting, err := adminSettingsRepo.GetSettingByKey(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, repository.ErrAdminSettingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get setting"})
		return
	}

	err = adminSettingsRepo.DeleteSetting(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, repository.ErrAdminSettingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete setting",
			"details": err.Error(),
		})
		return
	}

	// Log audit event
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			"admin.setting.delete",
			"SystemSetting",
			key,
			nil,
			models.SeverityWarn,
			map[string]interface{}{
				"key":          key,
				"setting_id":   setting.ID.String(),
				"is_encrypted": setting.IsEncrypted,
			},
			ipAddress,
			userAgent,
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "setting deleted successfully",
		"key":     key,
	})
}
