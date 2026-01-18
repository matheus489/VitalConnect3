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
	"github.com/vitalconnect/backend/internal/models"
)

var (
	// ErrAdminTenantNotFound is returned when a tenant is not found
	ErrAdminTenantNotFound = errors.New("tenant not found")

	// ErrAdminTenantSlugExists is returned when a tenant slug already exists
	ErrAdminTenantSlugExists = errors.New("tenant with this slug already exists")
)

// AdminTenantListParams contains parameters for listing tenants in admin view
type AdminTenantListParams struct {
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
	Search  string `form:"search"`
	Status  string `form:"status"` // "all", "active", "inactive"
}

// AdminTenantListResult contains paginated tenant results
type AdminTenantListResult struct {
	Tenants    []models.TenantWithMetrics `json:"tenants"`
	Total      int                        `json:"total"`
	Page       int                        `json:"page"`
	PerPage    int                        `json:"per_page"`
	TotalPages int                        `json:"total_pages"`
}

// AdminTenantRepository handles admin-level tenant data access (cross-tenant)
type AdminTenantRepository struct {
	db *sql.DB
}

// NewAdminTenantRepository creates a new admin tenant repository
func NewAdminTenantRepository(db *sql.DB) *AdminTenantRepository {
	return &AdminTenantRepository{db: db}
}

// ListAllTenants returns all tenants with pagination (no tenant_id filter)
func (r *AdminTenantRepository) ListAllTenants(ctx context.Context, params *AdminTenantListParams) (*AdminTenantListResult, error) {
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

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(t.name ILIKE $%d OR t.slug ILIKE $%d)", argIndex, argIndex+1))
		searchPattern := "%" + params.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	if params.Status == "active" {
		conditions = append(conditions, fmt.Sprintf("t.is_active = $%d", argIndex))
		args = append(args, true)
		argIndex++
	} else if params.Status == "inactive" {
		conditions = append(conditions, fmt.Sprintf("t.is_active = $%d", argIndex))
		args = append(args, false)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM tenants t %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	offset := (params.Page - 1) * params.PerPage

	// Get tenants with metrics
	query := fmt.Sprintf(`
		SELECT
			t.id, t.name, t.slug, t.theme_config, t.is_active, t.logo_url, t.favicon_url, t.created_at, t.updated_at,
			COALESCE((SELECT COUNT(*) FROM users u WHERE u.tenant_id = t.id), 0) AS user_count,
			COALESCE((SELECT COUNT(*) FROM hospitals h WHERE h.tenant_id = t.id AND h.deleted_at IS NULL), 0) AS hospital_count,
			COALESCE((SELECT COUNT(*) FROM occurrences o WHERE o.tenant_id = t.id), 0) AS occurrence_count
		FROM tenants t
		%s
		ORDER BY t.name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []models.TenantWithMetrics
	for rows.Next() {
		var t models.TenantWithMetrics
		var themeConfig, logoURL, faviconURL sql.NullString
		var isActive sql.NullBool

		err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.Slug,
			&themeConfig,
			&isActive,
			&logoURL,
			&faviconURL,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.UserCount,
			&t.HospitalCount,
			&t.OccurrenceCount,
		)
		if err != nil {
			return nil, err
		}

		if themeConfig.Valid {
			t.ThemeConfig = json.RawMessage(themeConfig.String)
		}
		if isActive.Valid {
			t.IsActive = isActive.Bool
		} else {
			t.IsActive = true // Default to active
		}
		if logoURL.Valid {
			t.LogoURL = &logoURL.String
		}
		if faviconURL.Valid {
			t.FaviconURL = &faviconURL.String
		}

		tenants = append(tenants, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &AdminTenantListResult{
		Tenants:    tenants,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetTenantByID retrieves a tenant by ID with user/hospital/occurrence counts
func (r *AdminTenantRepository) GetTenantByID(ctx context.Context, id uuid.UUID) (*models.TenantWithMetrics, error) {
	query := `
		SELECT
			t.id, t.name, t.slug, t.theme_config, t.is_active, t.logo_url, t.favicon_url, t.created_at, t.updated_at,
			COALESCE((SELECT COUNT(*) FROM users u WHERE u.tenant_id = t.id), 0) AS user_count,
			COALESCE((SELECT COUNT(*) FROM hospitals h WHERE h.tenant_id = t.id AND h.deleted_at IS NULL), 0) AS hospital_count,
			COALESCE((SELECT COUNT(*) FROM occurrences o WHERE o.tenant_id = t.id), 0) AS occurrence_count
		FROM tenants t
		WHERE t.id = $1
	`

	var t models.TenantWithMetrics
	var themeConfig, logoURL, faviconURL sql.NullString
	var isActive sql.NullBool

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.Name,
		&t.Slug,
		&themeConfig,
		&isActive,
		&logoURL,
		&faviconURL,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.UserCount,
		&t.HospitalCount,
		&t.OccurrenceCount,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminTenantNotFound
		}
		return nil, err
	}

	if themeConfig.Valid {
		t.ThemeConfig = json.RawMessage(themeConfig.String)
	}
	if isActive.Valid {
		t.IsActive = isActive.Bool
	} else {
		t.IsActive = true // Default to active
	}
	if logoURL.Valid {
		t.LogoURL = &logoURL.String
	}
	if faviconURL.Valid {
		t.FaviconURL = &faviconURL.String
	}

	return &t, nil
}

// CreateTenant creates a new tenant with default theme_config
func (r *AdminTenantRepository) CreateTenant(ctx context.Context, input *models.CreateTenantInput) (*models.Tenant, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Check if slug already exists
	exists, err := r.ExistsBySlug(ctx, input.Slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAdminTenantSlugExists
	}

	// Use provided theme config or default
	themeConfig := models.DefaultThemeConfig()
	if input.ThemeConfig != nil {
		themeConfig = *input.ThemeConfig
	}

	// Marshal theme config to JSON
	themeConfigJSON, err := json.Marshal(themeConfig)
	if err != nil {
		return nil, err
	}

	tenant := &models.Tenant{
		ID:          uuid.New(),
		Name:        input.Name,
		Slug:        input.Slug,
		ThemeConfig: themeConfigJSON,
		IsActive:    true,
		LogoURL:     input.LogoURL,
		FaviconURL:  input.FaviconURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO tenants (id, name, slug, theme_config, is_active, logo_url, favicon_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		string(tenant.ThemeConfig),
		tenant.IsActive,
		tenant.LogoURL,
		tenant.FaviconURL,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return tenant, nil
}

// UpdateTenant updates a tenant's basic info (name, slug, logo, favicon)
func (r *AdminTenantRepository) UpdateTenant(ctx context.Context, id uuid.UUID, input *models.UpdateTenantInput) (*models.Tenant, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Get existing tenant
	tenant, err := r.GetTenantByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if new slug conflicts (if changing)
	if input.Slug != nil && *input.Slug != tenant.Slug {
		exists, err := r.ExistsBySlug(ctx, *input.Slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrAdminTenantSlugExists
		}
		tenant.Slug = *input.Slug
	}

	// Apply updates
	if input.Name != nil {
		tenant.Name = *input.Name
	}
	if input.IsActive != nil {
		tenant.IsActive = *input.IsActive
	}
	if input.LogoURL != nil {
		tenant.LogoURL = input.LogoURL
	}
	if input.FaviconURL != nil {
		tenant.FaviconURL = input.FaviconURL
	}
	tenant.UpdatedAt = time.Now()

	query := `
		UPDATE tenants
		SET name = $1, slug = $2, is_active = $3, logo_url = $4, favicon_url = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		tenant.Name,
		tenant.Slug,
		tenant.IsActive,
		tenant.LogoURL,
		tenant.FaviconURL,
		tenant.UpdatedAt,
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
		return nil, ErrAdminTenantNotFound
	}

	return &tenant.Tenant, nil
}

// UpdateThemeConfig updates the tenant's theme_config JSONB field
func (r *AdminTenantRepository) UpdateThemeConfig(ctx context.Context, id uuid.UUID, themeConfig models.ThemeConfig) (*models.Tenant, error) {
	// Validate theme config
	if err := models.ValidateThemeConfig(&themeConfig); err != nil {
		return nil, err
	}

	// Marshal theme config to JSON
	themeConfigJSON, err := json.Marshal(themeConfig)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE tenants
		SET theme_config = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, name, slug, theme_config, is_active, logo_url, favicon_url, created_at, updated_at
	`

	var t models.Tenant
	var themeConfigStr, logoURL, faviconURL sql.NullString
	var isActive sql.NullBool

	err = r.db.QueryRowContext(ctx, query, string(themeConfigJSON), time.Now(), id).Scan(
		&t.ID,
		&t.Name,
		&t.Slug,
		&themeConfigStr,
		&isActive,
		&logoURL,
		&faviconURL,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminTenantNotFound
		}
		return nil, err
	}

	if themeConfigStr.Valid {
		t.ThemeConfig = json.RawMessage(themeConfigStr.String)
	}
	if isActive.Valid {
		t.IsActive = isActive.Bool
	} else {
		t.IsActive = true
	}
	if logoURL.Valid {
		t.LogoURL = &logoURL.String
	}
	if faviconURL.Valid {
		t.FaviconURL = &faviconURL.String
	}

	return &t, nil
}

// ToggleTenantActive toggles the is_active status of a tenant
func (r *AdminTenantRepository) ToggleTenantActive(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	query := `
		UPDATE tenants
		SET is_active = NOT COALESCE(is_active, true), updated_at = $1
		WHERE id = $2
		RETURNING id, name, slug, theme_config, is_active, logo_url, favicon_url, created_at, updated_at
	`

	var t models.Tenant
	var themeConfigStr, logoURL, faviconURL sql.NullString
	var isActive sql.NullBool

	err := r.db.QueryRowContext(ctx, query, time.Now(), id).Scan(
		&t.ID,
		&t.Name,
		&t.Slug,
		&themeConfigStr,
		&isActive,
		&logoURL,
		&faviconURL,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminTenantNotFound
		}
		return nil, err
	}

	if themeConfigStr.Valid {
		t.ThemeConfig = json.RawMessage(themeConfigStr.String)
	}
	if isActive.Valid {
		t.IsActive = isActive.Bool
	} else {
		t.IsActive = true
	}
	if logoURL.Valid {
		t.LogoURL = &logoURL.String
	}
	if faviconURL.Valid {
		t.FaviconURL = &faviconURL.String
	}

	return &t, nil
}

// UpdateAssets updates the logo_url and/or favicon_url for a tenant
func (r *AdminTenantRepository) UpdateAssets(ctx context.Context, id uuid.UUID, logoURL, faviconURL *string) (*models.Tenant, error) {
	// Build dynamic update query
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if logoURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("logo_url = $%d", argIndex))
		args = append(args, *logoURL)
		argIndex++
	}
	if faviconURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("favicon_url = $%d", argIndex))
		args = append(args, *faviconURL)
		argIndex++
	}

	if len(setClauses) == 0 {
		// No updates, just return the existing tenant
		return r.getTenantByIDBasic(ctx, id)
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE tenants
		SET %s
		WHERE id = $%d
		RETURNING id, name, slug, theme_config, is_active, logo_url, favicon_url, created_at, updated_at
	`, strings.Join(setClauses, ", "), argIndex)

	var t models.Tenant
	var themeConfigStr, logo, favicon sql.NullString
	var isActive sql.NullBool

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&t.ID,
		&t.Name,
		&t.Slug,
		&themeConfigStr,
		&isActive,
		&logo,
		&favicon,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminTenantNotFound
		}
		return nil, err
	}

	if themeConfigStr.Valid {
		t.ThemeConfig = json.RawMessage(themeConfigStr.String)
	}
	if isActive.Valid {
		t.IsActive = isActive.Bool
	} else {
		t.IsActive = true
	}
	if logo.Valid {
		t.LogoURL = &logo.String
	}
	if favicon.Valid {
		t.FaviconURL = &favicon.String
	}

	return &t, nil
}

// ExistsBySlug checks if a tenant with the given slug exists
func (r *AdminTenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE slug = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// getTenantByIDBasic is a helper to get a tenant without metrics
func (r *AdminTenantRepository) getTenantByIDBasic(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	query := `
		SELECT id, name, slug, theme_config, is_active, logo_url, favicon_url, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	var t models.Tenant
	var themeConfigStr, logoURL, faviconURL sql.NullString
	var isActive sql.NullBool

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.Name,
		&t.Slug,
		&themeConfigStr,
		&isActive,
		&logoURL,
		&faviconURL,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAdminTenantNotFound
		}
		return nil, err
	}

	if themeConfigStr.Valid {
		t.ThemeConfig = json.RawMessage(themeConfigStr.String)
	}
	if isActive.Valid {
		t.IsActive = isActive.Bool
	} else {
		t.IsActive = true
	}
	if logoURL.Valid {
		t.LogoURL = &logoURL.String
	}
	if faviconURL.Valid {
		t.FaviconURL = &faviconURL.String
	}

	return &t, nil
}

// DashboardMetrics contains global metrics for the admin dashboard
type DashboardMetrics struct {
	TotalTenants     int `json:"total_tenants"`
	ActiveTenants    int `json:"active_tenants"`
	InactiveTenants  int `json:"inactive_tenants"`
	TotalUsers       int `json:"total_users"`
	TotalHospitals   int `json:"total_hospitals"`
	TotalOccurrences int `json:"total_occurrences"`
}

// GetDashboardMetrics returns global metrics for the admin dashboard
func (r *AdminTenantRepository) GetDashboardMetrics(ctx context.Context) (*DashboardMetrics, error) {
	metrics := &DashboardMetrics{}

	// Get tenant counts
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true OR is_active IS NULL) as active,
			COUNT(*) FILTER (WHERE is_active = false) as inactive
		FROM tenants
	`).Scan(&metrics.TotalTenants, &metrics.ActiveTenants, &metrics.InactiveTenants)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant counts: %w", err)
	}

	// Get user count
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE ativo = true`).Scan(&metrics.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}

	// Get hospital count
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM hospitals WHERE ativo = true`).Scan(&metrics.TotalHospitals)
	if err != nil {
		return nil, fmt.Errorf("failed to get hospital count: %w", err)
	}

	// Get occurrence count
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM occurrences`).Scan(&metrics.TotalOccurrences)
	if err != nil {
		return nil, fmt.Errorf("failed to get occurrence count: %w", err)
	}

	return metrics, nil
}
