package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sidot/backend/internal/middleware"
	"github.com/sidot/backend/internal/repository"
)

var indicatorsRepo *repository.IndicatorsRepository

// SetIndicatorsRepository sets the indicators repository for handlers
func SetIndicatorsRepository(repo *repository.IndicatorsRepository) {
	indicatorsRepo = repo
}

// GetIndicators returns dashboard indicators metrics
// GET /api/v1/metrics/indicators
//
// Query params:
// - hospital_id (optional, UUID): Filter by hospital (ignored for operador role)
//
// Permissions:
// - admin/gestor: Can view all data or filter by hospital
// - operador: Automatically filtered by their hospital_id
func GetIndicators(c *gin.Context) {
	if indicatorsRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "indicators repository not configured"})
		return
	}

	ctx := c.Request.Context()

	// Get user claims from context
	claims, exists := middleware.GetUserClaims(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Determine hospital_id filter based on role
	var hospitalID *uuid.UUID

	// For operador role, always filter by their hospital_id
	if claims.Role == "operador" {
		if claims.HospitalID == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "operador must have a hospital_id assigned",
			})
			return
		}
		hID, err := uuid.Parse(claims.HospitalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid hospital_id in user claims",
			})
			return
		}
		hospitalID = &hID
	} else {
		// For admin/gestor, use optional query param
		hospitalIDParam := c.Query("hospital_id")
		if hospitalIDParam != "" {
			hID, err := uuid.Parse(hospitalIDParam)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid hospital_id format - must be a valid UUID",
				})
				return
			}
			hospitalID = &hID
		}
	}

	// Get all indicators
	metrics, err := indicatorsRepo.GetAllIndicators(ctx, hospitalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to fetch indicators",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
