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
	"github.com/vitalconnect/backend/internal/services/audit"
)

var triagemRuleRepo *repository.TriagemRuleRepository

// SetTriagemRuleRepository sets the triagem rule repository for handlers
func SetTriagemRuleRepository(repo *repository.TriagemRuleRepository) {
	triagemRuleRepo = repo
}

// ListTriagemRules returns all triagem rules
// GET /api/v1/triagem-rules
func ListTriagemRules(c *gin.Context) {
	if triagemRuleRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "triagem rule repository not configured"})
		return
	}

	rules, err := triagemRuleRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list triagem rules"})
		return
	}

	// Convert to response format
	response := make([]models.TriagemRuleResponse, 0, len(rules))
	for _, r := range rules {
		response = append(response, r.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"total": len(response),
	})
}

// CreateTriagemRule creates a new triagem rule
// POST /api/v1/triagem-rules
func CreateTriagemRule(c *gin.Context) {
	if triagemRuleRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "triagem rule repository not configured"})
		return
	}

	var input models.CreateTriagemRuleInput
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

	rule, err := triagemRuleRepo.Create(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create triagem rule"})
		return
	}

	// Log audit event for rule creation
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		// Get hospital_id from context if available
		var hospitalID *uuid.UUID
		claims, _ := middleware.GetUserClaims(c)
		if claims != nil && claims.HospitalID != "" {
			hid, _ := uuid.Parse(claims.HospitalID)
			hospitalID = &hid
		}

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			models.ActionRegraCreate,
			"Regra",
			rule.ID.String(),
			hospitalID,
			models.SeverityInfo,
			map[string]interface{}{
				"nome":       rule.Nome,
				"prioridade": rule.Prioridade,
			},
			ipAddress,
			userAgent,
		)
	}

	c.JSON(http.StatusCreated, rule.ToResponse())
}

// UpdateTriagemRule updates a triagem rule
// PATCH /api/v1/triagem-rules/:id
func UpdateTriagemRule(c *gin.Context) {
	if triagemRuleRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "triagem rule repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid triagem rule ID format"})
		return
	}

	// Get the old rule for audit comparison
	oldRule, err := triagemRuleRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrTriagemRuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "triagem rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get triagem rule"})
		return
	}

	var input models.UpdateTriagemRuleInput
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

	rule, err := triagemRuleRepo.Update(c.Request.Context(), id, &input)
	if err != nil {
		if errors.Is(err, repository.ErrTriagemRuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "triagem rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update triagem rule"})
		return
	}

	// Log audit event for rule update (CRITICAL severity as per spec)
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		// Get hospital_id from context if available
		var hospitalID *uuid.UUID
		claims, _ := middleware.GetUserClaims(c)
		if claims != nil && claims.HospitalID != "" {
			hid, _ := uuid.Parse(claims.HospitalID)
			hospitalID = &hid
		}

		// Build details with changes
		detalhes := map[string]interface{}{
			"nome_anterior": oldRule.Nome,
			"nome_novo":     rule.Nome,
		}

		// Track specific changes that are critical
		if input.Regras != nil {
			detalhes["regras_alteradas"] = true
		}
		if input.Ativo != nil {
			detalhes["ativo_anterior"] = oldRule.Ativo
			detalhes["ativo_novo"] = rule.Ativo
		}

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			models.ActionRegraUpdate,
			"Regra",
			rule.ID.String(),
			hospitalID,
			models.SeverityCritical,
			detalhes,
			ipAddress,
			userAgent,
		)
	}

	c.JSON(http.StatusOK, rule.ToResponse())
}

// DeleteTriagemRule performs a soft delete on a triagem rule
// DELETE /api/v1/triagem-rules/:id
func DeleteTriagemRule(c *gin.Context) {
	if triagemRuleRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "triagem rule repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid triagem rule ID format"})
		return
	}

	// Get the rule for audit log
	rule, err := triagemRuleRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrTriagemRuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "triagem rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get triagem rule"})
		return
	}

	// Perform soft delete
	err = triagemRuleRepo.SoftDelete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrTriagemRuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "triagem rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete triagem rule"})
		return
	}

	// Log audit event for rule deletion
	if auditService != nil {
		userID, actorName := audit.GetUserInfoFromContext(c)
		ipAddress, userAgent := audit.ExtractRequestInfo(c)

		var hospitalID *uuid.UUID
		claims, _ := middleware.GetUserClaims(c)
		if claims != nil && claims.HospitalID != "" {
			hid, _ := uuid.Parse(claims.HospitalID)
			hospitalID = &hid
		}

		auditService.LogEventWithUser(
			c.Request.Context(),
			userID,
			actorName,
			models.ActionRegraUpdate, // Using update action since it's a soft delete
			"Regra",
			rule.ID.String(),
			hospitalID,
			models.SeverityCritical,
			map[string]interface{}{
				"nome":   rule.Nome,
				"action": "soft_delete",
			},
			ipAddress,
			userAgent,
		)
	}

	c.Status(http.StatusNoContent)
}
