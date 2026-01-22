package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
)

var (
	ErrOccurrenceNotFound = errors.New("occurrence not found")
)

// OccurrenceRepository handles occurrence data access
type OccurrenceRepository struct {
	db *sql.DB
}

// NewOccurrenceRepository creates a new occurrence repository
func NewOccurrenceRepository(db *sql.DB) *OccurrenceRepository {
	return &OccurrenceRepository{db: db}
}

// List returns occurrences with pagination and filters for the current tenant
func (r *OccurrenceRepository) List(ctx context.Context, filters models.OccurrenceListFilters) ([]models.Occurrence, int, error) {
	tf := NewTenantFilter(ctx)

	// Build the WHERE clause
	where := "WHERE 1=1" + tf.AndClauseWithAlias("o")
	args := []interface{}{}
	argIndex := 1

	if filters.Status != nil && *filters.Status != "" {
		where += fmt.Sprintf(" AND o.status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.HospitalID != nil && *filters.HospitalID != "" {
		where += fmt.Sprintf(" AND o.hospital_id = $%d", argIndex)
		args = append(args, *filters.HospitalID)
		argIndex++
	}

	if filters.DateFrom != nil {
		where += fmt.Sprintf(" AND o.created_at >= $%d", argIndex)
		args = append(args, *filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != nil {
		where += fmt.Sprintf(" AND o.created_at <= $%d", argIndex)
		args = append(args, *filters.DateTo)
		argIndex++
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM occurrences o %s", where)
	var totalItems int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderBy := "o.created_at DESC"
	if filters.SortBy != "" {
		validSortColumns := map[string]bool{
			"created_at":       true,
			"score_priorizacao": true,
			"janela_expira_em": true,
			"data_obito":       true,
		}
		if validSortColumns[filters.SortBy] {
			order := "DESC"
			if filters.SortOrder == "asc" {
				order = "ASC"
			}
			orderBy = fmt.Sprintf("o.%s %s", filters.SortBy, order)
		}
	}

	// Pagination
	offset := (filters.Page - 1) * filters.PageSize
	limit := filters.PageSize

	// Main query with pagination
	query := fmt.Sprintf(`
		SELECT
			o.id, o.obito_id, o.hospital_id, o.status, o.score_priorizacao,
			o.nome_paciente_mascarado, o.dados_completos, o.created_at, o.updated_at,
			o.notificado_em, o.data_obito, o.janela_expira_em,
			h.id, h.nome, h.codigo, h.endereco, h.ativo
		FROM occurrences o
		LEFT JOIN hospitals h ON o.hospital_id = h.id
		%s
		ORDER BY %s
		LIMIT %d OFFSET %d
	`, where, orderBy, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var occurrences []models.Occurrence
	for rows.Next() {
		var o models.Occurrence
		var h models.Hospital
		var notificadoEm sql.NullTime
		var dadosCompletos string
		var hEndereco sql.NullString

		err := rows.Scan(
			&o.ID, &o.ObitoID, &o.HospitalID, &o.Status, &o.ScorePriorizacao,
			&o.NomePacienteMascarado, &dadosCompletos, &o.CreatedAt, &o.UpdatedAt,
			&notificadoEm, &o.DataObito, &o.JanelaExpiraEm,
			&h.ID, &h.Nome, &h.Codigo, &hEndereco, &h.Ativo,
		)
		if err != nil {
			return nil, 0, err
		}

		o.DadosCompletos = json.RawMessage(dadosCompletos)

		if notificadoEm.Valid {
			o.NotificadoEm = &notificadoEm.Time
		}

		if hEndereco.Valid {
			h.Endereco = &hEndereco.String
		}
		o.Hospital = &h

		occurrences = append(occurrences, o)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return occurrences, totalItems, nil
}

// GetByID retrieves an occurrence by ID with full data for the current tenant
func (r *OccurrenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Occurrence, error) {
	tf := NewTenantFilter(ctx)

	query := `
		SELECT
			o.id, o.obito_id, o.hospital_id, o.status, o.score_priorizacao,
			o.nome_paciente_mascarado, o.dados_completos, o.created_at, o.updated_at,
			o.notificado_em, o.data_obito, o.janela_expira_em,
			h.id, h.nome, h.codigo, h.endereco, h.ativo
		FROM occurrences o
		LEFT JOIN hospitals h ON o.hospital_id = h.id
		WHERE o.id = $1` + tf.AndClauseWithAlias("o") + `
	`

	var o models.Occurrence
	var h models.Hospital
	var notificadoEm sql.NullTime
	var dadosCompletos string
	var hEndereco sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID, &o.ObitoID, &o.HospitalID, &o.Status, &o.ScorePriorizacao,
		&o.NomePacienteMascarado, &dadosCompletos, &o.CreatedAt, &o.UpdatedAt,
		&notificadoEm, &o.DataObito, &o.JanelaExpiraEm,
		&h.ID, &h.Nome, &h.Codigo, &hEndereco, &h.Ativo,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOccurrenceNotFound
		}
		return nil, err
	}

	o.DadosCompletos = json.RawMessage(dadosCompletos)

	if notificadoEm.Valid {
		o.NotificadoEm = &notificadoEm.Time
	}

	if hEndereco.Valid {
		h.Endereco = &hEndereco.String
	}
	o.Hospital = &h

	return &o, nil
}

// UpdateStatus updates the status of an occurrence for the current tenant
func (r *OccurrenceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus models.OccurrenceStatus) error {
	tf := NewTenantFilter(ctx)

	query := `
		UPDATE occurrences
		SET status = $1, updated_at = $2
		WHERE id = $3` + tf.AndClause() + `
	`

	result, err := r.db.ExecContext(ctx, query, newStatus, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOccurrenceNotFound
	}

	return nil
}

// Create creates a new occurrence
func (r *OccurrenceRepository) Create(ctx context.Context, input *models.CreateOccurrenceInput) (*models.Occurrence, error) {
	occurrence := &models.Occurrence{
		ID:                    uuid.New(),
		ObitoID:               input.ObitoID,
		HospitalID:            input.HospitalID,
		Status:                models.StatusPendente,
		ScorePriorizacao:      input.ScorePriorizacao,
		NomePacienteMascarado: input.NomePacienteMascarado,
		DadosCompletos:        input.DadosCompletos,
		DataObito:             input.DataObito,
		JanelaExpiraEm:        input.DataObito.Add(6 * time.Hour),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	query := `
		INSERT INTO occurrences (
			id, obito_id, hospital_id, status, score_priorizacao,
			nome_paciente_mascarado, dados_completos, data_obito, janela_expira_em,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		occurrence.ID,
		occurrence.ObitoID,
		occurrence.HospitalID,
		occurrence.Status,
		occurrence.ScorePriorizacao,
		occurrence.NomePacienteMascarado,
		string(occurrence.DadosCompletos),
		occurrence.DataObito,
		occurrence.JanelaExpiraEm,
		occurrence.CreatedAt,
		occurrence.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return occurrence, nil
}

// SetNotificado marks the occurrence as notified for the current tenant
func (r *OccurrenceRepository) SetNotificado(ctx context.Context, id uuid.UUID) error {
	tf := NewTenantFilter(ctx)

	query := `
		UPDATE occurrences
		SET notificado_em = $1, updated_at = $1
		WHERE id = $2` + tf.AndClause() + `
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}

// GetPendingCount returns the count of pending occurrences for the current tenant
func (r *OccurrenceRepository) GetPendingCount(ctx context.Context) (int, error) {
	tf := NewTenantFilter(ctx)

	query := `SELECT COUNT(*) FROM occurrences WHERE status = 'PENDENTE'` + tf.AndClause()

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// GetEmAndamentoCount returns the count of in-progress occurrences for the current tenant
func (r *OccurrenceRepository) GetEmAndamentoCount(ctx context.Context) (int, error) {
	tf := NewTenantFilter(ctx)

	query := `SELECT COUNT(*) FROM occurrences WHERE status = 'EM_ANDAMENTO'` + tf.AndClause()

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// GetTodayEligibleCount returns the count of eligible occurrences created today for the current tenant
func (r *OccurrenceRepository) GetTodayEligibleCount(ctx context.Context) (int, error) {
	tf := NewTenantFilter(ctx)

	query := `
		SELECT COUNT(*) FROM occurrences
		WHERE DATE(created_at) = CURRENT_DATE` + tf.AndClause() + `
	`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// GetAverageNotificationTime returns the average notification time in seconds for the current tenant
func (r *OccurrenceRepository) GetAverageNotificationTime(ctx context.Context) (float64, error) {
	tf := NewTenantFilter(ctx)

	query := `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (notificado_em - created_at))), 0)
		FROM occurrences
		WHERE notificado_em IS NOT NULL
		AND DATE(created_at) = CURRENT_DATE` + tf.AndClause() + `
	`

	var avgTime float64
	err := r.db.QueryRowContext(ctx, query).Scan(&avgTime)
	return avgTime, err
}

// ExistsByObitoID checks if an occurrence already exists for a given obito (tenant-independent for triagem motor)
func (r *OccurrenceRepository) ExistsByObitoID(ctx context.Context, obitoID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM occurrences WHERE obito_id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, obitoID).Scan(&exists)
	return exists, err
}
