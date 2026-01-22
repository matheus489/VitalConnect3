package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
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

// List returns all active hospitals for the current tenant
func (r *HospitalRepository) List(ctx context.Context) ([]models.Hospital, error) {
	tf := NewTenantFilter(ctx)

	query := `
		SELECT id, nome, codigo, endereco, telefone, latitude, longitude, config_conexao, ativo, created_at, updated_at, deleted_at
		FROM hospitals
		WHERE deleted_at IS NULL` + tf.AndClause() + `
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
		var endereco, telefone, configConexao sql.NullString
		var latitude, longitude sql.NullFloat64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&h.ID,
			&h.Nome,
			&h.Codigo,
			&endereco,
			&telefone,
			&latitude,
			&longitude,
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
		if telefone.Valid {
			h.Telefone = &telefone.String
		}
		if latitude.Valid {
			h.Latitude = &latitude.Float64
		}
		if longitude.Valid {
			h.Longitude = &longitude.Float64
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

// GetByID retrieves a hospital by ID for the current tenant
func (r *HospitalRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Hospital, error) {
	tf := NewTenantFilter(ctx)

	query := `
		SELECT id, nome, codigo, endereco, telefone, latitude, longitude, config_conexao, ativo, created_at, updated_at, deleted_at
		FROM hospitals
		WHERE id = $1 AND deleted_at IS NULL` + tf.AndClause() + `
	`

	var h models.Hospital
	var endereco, telefone, configConexao sql.NullString
	var latitude, longitude sql.NullFloat64
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&h.ID,
		&h.Nome,
		&h.Codigo,
		&endereco,
		&telefone,
		&latitude,
		&longitude,
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
	if telefone.Valid {
		h.Telefone = &telefone.String
	}
	if latitude.Valid {
		h.Latitude = &latitude.Float64
	}
	if longitude.Valid {
		h.Longitude = &longitude.Float64
	}
	if configConexao.Valid {
		h.ConfigConexao = json.RawMessage(configConexao.String)
	}
	if deletedAt.Valid {
		h.DeletedAt = &deletedAt.Time
	}

	return &h, nil
}

// GetByCodigo retrieves a hospital by code for the current tenant
func (r *HospitalRepository) GetByCodigo(ctx context.Context, codigo string) (*models.Hospital, error) {
	tf := NewTenantFilter(ctx)

	query := `
		SELECT id, nome, codigo, endereco, telefone, latitude, longitude, config_conexao, ativo, created_at, updated_at, deleted_at
		FROM hospitals
		WHERE codigo = $1 AND deleted_at IS NULL` + tf.AndClause() + `
	`

	var h models.Hospital
	var endereco, telefone, configConexao sql.NullString
	var latitude, longitude sql.NullFloat64
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, codigo).Scan(
		&h.ID,
		&h.Nome,
		&h.Codigo,
		&endereco,
		&telefone,
		&latitude,
		&longitude,
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
	if telefone.Valid {
		h.Telefone = &telefone.String
	}
	if latitude.Valid {
		h.Latitude = &latitude.Float64
	}
	if longitude.Valid {
		h.Longitude = &longitude.Float64
	}
	if configConexao.Valid {
		h.ConfigConexao = json.RawMessage(configConexao.String)
	}
	if deletedAt.Valid {
		h.DeletedAt = &deletedAt.Time
	}

	return &h, nil
}

// Create creates a new hospital for the current tenant
func (r *HospitalRepository) Create(ctx context.Context, input *models.CreateHospitalInput) (*models.Hospital, error) {
	// Get tenant ID from context
	tenantID, err := RequireTenantID(ctx)
	if err != nil {
		return nil, err
	}
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}

	// Check if code already exists within the tenant
	exists, err := r.ExistsByCodigo(ctx, input.Codigo)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrHospitalExists
	}

	// Handle Ativo field with default value
	ativo := true
	if input.Ativo != nil {
		ativo = *input.Ativo
	}

	endereco := input.Endereco
	lat := input.Latitude
	lng := input.Longitude

	hospital := &models.Hospital{
		ID:        uuid.New(),
		TenantID:  tenantUUID,
		Nome:      input.Nome,
		Codigo:    input.Codigo,
		Endereco:  &endereco,
		Telefone:  input.Telefone,
		Latitude:  &lat,
		Longitude: &lng,
		Ativo:     ativo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if input.ConfigConexao != nil {
		hospital.ConfigConexao = input.ConfigConexao
	}

	query := `
		INSERT INTO hospitals (id, tenant_id, nome, codigo, endereco, telefone, latitude, longitude, config_conexao, ativo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	var configJSON interface{}
	if hospital.ConfigConexao != nil {
		configJSON = string(hospital.ConfigConexao)
	}

	_, err = r.db.ExecContext(ctx, query,
		hospital.ID,
		hospital.TenantID,
		hospital.Nome,
		hospital.Codigo,
		hospital.Endereco,
		hospital.Telefone,
		hospital.Latitude,
		hospital.Longitude,
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

// Update updates a hospital for the current tenant
func (r *HospitalRepository) Update(ctx context.Context, id uuid.UUID, input *models.UpdateHospitalInput) (*models.Hospital, error) {
	tf := NewTenantFilter(ctx)

	// Get existing hospital (already filtered by tenant)
	hospital, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if new code conflicts with existing hospital within the tenant
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
	if input.Telefone != nil {
		hospital.Telefone = input.Telefone
	}
	if input.Latitude != nil {
		hospital.Latitude = input.Latitude
	}
	if input.Longitude != nil {
		hospital.Longitude = input.Longitude
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
		SET nome = $1, codigo = $2, endereco = $3, telefone = $4, latitude = $5, longitude = $6, config_conexao = $7, ativo = $8, updated_at = $9
		WHERE id = $10 AND deleted_at IS NULL` + tf.AndClause() + `
	`

	var configJSON interface{}
	if hospital.ConfigConexao != nil {
		configJSON = string(hospital.ConfigConexao)
	}

	result, err := r.db.ExecContext(ctx, query,
		hospital.Nome,
		hospital.Codigo,
		hospital.Endereco,
		hospital.Telefone,
		hospital.Latitude,
		hospital.Longitude,
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

// SoftDelete performs a soft delete on a hospital for the current tenant
func (r *HospitalRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	tf := NewTenantFilter(ctx)

	query := `
		UPDATE hospitals
		SET deleted_at = $1, ativo = false, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL` + tf.AndClause() + `
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

// ExistsByCodigo checks if a hospital with the given code exists within the current tenant
func (r *HospitalRepository) ExistsByCodigo(ctx context.Context, codigo string) (bool, error) {
	tf := NewTenantFilter(ctx)

	query := `SELECT EXISTS(SELECT 1 FROM hospitals WHERE codigo = $1 AND deleted_at IS NULL` + tf.AndClause() + `)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, codigo).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetActiveHospitals returns all active hospitals for the current tenant
func (r *HospitalRepository) GetActiveHospitals(ctx context.Context) ([]models.Hospital, error) {
	tf := NewTenantFilter(ctx)

	query := `
		SELECT id, nome, codigo, endereco, telefone, latitude, longitude, config_conexao, ativo, created_at, updated_at
		FROM hospitals
		WHERE deleted_at IS NULL AND ativo = true` + tf.AndClause() + `
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
		var endereco, telefone, configConexao sql.NullString
		var latitude, longitude sql.NullFloat64

		err := rows.Scan(
			&h.ID,
			&h.Nome,
			&h.Codigo,
			&endereco,
			&telefone,
			&latitude,
			&longitude,
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
		if telefone.Valid {
			h.Telefone = &telefone.String
		}
		if latitude.Valid {
			h.Latitude = &latitude.Float64
		}
		if longitude.Valid {
			h.Longitude = &longitude.Float64
		}
		if configConexao.Valid {
			h.ConfigConexao = json.RawMessage(configConexao.String)
		}

		hospitals = append(hospitals, h)
	}

	return hospitals, nil
}

// GetActiveHospitalsWithCoordinates returns all active hospitals that have coordinates set for the current tenant
// This is used for the geographic map feature
func (r *HospitalRepository) GetActiveHospitalsWithCoordinates(ctx context.Context) ([]models.Hospital, error) {
	tf := NewTenantFilter(ctx)

	query := `
		SELECT id, nome, codigo, endereco, telefone, latitude, longitude, config_conexao, ativo, created_at, updated_at
		FROM hospitals
		WHERE deleted_at IS NULL
		  AND ativo = true
		  AND latitude IS NOT NULL
		  AND longitude IS NOT NULL` + tf.AndClause() + `
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
		var endereco, telefone, configConexao sql.NullString
		var latitude, longitude sql.NullFloat64

		err := rows.Scan(
			&h.ID,
			&h.Nome,
			&h.Codigo,
			&endereco,
			&telefone,
			&latitude,
			&longitude,
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
		if telefone.Valid {
			h.Telefone = &telefone.String
		}
		if latitude.Valid {
			h.Latitude = &latitude.Float64
		}
		if longitude.Valid {
			h.Longitude = &longitude.Float64
		}
		if configConexao.Valid {
			h.ConfigConexao = json.RawMessage(configConexao.String)
		}

		hospitals = append(hospitals, h)
	}

	return hospitals, nil
}
