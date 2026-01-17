package models

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidTenantSlug is returned when a tenant slug is invalid
	ErrInvalidTenantSlug = errors.New("invalid tenant slug: must be alphanumeric with hyphens only, 2-100 characters")

	// ErrTenantNotFound is returned when a tenant is not found
	ErrTenantNotFound = errors.New("tenant not found")

	// ErrTenantSlugExists is returned when a tenant slug already exists
	ErrTenantSlugExists = errors.New("tenant with this slug already exists")

	// slugRegex validates tenant slugs: alphanumeric with hyphens, no leading/trailing hyphens
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

// Tenant represents a Central de Transplantes (e.g., SES-GO, SES-PE, SES-SP)
type Tenant struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name" validate:"required,min=2,max=255"`
	Slug      string    `json:"slug" db:"slug" validate:"required,min=2,max=100"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateTenantInput represents input for creating a tenant
type CreateTenantInput struct {
	Name string `json:"name" validate:"required,min=2,max=255"`
	Slug string `json:"slug" validate:"required,min=2,max=100"`
}

// TenantResponse represents the API response for a tenant
type TenantResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ValidateSlug validates that a slug is URL-safe and follows naming conventions
// Valid slugs: lowercase alphanumeric with hyphens, no leading/trailing hyphens
// Examples: "ses-go", "ses-pe", "central-sp"
func ValidateSlug(slug string) error {
	if len(slug) < 2 || len(slug) > 100 {
		return ErrInvalidTenantSlug
	}

	if !slugRegex.MatchString(slug) {
		return ErrInvalidTenantSlug
	}

	return nil
}

// Validate validates the tenant data
func (t *Tenant) Validate() error {
	if t.Name == "" || len(t.Name) < 2 || len(t.Name) > 255 {
		return errors.New("tenant name must be between 2 and 255 characters")
	}

	return ValidateSlug(t.Slug)
}

// ToResponse converts Tenant to TenantResponse
func (t *Tenant) ToResponse() TenantResponse {
	return TenantResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// Validate validates the CreateTenantInput
func (i *CreateTenantInput) Validate() error {
	if i.Name == "" || len(i.Name) < 2 || len(i.Name) > 255 {
		return errors.New("tenant name must be between 2 and 255 characters")
	}

	return ValidateSlug(i.Slug)
}
