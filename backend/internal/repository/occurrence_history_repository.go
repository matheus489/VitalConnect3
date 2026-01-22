package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
)

var (
	ErrHistoryNotFound = errors.New("history entry not found")
)

// OccurrenceHistoryRepository handles occurrence history data access
type OccurrenceHistoryRepository struct {
	db *sql.DB
}

// NewOccurrenceHistoryRepository creates a new occurrence history repository
func NewOccurrenceHistoryRepository(db *sql.DB) *OccurrenceHistoryRepository {
	return &OccurrenceHistoryRepository{db: db}
}

// Create creates a new history entry
func (r *OccurrenceHistoryRepository) Create(ctx context.Context, input *models.CreateHistoryInput) (*models.OccurrenceHistory, error) {
	history := &models.OccurrenceHistory{
		ID:             uuid.New(),
		OccurrenceID:   input.OccurrenceID,
		UserID:         input.UserID,
		Acao:           input.Acao,
		StatusAnterior: input.StatusAnterior,
		StatusNovo:     input.StatusNovo,
		Observacoes:    input.Observacoes,
		Desfecho:       input.Desfecho,
		CreatedAt:      time.Now(),
	}

	query := `
		INSERT INTO occurrence_history (
			id, occurrence_id, user_id, acao, status_anterior, status_novo,
			observacoes, desfecho, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		history.ID,
		history.OccurrenceID,
		history.UserID,
		history.Acao,
		history.StatusAnterior,
		history.StatusNovo,
		history.Observacoes,
		history.Desfecho,
		history.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return history, nil
}

// GetByOccurrenceID retrieves all history entries for an occurrence
func (r *OccurrenceHistoryRepository) GetByOccurrenceID(ctx context.Context, occurrenceID uuid.UUID) ([]models.OccurrenceHistory, error) {
	query := `
		SELECT
			h.id, h.occurrence_id, h.user_id, h.acao, h.status_anterior, h.status_novo,
			h.observacoes, h.desfecho, h.created_at,
			u.nome as user_nome
		FROM occurrence_history h
		LEFT JOIN users u ON h.user_id = u.id
		WHERE h.occurrence_id = $1
		ORDER BY h.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, occurrenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []models.OccurrenceHistory
	for rows.Next() {
		var h models.OccurrenceHistory
		var userID sql.NullString
		var statusAnterior, statusNovo, observacoes, desfecho, userNome sql.NullString

		err := rows.Scan(
			&h.ID, &h.OccurrenceID, &userID, &h.Acao, &statusAnterior, &statusNovo,
			&observacoes, &desfecho, &h.CreatedAt, &userNome,
		)
		if err != nil {
			return nil, err
		}

		if userID.Valid {
			uid, err := uuid.Parse(userID.String)
			if err == nil {
				h.UserID = &uid
			}
		}

		if statusAnterior.Valid {
			status := models.OccurrenceStatus(statusAnterior.String)
			h.StatusAnterior = &status
		}

		if statusNovo.Valid {
			status := models.OccurrenceStatus(statusNovo.String)
			h.StatusNovo = &status
		}

		if observacoes.Valid {
			h.Observacoes = &observacoes.String
		}

		if desfecho.Valid {
			outcome := models.OutcomeType(desfecho.String)
			h.Desfecho = &outcome
		}

		if userNome.Valid {
			h.User = &models.User{Nome: userNome.String}
		}

		histories = append(histories, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return histories, nil
}

// GetLatestByOccurrenceID retrieves the latest history entry for an occurrence
func (r *OccurrenceHistoryRepository) GetLatestByOccurrenceID(ctx context.Context, occurrenceID uuid.UUID) (*models.OccurrenceHistory, error) {
	query := `
		SELECT
			h.id, h.occurrence_id, h.user_id, h.acao, h.status_anterior, h.status_novo,
			h.observacoes, h.desfecho, h.created_at
		FROM occurrence_history h
		WHERE h.occurrence_id = $1
		ORDER BY h.created_at DESC
		LIMIT 1
	`

	var h models.OccurrenceHistory
	var userID sql.NullString
	var statusAnterior, statusNovo, observacoes, desfecho sql.NullString

	err := r.db.QueryRowContext(ctx, query, occurrenceID).Scan(
		&h.ID, &h.OccurrenceID, &userID, &h.Acao, &statusAnterior, &statusNovo,
		&observacoes, &desfecho, &h.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHistoryNotFound
		}
		return nil, err
	}

	if userID.Valid {
		uid, err := uuid.Parse(userID.String)
		if err == nil {
			h.UserID = &uid
		}
	}

	if statusAnterior.Valid {
		status := models.OccurrenceStatus(statusAnterior.String)
		h.StatusAnterior = &status
	}

	if statusNovo.Valid {
		status := models.OccurrenceStatus(statusNovo.String)
		h.StatusNovo = &status
	}

	if observacoes.Valid {
		h.Observacoes = &observacoes.String
	}

	if desfecho.Valid {
		outcome := models.OutcomeType(desfecho.String)
		h.Desfecho = &outcome
	}

	return &h, nil
}

// GetOutcomeByOccurrenceID retrieves the outcome entry for an occurrence (if any)
func (r *OccurrenceHistoryRepository) GetOutcomeByOccurrenceID(ctx context.Context, occurrenceID uuid.UUID) (*models.OccurrenceHistory, error) {
	query := `
		SELECT
			h.id, h.occurrence_id, h.user_id, h.acao, h.status_anterior, h.status_novo,
			h.observacoes, h.desfecho, h.created_at
		FROM occurrence_history h
		WHERE h.occurrence_id = $1 AND h.desfecho IS NOT NULL
		ORDER BY h.created_at DESC
		LIMIT 1
	`

	var h models.OccurrenceHistory
	var userID sql.NullString
	var statusAnterior, statusNovo, observacoes, desfecho sql.NullString

	err := r.db.QueryRowContext(ctx, query, occurrenceID).Scan(
		&h.ID, &h.OccurrenceID, &userID, &h.Acao, &statusAnterior, &statusNovo,
		&observacoes, &desfecho, &h.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHistoryNotFound
		}
		return nil, err
	}

	if userID.Valid {
		uid, err := uuid.Parse(userID.String)
		if err == nil {
			h.UserID = &uid
		}
	}

	if statusAnterior.Valid {
		status := models.OccurrenceStatus(statusAnterior.String)
		h.StatusAnterior = &status
	}

	if statusNovo.Valid {
		status := models.OccurrenceStatus(statusNovo.String)
		h.StatusNovo = &status
	}

	if observacoes.Valid {
		h.Observacoes = &observacoes.String
	}

	if desfecho.Valid {
		outcome := models.OutcomeType(desfecho.String)
		h.Desfecho = &outcome
	}

	return &h, nil
}
