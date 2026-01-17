package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

// TenantRepository handles tenant data access
type TenantRepository struct {
	db *sql.DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *sql.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create creates a new tenant
func (r *TenantRepository) Create(ctx context.Context, input *models.CreateTenantInput) (*models.Tenant, error) {
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
		return nil, models.ErrTenantSlugExists
	}

	tenant := &models.Tenant{
		ID:        uuid.New(),
		Name:      input.Name,
		Slug:      input.Slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO tenants (id, name, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return tenant, nil
}

// CreateWithID creates a new tenant with a specific ID (used for seeding)
func (r *TenantRepository) CreateWithID(ctx context.Context, id uuid.UUID, name, slug string) (*models.Tenant, error) {
	// Validate slug
	if err := models.ValidateSlug(slug); err != nil {
		return nil, err
	}

	// Check if slug already exists
	exists, err := r.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, models.ErrTenantSlugExists
	}

	tenant := &models.Tenant{
		ID:        id,
		Name:      name,
		Slug:      slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO tenants (id, name, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return tenant, nil
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	var tenant models.Tenant
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrTenantNotFound
		}
		return nil, err
	}

	return &tenant, nil
}

// GetBySlug retrieves a tenant by slug
func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM tenants
		WHERE slug = $1
	`

	var tenant models.Tenant
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrTenantNotFound
		}
		return nil, err
	}

	return &tenant, nil
}

// List returns all tenants ordered by name
func (r *TenantRepository) List(ctx context.Context) ([]models.Tenant, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at
		FROM tenants
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []models.Tenant
	for rows.Next() {
		var tenant models.Tenant
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Slug,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tenants, nil
}

// ExistsBySlug checks if a tenant with the given slug exists
func (r *TenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE slug = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// ExistsByID checks if a tenant with the given ID exists
func (r *TenantRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// Update updates a tenant's name (slug cannot be changed)
func (r *TenantRepository) Update(ctx context.Context, id uuid.UUID, name string) (*models.Tenant, error) {
	// Validate name
	if name == "" || len(name) < 2 || len(name) > 255 {
		return nil, errors.New("tenant name must be between 2 and 255 characters")
	}

	query := `
		UPDATE tenants
		SET name = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, name, slug, created_at, updated_at
	`

	var tenant models.Tenant
	err := r.db.QueryRowContext(ctx, query, name, time.Now(), id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrTenantNotFound
		}
		return nil, err
	}

	return &tenant, nil
}
