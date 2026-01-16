package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
)

var hospitalRepo *repository.HospitalRepository

// SetHospitalRepository sets the hospital repository for handlers
func SetHospitalRepository(repo *repository.HospitalRepository) {
	hospitalRepo = repo
}

// ListHospitals returns all hospitals
// GET /api/v1/hospitals
func ListHospitals(c *gin.Context) {
	if hospitalRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hospital repository not configured"})
		return
	}

	hospitals, err := hospitalRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list hospitals"})
		return
	}

	// Convert to response format
	response := make([]models.HospitalResponse, 0, len(hospitals))
	for _, h := range hospitals {
		response = append(response, h.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"total": len(response),
	})
}

// GetHospital returns a hospital by ID
// GET /api/v1/hospitals/:id
func GetHospital(c *gin.Context) {
	if hospitalRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hospital repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hospital ID format"})
		return
	}

	hospital, err := hospitalRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrHospitalNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get hospital"})
		return
	}

	c.JSON(http.StatusOK, hospital.ToResponse())
}

// CreateHospital creates a new hospital
// POST /api/v1/hospitals
func CreateHospital(c *gin.Context) {
	if hospitalRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hospital repository not configured"})
		return
	}

	var input models.CreateHospitalInput
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

	hospital, err := hospitalRepo.Create(c.Request.Context(), &input)
	if err != nil {
		if errors.Is(err, repository.ErrHospitalExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "hospital with this code already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create hospital"})
		return
	}

	c.JSON(http.StatusCreated, hospital.ToResponse())
}

// UpdateHospital updates a hospital
// PATCH /api/v1/hospitals/:id
func UpdateHospital(c *gin.Context) {
	if hospitalRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hospital repository not configured"})
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

	hospital, err := hospitalRepo.Update(c.Request.Context(), id, &input)
	if err != nil {
		if errors.Is(err, repository.ErrHospitalNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
			return
		}
		if errors.Is(err, repository.ErrHospitalExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "hospital with this code already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update hospital"})
		return
	}

	c.JSON(http.StatusOK, hospital.ToResponse())
}

// DeleteHospital soft deletes a hospital
// DELETE /api/v1/hospitals/:id
func DeleteHospital(c *gin.Context) {
	if hospitalRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hospital repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hospital ID format"})
		return
	}

	err = hospitalRepo.SoftDelete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrHospitalNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete hospital"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "hospital deleted successfully"})
}
