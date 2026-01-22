package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
)

var (
	// ErrAdminHospitalNotFound is returned when a hospital is not found
	ErrAdminHospitalNotFound = errors.New("hospital not found")
)

// AdminHospitalListParams contains parameters for listing hospitals in admin view
type AdminHospitalListParams struct {
	Page     int        `form:"page"`
	PerPage  int        `form:"per_page"`
	Search   string     `form:"search"`
	Status   string     `form:"status"` // "all", "active", "inactive"
	TenantID *uuid.UUID `form:"tenant_id"`
}

// AdminHospitalListResult contains paginated hospital results
type AdminHospitalListResult struct {
	Hospitals  []models.HospitalWithTenant `json:"hospitals"`
	Total      int                         `json:"total"`
	Page       int                         `json:"page"`
	PerPage    int                         `json:"per_page"`
	TotalPages int                         `json:"total_pages"`
}

// AdminHospitalRepository handles admin-level hospital data access (cross-tenant)
type AdminHospitalRepository struct {
	db *sql.DB
}

// NewAdminHospitalRepository creates a new admin hospital repository
func NewAdminHospitalRepository(db *sql.DB) *AdminHospitalRepository {
	return &AdminHospitalRepository{db: db}
}

// ListAllHospitals returns all hospitals with pagination (cross-tenant, optional tenant filter)
func (r *AdminHospitalRepository) ListAllHospitals(ctx context.Context, params *AdminHospitalListParams) (*AdminHospitalListResult, error) {
	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}
	if params.Status == "" {
		params.Status = "all"
	}

	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Always exclude soft-deleted hospitals
	conditions = append(conditions, "h.deleted_at IS NULL")

	// Optional tenant filter
	if params.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("h.tenant_id = $%d", argIndex))
		args = append(args, *params.TenantID)
		argIndex++
	}

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(h.nome ILIKE $%d OR h.codigo ILIKE $%d OR h.endereco ILIKE $%d)", argIndex, argIndex+1, argIndex+2))
		searchPattern := "%" + params.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
		argIndex += 3
	}

	if params.Status == "active" {
		conditions = append(conditions, fmt.Sprintf("h.ativo = $%d", argIndex))
		args = append(args, true)
		argIndex++
	} else if params.Status == "inactive" {
		conditions = append(conditions, fmt.Sprintf("h.ativo = $%d", argIndex))
		args = append(args, false)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM hospitals h %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	offset := (params.Page - 1) * params.PerPage

	// Get hospitals with tenant info
	query := fmt.Sprintf(`
		SELECT
			h.id, h.tenant_id, h.nome, h.codigo, h.endereco, h.telefone,
			h.latitude, h.longitude, h.config_conexao, h.ativo,
			h.created_at, h.updated_at,
			t.name as tenant_name, t.slug as tenant_slug
		FROM hospitals h
		LEFT JOIN tenants t ON h.tenant_id = t.id
		%s
		ORDER BY h.nome ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hospitals []models.HospitalWithTenant
	for rows.Next() {
		var h models.HospitalWithTenant
		var endereco, telefone, configConexao, tenantName, tenantSlug sql.NullString
		var latitude, longitude sql.NullFloat64

		err := rows.Scan(
			&h.ID,
			&h.TenantID,
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
			&tenantName,
			&tenantSlug,
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
		if tenantName.Valid {
			h.TenantName = &tenantName.String
		}
		if tenantSlug.Valid {
			h.TenantSlug = &tenantSlug.String
		}

		hospitals = append(hospitals, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &AdminHospitalListResult{
		Hospitals:  hospitals,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetHospitalByID retrieves a hospital by ID (cross-tenant)
func (r *AdminHospitalRepository) GetHospitalByID(ctx context.Context, id uuid.UUID) (*models.HospitalWithTenant, error) {
	query := `
		SELECT
			h.id, h.tenant_id, h.nome, h.codigo, h.endereco, h.telefone,
			h.latitude, h.longitude, h.config_conexao, h.ativo,
			h.created_at, h.updated_at,
			t.name as tenant_name, t.slug as tenant_slug
		FROM hospitals h
		LEFT JOIN tenants t ON h.tenant_id = t.id
		WHERE h.id = $1 AND h.deleted_at IS NULL
	`

	var h models.HospitalWithTenant
	var endereco, telefone, configConexao, tenantName, tenantSlug sql.NullString
	var latitude, longitude sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&h.ID,
		&h.TenantID,
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
		&tenantName,
		&tenantSlug,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminHospitalNotFound
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
	if tenantName.Valid {
		h.TenantName = &tenantName.String
	}
	if tenantSlug.Valid {
		h.TenantSlug = &tenantSlug.String
	}

	return &h, nil
}

// UpdateHospital updates a hospital's details (cross-tenant)
func (r *AdminHospitalRepository) UpdateHospital(ctx context.Context, id uuid.UUID, input *models.UpdateHospitalInput) (*models.HospitalWithTenant, error) {
	// Get existing hospital
	hospital, err := r.GetHospitalByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if input.Nome != nil {
		hospital.Nome = *input.Nome
	}
	if input.Codigo != nil {
		hospital.Codigo = *input.Codigo
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
		SET nome = $1, codigo = $2, endereco = $3, telefone = $4,
		    latitude = $5, longitude = $6, config_conexao = $7, ativo = $8, updated_at = $9
		WHERE id = $10 AND deleted_at IS NULL
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
		return nil, ErrAdminHospitalNotFound
	}

	return r.GetHospitalByID(ctx, id)
}

// ReassignHospitalTenant moves a hospital to a different tenant
func (r *AdminHospitalRepository) ReassignHospitalTenant(ctx context.Context, hospitalID uuid.UUID, newTenantID uuid.UUID) (*models.HospitalWithTenant, error) {
	// Check if hospital exists
	_, err := r.GetHospitalByID(ctx, hospitalID)
	if err != nil {
		return nil, err
	}

	// Check if new tenant exists
	var tenantExists bool
	err = r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1)", newTenantID).Scan(&tenantExists)
	if err != nil {
		return nil, err
	}
	if !tenantExists {
		return nil, errors.New("target tenant not found")
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update hospital tenant
	query := `
		UPDATE hospitals
		SET tenant_id = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := tx.ExecContext(ctx, query, newTenantID, time.Now(), hospitalID)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, ErrAdminHospitalNotFound
	}

	// Update user_hospitals associations to new tenant
	_, err = tx.ExecContext(ctx, `
		UPDATE user_hospitals
		SET tenant_id = $1
		WHERE hospital_id = $2
	`, newTenantID, hospitalID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetHospitalByID(ctx, hospitalID)
}
