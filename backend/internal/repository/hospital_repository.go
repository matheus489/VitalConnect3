package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

var (
	ErrHospitalNotFound = errors.New("hospital not found")
	ErrHospitalExists   = errors.New("hospital with this code already exists")
)

// HospitalRepository handles hospital data access
type HospitalRepository struct {
	db *sql.DB
}

// NewHospitalRepository creates a new hospital repository
func NewHospitalRepository(db *sql.DB) *HospitalRepository {
	return &HospitalRepository{db: db}
}

// List returns all active hospitals
func (r *HospitalRepository) List(ctx context.Context) ([]models.Hospital, error) {
	query := `
		SELECT id, nome, codigo, endereco, config_conexao, ativo, created_at, updated_at, deleted_at
		FROM hospitals
		WHERE deleted_at IS NULL
		ORDER BY nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hospitals []models.Hospital
	for rows.Next() {
		var h models.Hospital
		var endereco, configConexao sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&h.ID,
			&h.Nome,
			&h.Codigo,
			&endereco,
			&configConexao,
			&h.Ativo,
			&h.CreatedAt,
			&h.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		if endereco.Valid {
			h.Endereco = &endereco.String
		}
		if configConexao.Valid {
			h.ConfigConexao = json.RawMessage(configConexao.String)
		}
		if deletedAt.Valid {
			h.DeletedAt = &deletedAt.Time
		}

		hospitals = append(hospitals, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hospitals, nil
}

// GetByID retrieves a hospital by ID
func (r *HospitalRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Hospital, error) {
	query := `
		SELECT id, nome, codigo, endereco, config_conexao, ativo, created_at, updated_at, deleted_at
		FROM hospitals
		WHERE id = $1 AND deleted_at IS NULL
	`

	var h models.Hospital
	var endereco, configConexao sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&h.ID,
		&h.Nome,
		&h.Codigo,
		&endereco,
		&configConexao,
		&h.Ativo,
		&h.CreatedAt,
		&h.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHospitalNotFound
		}
		return nil, err
	}

	if endereco.Valid {
		h.Endereco = &endereco.String
	}
	if configConexao.Valid {
		h.ConfigConexao = json.RawMessage(configConexao.String)
	}
	if deletedAt.Valid {
		h.DeletedAt = &deletedAt.Time
	}

	return &h, nil
}

// GetByCodigo retrieves a hospital by code
func (r *HospitalRepository) GetByCodigo(ctx context.Context, codigo string) (*models.Hospital, error) {
	query := `
		SELECT id, nome, codigo, endereco, config_conexao, ativo, created_at, updated_at, deleted_at
		FROM hospitals
		WHERE codigo = $1 AND deleted_at IS NULL
	`

	var h models.Hospital
	var endereco, configConexao sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, codigo).Scan(
		&h.ID,
		&h.Nome,
		&h.Codigo,
		&endereco,
		&configConexao,
		&h.Ativo,
		&h.CreatedAt,
		&h.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHospitalNotFound
		}
		return nil, err
	}

	if endereco.Valid {
		h.Endereco = &endereco.String
	}
	if configConexao.Valid {
		h.ConfigConexao = json.RawMessage(configConexao.String)
	}
	if deletedAt.Valid {
		h.DeletedAt = &deletedAt.Time
	}

	return &h, nil
}

// Create creates a new hospital
func (r *HospitalRepository) Create(ctx context.Context, input *models.CreateHospitalInput) (*models.Hospital, error) {
	// Check if code already exists
	exists, err := r.ExistsByCodigo(ctx, input.Codigo)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrHospitalExists
	}

	hospital := &models.Hospital{
		ID:        uuid.New(),
		Nome:      input.Nome,
		Codigo:    input.Codigo,
		Endereco:  input.Endereco,
		Ativo:     true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if input.ConfigConexao != nil {
		hospital.ConfigConexao = input.ConfigConexao
	}

	query := `
		INSERT INTO hospitals (id, nome, codigo, endereco, config_conexao, ativo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	var configJSON interface{}
	if hospital.ConfigConexao != nil {
		configJSON = string(hospital.ConfigConexao)
	}

	_, err = r.db.ExecContext(ctx, query,
		hospital.ID,
		hospital.Nome,
		hospital.Codigo,
		hospital.Endereco,
		configJSON,
		hospital.Ativo,
		hospital.CreatedAt,
		hospital.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return hospital, nil
}

// Update updates a hospital
func (r *HospitalRepository) Update(ctx context.Context, id uuid.UUID, input *models.UpdateHospitalInput) (*models.Hospital, error) {
	// Get existing hospital
	hospital, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if new code conflicts with existing hospital
	if input.Codigo != nil && *input.Codigo != hospital.Codigo {
		exists, err := r.ExistsByCodigo(ctx, *input.Codigo)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrHospitalExists
		}
		hospital.Codigo = *input.Codigo
	}

	if input.Nome != nil {
		hospital.Nome = *input.Nome
	}
	if input.Endereco != nil {
		hospital.Endereco = input.Endereco
	}
	if input.ConfigConexao != nil {
		hospital.ConfigConexao = input.ConfigConexao
	}
	if input.Ativo != nil {
		hospital.Ativo = *input.Ativo
	}

	hospital.UpdatedAt = time.Now()

	query := `
		UPDATE hospitals
		SET nome = $1, codigo = $2, endereco = $3, config_conexao = $4, ativo = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`

	var configJSON interface{}
	if hospital.ConfigConexao != nil {
		configJSON = string(hospital.ConfigConexao)
	}

	result, err := r.db.ExecContext(ctx, query,
		hospital.Nome,
		hospital.Codigo,
		hospital.Endereco,
		configJSON,
		hospital.Ativo,
		hospital.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, ErrHospitalNotFound
	}

	return hospital, nil
}

// SoftDelete performs a soft delete on a hospital
func (r *HospitalRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE hospitals
		SET deleted_at = $1, ativo = false, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrHospitalNotFound
	}

	return nil
}

// ExistsByCodigo checks if a hospital with the given code exists
func (r *HospitalRepository) ExistsByCodigo(ctx context.Context, codigo string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM hospitals WHERE codigo = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, codigo).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetActiveHospitals returns all active hospitals
func (r *HospitalRepository) GetActiveHospitals(ctx context.Context) ([]models.Hospital, error) {
	query := `
		SELECT id, nome, codigo, endereco, config_conexao, ativo, created_at, updated_at
		FROM hospitals
		WHERE deleted_at IS NULL AND ativo = true
		ORDER BY nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hospitals []models.Hospital
	for rows.Next() {
		var h models.Hospital
		var endereco, configConexao sql.NullString

		err := rows.Scan(
			&h.ID,
			&h.Nome,
			&h.Codigo,
			&endereco,
			&configConexao,
			&h.Ativo,
			&h.CreatedAt,
			&h.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if endereco.Valid {
			h.Endereco = &endereco.String
		}
		if configConexao.Valid {
			h.ConfigConexao = json.RawMessage(configConexao.String)
		}

		hospitals = append(hospitals, h)
	}

	return hospitals, nil
}
