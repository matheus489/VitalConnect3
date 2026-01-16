package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

// AuditLogRepository handles audit log data access
type AuditLogRepository struct {
	db *sql.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditLogRepository) Create(ctx context.Context, input *models.CreateAuditLogInput) (*models.AuditLog, error) {
	auditLog := &models.AuditLog{
		ID:           uuid.New(),
		Timestamp:    time.Now(),
		UsuarioID:    input.UsuarioID,
		ActorName:    input.ActorName,
		Acao:         input.Acao,
		EntidadeTipo: input.EntidadeTipo,
		EntidadeID:   input.EntidadeID,
		HospitalID:   input.HospitalID,
		Severity:     input.Severity,
		Detalhes:     input.Detalhes,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
	}

	// Ensure detalhes is valid JSON or null
	var detalhesArg interface{}
	if auditLog.Detalhes != nil && len(auditLog.Detalhes) > 0 {
		detalhesArg = auditLog.Detalhes
	} else {
		detalhesArg = nil
	}

	query := `
		INSERT INTO audit_logs (
			id, timestamp, usuario_id, actor_name, acao, entidade_tipo,
			entidade_id, hospital_id, severity, detalhes, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		auditLog.ID,
		auditLog.Timestamp,
		auditLog.UsuarioID,
		auditLog.ActorName,
		auditLog.Acao,
		auditLog.EntidadeTipo,
		auditLog.EntidadeID,
		auditLog.HospitalID,
		auditLog.Severity,
		detalhesArg,
		auditLog.IPAddress,
		auditLog.UserAgent,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	return auditLog, nil
}

// List retrieves audit logs with filters and pagination
func (r *AuditLogRepository) List(ctx context.Context, filters *models.AuditLogFilter) ([]models.AuditLog, int, error) {
	if filters == nil {
		filters = models.DefaultAuditLogFilters()
	}

	// Build WHERE conditions
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.DataInicio != nil {
		conditions = append(conditions, fmt.Sprintf("al.timestamp >= $%d", argIdx))
		args = append(args, *filters.DataInicio)
		argIdx++
	}

	if filters.DataFim != nil {
		conditions = append(conditions, fmt.Sprintf("al.timestamp <= $%d", argIdx))
		args = append(args, *filters.DataFim)
		argIdx++
	}

	if filters.UsuarioID != nil {
		conditions = append(conditions, fmt.Sprintf("al.usuario_id = $%d", argIdx))
		args = append(args, *filters.UsuarioID)
		argIdx++
	}

	if filters.Acao != nil && *filters.Acao != "" {
		conditions = append(conditions, fmt.Sprintf("al.acao = $%d", argIdx))
		args = append(args, *filters.Acao)
		argIdx++
	}

	if filters.EntidadeTipo != nil && *filters.EntidadeTipo != "" {
		conditions = append(conditions, fmt.Sprintf("al.entidade_tipo = $%d", argIdx))
		args = append(args, *filters.EntidadeTipo)
		argIdx++
	}

	if filters.EntidadeID != nil && *filters.EntidadeID != "" {
		conditions = append(conditions, fmt.Sprintf("al.entidade_id = $%d", argIdx))
		args = append(args, *filters.EntidadeID)
		argIdx++
	}

	if filters.Severity != nil {
		conditions = append(conditions, fmt.Sprintf("al.severity = $%d", argIdx))
		args = append(args, *filters.Severity)
		argIdx++
	}

	if filters.HospitalID != nil {
		conditions = append(conditions, fmt.Sprintf("al.hospital_id = $%d", argIdx))
		args = append(args, *filters.HospitalID)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total items
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM audit_logs al %s`, whereClause)
	var totalItems int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Calculate offset
	offset := (filters.Page - 1) * filters.PageSize

	// Main query with pagination
	query := fmt.Sprintf(`
		SELECT
			al.id, al.timestamp, al.usuario_id, al.actor_name, al.acao,
			al.entidade_tipo, al.entidade_id, al.hospital_id, al.severity,
			al.detalhes, al.ip_address, al.user_agent
		FROM audit_logs al
		%s
		ORDER BY al.timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filters.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var usuarioID, hospitalID sql.NullString
		var detalhes sql.NullString
		var ipAddress, userAgent sql.NullString

		err := rows.Scan(
			&log.ID, &log.Timestamp, &usuarioID, &log.ActorName, &log.Acao,
			&log.EntidadeTipo, &log.EntidadeID, &hospitalID, &log.Severity,
			&detalhes, &ipAddress, &userAgent,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if usuarioID.Valid {
			uid, err := uuid.Parse(usuarioID.String)
			if err == nil {
				log.UsuarioID = &uid
			}
		}

		if hospitalID.Valid {
			hid, err := uuid.Parse(hospitalID.String)
			if err == nil {
				log.HospitalID = &hid
			}
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

	return logs, totalItems, nil
}

// GetByEntityID retrieves audit logs for a specific entity (for timeline)
func (r *AuditLogRepository) GetByEntityID(ctx context.Context, entidadeTipo string, entidadeID string) ([]models.AuditLog, error) {
	query := `
		SELECT
			al.id, al.timestamp, al.usuario_id, al.actor_name, al.acao,
			al.entidade_tipo, al.entidade_id, al.hospital_id, al.severity,
			al.detalhes, al.ip_address, al.user_agent
		FROM audit_logs al
		WHERE al.entidade_tipo = $1 AND al.entidade_id = $2
		ORDER BY al.timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, entidadeTipo, entidadeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs by entity: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var usuarioID, hospitalID sql.NullString
		var detalhes sql.NullString
		var ipAddress, userAgent sql.NullString

		err := rows.Scan(
			&log.ID, &log.Timestamp, &usuarioID, &log.ActorName, &log.Acao,
			&log.EntidadeTipo, &log.EntidadeID, &hospitalID, &log.Severity,
			&detalhes, &ipAddress, &userAgent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if usuarioID.Valid {
			uid, err := uuid.Parse(usuarioID.String)
			if err == nil {
				log.UsuarioID = &uid
			}
		}

		if hospitalID.Valid {
			hid, err := uuid.Parse(hospitalID.String)
			if err == nil {
				log.HospitalID = &hid
			}
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
		return nil, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, nil
}

// ListWithHospitalNames retrieves audit logs with hospital names for display
func (r *AuditLogRepository) ListWithHospitalNames(ctx context.Context, filters *models.AuditLogFilter) ([]models.AuditLogResponse, int, error) {
	if filters == nil {
		filters = models.DefaultAuditLogFilters()
	}

	// Build WHERE conditions
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.DataInicio != nil {
		conditions = append(conditions, fmt.Sprintf("al.timestamp >= $%d", argIdx))
		args = append(args, *filters.DataInicio)
		argIdx++
	}

	if filters.DataFim != nil {
		conditions = append(conditions, fmt.Sprintf("al.timestamp <= $%d", argIdx))
		args = append(args, *filters.DataFim)
		argIdx++
	}

	if filters.UsuarioID != nil {
		conditions = append(conditions, fmt.Sprintf("al.usuario_id = $%d", argIdx))
		args = append(args, *filters.UsuarioID)
		argIdx++
	}

	if filters.Acao != nil && *filters.Acao != "" {
		conditions = append(conditions, fmt.Sprintf("al.acao = $%d", argIdx))
		args = append(args, *filters.Acao)
		argIdx++
	}

	if filters.EntidadeTipo != nil && *filters.EntidadeTipo != "" {
		conditions = append(conditions, fmt.Sprintf("al.entidade_tipo = $%d", argIdx))
		args = append(args, *filters.EntidadeTipo)
		argIdx++
	}

	if filters.EntidadeID != nil && *filters.EntidadeID != "" {
		conditions = append(conditions, fmt.Sprintf("al.entidade_id = $%d", argIdx))
		args = append(args, *filters.EntidadeID)
		argIdx++
	}

	if filters.Severity != nil {
		conditions = append(conditions, fmt.Sprintf("al.severity = $%d", argIdx))
		args = append(args, *filters.Severity)
		argIdx++
	}

	if filters.HospitalID != nil {
		conditions = append(conditions, fmt.Sprintf("al.hospital_id = $%d", argIdx))
		args = append(args, *filters.HospitalID)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total items
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM audit_logs al %s`, whereClause)
	var totalItems int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Calculate offset
	offset := (filters.Page - 1) * filters.PageSize

	// Main query with hospital join and pagination
	query := fmt.Sprintf(`
		SELECT
			al.id, al.timestamp, al.usuario_id, al.actor_name, al.acao,
			al.entidade_tipo, al.entidade_id, al.hospital_id, al.severity,
			al.detalhes, al.ip_address, al.user_agent,
			h.nome as hospital_nome
		FROM audit_logs al
		LEFT JOIN hospitals h ON al.hospital_id = h.id
		%s
		ORDER BY al.timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filters.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLogResponse
	for rows.Next() {
		var log models.AuditLogResponse
		var usuarioID, hospitalID sql.NullString
		var detalhes sql.NullString
		var ipAddress, userAgent, hospitalNome sql.NullString

		err := rows.Scan(
			&log.ID, &log.Timestamp, &usuarioID, &log.ActorName, &log.Acao,
			&log.EntidadeTipo, &log.EntidadeID, &hospitalID, &log.Severity,
			&detalhes, &ipAddress, &userAgent, &hospitalNome,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if usuarioID.Valid {
			uid, err := uuid.Parse(usuarioID.String)
			if err == nil {
				log.UsuarioID = &uid
			}
		}

		if hospitalID.Valid {
			hid, err := uuid.Parse(hospitalID.String)
			if err == nil {
				log.HospitalID = &hid
			}
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

		if hospitalNome.Valid {
			log.HospitalNome = &hospitalNome.String
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, totalItems, nil
}
