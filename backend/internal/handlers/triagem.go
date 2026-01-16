package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
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

	c.JSON(http.StatusOK, rule.ToResponse())
}
