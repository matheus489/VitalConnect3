package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/middleware"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
)

var (
	occurrenceRepo        *repository.OccurrenceRepository
	occurrenceHistoryRepo *repository.OccurrenceHistoryRepository
)

// SetOccurrenceRepository sets the occurrence repository for handlers
func SetOccurrenceRepository(repo *repository.OccurrenceRepository) {
	occurrenceRepo = repo
}

// SetOccurrenceHistoryRepository sets the occurrence history repository for handlers
func SetOccurrenceHistoryRepository(repo *repository.OccurrenceHistoryRepository) {
	occurrenceHistoryRepo = repo
}

// ListOccurrences returns occurrences with pagination and filters
// GET /api/v1/occurrences
func ListOccurrences(c *gin.Context) {
	if occurrenceRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "occurrence repository not configured"})
		return
	}

	// Parse filters from query parameters
	filters := models.DefaultFilters()

	// Status filter
	if status := c.Query("status"); status != "" {
		s := models.OccurrenceStatus(status)
		if !s.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status filter"})
			return
		}
		filters.Status = &s
	}

	// Hospital filter
	if hospitalID := c.Query("hospital_id"); hospitalID != "" {
		filters.HospitalID = &hospitalID
	}

	// Date filters
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		t, err := time.Parse(time.RFC3339, dateFrom)
		if err != nil {
			// Try date-only format
			t, err = time.Parse("2006-01-02", dateFrom)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_from format, use RFC3339 or YYYY-MM-DD"})
				return
			}
		}
		filters.DateFrom = &t
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		t, err := time.Parse(time.RFC3339, dateTo)
		if err != nil {
			// Try date-only format and set to end of day
			t, err = time.Parse("2006-01-02", dateTo)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_to format, use RFC3339 or YYYY-MM-DD"})
				return
			}
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
		filters.DateTo = &t
	}

	// Pagination
	if page := c.Query("page"); page != "" {
		p, err := strconv.Atoi(page)
		if err != nil || p < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
			return
		}
		filters.Page = p
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		ps, err := strconv.Atoi(pageSize)
		if err != nil || ps < 1 || ps > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size (1-100)"})
			return
		}
		filters.PageSize = ps
	}

	// Sorting
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		if sortOrder != "asc" && sortOrder != "desc" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sort_order (asc or desc)"})
			return
		}
		filters.SortOrder = sortOrder
	}

	occurrences, totalItems, err := occurrenceRepo.List(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list occurrences"})
		return
	}

	// Convert to list response format (with masked names)
	response := make([]models.OccurrenceListResponse, 0, len(occurrences))
	for _, o := range occurrences {
		response = append(response, o.ToListResponse())
	}

	c.JSON(http.StatusOK, models.NewPaginatedResponse(response, filters.Page, filters.PageSize, totalItems))
}

// GetOccurrence returns occurrence details with full data (unmasked name)
// GET /api/v1/occurrences/:id
func GetOccurrence(c *gin.Context) {
	if occurrenceRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "occurrence repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid occurrence ID format"})
		return
	}

	occurrence, err := occurrenceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrOccurrenceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "occurrence not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get occurrence"})
		return
	}

	// Log access to complete data for LGPD audit
	claims, _ := middleware.GetUserClaims(c)
	if claims != nil {
		// TODO: Log audit event for LGPD compliance
		_ = claims.UserID
	}

	c.JSON(http.StatusOK, occurrence.ToDetailResponse())
}

// GetOccurrenceHistory returns the action history for an occurrence
// GET /api/v1/occurrences/:id/history
func GetOccurrenceHistory(c *gin.Context) {
	if occurrenceHistoryRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "occurrence history repository not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid occurrence ID format"})
		return
	}

	// Verify occurrence exists
	if occurrenceRepo != nil {
		_, err = occurrenceRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrOccurrenceNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "occurrence not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify occurrence"})
			return
		}
	}

	histories, err := occurrenceHistoryRepo.GetByOccurrenceID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get occurrence history"})
		return
	}

	// Convert to response format
	response := make([]models.OccurrenceHistoryResponse, 0, len(histories))
	for _, h := range histories {
		response = append(response, h.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"total": len(response),
	})
}

// UpdateOccurrenceStatus updates the status of an occurrence
// PATCH /api/v1/occurrences/:id/status
func UpdateOccurrenceStatus(c *gin.Context) {
	if occurrenceRepo == nil || occurrenceHistoryRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repositories not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid occurrence ID format"})
		return
	}

	var input models.UpdateStatusInput
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

	// Get current occurrence
	occurrence, err := occurrenceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrOccurrenceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "occurrence not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get occurrence"})
		return
	}

	// Validate status transition
	if !occurrence.Status.CanTransitionTo(input.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "invalid status transition",
			"current_status": occurrence.Status,
			"target_status":  input.Status,
			"allowed":        models.StatusTransitions[occurrence.Status],
		})
		return
	}

	// Special validation for CONCLUIDA - requires outcome
	if input.Status == models.StatusConcluida {
		// Check if outcome is already registered
		outcome, _ := occurrenceHistoryRepo.GetOutcomeByOccurrenceID(c.Request.Context(), id)
		if outcome == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "outcome must be registered before completing the occurrence",
				"hint":  "use POST /api/v1/occurrences/:id/outcome first",
			})
			return
		}
	}

	// Update status
	err = occurrenceRepo.UpdateStatus(c.Request.Context(), id, input.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}

	// Get user claims for history
	claims, _ := middleware.GetUserClaims(c)
	var userID *uuid.UUID
	if claims != nil {
		uid, err := uuid.Parse(claims.UserID)
		if err == nil {
			userID = &uid
		}
	}

	// Determine action description
	action := models.ActionStatusChanged
	switch input.Status {
	case models.StatusEmAndamento:
		action = models.ActionOccurrenceAssigned
	case models.StatusAceita:
		action = models.ActionOccurrenceAccepted
	case models.StatusRecusada:
		action = models.ActionOccurrenceRefused
	case models.StatusCancelada:
		action = models.ActionOccurrenceCanceled
	case models.StatusConcluida:
		action = models.ActionOccurrenceConcluded
	}

	// Create history entry
	historyInput := &models.CreateHistoryInput{
		OccurrenceID:   id,
		UserID:         userID,
		Acao:           action,
		StatusAnterior: &occurrence.Status,
		StatusNovo:     &input.Status,
		Observacoes:    input.Observacoes,
	}

	_, err = occurrenceHistoryRepo.Create(c.Request.Context(), historyInput)
	if err != nil {
		// Log error but don't fail the request
		_ = err
	}

	// Get updated occurrence
	updatedOccurrence, _ := occurrenceRepo.GetByID(c.Request.Context(), id)
	if updatedOccurrence != nil {
		c.JSON(http.StatusOK, updatedOccurrence.ToDetailResponse())
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":    "status updated successfully",
			"new_status": input.Status,
		})
	}
}

// RegisterOutcome registers the outcome of an occurrence
// POST /api/v1/occurrences/:id/outcome
func RegisterOutcome(c *gin.Context) {
	if occurrenceRepo == nil || occurrenceHistoryRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repositories not configured"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid occurrence ID format"})
		return
	}

	var input models.RegisterOutcomeInput
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

	// Get current occurrence
	occurrence, err := occurrenceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrOccurrenceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "occurrence not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get occurrence"})
		return
	}

	// Check if occurrence is in a state that allows outcome registration
	// Outcome can be registered when status is ACEITA or RECUSADA (before transitioning to CONCLUIDA)
	allowedStatuses := []models.OccurrenceStatus{
		models.StatusAceita,
		models.StatusRecusada,
	}

	isAllowed := false
	for _, s := range allowedStatuses {
		if occurrence.Status == s {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "outcome can only be registered for occurrences in ACEITA or RECUSADA status",
			"current_status": occurrence.Status,
		})
		return
	}

	// Check if outcome is already registered
	existingOutcome, _ := occurrenceHistoryRepo.GetOutcomeByOccurrenceID(c.Request.Context(), id)
	if existingOutcome != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":            "outcome already registered for this occurrence",
			"existing_outcome": existingOutcome.Desfecho,
		})
		return
	}

	// Get user claims for history
	claims, _ := middleware.GetUserClaims(c)
	var userID *uuid.UUID
	if claims != nil {
		uid, err := uuid.Parse(claims.UserID)
		if err == nil {
			userID = &uid
		}
	}

	// Create history entry with outcome
	historyInput := &models.CreateHistoryInput{
		OccurrenceID: id,
		UserID:       userID,
		Acao:         models.ActionOutcomeRegistered,
		Observacoes:  input.Observacoes,
		Desfecho:     &input.Desfecho,
	}

	history, err := occurrenceHistoryRepo.Create(c.Request.Context(), historyInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register outcome"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "outcome registered successfully",
		"desfecho":  input.Desfecho,
		"history":   history.ToResponse(),
		"next_step": "PATCH /api/v1/occurrences/:id/status to transition to CONCLUIDA",
	})
}
