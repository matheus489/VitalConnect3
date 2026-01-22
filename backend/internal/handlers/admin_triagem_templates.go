package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
	"github.com/sidot/backend/internal/repository"
	"github.com/sidot/backend/internal/services/audit"
)

var adminTriagemRepo *repository.AdminTriagemTemplateRepository

// SetAdminTriagemTemplateRepository sets the admin triagem template repository for handlers
func SetAdminTriagemTemplateRepository(repo *repository.AdminTriagemTemplateRepository) {
	adminTriagemRepo = repo
}

// AdminListTriagemTemplates returns all triagem rule templates with pagination
// GET /api/v1/admin/triagem-templates
func AdminListTriagemTemplates(c *gin.Context) {
	if adminTriagemRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin triagem template repository not configured"})
		return
	}

	var params repository.AdminTriagemTemplateListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Parse ativo from query if provided
	if c.Query("ativo") != "" {
		ativo := c.Query("ativo") == "true"
		params.Ativo = &ativo
	}

	result, err := adminTriagemRepo.ListTemplates(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list triagem templates"})
		return
	}

	// Convert to response format
	response := make([]models.TriagemRuleTemplateWithUsageResponse, 0, len(result.Templates))
	for _, t := range result.Templates {
		response = append(response, t.ToWithUsageResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        response,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
	})
}

// AdminGetTriagemTemplate returns a single triagem template by ID
// GET /api/v1/admin/triagem-templates/:id
func AdminGetTriagemTemplate(c *gin.Context) {
	if adminTriagemRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin triagem template repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID format"})
		return
	}

	template, err := adminTriagemRepo.GetTemplateByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAdminTriagemTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "triagem template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get triagem template"})
		return
	}

	c.JSON(http.StatusOK, template.ToWithUsageResponse())
}

// AdminCreateTriagemTemplate creates a new triagem rule template
// POST /api/v1/admin/triagem-templates
func AdminCreateTriagemTemplate(c *gin.Context) {
	if adminTriagemRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin triagem template repository not configured"})
		return
	}

	var input models.CreateTriagemRuleTemplateInput
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

	template, err := adminTriagemRepo.CreateTemplate(c.Request.Context(), &input)
	if err != nil {
		if errors.Is(err, models.ErrInvalidTriagemRuleTemplateType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create triagem template",
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
			"admin.triagem_template.create",
			"TriagemRuleTemplate",
			template.ID.String(),
			nil,
			models.SeverityInfo,
			map[string]interface{}{
				"template_nome": template.Nome,
				"template_tipo": string(template.Tipo),
			},
			ipAddress,
			userAgent,
		)
	}

	c.JSON(http.StatusCreated, template.ToResponse())
}

// AdminUpdateTriagemTemplate updates an existing triagem template
// PUT /api/v1/admin/triagem-templates/:id
func AdminUpdateTriagemTemplate(c *gin.Context) {
	if adminTriagemRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin triagem template repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID format"})
		return
	}

	var input models.UpdateTriagemRuleTemplateInput
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

	template, err := adminTriagemRepo.UpdateTemplate(c.Request.Context(), id, &input)
	if err != nil {
		if errors.Is(err, repository.ErrAdminTriagemTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "triagem template not found"})
			return
		}
		if errors.Is(err, models.ErrInvalidTriagemRuleTemplateType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update triagem template",
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
			"admin.triagem_template.update",
			"TriagemRuleTemplate",
			template.ID.String(),
			nil,
			models.SeverityInfo,
			map[string]interface{}{
				"template_nome": template.Nome,
			},
			ipAddress,
			userAgent,
		)
	}

	c.JSON(http.StatusOK, template.ToResponse())
}

// AdminCloneTriagemTemplate clones a template to one or more tenants
// POST /api/v1/admin/triagem-templates/:id/clone
func AdminCloneTriagemTemplate(c *gin.Context) {
	if adminTriagemRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin triagem template repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID format"})
		return
	}

	var input models.CloneTriagemRuleTemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate input
	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation failed",
			"details": err.Error(),
		})
		return
	}

	result, err := adminTriagemRepo.CloneToTenant(c.Request.Context(), id, input.TenantIDs)
	if err != nil {
		if errors.Is(err, repository.ErrAdminTriagemTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "triagem template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to clone triagem template",
			"details": err.Error(),
		})
		return
	}

	// Log audit event
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		tenantIDStrs := make([]string, 0, len(result.ClonedToTenants))
		for _, tid := range result.ClonedToTenants {
			tenantIDStrs = append(tenantIDStrs, tid.String())
		}

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			"admin.triagem_template.clone",
			"TriagemRuleTemplate",
			id.String(),
			nil,
			models.SeverityInfo,
			map[string]interface{}{
				"cloned_to_tenants": tenantIDStrs,
				"success_count":     result.SuccessCount,
			},
			ipAddress,
			userAgent,
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "template cloned successfully",
		"template_id":       result.TemplateID,
		"cloned_to_tenants": result.ClonedToTenants,
		"success_count":     result.SuccessCount,
		"failed_tenants":    result.FailedTenants,
	})
}

// AdminGetTriagemTemplateUsage returns which tenants use a template
// GET /api/v1/admin/triagem-templates/:id/usage
func AdminGetTriagemTemplateUsage(c *gin.Context) {
	if adminTriagemRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin triagem template repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID format"})
		return
	}

	usage, err := adminTriagemRepo.GetTemplateUsage(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAdminTriagemTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "triagem template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get template usage"})
		return
	}

	c.JSON(http.StatusOK, usage)
}
