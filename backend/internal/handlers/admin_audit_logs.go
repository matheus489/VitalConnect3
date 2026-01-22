package handlers

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
)

// adminAuditLogDB is used to access the database for audit log queries
var adminAuditLogDB *sql.DB

// SetAdminAuditLogDB sets the database connection for admin audit log handlers
func SetAdminAuditLogDB(db *sql.DB) {
	adminAuditLogDB = db
}

// AdminAuditLogFilter represents filters for admin audit log queries (no tenant restriction)
type AdminAuditLogFilter struct {
	TenantID     *uuid.UUID       `form:"tenant_id"`
	DataInicio   *time.Time       `form:"data_inicio" time_format:"2006-01-02"`
	DataFim      *time.Time       `form:"data_fim" time_format:"2006-01-02"`
	UsuarioID    *uuid.UUID       `form:"usuario_id"`
	Acao         *string          `form:"acao"`
	EntidadeTipo *string          `form:"entidade_tipo"`
	Severity     *models.Severity `form:"severity"`
	Page         int              `form:"page"`
	PageSize     int              `form:"page_size"`
}

// AdminAuditLogResponse extends AuditLogResponse with tenant info
type AdminAuditLogResponse struct {
	ID           uuid.UUID       `json:"id"`
	TenantID     *uuid.UUID      `json:"tenant_id,omitempty"`
	TenantName   *string         `json:"tenant_name,omitempty"`
	TenantSlug   *string         `json:"tenant_slug,omitempty"`
	Timestamp    time.Time       `json:"timestamp"`
	UsuarioID    *uuid.UUID      `json:"usuario_id,omitempty"`
	ActorName    string          `json:"actor_name"`
	Acao         string          `json:"acao"`
	EntidadeTipo string          `json:"entidade_tipo"`
	EntidadeID   string          `json:"entidade_id"`
	HospitalID   *uuid.UUID      `json:"hospital_id,omitempty"`
	HospitalNome *string         `json:"hospital_nome,omitempty"`
	Severity     models.Severity `json:"severity"`
	Detalhes     json.RawMessage `json:"detalhes,omitempty"`
	IPAddress    *string         `json:"ip_address,omitempty"`
	UserAgent    *string         `json:"user_agent,omitempty"`
}

// AdminListAuditLogs returns all audit logs across all tenants with filtering
// GET /api/v1/admin/logs
func AdminListAuditLogs(c *gin.Context) {
	if adminAuditLogDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not configured"})
		return
	}

	var filter AdminAuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// Parse tenant_id from query if provided
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
			return
		}
		filter.TenantID = &tenantID
	}

	// Parse usuario_id from query if provided
	if userIDStr := c.Query("usuario_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid usuario_id format"})
			return
		}
		filter.UsuarioID = &userID
	}

	// Build query
	logs, total, err := queryAdminAuditLogs(c.Request.Context(), adminAuditLogDB, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to query audit logs",
			"details": err.Error(),
		})
		return
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"total":       total,
		"page":        filter.Page,
		"per_page":    filter.PageSize,
		"total_pages": totalPages,
	})
}

// AdminExportAuditLogs exports audit logs as CSV
// GET /api/v1/admin/logs/export
func AdminExportAuditLogs(c *gin.Context) {
	if adminAuditLogDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not configured"})
		return
	}

	var filter AdminAuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// For export, we want all records matching the filter (with reasonable limit)
	filter.Page = 1
	filter.PageSize = 10000 // Max 10k records for export

	// Parse tenant_id from query if provided
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant_id format"})
			return
		}
		filter.TenantID = &tenantID
	}

	logs, _, err := queryAdminAuditLogs(c.Request.Context(), adminAuditLogDB, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to query audit logs",
			"details": err.Error(),
		})
		return
	}

	// Set headers for CSV download
	filename := fmt.Sprintf("audit_logs_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Write CSV
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID",
		"Timestamp",
		"Tenant Name",
		"Tenant Slug",
		"Actor Name",
		"Action",
		"Entity Type",
		"Entity ID",
		"Hospital Name",
		"Severity",
		"IP Address",
		"Details",
	}
	if err := writer.Write(header); err != nil {
		return
	}

	// Write data rows
	for _, log := range logs {
		tenantName := ""
		if log.TenantName != nil {
			tenantName = *log.TenantName
		}
		tenantSlug := ""
		if log.TenantSlug != nil {
			tenantSlug = *log.TenantSlug
		}
		hospitalNome := ""
		if log.HospitalNome != nil {
			hospitalNome = *log.HospitalNome
		}
		ipAddress := ""
		if log.IPAddress != nil {
			ipAddress = *log.IPAddress
		}
		detalhes := ""
		if log.Detalhes != nil {
			detalhes = string(log.Detalhes)
		}

		row := []string{
			log.ID.String(),
			log.Timestamp.Format(time.RFC3339),
			tenantName,
			tenantSlug,
			log.ActorName,
			log.Acao,
			log.EntidadeTipo,
			log.EntidadeID,
			hospitalNome,
			string(log.Severity),
			ipAddress,
			detalhes,
		}
		if err := writer.Write(row); err != nil {
			return
		}
	}
}

// queryAdminAuditLogs executes the audit log query with tenant info
func queryAdminAuditLogs(ctx context.Context, db *sql.DB, filter *AdminAuditLogFilter) ([]AdminAuditLogResponse, int, error) {
	// Build WHERE clause - note: no tenant_id restriction by default
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("al.tenant_id = $%d", argIdx))
		args = append(args, *filter.TenantID)
		argIdx++
	}

	if filter.DataInicio != nil {
		conditions = append(conditions, fmt.Sprintf("al.timestamp >= $%d", argIdx))
		args = append(args, *filter.DataInicio)
		argIdx++
	}

	if filter.DataFim != nil {
		// Add one day to include the entire end date
		endDate := filter.DataFim.Add(24 * time.Hour)
		conditions = append(conditions, fmt.Sprintf("al.timestamp < $%d", argIdx))
		args = append(args, endDate)
		argIdx++
	}

	if filter.UsuarioID != nil {
		conditions = append(conditions, fmt.Sprintf("al.usuario_id = $%d", argIdx))
		args = append(args, *filter.UsuarioID)
		argIdx++
	}

	if filter.Acao != nil && *filter.Acao != "" {
		conditions = append(conditions, fmt.Sprintf("al.acao = $%d", argIdx))
		args = append(args, *filter.Acao)
		argIdx++
	}

	if filter.EntidadeTipo != nil && *filter.EntidadeTipo != "" {
		conditions = append(conditions, fmt.Sprintf("al.entidade_tipo = $%d", argIdx))
		args = append(args, *filter.EntidadeTipo)
		argIdx++
	}

	if filter.Severity != nil && *filter.Severity != "" {
		conditions = append(conditions, fmt.Sprintf("al.severity = $%d", argIdx))
		args = append(args, *filter.Severity)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM audit_logs al %s`, whereClause)
	var total int
	if err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Calculate offset
	offset := (filter.Page - 1) * filter.PageSize

	// Main query with tenant and hospital join
	query := fmt.Sprintf(`
		SELECT
			al.id, al.tenant_id, t.name, t.slug,
			al.timestamp, al.usuario_id, al.actor_name, al.acao,
			al.entidade_tipo, al.entidade_id, al.hospital_id,
			h.nome AS hospital_nome, al.severity, al.detalhes,
			al.ip_address, al.user_agent
		FROM audit_logs al
		LEFT JOIN tenants t ON al.tenant_id = t.id
		LEFT JOIN hospitals h ON al.hospital_id = h.id
		%s
		ORDER BY al.timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.PageSize, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AdminAuditLogResponse
	for rows.Next() {
		var log AdminAuditLogResponse
		var tenantID, usuarioID, hospitalID sql.NullString
		var tenantName, tenantSlug, hospitalNome sql.NullString
		var detalhes sql.NullString
		var ipAddress, userAgent sql.NullString

		err := rows.Scan(
			&log.ID, &tenantID, &tenantName, &tenantSlug,
			&log.Timestamp, &usuarioID, &log.ActorName, &log.Acao,
			&log.EntidadeTipo, &log.EntidadeID, &hospitalID,
			&hospitalNome, &log.Severity, &detalhes,
			&ipAddress, &userAgent,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if tenantID.Valid {
			tid, _ := uuid.Parse(tenantID.String)
			log.TenantID = &tid
		}
		if tenantName.Valid {
			log.TenantName = &tenantName.String
		}
		if tenantSlug.Valid {
			log.TenantSlug = &tenantSlug.String
		}
		if usuarioID.Valid {
			uid, _ := uuid.Parse(usuarioID.String)
			log.UsuarioID = &uid
		}
		if hospitalID.Valid {
			hid, _ := uuid.Parse(hospitalID.String)
			log.HospitalID = &hid
		}
		if hospitalNome.Valid {
			log.HospitalNome = &hospitalNome.String
		}
		if detalhes.Valid {
			log.Detalhes = json.RawMessage(detalhes.String)
		}
		if ipAddress.Valid {
			log.IPAddress = &ipAddress.String
		}
		if userAgent.Valid {
			log.UserAgent = &userAgent.String
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, total, nil
}
