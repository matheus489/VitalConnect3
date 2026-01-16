package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/middleware"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
	"github.com/vitalconnect/backend/internal/services/audit"
)

var (
	auditLogRepo    *repository.AuditLogRepository
	auditService    *audit.AuditService
)

// SetAuditLogRepository sets the audit log repository for handlers
func SetAuditLogRepository(repo *repository.AuditLogRepository) {
	auditLogRepo = repo
}

// SetAuditService sets the audit service for handlers
func SetAuditService(service *audit.AuditService) {
	auditService = service
}

// GetAuditService returns the global audit service
func GetAuditService() *audit.AuditService {
	return auditService
}

// ListAuditLogs returns audit logs with pagination and filters
// GET /api/v1/audit-logs
// Access: Admin (all logs), Gestor (only their hospital's logs)
// Operador: Access DENIED
func ListAuditLogs(c *gin.Context) {
	if auditLogRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "audit log repository not configured"})
		return
	}

	// Get user claims for authorization
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Operador does NOT have access to audit logs endpoint
	if claims.Role == "operador" {
		c.JSON(http.StatusForbidden, gin.H{
			"error":        "access denied",
			"message":      "operators do not have permission to access audit logs",
			"required_role": []string{"admin", "gestor"},
		})
		return
	}

	// Parse filters from query parameters
	filters := models.DefaultAuditLogFilters()

	// Date filters
	if dataInicio := c.Query("data_inicio"); dataInicio != "" {
		t, err := time.Parse(time.RFC3339, dataInicio)
		if err != nil {
			// Try date-only format
			t, err = time.Parse("2006-01-02", dataInicio)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data_inicio format, use RFC3339 or YYYY-MM-DD"})
				return
			}
		}
		filters.DataInicio = &t
	}

	if dataFim := c.Query("data_fim"); dataFim != "" {
		t, err := time.Parse(time.RFC3339, dataFim)
		if err != nil {
			// Try date-only format and set to end of day
			t, err = time.Parse("2006-01-02", dataFim)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data_fim format, use RFC3339 or YYYY-MM-DD"})
				return
			}
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
		filters.DataFim = &t
	}

	// Usuario filter
	if usuarioID := c.Query("usuario_id"); usuarioID != "" {
		uid, err := uuid.Parse(usuarioID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid usuario_id format"})
			return
		}
		filters.UsuarioID = &uid
	}

	// Acao filter
	if acao := c.Query("acao"); acao != "" {
		filters.Acao = &acao
	}

	// Entidade tipo filter
	if entidadeTipo := c.Query("entidade_tipo"); entidadeTipo != "" {
		filters.EntidadeTipo = &entidadeTipo
	}

	// Entidade ID filter
	if entidadeID := c.Query("entidade_id"); entidadeID != "" {
		filters.EntidadeID = &entidadeID
	}

	// Severity filter
	if severity := c.Query("severity"); severity != "" {
		s := models.Severity(severity)
		if !s.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         "invalid severity filter",
				"valid_values": []string{"INFO", "WARN", "CRITICAL"},
			})
			return
		}
		filters.Severity = &s
	}

	// Hospital filter - Gestor can only see their own hospital
	if hospitalID := c.Query("hospital_id"); hospitalID != "" {
		hid, err := uuid.Parse(hospitalID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hospital_id format"})
			return
		}
		filters.HospitalID = &hid
	}

	// For Gestor, enforce hospital_id filter
	if claims.Role == "gestor" {
		if claims.HospitalID == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "access denied",
				"message": "gestor must be associated with a hospital",
			})
			return
		}

		hospitalUUID, err := uuid.Parse(claims.HospitalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid hospital_id in claims"})
			return
		}

		// Override any hospital_id filter with gestor's hospital
		filters.HospitalID = &hospitalUUID
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

	// Query with hospital names for display
	logs, totalItems, err := auditLogRepo.ListWithHospitalNames(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list audit logs"})
		return
	}

	c.JSON(http.StatusOK, models.NewPaginatedResponse(logs, filters.Page, filters.PageSize, totalItems))
}

// GetOccurrenceTimeline returns audit logs for a specific occurrence (timeline view)
// GET /api/v1/occurrences/:id/timeline
// Access: Admin (all), Gestor (same hospital), Operador (their occurrences)
func GetOccurrenceTimeline(c *gin.Context) {
	if auditLogRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "audit log repository not configured"})
		return
	}

	idParam := c.Param("id")
	occurrenceID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid occurrence ID format"})
		return
	}

	// Get user claims for authorization
	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Verify occurrence exists and check access (using existing occurrence repository)
	if occurrenceRepo != nil {
		occurrence, err := occurrenceRepo.GetByID(c.Request.Context(), occurrenceID)
		if err != nil {
			if err == repository.ErrOccurrenceNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "occurrence not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify occurrence"})
			return
		}

		// Access control based on role
		switch claims.Role {
		case "admin":
			// Admin can view any occurrence timeline
		case "gestor":
			// Gestor can only view occurrences from their hospital
			if claims.HospitalID == "" || claims.HospitalID != occurrence.HospitalID.String() {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "access denied",
					"message": "you can only view occurrences from your hospital",
				})
				return
			}
		case "operador":
			// Operador can view occurrences from their hospital
			// Note: Could add additional restriction to only their assigned occurrences
			if claims.HospitalID == "" || claims.HospitalID != occurrence.HospitalID.String() {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "access denied",
					"message": "you can only view occurrences from your hospital",
				})
				return
			}
		default:
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	// Get audit logs for this occurrence
	logs, err := auditLogRepo.GetByEntityID(c.Request.Context(), "Ocorrencia", occurrenceID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get occurrence timeline"})
		return
	}

	// Convert to response format
	response := make([]models.AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		response = append(response, log.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"total": len(response),
	})
}
