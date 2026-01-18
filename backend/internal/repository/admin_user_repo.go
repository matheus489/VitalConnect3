package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

var (
	// ErrAdminUserNotFound is returned when a user is not found
	ErrAdminUserNotFound = errors.New("user not found")
)

// AdminUserListParams contains parameters for listing users in admin view
type AdminUserListParams struct {
	Page     int        `form:"page"`
	PerPage  int        `form:"per_page"`
	Search   string     `form:"search"`
	Status   string     `form:"status"` // "all", "active", "inactive", "banned"
	TenantID *uuid.UUID `form:"tenant_id"`
	Role     string     `form:"role"` // "operador", "gestor", "admin"
}

// AdminUserListResult contains paginated user results
type AdminUserListResult struct {
	Users      []models.UserWithTenant `json:"users"`
	Total      int                     `json:"total"`
	Page       int                     `json:"page"`
	PerPage    int                     `json:"per_page"`
	TotalPages int                     `json:"total_pages"`
}

// AdminUserRepository handles admin-level user data access (cross-tenant)
type AdminUserRepository struct {
	db *sql.DB
}

// NewAdminUserRepository creates a new admin user repository
func NewAdminUserRepository(db *sql.DB) *AdminUserRepository {
	return &AdminUserRepository{db: db}
}

// ListAllUsers returns all users with pagination (cross-tenant, optional tenant filter)
func (r *AdminUserRepository) ListAllUsers(ctx context.Context, params *AdminUserListParams) (*AdminUserListResult, error) {
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

	// Optional tenant filter
	if params.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("u.tenant_id = $%d", argIndex))
		args = append(args, *params.TenantID)
		argIndex++
	}

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(u.nome ILIKE $%d OR u.email ILIKE $%d)", argIndex, argIndex+1))
		searchPattern := "%" + params.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	if params.Status == "active" {
		conditions = append(conditions, fmt.Sprintf("u.ativo = $%d", argIndex))
		args = append(args, true)
		argIndex++
	} else if params.Status == "inactive" || params.Status == "banned" {
		conditions = append(conditions, fmt.Sprintf("u.ativo = $%d", argIndex))
		args = append(args, false)
		argIndex++
	}

	if params.Role != "" {
		conditions = append(conditions, fmt.Sprintf("u.role = $%d", argIndex))
		args = append(args, params.Role)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM users u %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	offset := (params.Page - 1) * params.PerPage

	// Get users with tenant info
	query := fmt.Sprintf(`
		SELECT
			u.id, u.email, u.nome, u.role, u.tenant_id, u.is_super_admin,
			u.mobile_phone, u.email_notifications, u.ativo, u.created_at, u.updated_at,
			t.name as tenant_name, t.slug as tenant_slug
		FROM users u
		LEFT JOIN tenants t ON u.tenant_id = t.id
		%s
		ORDER BY u.nome ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.UserWithTenant
	for rows.Next() {
		var u models.UserWithTenant
		var tenantID, mobilePhone, tenantName, tenantSlug sql.NullString
		var isSuperAdmin sql.NullBool

		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.Nome,
			&u.Role,
			&tenantID,
			&isSuperAdmin,
			&mobilePhone,
			&u.EmailNotifications,
			&u.Ativo,
			&u.CreatedAt,
			&u.UpdatedAt,
			&tenantName,
			&tenantSlug,
		)
		if err != nil {
			return nil, err
		}

		if tenantID.Valid {
			tid, err := uuid.Parse(tenantID.String)
			if err == nil {
				u.TenantID = &tid
			}
		}
		if mobilePhone.Valid {
			u.MobilePhone = &mobilePhone.String
		}
		if isSuperAdmin.Valid {
			u.IsSuperAdmin = isSuperAdmin.Bool
		}
		if tenantName.Valid {
			u.TenantName = &tenantName.String
		}
		if tenantSlug.Valid {
			u.TenantSlug = &tenantSlug.String
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &AdminUserListResult{
		Users:      users,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetUserByID retrieves a user by ID (cross-tenant)
func (r *AdminUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.UserWithTenant, error) {
	query := `
		SELECT
			u.id, u.email, u.nome, u.role, u.tenant_id, u.is_super_admin,
			u.mobile_phone, u.email_notifications, u.ativo, u.created_at, u.updated_at,
			t.name as tenant_name, t.slug as tenant_slug
		FROM users u
		LEFT JOIN tenants t ON u.tenant_id = t.id
		WHERE u.id = $1
	`

	var u models.UserWithTenant
	var tenantID, mobilePhone, tenantName, tenantSlug sql.NullString
	var isSuperAdmin sql.NullBool

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.Nome,
		&u.Role,
		&tenantID,
		&isSuperAdmin,
		&mobilePhone,
		&u.EmailNotifications,
		&u.Ativo,
		&u.CreatedAt,
		&u.UpdatedAt,
		&tenantName,
		&tenantSlug,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}

	if tenantID.Valid {
		tid, err := uuid.Parse(tenantID.String)
		if err == nil {
			u.TenantID = &tid
		}
	}
	if mobilePhone.Valid {
		u.MobilePhone = &mobilePhone.String
	}
	if isSuperAdmin.Valid {
		u.IsSuperAdmin = isSuperAdmin.Bool
	}
	if tenantName.Valid {
		u.TenantName = &tenantName.String
	}
	if tenantSlug.Valid {
		u.TenantSlug = &tenantSlug.String
	}

	// Load hospitals for the user
	hospitals, err := r.getUserHospitals(ctx, id)
	if err != nil {
		return nil, err
	}
	u.Hospitals = hospitals

	return &u, nil
}

// UpdateUserRole updates a user's role and/or super admin status
func (r *AdminUserRepository) UpdateUserRole(ctx context.Context, id uuid.UUID, role *models.UserRole, isSuperAdmin *bool) (*models.UserWithTenant, error) {
	// Build dynamic update query
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if role != nil {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, *role)
		argIndex++
	}

	if isSuperAdmin != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_super_admin = $%d", argIndex))
		args = append(args, *isSuperAdmin)
		argIndex++
	}

	if len(setClauses) == 0 {
		// No updates, just return the existing user
		return r.GetUserByID(ctx, id)
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, ErrAdminUserNotFound
	}

	return r.GetUserByID(ctx, id)
}

// UpdateUserStatus updates a user's active status (ban/unban)
func (r *AdminUserRepository) UpdateUserStatus(ctx context.Context, id uuid.UUID, active bool, banReason *string) (*models.UserWithTenant, error) {
	query := `
		UPDATE users
		SET ativo = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, active, time.Now(), id)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, ErrAdminUserNotFound
	}

	return r.GetUserByID(ctx, id)
}

// ResetUserPassword resets a user's password
func (r *AdminUserRepository) ResetUserPassword(ctx context.Context, id uuid.UUID, newPasswordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, newPasswordHash, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAdminUserNotFound
	}

	return nil
}

// getUserHospitals retrieves all hospitals associated with a user
func (r *AdminUserRepository) getUserHospitals(ctx context.Context, userID uuid.UUID) ([]models.Hospital, error) {
	query := `
		SELECT h.id, h.nome, h.codigo, h.endereco, h.ativo, h.created_at, h.updated_at
		FROM hospitals h
		INNER JOIN user_hospitals uh ON h.id = uh.hospital_id
		WHERE uh.user_id = $1
		ORDER BY h.nome ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hospitals []models.Hospital
	for rows.Next() {
		var h models.Hospital
		var endereco sql.NullString

		err := rows.Scan(
			&h.ID, &h.Nome, &h.Codigo, &endereco, &h.Ativo, &h.CreatedAt, &h.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if endereco.Valid {
			h.Endereco = &endereco.String
		}

		hospitals = append(hospitals, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hospitals, nil
}
