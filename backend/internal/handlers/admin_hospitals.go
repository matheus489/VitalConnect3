package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
	"github.com/vitalconnect/backend/internal/services/audit"
	"github.com/vitalconnect/backend/internal/services/auth"
)

var adminHospitalRepo *repository.AdminHospitalRepository

// SetAdminHospitalRepository sets the admin hospital repository for handlers
func SetAdminHospitalRepository(repo *repository.AdminHospitalRepository) {
	adminHospitalRepo = repo
}

// AdminListHospitals returns all hospitals with pagination (cross-tenant, optional tenant filter)
// GET /api/v1/admin/hospitals
func AdminListHospitals(c *gin.Context) {
	if adminHospitalRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin hospital repository not configured"})
		return
	}

	var params repository.AdminHospitalListParams
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

	result, err := adminHospitalRepo.ListAllHospitals(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list hospitals"})
		return
	}

	// Convert to response format
	response := make([]models.HospitalWithTenantResponse, 0, len(result.Hospitals))
	for _, h := range result.Hospitals {
		response = append(response, h.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        response,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
	})
}

// AdminGetHospital returns a single hospital by ID (cross-tenant)
// GET /api/v1/admin/hospitals/:id
func AdminGetHospital(c *gin.Context) {
	if adminHospitalRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin hospital repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hospital ID format"})
		return
	}

	hospital, err := adminHospitalRepo.GetHospitalByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAdminHospitalNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get hospital"})
		return
	}

	c.JSON(http.StatusOK, hospital.ToResponse())
}

// AdminUpdateHospital updates a hospital's details (cross-tenant)
// PUT /api/v1/admin/hospitals/:id
func AdminUpdateHospital(c *gin.Context) {
	if adminHospitalRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin hospital repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hospital ID format"})
		return
	}

	var input models.UpdateHospitalInput
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

	// Get existing hospital for audit
	existingHospital, err := adminHospitalRepo.GetHospitalByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAdminHospitalNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get hospital"})
		return
	}

	// Update hospital
	hospital, err := adminHospitalRepo.UpdateHospital(c.Request.Context(), id, &input)
	if err != nil {
		if errors.Is(err, repository.ErrAdminHospitalNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update hospital"})
		return
	}

	// Log audit event
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		detalhes := map[string]interface{}{
			"hospital_id":   id.String(),
			"hospital_nome": existingHospital.Nome,
		}
		if input.Nome != nil {
			detalhes["nome_anterior"] = existingHospital.Nome
			detalhes["nome_novo"] = *input.Nome
		}
		if input.Ativo != nil {
			detalhes["ativo_anterior"] = existingHospital.Ativo
			detalhes["ativo_novo"] = *input.Ativo
		}

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			"admin.hospital.update",
			"Hospital",
			id.String(),
			nil,
			models.SeverityInfo,
			detalhes,
			ipAddress,
			userAgent,
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "hospital updated successfully",
		"hospital": hospital.ToResponse(),
	})
}

// AdminReassignHospitalTenant reassigns a hospital to a different tenant
// PUT /api/v1/admin/hospitals/:id/reassign
func AdminReassignHospitalTenant(c *gin.Context) {
	if adminHospitalRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin hospital repository not configured"})
		return
	}

	idParam := c.Param("id")
	hospitalID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hospital ID format"})
		return
	}

	var input models.AdminReassignHospitalInput
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

	// Get existing hospital for audit and validation
	existingHospital, err := adminHospitalRepo.GetHospitalByID(c.Request.Context(), hospitalID)
	if err != nil {
		if errors.Is(err, repository.ErrAdminHospitalNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get hospital"})
		return
	}

	// Check if already in the target tenant
	if existingHospital.TenantID == input.TenantID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hospital is already in the target tenant"})
		return
	}

	// Reassign hospital
	hospital, err := adminHospitalRepo.ReassignHospitalTenant(c.Request.Context(), hospitalID, input.TenantID)
	if err != nil {
		if errors.Is(err, repository.ErrAdminHospitalNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
			return
		}
		if err.Error() == "target tenant not found" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "target tenant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reassign hospital"})
		return
	}

	// Log audit event
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		detalhes := map[string]interface{}{
			"hospital_id":         hospitalID.String(),
			"hospital_nome":       existingHospital.Nome,
			"tenant_id_anterior":  existingHospital.TenantID.String(),
			"tenant_nome_anterior": existingHospital.TenantName,
			"tenant_id_novo":      input.TenantID.String(),
		}

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			auth.ActionHospitalReassign,
			"Hospital",
			hospitalID.String(),
			nil,
			models.SeverityWarn,
			detalhes,
			ipAddress,
			userAgent,
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "hospital reassigned successfully",
		"hospital": hospital.ToResponse(),
	})
}
