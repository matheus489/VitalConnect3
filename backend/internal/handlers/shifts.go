package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/middleware"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
)

// ShiftHandler handles shift-related HTTP requests
type ShiftHandler struct {
	shiftRepo *repository.ShiftRepository
	userRepo  *repository.UserRepository
}

// NewShiftHandler creates a new shift handler
func NewShiftHandler(shiftRepo *repository.ShiftRepository, userRepo *repository.UserRepository) *ShiftHandler {
	return &ShiftHandler{
		shiftRepo: shiftRepo,
		userRepo:  userRepo,
	}
}

// Create creates a new shift
// @Summary Create a new shift
// @Description Create a new shift schedule
// @Tags shifts
// @Accept json
// @Produce json
// @Param shift body models.CreateShiftInput true "Shift data"
// @Success 201 {object} models.ShiftResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/shifts [post]
func (h *ShiftHandler) Create(c *gin.Context) {
	// Check permission
	claims, exists := middleware.GetUserClaims(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	if claims.Role != string(models.RoleAdmin) && claims.Role != string(models.RoleGestor) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para criar escalas"})
		return
	}

	var input models.CreateShiftInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Gestor can only create shifts for their hospital
	if claims.Role == string(models.RoleGestor) && claims.HospitalID != "" {
		claimHospitalID, err := uuid.Parse(claims.HospitalID)
		if err == nil && input.HospitalID != claimHospitalID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Gestores só podem criar escalas do próprio hospital"})
			return
		}
	}

	shift, err := h.shiftRepo.Create(c.Request.Context(), &input)
	if err != nil {
		if err == models.ErrShiftExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Já existe uma escala para este operador neste horário"})
			return
		}
		if err == models.ErrInvalidDayOfWeek {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dia da semana inválido"})
			return
		}
		if err == models.ErrInvalidStartTime {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Horário de início inválido (use formato HH:MM)"})
			return
		}
		if err == models.ErrInvalidEndTime {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Horário de fim inválido (use formato HH:MM)"})
			return
		}
		// Log the actual error for debugging
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar escala: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, shift.ToResponse())
}

// GetByID retrieves a shift by ID
// @Summary Get shift by ID
// @Description Get a shift by its ID
// @Tags shifts
// @Produce json
// @Param id path string true "Shift ID"
// @Success 200 {object} models.ShiftResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/shifts/{id} [get]
func (h *ShiftHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	shift, err := h.shiftRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == models.ErrShiftNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Escala não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar escala"})
		return
	}

	c.JSON(http.StatusOK, shift.ToResponse())
}

// Update updates a shift
// @Summary Update a shift
// @Description Update an existing shift
// @Tags shifts
// @Accept json
// @Produce json
// @Param id path string true "Shift ID"
// @Param shift body models.UpdateShiftInput true "Shift data"
// @Success 200 {object} models.ShiftResponse
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/shifts/{id} [put]
func (h *ShiftHandler) Update(c *gin.Context) {
	claims, exists := middleware.GetUserClaims(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	if claims.Role != string(models.RoleAdmin) && claims.Role != string(models.RoleGestor) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para editar escalas"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Check if gestor owns the shift's hospital
	if claims.Role == string(models.RoleGestor) {
		existingShift, err := h.shiftRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			if err == models.ErrShiftNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Escala não encontrada"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar escala"})
			return
		}
		if claims.HospitalID != "" {
			claimHospitalID, _ := uuid.Parse(claims.HospitalID)
			if existingShift.HospitalID != claimHospitalID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Gestores só podem editar escalas do próprio hospital"})
				return
			}
		}
	}

	var input models.UpdateShiftInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shift, err := h.shiftRepo.Update(c.Request.Context(), id, &input)
	if err != nil {
		if err == models.ErrShiftNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Escala não encontrada"})
			return
		}
		if err == models.ErrShiftExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Conflito de horário com outra escala"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar escala"})
		return
	}

	c.JSON(http.StatusOK, shift.ToResponse())
}

// Delete deletes a shift
// @Summary Delete a shift
// @Description Delete a shift by ID
// @Tags shifts
// @Param id path string true "Shift ID"
// @Success 204
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/shifts/{id} [delete]
func (h *ShiftHandler) Delete(c *gin.Context) {
	claims, exists := middleware.GetUserClaims(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	if claims.Role != string(models.RoleAdmin) && claims.Role != string(models.RoleGestor) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para excluir escalas"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Check if gestor owns the shift's hospital
	if claims.Role == string(models.RoleGestor) {
		existingShift, err := h.shiftRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			if err == models.ErrShiftNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Escala não encontrada"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar escala"})
			return
		}
		if claims.HospitalID != "" {
			claimHospitalID, _ := uuid.Parse(claims.HospitalID)
			if existingShift.HospitalID != claimHospitalID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Gestores só podem excluir escalas do próprio hospital"})
				return
			}
		}
	}

	err = h.shiftRepo.Delete(c.Request.Context(), id)
	if err != nil {
		if err == models.ErrShiftNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Escala não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir escala"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListByHospital lists all shifts for a hospital
// @Summary List shifts by hospital
// @Description Get all shifts for a specific hospital
// @Tags shifts
// @Produce json
// @Param hospital_id path string true "Hospital ID"
// @Success 200 {array} models.ShiftResponse
// @Router /api/v1/hospitals/{hospital_id}/shifts [get]
func (h *ShiftHandler) ListByHospital(c *gin.Context) {
	hospitalIDStr := c.Param("id")
	hospitalID, err := uuid.Parse(hospitalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do hospital inválido"})
		return
	}

	shifts, err := h.shiftRepo.ListByHospitalID(c.Request.Context(), hospitalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar escalas"})
		return
	}

	responses := make([]models.ShiftResponse, len(shifts))
	for i, shift := range shifts {
		responses[i] = shift.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
}

// GetMyShifts returns shifts for the current user
// @Summary Get my shifts
// @Description Get shifts for the currently authenticated user
// @Tags shifts
// @Produce json
// @Success 200 {array} models.ShiftResponse
// @Router /api/v1/shifts/me [get]
func (h *ShiftHandler) GetMyShifts(c *gin.Context) {
	claims, exists := middleware.GetUserClaims(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autorizado"})
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID de usuário inválido"})
		return
	}

	shifts, err := h.shiftRepo.GetShiftsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar escalas"})
		return
	}

	responses := make([]models.ShiftResponse, len(shifts))
	for i, shift := range shifts {
		responses[i] = shift.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
}

// GetTodayShifts returns today's shifts for a hospital
// @Summary Get today's shifts
// @Description Get all shifts scheduled for today for a hospital
// @Tags shifts
// @Produce json
// @Param hospital_id path string true "Hospital ID"
// @Success 200 {array} object
// @Router /api/v1/hospitals/{hospital_id}/shifts/today [get]
func (h *ShiftHandler) GetTodayShifts(c *gin.Context) {
	hospitalIDStr := c.Param("id")
	hospitalID, err := uuid.Parse(hospitalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do hospital inválido"})
		return
	}

	shifts, err := h.shiftRepo.GetTodayShifts(c.Request.Context(), hospitalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar escalas de hoje"})
		return
	}

	responses := make([]map[string]interface{}, len(shifts))
	for i, shift := range shifts {
		responses[i] = shift.ToTodayResponse()
	}

	c.JSON(http.StatusOK, responses)
}

// GetCoverageGaps returns coverage analysis for a hospital
// @Summary Get coverage gaps
// @Description Analyze shift coverage and find gaps for a hospital
// @Tags shifts
// @Produce json
// @Param hospital_id path string true "Hospital ID"
// @Success 200 {object} models.CoverageAnalysis
// @Router /api/v1/hospitals/{hospital_id}/shifts/coverage [get]
func (h *ShiftHandler) GetCoverageGaps(c *gin.Context) {
	hospitalIDStr := c.Param("id")
	hospitalID, err := uuid.Parse(hospitalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do hospital inválido"})
		return
	}

	analysis, err := h.shiftRepo.GetCoverageGaps(c.Request.Context(), hospitalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao analisar cobertura"})
		return
	}

	c.JSON(http.StatusOK, analysis)
}
